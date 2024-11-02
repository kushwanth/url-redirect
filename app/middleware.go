package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oschwald/geoip2-golang"
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
		apiKeyHeader := r.Header.Get("x-url-redirect-token")
		if apiKeyHeader != apiKey {
			http.Error(w, errorMessage, http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func logRequest(db *pgxpool.Pool, geoIpDb *geoip2.Reader) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()
			appResWriter := &AppResponseWriter{ResponseWriter: w}
			rPath := r.URL.Path
			rIP := sql.NullString{String: r.Header.Get("X-Forwarded-For"), Valid: len(r.Header.Get("X-Forwarded-For")) > 0}
			rCountry := sql.NullString{String: "", Valid: false}
			if rIP.Valid && geoIpDb != nil {
				validIP := net.ParseIP(rIP.String)
				geoIpCountry, countryErr := geoIpDb.Country(validIP)
				if countryErr == nil {
					rCountry.String = geoIpCountry.Country.IsoCode
					rCountry.Valid = len(geoIpCountry.Country.IsoCode) <= 3 && len(geoIpCountry.Country.IsoCode) > 0
				}
			}
			next.ServeHTTP(appResWriter, r)
			processingTime := time.Since(startTime).Milliseconds()
			deviceType, browser, os := getRequestDeviceType(r.Header.Get("User-Agent"), r.Header.Get("x-url-redirect-version"))
			_, dbEventErr := db.Exec(context.Background(), `INSERT INTO UrlRedirects_Analytics (path, log_timestamp, status, country, processing_time, ip_address, browser, os, device_type) VALUES ($1,now(),$2,$3,$4,$5,$6,$7,$8) RETURNING id`, rPath, appResWriter.statusCode, rCountry, processingTime, rIP, browser, os, deviceType)
			if dbEventErr != nil {
				log.Println(dbEventErr.Error())
			}
			log.Printf("%s %s %d %vms\n", r.Method, rPath, appResWriter.statusCode, processingTime)
		})
	}
}
