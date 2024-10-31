package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AppResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *AppResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *AppResponseWriter) Status() int {
	return w.statusCode
}

func verifyApiKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := os.Getenv("API_KEY")
		apiKeyHeader := r.Header.Get("X-Redirect-API-KEY")
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
			rIP := sql.NullString{String: r.Header.Get("Cf-Connecting-Ip"), Valid: len(r.Header.Get("Cf-Connecting-Ip")) > 0}
			next.ServeHTTP(appResWriter, r)
			processingTime := time.Since(startTime).Milliseconds()
			rUA := uaParser.Parse(r.Header.Get("User-Agent"))
			deviceType := getRequestDeviceType(rUA)
			_, dbEventErr := db.Exec(context.Background(), `INSERT INTO UrlRedirects_Analytics (path, event_date, status, country, processing_time, ip_address, browser, browser_version, os, device_type) VALUES ($1,now(),$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id, path, event_date, status, country, processing_time, ip_address, browser, browser_version, os, device_type`, rPath, appResWriter.statusCode, r.Header.Get("Cf-Ipcountry"), processingTime, rIP, rUA.GetBrowser(), rUA.GetMajorVersion(), rUA.GetOS(), deviceType) //.Scan(&responseData.Id, &responseData.Path, &responseData.Url, &responseData.LastUpdated, &responseData.Inactive)
			if dbEventErr != nil {
				log.Println(dbEventErr.Error())
			}
			log.Printf("%s %s %d %v\n", r.Method, rPath, appResWriter.statusCode, processingTime)
		})
	}
}
