package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

func listall(db *pgxpool.Pool) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var responseData []Redirect
		page, pageErr := strconv.Atoi(r.URL.Query().Get("page"))
		if pageErr != nil {
			page = 0
		}
		min, max := page, page+pageLimit
		rows, db_err := db.Query(context.Background(), "SELECT id, path, url, updated_at::TEXT, inactive FROM UrlRedirects WHERE id>$1 AND id<=$2 LIMIT $3", min, max, pageLimit)
		if db_err != nil {
			http.Error(w, notFoundMessage, http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		for rows.Next() {
			var temp Redirect
			rowErr := rows.Scan(&temp.Id, &temp.Path, &temp.Url, &temp.LastUpdated, &temp.Inactive)
			if rowErr == nil {
				responseData = append(responseData, temp)
			}
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
			http.Error(w, badRequest, http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		var responseData []Redirect
		pathMatchPattern := "%" + requestData.Data + "%"
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		rows, db_err := db.Query(context.Background(), "SELECT id, path, url, updated_at::TEXT, inactive FROM UrlRedirects WHERE path ILIKE $1 AND inactive=$2 LIMIT $3 OFFSET $4", pathMatchPattern, false, pageLimit, page)
		if db_err != nil {
			http.Error(w, dbError, http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		for rows.Next() {
			var temp Redirect
			rowErr := rows.Scan(&temp.Id, &temp.Path, &temp.Url, &temp.LastUpdated, &temp.Inactive)
			if rowErr == nil {
				responseData = append(responseData, temp)
			}
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
			http.Error(w, badRequest, http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		var responseData Redirect
		db_err := db.QueryRow(context.Background(), "SELECT id, path, url, updated_at::TEXT, inactive FROM UrlRedirects WHERE url=$1 LIMIT $2", requestData.Data, dbLimit).Scan(&responseData.Id, &responseData.Path, &responseData.Url, &responseData.LastUpdated, &responseData.Inactive)
		if db_err != nil {
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
			http.Error(w, dbError, http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(toJson(responseData))
	})
}

func stats(db *pgxpool.Pool) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var statsQueryPeriod StatsTime
		err := json.NewDecoder(r.Body).Decode(&statsQueryPeriod)
		startTime := time.Unix(statsQueryPeriod.Start, 0)
		endTime := time.Unix(statsQueryPeriod.End, 0)
		if err != nil {
			http.Error(w, badRequest, http.StatusBadRequest)
			return
		}
		statsData := LogStatsData{}
		queryResults, queryErr := db.Query(context.Background(),
			`SELECT 'path' AS col, path AS stat_key, count(id) AS stat_count FROM urlredirects_analytics WHERE log_timestamp BETWEEN TO_TIMESTAMP($1) AND TO_TIMESTAMP($2) GROUP BY path UNION ALL 
			 SELECT 'status' AS col, CAST(status AS VARCHAR) AS stat_key, count(id) AS stat_count FROM urlredirects_analytics WHERE log_timestamp BETWEEN TO_TIMESTAMP($1) AND TO_TIMESTAMP($2) GROUP BY status UNION ALL 
			 SELECT 'time' AS col, CAST(status AS VARCHAR) AS stat_key, CAST(avg(processing_time) AS INTEGER) AS stat_count FROM urlredirects_analytics WHERE log_timestamp BETWEEN TO_TIMESTAMP($1) AND TO_TIMESTAMP($2) GROUP BY status;`,
			startTime.Unix(), endTime.Unix())
		if queryErr != nil {
			http.Error(w, internalError, http.StatusInternalServerError)
			return
		}
		defer queryResults.Close()
		for queryResults.Next() {
			var dataItem LogQueryData
			var dataKey string
			var statKey pgtype.Text
			rowErr := queryResults.Scan(&dataKey, &statKey, &dataItem.StatCount)
			if rowErr != nil {
				continue
			}
			if len(strings.TrimSpace(statKey.String)) <= 0 {
				dataItem.StatKey = "Other"
			} else {
				dataItem.StatKey = strings.TrimSpace(statKey.String)
			}
			switch {
			case dataKey == "path":
				statsData.Path = append(statsData.Path, dataItem)
			case dataKey == "status":
				statsData.Status = append(statsData.Status, dataItem)
			case dataKey == "time":
				statsData.Time = append(statsData.Time, dataItem)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(toJson(statsData))
	})
}
