package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

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
	Buckets: []float64{0.9, 2, 9},
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

func verifyApiKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := os.Getenv("API_KEY")
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
			appResWriter := &AppResponseWriter{ResponseWriter: w}
			rPath := r.URL.Path
			// rIP := sql.NullString{String: r.Header.Get("X-Forwarded-For"), Valid: len(r.Header.Get("X-Forwarded-For")) > 0}
			// rCountry := sql.NullString{String: "", Valid: false}
			// if rIP.Valid && geoIpDb != nil {
			// 	validIP := net.ParseIP(rIP.String)
			// 	geoIpCountry, countryErr := geoIpDb.Country(validIP)
			// 	if countryErr == nil {
			// 		rCountry.String = geoIpCountry.Country.IsoCode
			// 		rCountry.Valid = len(geoIpCountry.Country.IsoCode) <= 3 && len(geoIpCountry.Country.IsoCode) > 0
			// 	}
			// }
			next.ServeHTTP(appResWriter, r)
			processingTime := time.Since(startTime).Milliseconds()
			log.Printf("%s %s %d %vms\n", r.Method, rPath, appResWriter.statusCode, processingTime)
			// deviceType, browser, os := getRequestDeviceType(r.Header.Get("User-Agent"), r.Header.Get("x-url-redirect-version"))
			if !skipMiddleware(r.URL.Path) {
				_, dbEventErr := db.Exec(context.Background(), `INSERT INTO UrlRedirects_Analytics (path, log_timestamp, status, processing_time) VALUES ($1,now(),$2,$3) RETURNING id`, rPath, appResWriter.statusCode, processingTime)
				if dbEventErr != nil {
					log.Println(dbEventErr.Error())
				}
			}
		})
	}
}

func prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		activeRequestsGauge.Inc()
		reqPath := r.URL.Path
		appResWriter := &AppResponseWriter{ResponseWriter: w}
		startTime := time.Now()
		next.ServeHTTP(appResWriter, r)
		endTime := time.Since(startTime).Seconds()
		statusCode := appResWriter.statusCode
		activeRequestsGauge.Dec()
		if skipMiddleware(reqPath) {
			return
		}
		responseStatus.WithLabelValues(strconv.Itoa(statusCode)).Inc()
		totalRequestsByPath.WithLabelValues(reqPath).Inc()
		totalRequests.Inc()
		reqFunc, isApiReq := getRequestFunction(reqPath, statusCode)
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
	})
}
