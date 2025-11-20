package main

import (
	"log"
	"strings"

	"health-check-on-go/libs/config"
	"health-check-on-go/libs/health_check"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	buildHTTPServices(cfg.HTTP)
	buildSFTPServices(cfg.SFTP)

	Execute()
}

func buildHTTPServices(domains []config.HTTPDomainConfig) health_check.HTTPServices {
	services := health_check.HTTPServices{}

	for _, entry := range domains {
		host := strings.TrimSpace(entry.Host)
		if host == "" {
			continue
		}

		domain := health_check.Domain{Host: host}

		for _, endpoint := range entry.Endpoints {
			path := strings.TrimSpace(endpoint.Path)
			if path == "" {
				continue
			}
			method := strings.TrimSpace(endpoint.Method)

			schedule := strings.TrimSpace(endpoint.Schedule)

			expectedStatus := endpoint.ExpectedStatus

			if method == "" {
				method = "GET"
			}
			domain.SetUrl(path, strings.ToUpper(method), nil, schedule, expectedStatus)
		}

		if len(domain.Urls) == 0 {
			continue
		}

		services.SetDomain(domain)
	}

	globalContext.healthCheckContext.AddHttp(services)

	return services
}

func buildSFTPServices(servers []config.SFTPCheckConfig) health_check.SFTPServices {
	services := health_check.SFTPServices{}

	for _, entry := range servers {
		host := strings.TrimSpace(entry.Host)
		username := strings.TrimSpace(entry.Username)
		if host == "" || username == "" {
			continue
		}

		server := health_check.Server{
			Name:     strings.TrimSpace(entry.Name),
			Host:     host,
			Port:     entry.Port,
			Username: username,
			Password: entry.Password,
		}

		server.SetCommand(health_check.Command{
			Path: entry.Path,
			Mode: parseCommandMode(entry.Mode),
		})

		services.SetServer(server)
	}

	globalContext.healthCheckContext.AddSftp(services)

	return services
}

func parseCommandMode(value string) health_check.CommandMode {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(health_check.CommandModeList):
		return health_check.CommandModeList
	case string(health_check.CommandModeRead):
		return health_check.CommandModeRead
	default:
		return health_check.CommandModeStat
	}
}
