package configRoutes

import (
	"fmt"
	"health-check-on-go/libs/config"
	"health-check-on-go/libs/health_check"
	"net/http"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func GetIndex(c *fiber.Ctx) error {
	cfg, err := config.Load()

	if err != nil {
		return err
	}

	return c.JSON(cfg)
}

func SaveHTTPConfig(c *fiber.Ctx) error {
	cfg, err := config.Load()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("failed to load configuration: %v", err))
	}

	var entry config.HTTPDomainConfig
	if err := c.BodyParser(&entry); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(fmt.Sprintf("invalid payload: %v", err))
	}

	fmt.Println(entry)

	entry.Name = strings.TrimSpace(entry.Name)
	entry.Host = strings.TrimSpace(entry.Host)

	for idx := range entry.Endpoints {
		endpoint := &entry.Endpoints[idx]
		endpoint.Path = strings.TrimSpace(endpoint.Path)
		endpoint.Schedule = strings.TrimSpace(endpoint.Schedule)
		endpoint.Method = strings.ToUpper(strings.TrimSpace(endpoint.Method))

		if endpoint.Method == "" {
			endpoint.Method = http.MethodGet
		}

		if endpoint.ExpectedStatus == 0 {
			endpoint.ExpectedStatus = http.StatusOK
		}
	}

	if err := entry.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	cfg.HTTP = append(cfg.HTTP, entry)

	if err := config.Save(cfg); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("failed to save configuration: %v", err))
	}

	return c.JSON(cfg)
}

func SaveSFTPConfig(c *fiber.Ctx) error {
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

	return c.JSON(cfg)
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
