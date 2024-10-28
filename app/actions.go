package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func queryDB(path string, limit int, status bool, db *pgxpool.Pool) (Redirect, error) {
	var responseData Redirect
	db_err := db.QueryRow(context.Background(), "SELECT id, path, url, updated_at::TEXT, inactive FROM UrlRedirects WHERE path=$1 AND inactive=$2 LIMIT $3", path, status, limit).Scan(&responseData.Id, &responseData.Path, &responseData.Url, &responseData.LastUpdated, &responseData.Inactive)
	if db_err != nil {
		return responseData, errors.New("database error")
	}
	return responseData, nil
}

func queryDbWithId(id int, limit int, status bool, db *pgxpool.Pool) (Redirect, error) {
	var responseData Redirect
	db_err := db.QueryRow(context.Background(), "SELECT id, path, url, updated_at::TEXT, inactive FROM UrlRedirects WHERE id=$1 AND inactive=$2 LIMIT $3", id, status, limit).Scan(&responseData.Id, &responseData.Path, &responseData.Url, &responseData.LastUpdated, &responseData.Inactive)
	if db_err != nil {
		return responseData, errors.New("database error")
	}
	return responseData, nil
}

func handleRedirect(db *pgxpool.Pool) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		validPath, isPathValid := validateAndFormatPath(path)
		if !isPathValid {
			w.WriteHeader(http.StatusNotFound)
			log.Print("Path Invalid")
			return
		}
		dbResponse, err := queryDB(validPath, dbLimit, false, db)
		if err != nil {
			http.Redirect(w, r, WebsiteUrl, http.StatusFound)
			return
		}
		http.Redirect(w, r, buildUri(dbResponse.Url), http.StatusFound)
	}
	return http.HandlerFunc(fn)
}

func redirectPathInfo(db *pgxpool.Pool) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var response ResponseMessage
		validPath, isPathValid := validateAndFormatPath(r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		if !isPathValid {
			w.WriteHeader(http.StatusBadRequest)
			response.Message = badRequest
			w.Write(toJson(response))
			log.Print("Path Invalid")
			return
		}
		dbResponse, dbErr := queryDB(validPath, dbLimit, false, db)
		if dbErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			response.Message = notExistMessage
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(toJson(dbResponse))
	}
	return http.HandlerFunc(fn)
}

func redirectInfo(db *pgxpool.Pool) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var response ResponseMessage
		redirectId, idErr := strconv.Atoi(chi.URLParam(r, "id"))
		if idErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			response.Message = badRequest
			w.Write(toJson(response))
			log.Print(idErr.Error())
			return
		}
		dbResponse, dbErr := queryDbWithId(redirectId, dbLimit, false, db)
		if dbErr != nil || dbResponse.Id != redirectId {
			w.WriteHeader(http.StatusBadRequest)
			response.Message = notExistMessage
			w.Write(toJson(response))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(toJson(dbResponse))
	}
	return http.HandlerFunc(fn)
}

func addRedirect(db *pgxpool.Pool) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var requestData UrlData
		var response ResponseMessage
		err := json.NewDecoder(r.Body).Decode(&requestData)
		validUrl, isUrlValid := validateAndFormatURL(requestData.Url)
		validPath, isPathValid := validateAndFormatPath(requestData.Path)
		w.Header().Set("Content-Type", "application/json")
		if !isUrlValid || !isPathValid || err != nil {
			w.WriteHeader(http.StatusBadRequest)
			response.Message = badRequest
			w.Write(toJson(response))
			log.Print("Either Path or URL is Invalid", err.Error())
			return
		}
		var responseData Redirect
		db_err := db.QueryRow(context.Background(), "INSERT INTO UrlRedirects (path, url, updated_at) VALUES ($1,$2,now()) RETURNING id, path, url, updated_at::TEXT, inactive", validPath, validUrl).Scan(&responseData.Id, &responseData.Path, &responseData.Url, &responseData.LastUpdated, &responseData.Inactive)
		if db_err != nil {
			_, duplicateErr := queryDB(validPath, dbLimit, false, db)
			if duplicateErr == nil {
				w.WriteHeader(http.StatusBadRequest)
				response.Message = alreadyExistMessage
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				response.Message = internalError
			}
			w.Write(toJson(response))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(toJson(responseData))
	}
	return http.HandlerFunc(fn)
}

func patchRedirect(db *pgxpool.Pool) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var requestData UrlData
		var response ResponseMessage
		err := json.NewDecoder(r.Body).Decode(&requestData)
		validUrl, isUrlValid := validateAndFormatURL(requestData.Url)
		validPath, isPathValid := validateAndFormatPath(requestData.Path)
		w.Header().Set("Content-Type", "application/json")
		if !isUrlValid || !isPathValid || err != nil {
			w.WriteHeader(http.StatusBadRequest)
			response.Message = badRequest
			w.Write(toJson(response))
			log.Print("Either Path or URL is Invalid", err.Error())
			return
		}
		var responseData Redirect
		dbResponse, dbErr := queryDB(validPath, dbLimit, false, db)
		if dbErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			response.Message = notExistMessage
			w.Write(toJson(response))
			return
		}
		db_err := db.QueryRow(context.Background(), "UPDATE UrlRedirects SET url=$1, updated_at=now() WHERE id=$2 RETURNING id, path, url, updated_at::TEXT, inactive", validUrl, dbResponse.Id).Scan(&responseData.Id, &responseData.Path, &responseData.Url, &responseData.LastUpdated, &responseData.Inactive)
		if db_err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			response.Message = internalError
			w.Write(toJson(response))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(toJson(responseData))
	}
	return http.HandlerFunc(fn)
}

func updateRedirect(db *pgxpool.Pool) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var requestData Redirect
		var response ResponseMessage
		redirectId, idErr := strconv.Atoi(chi.URLParam(r, "id"))
		if idErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			response.Message = badRequest
			w.Write(toJson(response))
			log.Print(idErr.Error())
			return
		}
		err := json.NewDecoder(r.Body).Decode(&requestData)
		validUrl, isUrlValid := validateAndFormatURL(requestData.Url)
		validPath, isPathValid := validateAndFormatPath(requestData.Path)
		w.Header().Set("Content-Type", "application/json")
		if !isUrlValid || !isPathValid || err != nil {
			w.WriteHeader(http.StatusBadRequest)
			response.Message = badRequest
			w.Write(toJson(response))
			log.Print("Either Path or URL is Invalid", err.Error())
			return
		}
		var responseData Redirect
		dbResponse, dbErr := queryDbWithId(redirectId, dbLimit, requestData.Inactive, db)
		if dbErr != nil || dbResponse.Id != redirectId {
			w.WriteHeader(http.StatusBadRequest)
			response.Message = notExistMessage
			w.Write(toJson(response))
			return
		}
		db_err := db.QueryRow(context.Background(), "UPDATE UrlRedirects SET path=$1, url=$2, updated_at=now() WHERE id=$3 RETURNING id, path, url, updated_at::TEXT, inactive", validPath, validUrl, dbResponse.Id).Scan(&responseData.Id, &responseData.Path, &responseData.Url, &responseData.LastUpdated, &responseData.Inactive)
		if db_err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			response.Message = internalError
			w.Write(toJson(response))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(toJson(responseData))
	}
	return http.HandlerFunc(fn)
}

func deleteRedirect(db *pgxpool.Pool) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var response ResponseMessage
		redirectId, idErr := strconv.Atoi(chi.URLParam(r, "id"))
		if idErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			response.Message = badRequest
			w.Write(toJson(response))
			log.Print(idErr.Error())
			return
		}
		var responseData Redirect
		dbResponse, dbErr := queryDbWithId(redirectId, dbLimit, false, db)
		if dbErr != nil || dbResponse.Id != redirectId {
			w.WriteHeader(http.StatusBadRequest)
			response.Message = notExistMessage
			w.Write(toJson(response))
			return
		}
		db_err := db.QueryRow(context.Background(), "UPDATE UrlRedirects SET inactive=$1, updated_at=now() WHERE id=$2 RETURNING id, path, url, updated_at::TEXT, inactive", true, dbResponse.Id).Scan(&responseData.Id, &responseData.Path, &responseData.Url, &responseData.LastUpdated, &responseData.Inactive)
		if db_err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			response.Message = internalError
			w.Write(toJson(response))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(toJson(responseData))
	}
	return http.HandlerFunc(fn)
}
