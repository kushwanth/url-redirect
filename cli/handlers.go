package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/urfave/cli/v2"
)

func getStatus(*cli.Context) error {
	if isAPIUp() {
		fmt.Println("API is running")
	} else {
		fmt.Println("API is Down")
	}
	return nil
}

func getUrlRedirect(cCtx *cli.Context) error {
	idStr := cCtx.Args().Get(0)
	id, idErr := strconv.Atoi(idStr)
	if idErr != nil {
		respondAndExit("Args Error", idErr)
	}
	var response Redirect
	endPoint := httpsProtocol + apiHost + actionsApiEndpoint + "info/" + strconv.Itoa(id)
	res := apiService(http.MethodGet, endPoint, nil)
	if res.StatusCode != http.StatusOK {
		respondAndExit(res.Status)
	}
	json.NewDecoder(res.Body).Decode(&response)
	consoleDataWriter(response)
	return nil
}

func addUrlRedirect(cCtx *cli.Context) error {
	path, uri := cCtx.Args().Get(0), cCtx.Args().Get(1)
	_, pathErr := url.Parse(path)
	_, uriErr := url.Parse(uri)
	if pathErr != nil || uriErr != nil {
		respondAndExit("Args Error", pathErr, uriErr)
	}
	var response Redirect
	reqBody := UrlData{Url: uri, Path: path}
	endPoint := httpsProtocol + apiHost + actionsApiEndpoint + "create"
	reqBodyBytes := bytes.NewBuffer(toJson(reqBody))
	res := apiService(http.MethodPost, endPoint, reqBodyBytes)
	if res.StatusCode != http.StatusOK {
		respondAndExit(res.Status)
	}
	json.NewDecoder(res.Body).Decode(&response)
	consoleDataWriter(response)
	return nil
}

func updateUrlRedirect(cCtx *cli.Context) error {
	idStr, path, uri := cCtx.Args().Get(0), cCtx.Args().Get(1), cCtx.Args().Get(2)
	id, idErr := strconv.Atoi(idStr)
	_, pathErr := url.Parse(path)
	_, uriErr := url.Parse(uri)
	if idErr != nil || pathErr != nil || uriErr != nil {
		respondAndExit("Args Error", idErr, pathErr, uriErr)
	}
	var response Redirect
	reqBody := Redirect{Id: id, Url: uri, Path: path, LastUpdated: time.Now().Format("YYYY-MM-DD hh:mm:ss"), Inactive: false}
	endPoint := httpsProtocol + apiHost + actionsApiEndpoint + "update/" + strconv.Itoa(id)
	reqBodyBytes := bytes.NewBuffer(toJson(reqBody))
	res := apiService(http.MethodPut, endPoint, reqBodyBytes)
	if res.StatusCode != http.StatusOK {
		respondAndExit(res.Status)
	}
	json.NewDecoder(res.Body).Decode(&response)
	consoleDataWriter(response)
	return nil
}

func fixUrlRedirect(cCtx *cli.Context) error {
	path, uri := cCtx.Args().Get(0), cCtx.Args().Get(1)
	_, pathErr := url.Parse(path)
	_, uriErr := url.Parse(uri)
	if pathErr != nil || uriErr != nil {
		respondAndExit("Args Error", pathErr, uriErr)
	}
	var response Redirect
	reqBody := UrlData{Url: uri, Path: path}
	endPoint := httpsProtocol + apiHost + actionsApiEndpoint + "fix"
	reqBodyBytes := bytes.NewBuffer(toJson(reqBody))
	res := apiService(http.MethodPatch, endPoint, reqBodyBytes)
	if res.StatusCode != http.StatusOK {
		respondAndExit(res.Status)
	}
	json.NewDecoder(res.Body).Decode(&response)
	consoleDataWriter(response)
	return nil
}

func disableUrlRedirect(cCtx *cli.Context) error {
	idStr := cCtx.Args().Get(0)
	id, idErr := strconv.Atoi(idStr)
	if idErr != nil {
		respondAndExit("Args Error", idErr)
	}
	var response Redirect
	endPoint := httpsProtocol + apiHost + actionsApiEndpoint + "disable/" + strconv.Itoa(id)
	res := apiService(http.MethodDelete, endPoint, nil)
	if res.StatusCode != http.StatusOK {
		respondAndExit(res.Status)
	}
	json.NewDecoder(res.Body).Decode(&response)
	consoleDataWriter(response)
	return nil
}

func listUrlRedirects(cCtx *cli.Context) error {
	pageStr := cCtx.Args().Get(0)
	page, pageErr := strconv.Atoi(pageStr)
	if pageErr != nil {
		page = 0
	} else {
		page = page * 10
	}
	var response []Redirect
	endPoint := httpsProtocol + apiHost + operationsApiEndpoint + "list?page=" + strconv.Itoa(page)
	res := apiService(http.MethodGet, endPoint, nil)
	if res.StatusCode != http.StatusOK {
		respondAndExit(res.Status)
	}
	json.NewDecoder(res.Body).Decode(&response)
	consoleDataListWriter(response)
	return nil
}

func searchUrlRedirect(cCtx *cli.Context) error {
	path, pageStr := cCtx.Args().Get(0), cCtx.Args().Get(1)
	page, pageErr := strconv.Atoi(pageStr)
	_, pathErr := url.Parse(path)
	if pathErr != nil {
		respondAndExit("Args Error", pathErr)
	}
	if pageErr != nil {
		page = 0
	} else {
		page = page * 10
	}
	var response []Redirect
	reqBody := OpsData{Data: path}
	endPoint := httpsProtocol + apiHost + operationsApiEndpoint + "search?page=" + strconv.Itoa(page)
	reqBodyBytes := bytes.NewBuffer(toJson(reqBody))
	res := apiService(http.MethodPost, endPoint, reqBodyBytes)
	if res.StatusCode != http.StatusOK {
		respondAndExit(res.Status)
	}
	json.NewDecoder(res.Body).Decode(&response)
	consoleDataListWriter(response)
	return nil
}

func urlRedirectExists(cCtx *cli.Context) error {
	uri := cCtx.Args().Get(0)
	_, uriErr := url.Parse(uri)
	if uriErr != nil {
		respondAndExit("Args Error", uriErr)
	}
	var response Redirect
	reqBody := OpsData{Data: uri}
	endPoint := httpsProtocol + apiHost + operationsApiEndpoint + "check"
	reqBodyBytes := bytes.NewBuffer(toJson(reqBody))
	res := apiService(http.MethodPost, endPoint, reqBodyBytes)
	if res.StatusCode != http.StatusOK {
		respondAndExit(res.Status)
	}
	json.NewDecoder(res.Body).Decode(&response)
	consoleDataWriter(response)
	return nil
}

func generateShortRedirect(cCtx *cli.Context) error {
	uri := cCtx.Args().Get(0)
	_, uriErr := url.Parse(uri)
	if uriErr != nil {
		respondAndExit("Args Error", uriErr)
	}
	var response Redirect
	reqBody := OpsData{Data: uri}
	endPoint := httpsProtocol + apiHost + operationsApiEndpoint + "generate"
	reqBodyBytes := bytes.NewBuffer(toJson(reqBody))
	res := apiService(http.MethodPost, endPoint, reqBodyBytes)
	if res.StatusCode != http.StatusOK {
		respondAndExit(res.Status)
	}
	json.NewDecoder(res.Body).Decode(&response)
	consoleDataWriter(response)
	return nil
}

func getRedirectStats(cCtx *cli.Context) error {
	timeFrame := ((int(cCtx.Value("days").(int)) * 24) + int(cCtx.Value("hours").(int))) * -1
	reqBody := StatsTime{Start: time.Now().Add(time.Duration(timeFrame) * time.Hour).Unix(), End: time.Now().Unix()}
	endPoint := httpsProtocol + apiHost + operationsApiEndpoint + "stats"
	reqBodyBytes := bytes.NewBuffer(toJson(reqBody))
	var response LogStatsData
	res := apiService(http.MethodPost, endPoint, reqBodyBytes)
	if res.StatusCode != http.StatusOK {
		respondAndExit(res.Status)
	}
	json.NewDecoder(res.Body).Decode(&response)
	consoleStatsWriter(response)
	return nil
}
