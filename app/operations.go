package main

import (
	"context"
	"encoding/json"
	"log"
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
			log.Println("listall -> ", pageErr.Error())
			page = 0
		}
		min, max := page, page+pageLimit
		rows, db_err := db.Query(context.Background(), "SELECT id, path, url, updated_at::TEXT, inactive FROM UrlRedirects WHERE id>$1 AND id<=$2 LIMIT $3", min, max, pageLimit)
		if db_err != nil {
			log.Println("listall -> ", db_err.Error())
			http.Error(w, notFoundMessage, http.StatusNotFound)
			return
		}
		defer rows.Close()
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
		defer rows.Close()
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

func stats(db *pgxpool.Pool) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var statsQueryPeriod StatsTime
		err := json.NewDecoder(r.Body).Decode(&statsQueryPeriod)
		startTime := time.Unix(statsQueryPeriod.Start, 0)
		endTime := time.Unix(statsQueryPeriod.End, 0)
		if err != nil {
			log.Println("stats -> ", err.Error())
			http.Error(w, badRequest, http.StatusBadRequest)
			return
		}
		// statsKey := [...]string{"path", "status", "browser", "os", "country", "devices", "time"}
		// statsQueries := [...]string{
		// 	"SELECT path AS stat_key, count(id) AS stat_count FROM urlredirects_analytics WHERE log_timestamp BETWEEN TO_TIMESTAMP($1) AND TO_TIMESTAMP($2) GROUP BY path;",
		// 	"SELECT status AS stat_key, count(id) AS stat_count FROM urlredirects_analytics WHERE log_timestamp BETWEEN TO_TIMESTAMP($1) AND TO_TIMESTAMP($2) GROUP BY status;",
		// 	"SELECT browser AS stat_key, count(id) AS stat_count FROM urlredirects_analytics WHERE log_timestamp BETWEEN TO_TIMESTAMP($1) AND TO_TIMESTAMP($2) GROUP BY browser;",
		// 	"SELECT os AS stat_key, count(id) AS stat_count FROM urlredirects_analytics WHERE log_timestamp BETWEEN TO_TIMESTAMP($1) AND TO_TIMESTAMP($2) GROUP BY os;",
		// 	"SELECT country AS stat_key, count(id) AS stat_count FROM urlredirects_analytics WHERE log_timestamp BETWEEN TO_TIMESTAMP($1) AND TO_TIMESTAMP($2) GROUP BY country;",
		// 	"SELECT device_type AS stat_key, count(id) AS stat_count FROM urlredirects_analytics WHERE log_timestamp BETWEEN TO_TIMESTAMP($1) AND TO_TIMESTAMP($2) GROUP BY device_type;",
		// 	"SELECT status AS stat_key, avg(processing_time) AS stat_count FROM urlredirects_analytics WHERE log_timestamp BETWEEN TO_TIMESTAMP($1) AND TO_TIMESTAMP($2) GROUP BY status;",
		// }
		// queryBatch := pgx.Batch{}
		// for _, statsQuery := range statsQueries {
		// 	queryBatch.Queue(statsQuery, startTime.Unix(), endTime.Unix())
		// }
		// queryResults := db.SendBatch(context.Background(), &queryBatch)
		// defer queryResults.Close()
		// for i, _ := range statsQueries {
		// 	var dataQueryList []LogQueryData
		// 	rows, err := queryResults.Query()
		// 	if err != nil {
		// 		log.Println(err)
		// 		continue
		// 	}
		// 	defer rows.Close()
		// 	for rows.Next() {
		// 		var dataItem LogQueryData
		// 		rowErr := rows.Scan(&dataItem.DataItem, &dataItem.ItemCount)
		// 		if rowErr != nil {
		// 			log.Println("stats ->", rowErr.Error())
		// 			continue
		// 		}
		// 		if len(strings.TrimSpace(dataItem.DataItem)) <= 0 {
		// 			dataItem.DataItem = "Other"
		// 		}
		// 		dataQueryList = append(dataQueryList, dataItem)
		// 	}
		// 	statsData[statsKey[i]] = dataQueryList
		// }
		statsData := LogStatsData{}
		queryResults, queryErr := db.Query(context.Background(),
			`SELECT 'path' AS col, path AS stat_key, count(id) AS stat_count FROM urlredirects_analytics WHERE log_timestamp BETWEEN TO_TIMESTAMP($1) AND TO_TIMESTAMP($2) GROUP BY path UNION ALL 
		 SELECT 'status' AS col, CAST(status AS VARCHAR) AS stat_key, count(id) AS stat_count FROM urlredirects_analytics WHERE log_timestamp BETWEEN TO_TIMESTAMP($1) AND TO_TIMESTAMP($2) GROUP BY status UNION ALL 
		 SELECT 'browser' AS col, browser AS stat_key, count(id) AS stat_count FROM urlredirects_analytics WHERE log_timestamp BETWEEN TO_TIMESTAMP($1) AND TO_TIMESTAMP($2) GROUP BY browser UNION ALL 
		 SELECT 'os' AS col, os AS stat_key, count(id) AS stat_count FROM urlredirects_analytics WHERE log_timestamp BETWEEN TO_TIMESTAMP($1) AND TO_TIMESTAMP($2) GROUP BY os UNION ALL 
		 SELECT 'country' AS col, country AS stat_key, count(id) AS stat_count FROM urlredirects_analytics WHERE log_timestamp BETWEEN TO_TIMESTAMP($1) AND TO_TIMESTAMP($2) GROUP BY country UNION ALL 
		 SELECT 'devices' AS col, device_type AS stat_key, count(id) AS stat_count FROM urlredirects_analytics WHERE log_timestamp BETWEEN TO_TIMESTAMP($1) AND TO_TIMESTAMP($2) GROUP BY device_type UNION ALL 
		 SELECT 'time' AS col, CAST(status AS VARCHAR) AS stat_key, CAST(avg(processing_time) AS INTEGER) AS stat_count FROM urlredirects_analytics WHERE log_timestamp BETWEEN TO_TIMESTAMP($1) AND TO_TIMESTAMP($2) GROUP BY status;`,
			startTime.Unix(), endTime.Unix())
		if queryErr != nil {
			log.Println(queryErr.Error())
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
				log.Println("stats ->", rowErr.Error())
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
			case dataKey == "browser":
				statsData.Browser = append(statsData.Browser, dataItem)
			case dataKey == "status":
				statsData.Status = append(statsData.Status, dataItem)
			case dataKey == "os":
				statsData.Os = append(statsData.Os, dataItem)
			case dataKey == "country":
				statsData.Country = append(statsData.Country, dataItem)
			case dataKey == "devices":
				statsData.Devices = append(statsData.Devices, dataItem)
			case dataKey == "time":
				statsData.Time = append(statsData.Time, dataItem)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(toJson(statsData))
	})
}
