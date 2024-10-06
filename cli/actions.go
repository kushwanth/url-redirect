package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const endPointPath = "/api/action/"

var apiHost = os.Getenv("API_HOST")
var apiKey = os.Getenv("API_KEY")

func getRedirect(id int) {
	var response Redirect
	endPoint := httpsProtocol + apiHost + endPointPath + "info/" + strconv.Itoa(id)
	client := &http.Client{}
	req, err := http.NewRequest("GET", endPoint, nil)
	if err != nil {
		respondAndExit("Request Creation failed")
	}
	req.Header.Add("X-Redirect-API-KEY", apiKey)
	res, err := client.Do(req)
	if err != nil || !strings.Contains(res.Status, successfulRequest) {
		respondAndExit("Info: Request Failed")
	}
	json.NewDecoder(res.Body).Decode(&response)
	responseString(response)
}

func createRedirect(url string, path string) {
	var response Redirect
	reqBody := UrlData{Url: url, Path: path}
	endPoint := httpsProtocol + apiHost + endPointPath + "create"
	reqBodyBytes := bytes.NewBuffer(toJson(reqBody))
	client := &http.Client{}
	req, err := http.NewRequest("POST", endPoint, reqBodyBytes)
	if err != nil {
		respondAndExit("Request Creation failed")
	}
	req.Header.Set("X-Redirect-API-KEY", apiKey)
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil || !strings.Contains(res.Status, successfulRequest) {
		respondAndExit("Create: Request Failed")
	}
	json.NewDecoder(res.Body).Decode(&response)
	responseString(response)
}

func fixRedirect(url string, path string) {
	var response Redirect
	reqBody := UrlData{Url: url, Path: path}
	endPoint := httpsProtocol + apiHost + endPointPath + "fix"
	reqBodyBytes := bytes.NewBuffer(toJson(reqBody))
	client := &http.Client{}
	req, err := http.NewRequest("PATCH", endPoint, reqBodyBytes)
	if err != nil {
		fmt.Print("Request Creation failed")
	}
	req.Header.Set("X-Redirect-API-KEY", apiKey)
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil || !strings.Contains(res.Status, successfulRequest) {
		respondAndExit("Fix: Request Failed")
	}
	json.NewDecoder(res.Body).Decode(&response)
	responseString(response)
}
func updateRedirect(url string, id int, path string) {
	var response Redirect
	reqBody := Redirect{Id: id, Url: url, Path: path, LastUpdated: time.Now().Format("YYYY-MM-DD hh:mm:ss"), Inactive: false}
	endPoint := httpsProtocol + apiHost + endPointPath + "update/" + strconv.Itoa(id)
	reqBodyBytes := bytes.NewBuffer(toJson(reqBody))
	client := &http.Client{}
	req, err := http.NewRequest("PUT", endPoint, reqBodyBytes)
	if err != nil {
		fmt.Print("Request Creation failed")
	}
	req.Header.Set("X-Redirect-API-KEY", apiKey)
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil || !strings.Contains(res.Status, successfulRequest) {
		respondAndExit("Update: Request Failed")
	}
	json.NewDecoder(res.Body).Decode(&response)
	responseString(response)
}

func deleteRedirect(id int) {
	var response Redirect
	endPoint := httpsProtocol + apiHost + endPointPath + "disable/" + strconv.Itoa(id)
	client := &http.Client{}
	req, err := http.NewRequest("DELETE", endPoint, nil)
	if err != nil {
		fmt.Print("Request Creation failed")
	}
	req.Header.Add("X-Redirect-API-KEY", apiKey)
	res, err := client.Do(req)
	if err != nil || !strings.Contains(res.Status, successfulRequest) {
		respondAndExit("Disable: Request Failed")
	}
	json.NewDecoder(res.Body).Decode(&response)
	responseString(response)
}
