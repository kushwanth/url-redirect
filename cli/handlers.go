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
	id := cCtx.Value("id").(int)
	if id < 0 {
		respondAndExit("Args Error", id)
	}
	var redirectData Redirect
	endPoint := httpsProtocol + apiHost + redirectorApiEndpoint + "info/" + strconv.Itoa(id)
	res := apiService(http.MethodGet, endPoint, nil)
	if res.StatusCode != http.StatusOK {
		respondAndExit(res.Status)
	}
	json.NewDecoder(res.Body).Decode(&redirectData)
	consoleDataWriter(redirectData)
	return nil
}

func addUrlRedirect(cCtx *cli.Context) error {
	path, uri := cCtx.Args().Get(0), cCtx.Args().Get(1)
	_, pathErr := url.Parse(path)
	_, uriErr := url.Parse(uri)
	if pathErr != nil || uriErr != nil {
		respondAndExit("Args Error", pathErr, uriErr)
	}
	var redirectData Redirect
	reqBody := UrlData{Url: uri, Path: path}
	endPoint := httpsProtocol + apiHost + redirectorApiEndpoint + "create"
	reqBodyBytes := bytes.NewBuffer(toJson(reqBody))
	res := apiService(http.MethodPost, endPoint, reqBodyBytes)
	if res.StatusCode != http.StatusOK {
		respondAndExit(res.Status)
	}
	json.NewDecoder(res.Body).Decode(&redirectData)
	consoleDataWriter(redirectData)
	return nil
}

func updateUrlRedirect(cCtx *cli.Context) error {
	path, uri := cCtx.Args().Get(0), cCtx.Args().Get(1)
	_, pathErr := url.Parse(path)
	_, uriErr := url.Parse(uri)
	id := cCtx.Value("id").(int)
	if id < 0 || pathErr != nil || uriErr != nil {
		respondAndExit("Args Error", id, pathErr, uriErr)
	}
	var redirectData Redirect
	reqBody := Redirect{Id: id, Url: uri, Path: path, LastUpdated: time.Now().Format("YYYY-MM-DD hh:mm:ss"), Inactive: false}
	endPoint := httpsProtocol + apiHost + redirectorApiEndpoint + "update/" + strconv.Itoa(id)
	reqBodyBytes := bytes.NewBuffer(toJson(reqBody))
	res := apiService(http.MethodPut, endPoint, reqBodyBytes)
	if res.StatusCode != http.StatusOK {
		respondAndExit(res.Status)
	}
	json.NewDecoder(res.Body).Decode(&redirectData)
	consoleDataWriter(redirectData)
	return nil
}

func fixUrlRedirect(cCtx *cli.Context) error {
	path, uri := cCtx.Args().Get(0), cCtx.Args().Get(1)
	_, pathErr := url.Parse(path)
	_, uriErr := url.Parse(uri)
	if pathErr != nil || uriErr != nil {
		respondAndExit("Args Error", pathErr, uriErr)
	}
	var redirectData Redirect
	reqBody := UrlData{Url: uri, Path: path}
	endPoint := httpsProtocol + apiHost + redirectorApiEndpoint + "fix"
	reqBodyBytes := bytes.NewBuffer(toJson(reqBody))
	res := apiService(http.MethodPatch, endPoint, reqBodyBytes)
	if res.StatusCode != http.StatusOK {
		respondAndExit(res.Status)
	}
	json.NewDecoder(res.Body).Decode(&redirectData)
	consoleDataWriter(redirectData)
	return nil
}

func disableUrlRedirect(cCtx *cli.Context) error {
	id := cCtx.Value("id").(int)
	if id < 0 {
		respondAndExit("Args Error", id)
	}
	var redirectData Redirect
	endPoint := httpsProtocol + apiHost + redirectorApiEndpoint + "disable/" + strconv.Itoa(id)
	res := apiService(http.MethodDelete, endPoint, nil)
	if res.StatusCode != http.StatusOK {
		respondAndExit(res.Status)
	}
	json.NewDecoder(res.Body).Decode(&redirectData)
	consoleDataWriter(redirectData)
	return nil
}

func listUrlRedirects(cCtx *cli.Context) error {
	page := cCtx.Value("page").(int)
	if page > 0 {
		page = page * 10
	} else {
		page = 0
	}
	var redirectDataList []Redirect
	endPoint := httpsProtocol + apiHost + redirectorApiEndpoint + "list?page=" + strconv.Itoa(page)
	res := apiService(http.MethodGet, endPoint, nil)
	if res.StatusCode != http.StatusOK {
		respondAndExit(res.Status)
	}
	json.NewDecoder(res.Body).Decode(&redirectDataList)
	consoleDataListWriter(redirectDataList)
	return nil
}

func searchUrlRedirect(cCtx *cli.Context) error {
	path := cCtx.Args().Get(0)
	page := cCtx.Value("page").(int)
	_, pathErr := url.Parse(path)
	if pathErr != nil {
		respondAndExit("Args Error", pathErr)
	}
	if page > 0 {
		page = page * 10
	} else {
		page = 0
	}
	var redirectDataList []Redirect
	reqBody := OpsData{Data: path}
	endPoint := httpsProtocol + apiHost + redirectorApiEndpoint + "search?page=" + strconv.Itoa(page)
	reqBodyBytes := bytes.NewBuffer(toJson(reqBody))
	res := apiService(http.MethodPost, endPoint, reqBodyBytes)
	if res.StatusCode != http.StatusOK {
		respondAndExit(res.Status)
	}
	json.NewDecoder(res.Body).Decode(&redirectDataList)
	consoleDataListWriter(redirectDataList)
	return nil
}

func urlRedirectExists(cCtx *cli.Context) error {
	uri := cCtx.Args().Get(0)
	_, uriErr := url.Parse(uri)
	if uriErr != nil {
		respondAndExit("Args Error", uriErr)
	}
	var redirectData Redirect
	reqBody := OpsData{Data: uri}
	endPoint := httpsProtocol + apiHost + redirectorApiEndpoint + "check"
	reqBodyBytes := bytes.NewBuffer(toJson(reqBody))
	res := apiService(http.MethodPost, endPoint, reqBodyBytes)
	if res.StatusCode != http.StatusOK {
		respondAndExit(res.Status)
	}
	json.NewDecoder(res.Body).Decode(&redirectData)
	consoleDataWriter(redirectData)
	return nil
}

func generateShortRedirect(cCtx *cli.Context) error {
	uri := cCtx.Args().Get(0)
	_, uriErr := url.Parse(uri)
	if uriErr != nil {
		respondAndExit("Args Error", uriErr)
	}
	var redirectData Redirect
	reqBody := OpsData{Data: uri}
	endPoint := httpsProtocol + apiHost + redirectorApiEndpoint + "generate"
	reqBodyBytes := bytes.NewBuffer(toJson(reqBody))
	res := apiService(http.MethodPost, endPoint, reqBodyBytes)
	if res.StatusCode != http.StatusOK {
		respondAndExit(res.Status)
	}
	json.NewDecoder(res.Body).Decode(&redirectData)
	consoleDataWriter(redirectData)
	return nil
}

func getRedirectStats(cCtx *cli.Context) error {
	timeFrame := ((int(cCtx.Value("days").(int)) * 24) + int(cCtx.Value("hours").(int))) * -1
	reqBody := StatsTime{Start: time.Now().Add(time.Duration(timeFrame) * time.Hour).Unix(), End: time.Now().Unix()}
	endPoint := httpsProtocol + apiHost + redirectorApiEndpoint + "stats"
	reqBodyBytes := bytes.NewBuffer(toJson(reqBody))
	var LogStatsDataList LogStatsDataList
	res := apiService(http.MethodPost, endPoint, reqBodyBytes)
	if res.StatusCode != http.StatusOK {
		respondAndExit(res.Status)
	}
	json.NewDecoder(res.Body).Decode(&LogStatsDataList)
	consoleStatsWriter(LogStatsDataList)
	return nil
}
