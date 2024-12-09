package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime/metrics"
	"slices"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/exp/rand"
)

var errorBytes []byte

const errorMessage = "Error"
const notFoundMessage = "Are you Lost??"
const alreadyExistMessage = "URL Redirect Exists"
const notExistMessage = "URL Redirect for Path doesn't Exists"
const badRequest = "Bad Request"
const dbError = "DataBase Error"
const internalError = "Internal Error"
const dbLimit = 1
const pageLimit = 10
const httpsProtocol = "https://"

var metricsList = map[string]string{
	"/gc/heap/allocs:bytes":               "go_memstats_alloc_bytes_total",
	"/gc/heap/frees:bytes":                "go_memstats_free_bytes_total",
	"/memory/classes/heap/free:bytes":     "go_memstats_heap_bytes_free",
	"/memory/classes/heap/released:bytes": "go_memstats_heap_bytes_released",
	"/memory/classes/heap/stacks:bytes":   "go_memstats_stack_bytes_reserved",
	"/memory/classes/total:bytes":         "go_memstats_sys_bytes",
	"/memory/classes/heap/unused:bytes":   "go_memstats_heap_bytes_unused",
	"/sched/goroutines:goroutines":        "go_goroutines",
}

var pathsToSkipMiddleware = []string{"/metrics"}

const urlredirectSchema = `CREATE TABLE IF NOT EXISTS UrlRedirects (
    id SERIAL PRIMARY KEY,
    path VARCHAR(29) NOT NULL UNIQUE,
    url VARCHAR(100) NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    inactive BOOLEAN NOT NULL DEFAULT FALSE,
	is_private BOOLEAN NOT NULL DEFAULT FALSE
);`

const urlredirectAnalyticsSchema = `CREATE TABLE IF NOT EXISTS UrlRedirects_Analytics (
  id SERIAL PRIMARY KEY,
  path VARCHAR(29) NOT NULL,
  log_timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
  status int NOT NULL,
  country VARCHAR(3),
  processing_time bigint,
  ip_address inet,
  browser VARCHAR(16),
  os VARCHAR(16),
  device_type CHAR(1)
);`

type Redirect struct {
	Id          int    `json:"id,omitempty"`
	Path        string `json:"path,omitempty"`
	Url         string `json:"url,omitempty"`
	LastUpdated string `json:"lastUpdated,omitempty"`
	Inactive    bool   `json:"inactive,omitempty"`
}

type LogStatsData struct {
	Path    []LogQueryData `json:"path,omitempty"`
	Status  []LogQueryData `json:"status,omitempty"`
	Country []LogQueryData `json:"country,omitempty"`
	Time    []LogQueryData `json:"time,omitempty"`
	Browser []LogQueryData `json:"browser,omitempty"`
	Os      []LogQueryData `json:"os,omitempty"`
	Devices []LogQueryData `json:"devices,omitempty"`
}

type LogQueryData struct {
	StatKey   string `json:"stat_key,omitempty"`
	StatCount any    `json:"stat_count,omitempty"`
}

type UrlData struct {
	Url  string `json:"url,omitempty"`
	Path string `json:"path,omitempty"`
}

type OpsData struct {
	Data string `json:"data,omitempty"`
}

type StatsTime struct {
	Start int64 `json:"start,omitempty"`
	End   int64 `json:"end,omitempty"`
}

func validateAndFormatURL(uri string) (string, bool) {
	validUri, err := url.ParseRequestURI(uri)
	if err != nil {
		return errorMessage, false
	}
	formattedUri := validUri.Host + validUri.Path
	if len(validUri.Query()) > 0 {
		formattedUri = formattedUri + "?" + validUri.RawQuery
	}
	return formattedUri, err == nil
}

func validateAndFormatPath(path string) (string, bool) {
	var formattedPath string
	validPath, err := url.ParseRequestURI(path)
	if err != nil {
		return errorMessage, false
	}
	formattedPath = strings.Trim(validPath.Path, "/")
	return formattedPath, err == nil
}

func getRedirectUsingPath(path string, db *pgxpool.Pool) (Redirect, error) {
	var responseData Redirect
	db_err := db.QueryRow(context.Background(), "SELECT id, path, url, updated_at::TEXT, inactive FROM UrlRedirects WHERE path=$1 LIMIT $2", path, dbLimit).Scan(&responseData.Id, &responseData.Path, &responseData.Url, &responseData.LastUpdated, &responseData.Inactive)
	if db_err != nil {
		log.Println("getRedirectUsingPath ->", db_err.Error())
		return responseData, errors.New("database error")
	}
	return responseData, nil
}

func getRedirectUsingId(id int, db *pgxpool.Pool) (Redirect, error) {
	var responseData Redirect
	db_err := db.QueryRow(context.Background(), "SELECT id, path, url, updated_at::TEXT, inactive FROM UrlRedirects WHERE id=$1 LIMIT $2", id, dbLimit).Scan(&responseData.Id, &responseData.Path, &responseData.Url, &responseData.LastUpdated, &responseData.Inactive)
	if db_err != nil {
		log.Println("getRedirectUsingId ->", db_err.Error())
		return responseData, errors.New("database error")
	}
	return responseData, nil
}

func doesUrlExists(url string, db *pgxpool.Pool) bool {
	var possibleId int
	db_err := db.QueryRow(context.Background(), "SELECT id FROM UrlRedirects WHERE url=$1 LIMIT $2", url, dbLimit).Scan(&possibleId) //.Scan(&responseData.Id, &responseData.Path, &responseData.Url, &responseData.LastUpdated, &responseData.Inactive)
	if db_err == nil {
		return true
	}
	log.Println("doesUrlExists ->", db_err.Error())
	return false
}

func generateShortRedirectPath() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	length, parseErr := strconv.Atoi(os.Getenv("PATH_LENGTH"))
	if parseErr != nil {
		length = 9
	}
	shortPath := make([]byte, length)
	for i := range shortPath {
		shortPath[i] = charset[rand.Intn(len(charset))]
	}
	return string(shortPath)
}

func buildUri(url string) string {
	return httpsProtocol + url
}

func toJson(data interface{}) []byte {
	responseJson, err := json.Marshal(data)
	if err != nil {
		log.Println("JSONwriter ->", err.Error())
		return errorBytes
	} else {
		return responseJson
	}
}

// func getRequestDeviceType(UA, redirectorVersion string) (string, any, any) {
// 	deviceType := ""
// 	rUA := uaParser.Parse(UA)
// 	match, matchErr := regexp.MatchString("^([0-9]+.)?([0-9]+)$", redirectorVersion)
// 	if match && matchErr == nil {
// 		deviceType = "C"
// 	} else if rUA.IsDesktop() {
// 		deviceType = "D"
// 	} else if rUA.IsMobile() || rUA.IsTablet() {
// 		deviceType = "M"
// 	} else if rUA.IsBot() {
// 		deviceType = "B"
// 	} else {
// 		deviceType = "U"
// 	}
// 	return string(deviceType), rUA.GetBrowser(), rUA.GetOS()
// }

func getRequestFunction(path string, statusCode int) (string, bool) {
	requestFunction := "misc"
	isApiRequest := strings.Contains(path, "api/")
	if statusCode >= http.StatusMovedPermanently && statusCode <= http.StatusPermanentRedirect {
		requestFunction = "redirect"
	} else if statusCode == http.StatusNotFound {
		requestFunction = "not_found"
	} else if statusCode >= http.StatusOK && statusCode <= http.StatusAlreadyReported {
		requestFunction = "request_success"
	} else if statusCode == http.StatusUnauthorized {
		requestFunction = "request_unauthorized"
	} else if statusCode == http.StatusForbidden {
		requestFunction = "request_forbidden"
	} else if statusCode == http.StatusMethodNotAllowed {
		requestFunction = "request_method_not_allowed"
	} else if statusCode >= http.StatusBadRequest && statusCode <= http.StatusUnavailableForLegalReasons {
		requestFunction = "request_client_error"
	} else if statusCode >= http.StatusInternalServerError && statusCode <= http.StatusNetworkAuthenticationRequired {
		requestFunction = "request_server_error"
	} else {
		requestFunction = "request_misc"
	}
	if isApiRequest {
		requestFunction = "api_" + requestFunction
	}
	return requestFunction, isApiRequest
}

func getRuntimeMetrics() map[string]float64 {
	samples := make([]metrics.Sample, len(metricsList))
	i := 0
	for k := range metricsList {
		samples[i].Name = k
		i++
	}
	metrics.Read(samples)
	runtimeMetrics := map[string]float64{}
	for _, sample := range samples {
		value := sample.Value
		name := metricsList[sample.Name]
		switch value.Kind() {
		case metrics.KindUint64:
			runtimeMetrics[name] = float64(value.Uint64())
		case metrics.KindFloat64:
			runtimeMetrics[name] = value.Float64()
		case metrics.KindFloat64Histogram:
			runtimeMetrics[name] = medianBucket(value.Float64Histogram())
		case metrics.KindBad:
			panic("bug in runtime/metrics package!")
		default:
			fmt.Printf("%s: unexpected metric Kind: %v\n", name, value.Kind())
		}
	}
	return runtimeMetrics
}

func medianBucket(h *metrics.Float64Histogram) float64 {
	total := uint64(0)
	for _, count := range h.Counts {
		total += count
	}
	thresh := total / 2
	total = 0
	for i, count := range h.Counts {
		total += count
		if total >= thresh {
			return h.Buckets[i]
		}
	}
	panic("should not happen")
}

func skipMiddleware(path string) bool {
	return slices.Contains(pathsToSkipMiddleware, path)
}
