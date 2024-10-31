package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func handleRedirect(db *pgxpool.Pool) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		validPath, isPathValid := validateAndFormatPath(path)
		if !isPathValid {
			log.Println("handleRedirect -> Path Invalid")
			http.Error(w, notFoundMessage, http.StatusNotFound)
			return
		}
		dbResponse, err := getRedirectUsingPath(validPath, db)
		if err != nil || dbResponse.Inactive {
			log.Println("handleRedirect ->", err.Error())
			http.Error(w, notFoundMessage, http.StatusNotFound)
			return
		}
		http.Redirect(w, r, buildUri(dbResponse.Url), http.StatusFound)
	})
}

func redirectInfo(db *pgxpool.Pool) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		redirectId, idErr := strconv.Atoi(chi.URLParam(r, "id"))
		if idErr != nil {
			log.Println("redirectInfo ->", idErr.Error())
			http.Error(w, badRequest, http.StatusBadRequest)
			return
		}
		dbResponse, dbErr := getRedirectUsingId(redirectId, db)
		if dbErr != nil || dbResponse.Id != redirectId {
			log.Println("redirectInfo ->", dbErr.Error())
			http.Error(w, dbError, http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(toJson(dbResponse))
	})
}

func addRedirect(db *pgxpool.Pool) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var requestData UrlData
		err := json.NewDecoder(r.Body).Decode(&requestData)
		validUrl, isUrlValid := validateAndFormatURL(requestData.Url)
		validPath, isPathValid := validateAndFormatPath(requestData.Path)
		if !isUrlValid || !isPathValid || err != nil {
			log.Println("addRedirect -> ", err.Error(), isPathValid, isUrlValid)
			http.Error(w, badRequest, http.StatusBadRequest)
			return
		}
		_, duplicateErr := getRedirectUsingPath(validPath, db)
		urlExists := doesUrlExists(validUrl, db)
		if urlExists || duplicateErr == nil {
			http.Error(w, alreadyExistMessage, http.StatusPreconditionFailed)
			return
		}
		var responseData Redirect
		db_err := db.QueryRow(context.Background(), "INSERT INTO UrlRedirects (path, url, updated_at) VALUES ($1,$2,now()) RETURNING id, path, url, updated_at::TEXT, inactive", validPath, validUrl).Scan(&responseData.Id, &responseData.Path, &responseData.Url, &responseData.LastUpdated, &responseData.Inactive)
		if db_err != nil {
			log.Println("addRedirect -> ", db_err.Error())
			http.Error(w, dbError, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(toJson(responseData))
	})
}

func patchRedirect(db *pgxpool.Pool) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var requestData UrlData
		err := json.NewDecoder(r.Body).Decode(&requestData)
		validUrl, isUrlValid := validateAndFormatURL(requestData.Url)
		validPath, isPathValid := validateAndFormatPath(requestData.Path)
		w.Header().Set("Content-Type", "application/json")
		if !isUrlValid || !isPathValid || err != nil {
			log.Println("patchRedirect -> ", err.Error(), isPathValid, isUrlValid)
			http.Error(w, badRequest, http.StatusBadRequest)
			return
		}
		var responseData Redirect
		dbResponse, dbErr := getRedirectUsingPath(validPath, db)
		if dbErr != nil {
			log.Println("patchRedirect -> ", dbErr.Error())
			http.Error(w, notExistMessage, http.StatusPreconditionFailed)
			return
		}
		db_err := db.QueryRow(context.Background(), "UPDATE UrlRedirects SET url=$1, updated_at=now(), inactive=$2 WHERE id=$3 RETURNING id, path, url, updated_at::TEXT, inactive, is_private", validUrl, false, dbResponse.Id).Scan(&responseData.Id, &responseData.Path, &responseData.Url, &responseData.LastUpdated, &responseData.Inactive)
		if db_err != nil {
			log.Println("patchRedirect -> ", db_err.Error())
			http.Error(w, dbError, http.StatusPreconditionFailed)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(toJson(responseData))
	})
}

func updateRedirect(db *pgxpool.Pool) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var requestData Redirect
		redirectId, idErr := strconv.Atoi(chi.URLParam(r, "id"))
		if idErr != nil {
			log.Println("updateRedirect ->", idErr.Error())
			http.Error(w, badRequest, http.StatusBadRequest)
			return
		}
		err := json.NewDecoder(r.Body).Decode(&requestData)
		validUrl, isUrlValid := validateAndFormatURL(requestData.Url)
		validPath, isPathValid := validateAndFormatPath(requestData.Path)
		w.Header().Set("Content-Type", "application/json")
		if !isUrlValid || !isPathValid || err != nil {
			log.Println("updateRedirect ->", err.Error(), isPathValid, isUrlValid)
			http.Error(w, badRequest, http.StatusBadRequest)
			return
		}
		var responseData Redirect
		dbResponse, dbErr := getRedirectUsingId(redirectId, db)
		if dbErr != nil || dbResponse.Id != redirectId {
			log.Println("updateRedirect -> ", dbErr.Error())
			http.Error(w, notExistMessage, http.StatusPreconditionFailed)
			return
		}
		db_err := db.QueryRow(context.Background(), "UPDATE UrlRedirects SET path=$1, url=$2, updated_at=now(), inactive=$3 WHERE id=$4 RETURNING id, path, url, updated_at::TEXT, inactive", validPath, validUrl, false, dbResponse.Id).Scan(&responseData.Id, &responseData.Path, &responseData.Url, &responseData.LastUpdated, &responseData.Inactive)
		if db_err != nil {
			log.Println("updateRedirect -> ", db_err.Error())
			http.Error(w, dbError, http.StatusPreconditionFailed)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(toJson(responseData))
	})
}

func deleteRedirect(db *pgxpool.Pool) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		redirectId, idErr := strconv.Atoi(chi.URLParam(r, "id"))
		if idErr != nil {
			log.Println("deleteRedirect ->", idErr.Error())
			http.Error(w, badRequest, http.StatusBadRequest)
			return
		}
		var responseData Redirect
		dbResponse, dbErr := getRedirectUsingId(redirectId, db)
		if dbErr != nil || dbResponse.Id != redirectId {
			log.Println("deleteRedirect -> ", dbErr.Error())
			http.Error(w, notExistMessage, http.StatusPreconditionFailed)
			return
		}
		db_err := db.QueryRow(context.Background(), "UPDATE UrlRedirects SET inactive=$1, updated_at=now() WHERE id=$2 RETURNING id, path, url, updated_at::TEXT, inactive", true, dbResponse.Id).Scan(&responseData.Id, &responseData.Path, &responseData.Url, &responseData.LastUpdated, &responseData.Inactive)
		if db_err != nil {
			log.Println("deleteRedirect -> ", db_err.Error())
			http.Error(w, dbError, http.StatusPreconditionFailed)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(toJson(responseData))
	})
}
