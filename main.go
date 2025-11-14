package main

import (
	"encoding/json"
	"health-check-on-go/libs/health_check"
	"strings"
)

type catFact struct {
	Fact   string `json:"fact"`
	Length int    `json:"length"`
}

type catFactsResponse struct {
	Data []catFact `json:"data"`
}

func main() {
	buildHTTPServices()
	buildSFTPServices()

	Execute()
}

func buildHTTPServices() health_check.HTTPServices {
	httpServices := health_check.HTTPServices{}

	catFacts := health_check.Domain{
		Host: "https://catfact.ninja",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	catFacts.SetUrl(
		"/fact",
		"GET",
		nil,
		[]health_check.Assert{
			{
				Fn: func(response string) bool {
					var fact catFact
					if err := json.Unmarshal([]byte(response), &fact); err != nil {
						return false
					}
					return strings.TrimSpace(fact.Fact) != ""
				},
				Description: "Fact endpoint returns text",
			},
		},
	)

	catFacts.SetUrl(
		"/facts",
		"GET",
		nil,
		[]health_check.Assert{
			{
				Fn: func(response string) bool {
					var facts catFactsResponse
					if err := json.Unmarshal([]byte(response), &facts); err != nil {
						return false
					}
					return len(facts.Data) > 0
				},
				Description: "Facts endpoint returns at least one entry",
			},
			{
				Fn: func(response string) bool {
					var facts catFactsResponse
					if err := json.Unmarshal([]byte(response), &facts); err != nil {
						return false
					}
					for _, fact := range facts.Data {
						if len(strings.TrimSpace(fact.Fact)) == 0 {
							return false
						}
					}
					return true
				},
				Description: "Facts payload includes valid text entries",
			},
		},
	)

	httpServices.SetDomain(catFacts)

	globalContext.healthCheckContext.AddHttp(httpServices)

	return httpServices
}

func buildSFTPServices() health_check.SFTPServices {
	server := health_check.Server{
		Name:     "Free SFTP server",
		Host:     "test.rebex.net",
		Username: "demo",
		Password: "password",
	}

	server.SetCommand(health_check.Command{
		Name: "List directory",
		Path: "/",
		Mode: health_check.CommandModeList,
		Asserts: []health_check.Assert{
			{
				Fn: func(output string) bool {
					return strings.TrimSpace(output) != ""
				},
				Description: "Directory is not empty",
			},
		},
	})

	server.SetCommand(health_check.Command{
		Name: "Validate path exists",
		Path: "/",
		Mode: health_check.CommandModeStat,
		Asserts: []health_check.Assert{
			{
				Fn: func(output string) bool {
					return strings.Contains(output, "dir=true") || strings.Contains(output, "dir=false")
				},
				Description: "Stat response returned metadata",
			},
		},
	})

	services := health_check.SFTPServices{}
	services.SetServer(server)

	globalContext.healthCheckContext.AddSftp(services)

	return services
}
