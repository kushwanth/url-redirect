package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
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
const dbLimit = 1
const pageLimit = 10
const httpsProtocol = "https://"

const urlredirectSchema = `CREATE TABLE IF NOT EXISTS UrlRedirects (
    id SERIAL PRIMARY KEY,
    path VARCHAR(29) NOT NULL UNIQUE,
    url VARCHAR(100) NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    inactive BOOLEAN NOT NULL DEFAULT FALSE,
	is_private BOOLEAN NOT NULL DEFAULT FALSE
);`

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

type OpsData struct {
	Data string `json:"data,omitempty"`
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
	db_err := db.QueryRow(context.Background(), "SELECT id, path, url, updated_at::TEXT, inactive FROM UrlRedirects WHERE id=$1 LIMIT $3", id, dbLimit).Scan(&responseData.Id, &responseData.Path, &responseData.Url, &responseData.LastUpdated, &responseData.Inactive)
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
