package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"text/tabwriter"
)

var errorBytes []byte

const httpsProtocol = "https://"
const successfulRequest = "200"
const actionsApiEndpoint = "/api/action/"
const operationsApiEndpoint = "/api/operations/"

const commandHelpText = `NAME: 
   {{.Name}} - {{.Usage}}

USAGE: 
   {{.HelpName}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}

OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}
`

const appHelpText = `NAME: 
   {{.Name}} - {{.Usage}}

USAGE: 
   {{.HelpName}} {{if .Commands}}command{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}
   {{if len .Authors}}
AUTHOR:
   {{range .Authors}}{{ . }}{{end}}
   {{end}}{{if .Commands}}
COMMANDS:
{{range .Commands}}{{if not .HideHelp}}   {{join .Names ", "}}{{ "\t"}}{{.Usage}}{{ "\n" }}{{end}}{{end}}{{end}}{{if .VisibleFlags}}
GLOBAL OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}
`

var apiHost = os.Getenv("API_HOST")
var apiKey = os.Getenv("API_KEY")

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

type UrlParams struct {
	page int
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

func consoleDataWriter(r Redirect) {
	absoluteUrl := httpsProtocol + r.Url
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "ID:\t%d\n", r.Id)
	fmt.Fprintf(w, "Path:\t%s\n", r.Path)
	fmt.Fprintf(w, "URL:\t%s\n", absoluteUrl)
	fmt.Fprintf(w, "Inactive:\t%t\n", r.Inactive)
	fmt.Fprintf(w, "Updated:\t%s\n", r.LastUpdated)
	w.Flush()
	defer os.Exit(0)
}

func consoleDataListWriter(redirectList []Redirect) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tPath\tUrl\tInactive")
	fmt.Fprintln(w, "--\t----\t---\t--------")
	for _, r := range redirectList {
		fmt.Fprintf(w, "%d\t%s\t%s\t%t\n", r.Id, r.Path, r.Url, r.Inactive)
	}
	w.Flush()
	defer os.Exit(0)
}

func respondAndExit(msg string, args ...any) {
	log.Fatalln(msg, args)
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
