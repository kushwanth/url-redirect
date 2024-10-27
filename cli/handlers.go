package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func getRedirect(id int) {
	var response Redirect
	endPoint := httpsProtocol + apiHost + actionsApiEndpoint + "info/" + strconv.Itoa(id)
	client := &http.Client{}
	req, err := http.NewRequest("GET", endPoint, nil)
	if err != nil {
		respondAndExit("Request Creation failed", err)
	}
	req.Header.Add("X-Redirect-API-KEY", apiKey)
	res, err := client.Do(req)
	if err != nil || !strings.Contains(res.Status, successfulRequest) {
		respondAndExit("Info: Request Failed", res.StatusCode, res.Body)
	}
	json.NewDecoder(res.Body).Decode(&response)
	consoleDataWriter(response)
}

func createRedirect(url string, path string) {
	var response Redirect
	reqBody := UrlData{Url: url, Path: path}
	endPoint := httpsProtocol + apiHost + actionsApiEndpoint + "create"
	reqBodyBytes := bytes.NewBuffer(toJson(reqBody))
	client := &http.Client{}
	req, err := http.NewRequest("POST", endPoint, reqBodyBytes)
	if err != nil {
		respondAndExit("Request Creation failed", err)
	}
	req.Header.Set("X-Redirect-API-KEY", apiKey)
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil || !strings.Contains(res.Status, successfulRequest) {
		respondAndExit("Create: Request Failed", res.StatusCode, res.Body)
	}
	json.NewDecoder(res.Body).Decode(&response)
	consoleDataWriter(response)
}

func fixRedirect(url string, path string) {
	var response Redirect
	reqBody := UrlData{Url: url, Path: path}
	endPoint := httpsProtocol + apiHost + actionsApiEndpoint + "fix"
	reqBodyBytes := bytes.NewBuffer(toJson(reqBody))
	client := &http.Client{}
	req, err := http.NewRequest("PATCH", endPoint, reqBodyBytes)
	if err != nil {
		respondAndExit("Request Creation failed", err)
	}
	req.Header.Set("X-Redirect-API-KEY", apiKey)
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil || !strings.Contains(res.Status, successfulRequest) {
		respondAndExit("Fix: Request Failed", res.StatusCode, res.Body)
	}
	json.NewDecoder(res.Body).Decode(&response)
	consoleDataWriter(response)
}
func updateRedirect(id int, path string, url string) {
	var response Redirect
	reqBody := Redirect{Id: id, Url: url, Path: path, LastUpdated: time.Now().Format("YYYY-MM-DD hh:mm:ss"), Inactive: false}
	endPoint := httpsProtocol + apiHost + actionsApiEndpoint + "update/" + strconv.Itoa(id)
	reqBodyBytes := bytes.NewBuffer(toJson(reqBody))
	client := &http.Client{}
	req, err := http.NewRequest("PUT", endPoint, reqBodyBytes)
	if err != nil {
		respondAndExit("Request Creation failed", err)
	}
	req.Header.Set("X-Redirect-API-KEY", apiKey)
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil || !strings.Contains(res.Status, successfulRequest) {
		respondAndExit("Update: Request Failed", res.StatusCode, res.Body)
	}
	json.NewDecoder(res.Body).Decode(&response)
	consoleDataWriter(response)
}

func deleteRedirect(id int) {
	var response Redirect
	endPoint := httpsProtocol + apiHost + actionsApiEndpoint + "disable/" + strconv.Itoa(id)
	client := &http.Client{}
	req, err := http.NewRequest("DELETE", endPoint, nil)
	if err != nil {
		respondAndExit("Request Creation failed", err)
	}
	req.Header.Add("X-Redirect-API-KEY", apiKey)
	res, err := client.Do(req)
	if err != nil || !strings.Contains(res.Status, successfulRequest) {
		respondAndExit("Disable: Request Failed", res.StatusCode, res.Body)
	}
	json.NewDecoder(res.Body).Decode(&response)
	consoleDataWriter(response)
}

func listRedirect(page int) {
	var response []Redirect
	endPoint := httpsProtocol + apiHost + operationsApiEndpoint + "list?page=" + strconv.Itoa(page)
	client := &http.Client{}
	req, err := http.NewRequest("GET", endPoint, nil)
	if err != nil {
		respondAndExit("Request Creation failed", err)
	}
	req.Header.Add("X-Redirect-API-KEY", apiKey)
	res, err := client.Do(req)
	if err != nil || !strings.Contains(res.Status, successfulRequest) {
		respondAndExit("List: Request Failed", res.StatusCode, res.Body)
	}
	json.NewDecoder(res.Body).Decode(&response)
	consoleDataListWriter(response)
}

func searchRedirect(path string, page int) {
	var response []Redirect
	reqBody := SearchQuery{Data: path}
	endPoint := httpsProtocol + apiHost + operationsApiEndpoint + "searchPath?page=" + strconv.Itoa(page)
	reqBodyBytes := bytes.NewBuffer(toJson(reqBody))
	client := &http.Client{}
	req, err := http.NewRequest("POST", endPoint, reqBodyBytes)
	if err != nil {
		respondAndExit("Request Creation failed", err)
	}
	req.Header.Set("X-Redirect-API-KEY", apiKey)
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil || !strings.Contains(res.Status, successfulRequest) {
		respondAndExit("Search: Request Failed", res.StatusCode, res.Body)
	}
	json.NewDecoder(res.Body).Decode(&response)
	consoleDataListWriter(response)
}

func redirectExists(path string) {
	var response Redirect
	reqBody := SearchQuery{Data: path}
	endPoint := httpsProtocol + apiHost + operationsApiEndpoint + "destinationExists"
	reqBodyBytes := bytes.NewBuffer(toJson(reqBody))
	client := &http.Client{}
	req, err := http.NewRequest("POST", endPoint, reqBodyBytes)
	if err != nil {
		respondAndExit("Request Creation failed", err)
	}
	req.Header.Set("X-Redirect-API-KEY", apiKey)
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil || !strings.Contains(res.Status, successfulRequest) {
		respondAndExit("Exists: Request Failed", res.StatusCode, res.Body)
	}
	json.NewDecoder(res.Body).Decode(&response)
	consoleDataWriter(response)
}
