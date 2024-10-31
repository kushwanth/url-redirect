package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
)

func listall(db *pgxpool.Pool) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var responseData []Redirect
		page, pageErr := strconv.Atoi(r.URL.Query().Get("page"))
		if pageErr != nil {
			log.Println("listall -> ", pageErr.Error())
			page = 0
		}
		min, max := page, page+pageLimit
		rows, db_err := db.Query(context.Background(), "SELECT id, path, url, updated_at::TEXT, inactive FROM UrlRedirects where id>$1 AND id<=$2 LIMIT $3", min, max, pageLimit)
		if db_err != nil {
			log.Println("listall -> ", db_err.Error())
			http.Error(w, notFoundMessage, http.StatusNotFound)
			return
		}
		for rows.Next() {
			var temp Redirect
			err := rows.Scan(&temp.Id, &temp.Path, &temp.Url, &temp.LastUpdated, &temp.Inactive)
			if err != nil {
				log.Println("listall -> ", temp.Id, err.Error())
			}
			responseData = append(responseData, temp)
		}
		w.WriteHeader(http.StatusOK)
		w.Write(toJson(responseData))
	})
}

func searchPath(db *pgxpool.Pool) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var requestData OpsData
		err := json.NewDecoder(r.Body).Decode(&requestData)
		if err != nil {
			log.Println("searchPath -> ", err.Error())
			http.Error(w, badRequest, http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		var responseData []Redirect
		pathMatchPattern := "%" + requestData.Data + "%"
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		rows, db_err := db.Query(context.Background(), "SELECT id, path, url, updated_at::TEXT, inactive FROM UrlRedirects WHERE path ILIKE $1 AND inactive=$2 LIMIT $3 OFFSET $4", pathMatchPattern, false, pageLimit, page)
		if db_err != nil {
			log.Println("searchPath -> ", db_err.Error())
			http.Error(w, dbError, http.StatusPreconditionFailed)
			return
		}
		for rows.Next() {
			var temp Redirect
			err := rows.Scan(&temp.Id, &temp.Path, &temp.Url, &temp.LastUpdated, &temp.Inactive)
			if err != nil {
				log.Println("searchPath -> ", temp.Id, err.Error())
			}
			responseData = append(responseData, temp)
		}
		w.WriteHeader(http.StatusOK)
		w.Write(toJson(responseData))
	})
}

func redirectExists(db *pgxpool.Pool) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var requestData OpsData
		err := json.NewDecoder(r.Body).Decode(&requestData)
		if err != nil {
			log.Println("redirectExists -> ", err.Error())
			http.Error(w, badRequest, http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		var responseData Redirect
		db_err := db.QueryRow(context.Background(), "SELECT id, path, url, updated_at::TEXT, inactive FROM UrlRedirects WHERE url=$1 LIMIT $2", requestData.Data, dbLimit).Scan(&responseData.Id, &responseData.Path, &responseData.Url, &responseData.LastUpdated, &responseData.Inactive)
		if db_err != nil {
			log.Println("redirectExists -> ", db_err.Error())
			http.Error(w, dbError, http.StatusPreconditionFailed)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(toJson(responseData))
	})
}

func generateRedirect(db *pgxpool.Pool) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var requestData OpsData
		err := json.NewDecoder(r.Body).Decode(&requestData)
		validUrl, isUrlValid := validateAndFormatURL(requestData.Data)
		w.Header().Set("Content-Type", "application/json")
		if !isUrlValid || err != nil {
			log.Println("generateRedirect -> ", err.Error(), isUrlValid)
			http.Error(w, badRequest, http.StatusBadRequest)
			return
		}
		generatedShortPath := generateShortRedirectPath()
		_, duplicateErr := getRedirectUsingPath(generatedShortPath, db)
		urlExists := doesUrlExists(validUrl, db)
		if urlExists || duplicateErr == nil {
			http.Error(w, alreadyExistMessage, http.StatusPreconditionFailed)
			return
		}
		var responseData Redirect
		db_err := db.QueryRow(context.Background(), "INSERT INTO UrlRedirects (path, url, updated_at) VALUES ($1,$2,now()) RETURNING id, path, url, updated_at::TEXT, inactive", generatedShortPath, validUrl).Scan(&responseData.Id, &responseData.Path, &responseData.Url, &responseData.LastUpdated, &responseData.Inactive)
		if db_err != nil {
			log.Println("generateRedirect -> ", db_err.Error())
			http.Error(w, dbError, http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(toJson(responseData))
	})
}
