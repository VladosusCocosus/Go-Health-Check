package server

import (
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	html "github.com/gofiber/template/html/v2"

	"health-check-on-go/libs/config"
	"health-check-on-go/libs/health_check"
)

type templateData struct {
	Saved       string
	HTTPDomains []config.HTTPDomainConfig
	SFTPList    []config.SFTPCheckConfig
	HTTPForm    config.HTTPDomainConfig
	SFTPForm    config.SFTPCheckConfig
}

func RunServer() {
	engine := html.New("./libs/server/templates", ".html")
	app := fiber.New(fiber.Config{Views: engine})

	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	app.Get("/", getIndex)
	app.Get("/config", getConfig)
	app.Post("/config/http", postHTTPConfig)
	app.Post("/config/sftp", postSFTPConfig)

	log.Fatal(app.Listen(":3000"))
}

func getIndex(c *fiber.Ctx) error {
	cfg, err := config.Load()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("failed to load configuration: %v", err))
	}

	data := templateData{
		Saved:       c.Query("saved"),
		HTTPDomains: cfg.HTTP,
		SFTPList:    cfg.SFTP,
		HTTPForm:    defaultHTTPDomainConfig(),
		SFTPForm:    defaultSFTPConfig(),
	}

	return c.Render("index", data)
}

func getConfig(c *fiber.Ctx) error {
	cfg, err := config.Load()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(cfg)
}

func postHTTPConfig(c *fiber.Ctx) error {
	cfg, err := config.Load()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("failed to load configuration: %v", err))
	}

	entry, err := parseHTTPDomainForm(c.Body())
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	cfg.HTTP = append(cfg.HTTP, entry)
	if err := config.Save(cfg); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("failed to save configuration: %v", err))
	}

	return c.Redirect("/?saved=http", fiber.StatusSeeOther)
}

func postSFTPConfig(c *fiber.Ctx) error {
	cfg, err := config.Load()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("failed to load configuration: %v", err))
	}

	entry := config.SFTPCheckConfig{
		Name:     c.FormValue("sftp_name"),
		Host:     strings.TrimSpace(c.FormValue("sftp_host")),
		Port:     parsePort(c.FormValue("sftp_port")),
		Username: strings.TrimSpace(c.FormValue("sftp_username")),
		Password: c.FormValue("sftp_password"),
		Path:     strings.TrimSpace(c.FormValue("sftp_path")),
		Mode:     normalizeSFTPMode(c.FormValue("sftp_mode")),
	}

	if err := entry.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	cfg.SFTP = append(cfg.SFTP, entry)
	if err := config.Save(cfg); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("failed to save configuration: %v", err))
	}

	return c.Redirect("/?saved=sftp", fiber.StatusSeeOther)
}

func parseHTTPDomainForm(body []byte) (config.HTTPDomainConfig, error) {
	values, err := url.ParseQuery(string(body))
	if err != nil {
		return config.HTTPDomainConfig{}, fmt.Errorf("invalid form payload: %w", err)
	}

	entry := config.HTTPDomainConfig{
		Name: strings.TrimSpace(values.Get("http_name")),
		Host: strings.TrimSpace(values.Get("http_host")),
	}

	paths := values["http_endpoint_path"]
	methods := values["http_endpoint_method"]

	for idx, path := range paths {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}

		method := "GET"
		if idx < len(methods) {
			m := strings.TrimSpace(methods[idx])
			if m != "" {
				method = strings.ToUpper(m)
			}
		}

		entry.Endpoints = append(entry.Endpoints, config.HTTPEndpoint{Path: path, Method: method})
	}

	if err := entry.Validate(); err != nil {
		return config.HTTPDomainConfig{}, err
	}

	return entry, nil
}

func defaultHTTPDomainConfig() config.HTTPDomainConfig {
	return config.HTTPDomainConfig{
		Endpoints: []config.HTTPEndpoint{{Method: "GET"}},
	}
}

func defaultSFTPConfig() config.SFTPCheckConfig {
	return config.SFTPCheckConfig{Port: 22, Mode: string(health_check.CommandModeStat)}
}

func normalizeSFTPMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case string(health_check.CommandModeList):
		return string(health_check.CommandModeList)
	case string(health_check.CommandModeRead):
		return string(health_check.CommandModeRead)
	default:
		return string(health_check.CommandModeStat)
	}
}

func parsePort(value string) int {
	if value == "" {
		return 22
	}
	if port, err := strconv.Atoi(value); err == nil && port > 0 {
		return port
	}
	return 22
}
