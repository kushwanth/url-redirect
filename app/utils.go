package main

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var errorBytes []byte

const errorMessage = "Error"
const notFoundMessage = "Not Found"
const alreadyExistMessage = "URL Redirect for Path Exists"
const notExistMessage = "URL Redirect for Path doesn't Exists"
const internalError = "Internal Error"
const badRequest = "Bad Request"
const dbError = "DataBase Error"
const dbLimit = 1
const pageLimit = 10
const httpsProtocol = "https://"

const urlredirectSchema = `CREATE TABLE IF NOT EXISTS UrlRedirects (
    id SERIAL PRIMARY KEY,
    path VARCHAR(10) NOT NULL UNIQUE,
    url VARCHAR(100) NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    inactive BOOLEAN NOT NULL DEFAULT FALSE
);`

var WebsiteUrl string = os.Getenv("WebsiteUrl")

type ResponseMessage struct {
	Message string
}

type Redirect struct {
	Id          int    `json:"id,omitempty"`
	Path        string `json:"path,omitempty"`
	Url         string `json:"url,omitempty"`
	LastUpdated string `json:"lastUpdated,omitempty"`
	Inactive    bool   `json:"inactive,omitempty"`
}

type UrlData struct {
	Url  string `json:"url,omitempty"`
	Path string `json:"path,omitempty"`
}

type SearchQuery struct {
	Data string `json:"data,omitempty"`
}

func validateAndFormatURL(uri string) (string, bool) {
	validUri, err := url.ParseRequestURI(uri)
	if err != nil {
		return errorMessage, false
	}
	formattedUri := validUri.Host + validUri.Path
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

func buildUri(url string) string {
	return httpsProtocol + url
}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(notFoundMessage))
}

func toJson(struc interface{}) []byte {
	responseMessageJson, err := json.Marshal(struc)
	if err != nil {
		return errorBytes
	} else {
		return responseMessageJson
	}
}

func about(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "text/plain")
	w.Write([]byte("Hello, I Redirect URL's"))
}
