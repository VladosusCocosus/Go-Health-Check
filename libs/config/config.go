package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const DefaultFilePath = "configs/health_checks.json"

type HealthCheckConfig struct {
	HTTP []HTTPDomainConfig `json:"http"`
	SFTP []SFTPCheckConfig  `json:"sftp"`
}

type HTTPDomainConfig struct {
	Name      string         `json:"name"`
	Host      string         `json:"host"`
	Endpoints []HTTPEndpoint `json:"endpoints"`
}

type HTTPEndpoint struct {
	Path   string `json:"path"`
	Method string `json:"method"`
}

type SFTPCheckConfig struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Path     string `json:"path"`
	Mode     string `json:"mode"`
}

func Load() (HealthCheckConfig, error) {
	return LoadFrom(DefaultFilePath)
}

func LoadFrom(path string) (HealthCheckConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return HealthCheckConfig{}, nil
		}
		return HealthCheckConfig{}, err
	}

	var cfg HealthCheckConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return HealthCheckConfig{}, err
	}

	if shouldMigrateHTTP(cfg.HTTP) {
		if upgraded, err := migrateLegacyHTTP(data); err == nil && len(upgraded) > 0 {
			cfg.HTTP = upgraded
		}
	}

	return cfg, nil
}

func Save(cfg HealthCheckConfig) error {
	return SaveTo(DefaultFilePath, cfg)
}

func SaveTo(path string, cfg HealthCheckConfig) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	payload, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, payload, 0o644)
}

type legacyHTTPEntry struct {
	Name   string `json:"name"`
	Host   string `json:"host"`
	Path   string `json:"path"`
	Method string `json:"method"`
}

type legacyConfig struct {
	HTTP []legacyHTTPEntry `json:"http"`
}

func shouldMigrateHTTP(domains []HTTPDomainConfig) bool {
	if len(domains) == 0 {
		return false
	}
	for _, d := range domains {
		if len(d.Endpoints) > 0 {
			return false
		}
	}
	return true
}

func migrateLegacyHTTP(data []byte) ([]HTTPDomainConfig, error) {
	var legacy legacyConfig
	if err := json.Unmarshal(data, &legacy); err != nil {
		return nil, err
	}
	if len(legacy.HTTP) == 0 {
		return nil, nil
	}

	result := make([]HTTPDomainConfig, 0)
	index := make(map[string]int)

	for _, entry := range legacy.HTTP {
		host := strings.TrimSpace(entry.Host)
		path := strings.TrimSpace(entry.Path)
		if host == "" || path == "" {
			continue
		}
		name := strings.TrimSpace(entry.Name)
		key := fmt.Sprintf("%s|%s", host, name)
		pos, ok := index[key]
		if !ok {
			result = append(result, HTTPDomainConfig{Name: name, Host: host})
			pos = len(result) - 1
			index[key] = pos
		}
		method := strings.TrimSpace(entry.Method)
		if method == "" {
			method = "GET"
		}
		result[pos].Endpoints = append(result[pos].Endpoints, HTTPEndpoint{
			Path:   path,
			Method: strings.ToUpper(method),
		})
	}

	return result, nil
}

func (cfg HTTPDomainConfig) Validate() error {
	if strings.TrimSpace(cfg.Host) == "" {
		return fmt.Errorf("HTTP host is required")
	}
	if len(cfg.Endpoints) == 0 {
		return fmt.Errorf("add at least one endpoint")
	}
	for _, ep := range cfg.Endpoints {
		if err := ep.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (ep HTTPEndpoint) Validate() error {
	if strings.TrimSpace(ep.Path) == "" {
		return fmt.Errorf("HTTP endpoint path is required")
	}
	if strings.TrimSpace(ep.Method) == "" {
		return fmt.Errorf("HTTP method is required")
	}
	return nil
}

func (cfg SFTPCheckConfig) Validate() error {
	if strings.TrimSpace(cfg.Host) == "" {
		return fmt.Errorf("SFTP host is required")
	}
	if strings.TrimSpace(cfg.Username) == "" {
		return fmt.Errorf("SFTP username is required")
	}
	return nil
}
