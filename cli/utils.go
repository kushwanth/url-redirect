package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"text/tabwriter"
)

var errorBytes []byte

const httpsProtocol = "https://"
const actionsApiEndpoint = "/api/action/"
const operationsApiEndpoint = "/api/operations/"
const cliVersion = "2.1"

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

var apiHost, apiHostAvailable = os.LookupEnv("API_HOST")
var apiKey, apiKeyAvailable = os.LookupEnv("API_KEY")

var deviceTypeMapper = map[string]string{
	"D": "Desktop/PC",
	"M": "Mobile/Tablet",
	"C": "Redirector CLI",
	"B": "Bot",
	"U": "Unknown",
}

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

type OpsData struct {
	Data string `json:"data,omitempty"`
}

type StatsTime struct {
	Start int64 `json:"start,omitempty"`
	End   int64 `json:"end,omitempty"`
}

type LogStatsData struct {
	Path    []LogStatsDataList `json:"path,omitempty"`
	Status  []LogStatsDataList `json:"status,omitempty"`
	Country []LogStatsDataList `json:"country,omitempty"`
	Time    []LogStatsDataList `json:"time,omitempty"`
	Browser []LogStatsDataList `json:"browser,omitempty"`
	Os      []LogStatsDataList `json:"os,omitempty"`
	Devices []LogStatsDataList `json:"devices,omitempty"`
}

type LogStatsDataList struct {
	StatKey   string `json:"stat_key,omitempty"`
	StatCount int    `json:"stat_count,omitempty"`
}

func getCliVersion() string {
	return fmt.Sprintf("v%s", cliVersion)
}

func isAPIUp() bool {
	areApiEnvAvailable := apiHostAvailable && apiKeyAvailable
	if !areApiEnvAvailable {
		defer os.Exit(0)
		return false
	}
	url := httpsProtocol + apiHost + "/app/health"
	res, err := http.Get(url)
	if err != nil {
		return false && areApiEnvAvailable
	}
	return res.StatusCode == http.StatusOK && areApiEnvAvailable
}

func apiService(method string, apiEndpoint string, data io.Reader) *http.Response {
	client := &http.Client{}
	req, err := http.NewRequest(method, apiEndpoint, data)
	if err != nil {
		respondAndExit("Request Creation failed", err)
	}
	req.Header.Set("x-url-redirect-token", apiKey)
	req.Header.Set("x-url-redirect-version", cliVersion)
	req.Header.Set("Content-Type", "application/json")
	if !isAPIUp() {
		respondAndExit("API Down")
	}
	res, err := client.Do(req)
	if err != nil {
		respondAndExit("API Request Failed", res.StatusCode)
	}
	return res
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

func consoleStatsListWriter(statCategory string, statsList []LogStatsDataList) {
	sort.Slice(statsList, func(x, y int) bool {
		return statsList[x].StatCount > statsList[y].StatCount
	})
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "%s\t%s\n", statCategory, "Count")
	fmt.Fprintln(w, "-------\t-----")
	for _, statItem := range statsList {
		fmt.Fprintf(w, "%s\t%v\n", statItem.StatKey, statItem.StatCount)
	}
	fmt.Fprintln(w)
	defer w.Flush()
}

func consoleStatsWriter(statsData LogStatsData) {
	consoleStatsListWriter("Path", statsData.Path)
	consoleStatsListWriter("Status", statsData.Status)
	// consoleStatsListWriter("Browser", statsData.Browser)
	// consoleStatsListWriter("OS", statsData.Os)
	// consoleStatsListWriter("Time", statsData.Time)
	// consoleStatsListWriter("Country", statsData.Country)
	// sort.Slice(statsData.Devices, func(x, y int) bool {
	// 	return statsData.Devices[x].StatCount > statsData.Devices[y].StatCount
	// })
	// w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	// fmt.Fprintln(w, "Device\tCount")
	// fmt.Fprintln(w, "------\t-----")
	// for _, deviceItem := range statsData.Devices {
	// 	deviceKey, err := deviceTypeMapper[deviceItem.StatKey]
	// 	if !err {
	// 		deviceKey = deviceItem.StatKey
	// 	}
	// 	fmt.Fprintf(w, "%s\t%v\n", deviceKey, deviceItem.StatCount)
	// }
	// fmt.Fprintln(w)
	// w.Flush()
	defer os.Exit(0)
}

func respondAndExit(msg string, args ...any) {
	fmt.Println(msg, args)
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
