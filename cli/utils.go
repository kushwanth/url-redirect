package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

var errorBytes []byte

const httpsProtocol = "https://"
const successfulRequest = "200"

type ResponseMessage struct {
	Message string
}

type Redirect struct {
	Id          int    `json:"id,omitempty"`
	Path        string `json:"path,omitempty"`
	Url         string `json:"url,omitempty"`
	LastUpdated string `json:"lastUpdated,omitempty"`
	Inactive    bool   `json:"inactive,omitempty"`
}

type UrlData struct {
	Url  string `json:"url,omitempty"`
	Path string `json:"path,omitempty"`
}

type SearchQuery struct {
	Data string `json:"data,omitempty"`
}

func isAPIUp() bool {
	url := httpsProtocol + apiHost + "/app/health"
	res, err := http.Get(url)
	if err != nil {
		return false
	}
	return strings.Contains(res.Status, successfulRequest)
}

func responseString(r Redirect) {
	fmt.Printf("ID: %d\nPath: %s\nURL: %s\nLast Updated: %s\nRedirect Inactive: %t\n\n", r.Id, r.Path, r.Url, r.LastUpdated, r.Inactive)
	defer os.Exit(0)
}

func respondAndExit(msg string) {
	fmt.Println(msg)
	defer os.Exit(1)
}

func toJson(struc interface{}) []byte {
	responseMessageJson, err := json.Marshal(struc)
	if err != nil {
		return errorBytes
	} else {
		return responseMessageJson
	}
}
