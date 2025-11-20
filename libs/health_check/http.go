package health_check

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	url2 "net/url"
	"os"
	path2 "path"
	"time"

	"github.com/robfig/cron/v3"
)

const DefaultResultsPath = "results/"

type HTTPServices struct {
	Domains []Domain
}

type Domain struct {
	Host    string
	Headers map[string]string
	Urls    []url
}

type url struct {
	Path           string
	Method         string
	Headers        map[string]string
	ExpectedStatus int
	Schedule       string
}

type Result struct {
	Success    bool   `json:"success"`
	Path       string `json:"path"`
	Domain     string `json:"domain"`
	StatusCode int    `json:"status_code"`
	Body       string `json:"body"`
	CreatedAt  int64  `json:"created_at"`
}

type ServicesTestResult struct {
	domain  string
	results []Result
}

func (httpServices *HTTPServices) SetDomain(d Domain) {
	httpServices.Domains = append(httpServices.Domains, d)
}

func (d *Domain) SetUrl(path string, method string, headers map[string]string, schedule string, expectedStatus int) {
	d.Urls = append(d.Urls, url{
		Path:           path,
		Method:         method,
		Headers:        headers,
		Schedule:       schedule,
		ExpectedStatus: expectedStatus,
	})
}

func saveRunInfo(result *Result, resultPath string) {
	payload, err := json.MarshalIndent(&result, "", "  ")
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(resultPath, payload, 0o644)
}

func saveRecordInIndex(result *Result, previousResults *[]Result, indexPath string) {
	*previousResults = append(*previousResults, *result)

	previousResultsText, err := json.MarshalIndent(previousResults, "", "  ")

	if err != nil {
		panic(err)
	}

	os.WriteFile(indexPath, previousResultsText, 0o644)
}

func runRequestSync(previousResults *[]Result, d Domain, u url, client http.Client) {
	fmt.Println(previousResults)
	dir := path2.Join(DefaultResultsPath, u.Path)

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}

	resultPath := path2.Join(dir, time.Now().Format(time.RFC3339))

	indexPath := path2.Join(dir, "index")

	fullURL, _ := url2.JoinPath(d.Host, u.Path)

	result := &Result{
		Path:      u.Path,
		Success:   false,
		Domain:    d.Host,
		CreatedAt: time.Now().Unix(),
	}

	defer saveRunInfo(result, resultPath)
	defer saveRecordInIndex(result, previousResults, indexPath)

	req, err := http.NewRequest(u.Method, fullURL, nil)

	for key, value := range d.Headers {
		req.Header.Add(key, value)
	}

	for key, value := range u.Headers {
		req.Header.Add(key, value)
	}

	if err != nil {
		result.Body = err.Error()
		result.StatusCode = 604

		return
	}

	res, err := client.Do(req)
	if err != nil {
		result.Body = err.Error()
		result.StatusCode = 602
		return
	}

	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		result.Body = err.Error()
		result.StatusCode = res.StatusCode
		return
	} else {
		result.StatusCode = res.StatusCode
		result.Body = string(body)
		result.Success = res.StatusCode == u.ExpectedStatus
	}
}

func loadPreviousRuns(u url) *[]Result {
	filePath := path2.Join(DefaultResultsPath, u.Path, "index")
	file, err := os.ReadFile(filePath)

	var previousRuns []Result

	err = json.Unmarshal(file, &previousRuns)

	if err != nil {
		return &previousRuns
	}

	return &previousRuns
}

func (httpServices *HTTPServices) RunHttpCrons() {
	c := cron.New()

	client := http.Client{}

	for _, domain := range httpServices.Domains {
		for _, url := range domain.Urls {
			previousRuns := loadPreviousRuns(url)

			_, err := c.AddFunc(url.Schedule, func() {
				runRequestSync(previousRuns, domain, url, client)
			})
			if err != nil {
				fmt.Println(err)
			}
		}
	}

	c.Start()

	select {}
}
