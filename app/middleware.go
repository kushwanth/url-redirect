package main

import (
	"context"
	"log"
	"net/http"
	"net/textproto"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/httprate"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var metricsRegistry = prometheus.NewRegistry()

type AppResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

var totalRequests = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total requests",
	},
)

var totalRequestsByPath = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total_by_path",
		Help: "Total requests by Path",
	},
	[]string{"path"},
)

var responseStatus = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_response_status",
		Help: "Status of HTTP response",
	},
	[]string{"status"},
)

var httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "http_response_time_seconds",
	Help:    "Duration of HTTP requests.",
	Buckets: []float64{0.1, 0.5, 1, 5, 10},
},
	[]string{"function"},
)

var activeRequestsGauge = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "http_active_requests",
		Help: "Number of active connections to the service",
	},
)

var apiLatencySummary = prometheus.NewSummary(
	prometheus.SummaryOpts{
		Name: "api_request_duration_seconds",
		Help: "Duration of API requests",
		Objectives: map[float64]float64{
			0.5:  0.05,
			0.9:  0.01,
			0.99: 0.001,
		},
	},
)

var redirectLatencySummary = prometheus.NewSummary(
	prometheus.SummaryOpts{
		Name: "redirect_duration_seconds",
		Help: "Duration of redirecting requests",
		Objectives: map[float64]float64{
			0.5:  0.05,
			0.9:  0.01,
			0.99: 0.001,
		},
	},
)

var runtimeMetricsGuages = map[string]prometheus.Gauge{}

func (w *AppResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *AppResponseWriter) Status() int {
	return w.statusCode
}

func initRuntimeMetrics() {
	for k, v := range metricsList {
		runtimeMetricsGuages[v] = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: v,
				Help: k,
			},
		)
	}
}

func initMetrics() {
	initRuntimeMetrics()
	metricsRegistry.MustRegister(totalRequests)
	metricsRegistry.MustRegister(totalRequestsByPath)
	metricsRegistry.MustRegister(responseStatus)
	metricsRegistry.MustRegister(httpDuration)
	metricsRegistry.MustRegister(activeRequestsGauge)
	metricsRegistry.MustRegister(apiLatencySummary)
	metricsRegistry.MustRegister(redirectLatencySummary)
	for _, runtimeMetricsGuage := range runtimeMetricsGuages {
		metricsRegistry.MustRegister(runtimeMetricsGuage)
	}
}

func httpRateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httprate.Limit(getHttpRateLimit(), time.Second)
		next.ServeHTTP(w, r)
	})
}

func verifyApiKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKeyHeader := r.Header.Get("x-url-redirect-token")
		if apiKeyHeader != apiKey {
			http.Error(w, errorMessage, http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func logRequest(db *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()
			reqPath := r.URL.Path
			appResponse := &AppResponseWriter{ResponseWriter: w}
			next.ServeHTTP(appResponse, r)
			processingTime := time.Since(startTime).Milliseconds()
			if !skipLogging(r.URL.Path) {
				log.Printf("%s %s %d %vms\n", r.Method, reqPath, appResponse.Status(), processingTime)
				additionalHeaders := make([]string, len(logAdditionalHeaders))
				for i, logAdditionalHeader := range logAdditionalHeaders {
					additionalHeaders[i] = r.Header.Get(textproto.CanonicalMIMEHeaderKey(logAdditionalHeader))
				}
				_, logReqDbErr := db.Exec(context.Background(), `INSERT INTO UrlRedirects_Analytics (path, log_timestamp, status, processing_time, additional_headers) VALUES ($1,now(),$2,$3,$4) RETURNING id`, reqPath, appResponse.Status(), processingTime, strings.Join(additionalHeaders, "|"))
				if logReqDbErr != nil {
					log.Println(logReqDbErr.Error())
				}
			}
		})
	}
}

func prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		activeRequestsGauge.Inc()
		reqPath := r.URL.Path
		appResWriter := &AppResponseWriter{ResponseWriter: w}
		next.ServeHTTP(appResWriter, r)
		endTime := time.Since(startTime).Seconds()
		statusCode := appResWriter.Status()
		activeRequestsGauge.Dec()
		if !skipLogging(reqPath) {
			responseStatus.WithLabelValues(strconv.Itoa(statusCode)).Inc()
			totalRequests.Inc()
			reqFunc, isApiReq := getRequestFunction(reqPath, statusCode)
			totalRequestsByPath.WithLabelValues(reqFunc).Inc()
			httpDuration.WithLabelValues(reqFunc).Observe(endTime)
			if isApiReq {
				apiLatencySummary.Observe(endTime)
			}
			if statusCode == http.StatusFound || statusCode == http.StatusNotFound {
				redirectLatencySummary.Observe(endTime)
			}
			for runtimeMetricKey, runtimeMetricValue := range getRuntimeMetrics() {
				runtimeMetricsGuages[runtimeMetricKey].Set(runtimeMetricValue)
			}
		}
	})
}
