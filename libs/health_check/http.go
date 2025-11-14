package health_check

import (
	"fmt"
	"health-check-on-go/libs/utility"
	"io"
	"net/http"
	url2 "net/url"
	"strings"
	"sync"

	"github.com/fatih/color"
)

type HTTPServices struct {
	Domains []Domain
}

type Domain struct {
	Host    string
	Headers map[string]string
	Urls    []url
}

type Assert struct {
	Fn          func(body string) bool
	Description string
}

type url struct {
	Path    string
	Method  string
	Headers map[string]string
	Asserts []Assert
}

type Result struct {
	Path       string
	StatusCode int
	Body       string
	Asserts    []map[string]bool
}

type ServicesTestResult struct {
	domain  string
	results []Result
}

func (httpServices *HTTPServices) SetDomain(d Domain) {
	httpServices.Domains = append(httpServices.Domains, d)
}

func (d *Domain) SetUrl(path string, method string, headers map[string]string, asserts []Assert) {
	d.Urls = append(d.Urls, url{Path: path, Method: method, Headers: headers, Asserts: asserts})
}

func runRequest(d Domain, u url, client http.Client, ch chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()

	result := Result{
		Path: u.Path,
	}

	fullURL, _ := url2.JoinPath(d.Host, u.Path)

	req, err := http.NewRequest(u.Method, fullURL, nil)

	for key, value := range d.Headers {
		req.Header.Add(key, value)
	}

	for key, value := range u.Headers {
		req.Header.Add(key, value)
	}

	if err != nil {
		ch <- Result{StatusCode: 600, Body: err.Error()}
		return
	}

	res, err := client.Do(req)
	if err != nil {
		ch <- Result{StatusCode: 602, Body: err.Error()}
		return
	}

	// Don't forget to close the body
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		ch <- Result{StatusCode: res.StatusCode, Body: err.Error()}
	} else {
		result.StatusCode = res.StatusCode
		result.Body = string(body)

	}

	for _, assert := range u.Asserts {
		success := assert.Fn(string(body))

		result.Asserts = append(result.Asserts, map[string]bool{
			assert.Description: success,
		})
	}

	ch <- result
}

func getChanLen(d Domain) int {
	var sum int
	for _, u := range d.Urls {
		sum += len(u.Asserts)
	}

	return sum
}

func (d *Domain) testUrls(globalResults chan<- ServicesTestResult, wg *sync.WaitGroup) {
	defer wg.Done()
	client := &http.Client{}

	var requestsWG sync.WaitGroup

	chanLen := getChanLen(*d)

	urlResults := make(chan Result, chanLen)

	for _, u := range d.Urls {
		requestsWG.Add(1)
		go runRequest(*d, u, *client, urlResults, &requestsWG)
	}

	go func() {
		requestsWG.Wait()
		close(urlResults)
	}()

	var domainResults []Result

	for result := range urlResults {
		domainResults = append(domainResults, result)
	}

	globalResults <- ServicesTestResult{
		domain:  d.Host,
		results: domainResults,
	}
}

const (
	urlLabelWidth = 40
)

func httpStatusBadge(code int) string {
	switch {
	case code >= 200 && code < 300:
		return color.GreenString("%d", code)
	case code >= 300 && code < 400:
		return color.YellowString("%d", code)
	case code == 0:
		return color.YellowString("pending")
	default:
		return color.RedString("%d", code)
	}
}

func printDomainHeader(domain string, success bool) {
	fmt.Println(utility.DividerLine())
	fmt.Printf("%s %s\n", utility.StatusBadge(success), color.CyanString(domain))
}

func printURLSummary(path string, statusCode int, success bool) {
	fmt.Printf("%s%s %-*s %s\n", utility.Indent, utility.StatusBadge(success), urlLabelWidth, path, httpStatusBadge(statusCode))
}

func (httpServices *HTTPServices) RunTesting() {
	var wg sync.WaitGroup
	domainsChan := make(chan ServicesTestResult, len(httpServices.Domains))

	for _, d := range httpServices.Domains {
		wg.Add(1)
		go d.testUrls(domainsChan, &wg)
	}

	go func() {
		wg.Wait()
		close(domainsChan)
	}()

	for domainResult := range domainsChan {
		success := true
		for _, result := range domainResult.results {
			for _, assert := range result.Asserts {
				for _, value := range assert {
					if !value {
						success = false
					}
				}
			}
		}

		printDomainHeader(domainResult.domain, success)

		for _, url := range domainResult.results {
			domainSuccess := true
			for _, assert := range url.Asserts {
				for _, value := range assert {
					if !value {
						domainSuccess = false
					}
				}
			}

			printURLSummary(url.Path, url.StatusCode, domainSuccess)

			if !domainSuccess && strings.TrimSpace(url.Body) != "" {
				fmt.Printf("%s%sBody: %s\n", utility.Indent, utility.Indent, color.YellowString(utility.FormatSnippet(url.Body, "no response body", utility.DefaultSnippetLimit)))
			}

			for _, value := range url.Asserts {
				for description, r := range value {
					fmt.Printf("%s%s %s\n", utility.DoubleIndent, utility.StatusBadge(r), description)
				}
			}

			fmt.Println()
		}

		fmt.Println()
	}
}
