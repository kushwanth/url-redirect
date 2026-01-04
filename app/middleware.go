package main

import (
	"context"
	"crypto/subtle"
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

type AnalyticsLog struct {
	Path              string
	Status            int
	ProcessingTime    int64
	AdditionalHeaders string
}

var analyticsChan = make(chan AnalyticsLog, 1000)

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

var requestsByResponseStatus = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "requests_by_response_status",
		Help: "Requests by response status",
	},
	[]string{"status"},
)

var httpStatusDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "http_response_status_time_ms",
	Help:    "Duration of HTTP Requests in milliseconds and Response Status",
	Buckets: []float64{0, 0.5, 1, 2, 5, 10},
},
	[]string{"handler", "status"},
)

var activeRequestsGauge = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "http_active_requests",
		Help: "Number of active connections to the service",
	},
)

var apiLatencySummary = prometheus.NewSummary(
	prometheus.SummaryOpts{
		Name: "api_request_duration",
		Help: "Duration of API requests",
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
	metricsRegistry.MustRegister(requestsByResponseStatus)
	metricsRegistry.MustRegister(httpStatusDuration)
	metricsRegistry.MustRegister(activeRequestsGauge)
	metricsRegistry.MustRegister(apiLatencySummary)
	for _, runtimeMetricsGuage := range runtimeMetricsGuages {
		metricsRegistry.MustRegister(runtimeMetricsGuage)
	}
}

func startAnalyticsWorker(db *pgxpool.Pool) {
	go func() {
		for logEntry := range analyticsChan {
			_, err := db.Exec(context.Background(), `INSERT INTO UrlRedirects_Analytics (path, log_timestamp, status, processing_time, additional_headers) VALUES ($1,now(),$2,$3,$4)`, logEntry.Path, logEntry.Status, logEntry.ProcessingTime, logEntry.AdditionalHeaders)
			if err != nil {
				log.Println("Analytics Insert Error:", err)
			}
		}
	}()
}

func httpRateLimit(next http.Handler) http.Handler {
	return httprate.Limit(
		getHttpRateLimit(),
		time.Second,
		httprate.WithKeyFuncs(httprate.KeyByIP),
	)(next)
}

func verifyApiKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKeyHeader := r.Header.Get("x-url-redirect-token")
		if subtle.ConstantTimeCompare([]byte(apiKeyHeader), []byte(apiKey)) != 1 {
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

				analyticsChan <- AnalyticsLog{
					Path:              reqPath,
					Status:            appResponse.Status(),
					ProcessingTime:    processingTime,
					AdditionalHeaders: strings.Join(additionalHeaders, "|"),
				}
			}
		})
	}
}

func prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		activeRequestsGauge.Inc()
		startTime := time.Now()
		reqPath := r.URL.Path
		appResWriter := &AppResponseWriter{ResponseWriter: w}
		next.ServeHTTP(appResWriter, r)
		endTime := float64(time.Since(startTime).Milliseconds())
		if !skipLogging(reqPath) {
			totalRequests.Inc()
			statusCode := strconv.Itoa(appResWriter.Status())
			requestsByResponseStatus.WithLabelValues(statusCode).Inc()
			reqFunc, isApiReq := getRequestFunction(reqPath, appResWriter.Status())
			httpStatusDuration.WithLabelValues(reqFunc, statusCode).Observe(endTime)
			if isApiReq {
				apiLatencySummary.Observe(endTime)
			}
			for runtimeMetricKey, runtimeMetricValue := range getRuntimeMetrics() {
				runtimeMetricsGuages[runtimeMetricKey].Set(runtimeMetricValue)
			}
		}
		activeRequestsGauge.Dec()
	})
}
