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
	fn := func(w http.ResponseWriter, r *http.Request) {
		var responseData []Redirect
		var response ResponseMessage
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		min, max := page, page+pageLimit
		rows, db_err := db.Query(context.Background(), "SELECT id, path, url, updated_at::TEXT, inactive FROM redirects where id>$1 AND id<=$2 LIMIT $3", min, max, pageLimit)
		if db_err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			response.Message = dbError
			w.Write(toJson(response))
			log.Print(db_err.Error())
			return
		}
		for rows.Next() {
			var temp Redirect
			err := rows.Scan(&temp.Id, &temp.Path, &temp.Url, &temp.LastUpdated, &temp.Inactive)
			if err != nil {
				log.Print(temp.Id, err.Error())
			}
			responseData = append(responseData, temp)
		}
		w.WriteHeader(http.StatusOK)
		w.Write(toJson(responseData))
	}
	return http.HandlerFunc(fn)
}

func searchPath(db *pgxpool.Pool) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var requestData SearchQuery
		var response ResponseMessage
		err := json.NewDecoder(r.Body).Decode(&requestData)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			response.Message = badRequest
			w.Write(toJson(response))
			log.Print(err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		var responseData []Redirect
		pathMatchPattern := "%" + requestData.Data + "%"
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		rows, db_err := db.Query(context.Background(), "SELECT id, path, url, updated_at::TEXT, inactive FROM redirects WHERE path ILIKE $1 AND inactive=$2 LIMIT $3 OFFSET $4", pathMatchPattern, false, pageLimit, page)
		if db_err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			response.Message = dbError
			w.Write(toJson(response))
			log.Print(db_err.Error())
			return
		}
		for rows.Next() {
			var temp Redirect
			err := rows.Scan(&temp.Id, &temp.Path, &temp.Url, &temp.LastUpdated, &temp.Inactive)
			if err != nil {
				log.Print(temp.Id, err.Error())
			}
			responseData = append(responseData, temp)
		}
		w.WriteHeader(http.StatusOK)
		w.Write(toJson(responseData))
	}
	return http.HandlerFunc(fn)
}

func redirectExists(db *pgxpool.Pool) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var requestData SearchQuery
		var response ResponseMessage
		err := json.NewDecoder(r.Body).Decode(&requestData)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			response.Message = badRequest
			w.Write(toJson(response))
			log.Print(err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		var responseData Redirect
		db_err := db.QueryRow(context.Background(), "SELECT id, path, url, updated_at::TEXT, inactive FROM redirects WHERE url=$1 AND inactive=$2 LIMIT $3", requestData.Data, true, dbLimit).Scan(&responseData.Id, &responseData.Path, &responseData.Url, &responseData.LastUpdated, &responseData.Inactive)
		if db_err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			response.Message = dbError
			w.Write(toJson(response))
			log.Print(db_err.Error())
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(toJson(responseData))
	}
	return http.HandlerFunc(fn)
}
