package otelsdk

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	ServiceName             string            `json:"ServiceName"`
	Host                    string            `json:"Host"`
	Port                    int               `json:"Port"`
	Protocol                string            `json:"Protocol"`
	ShowConsoleMetrics      bool              `json:"ShowConsoleMetrics"`
	ShowConsoleTrace        bool              `json:"ShowConsoleTrace"`
	ExtraResourceAttributes map[string]string `json:"ExtraResourceAttributes"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}
	return ParseConfig(data)
}

func ParseConfig(data []byte) (*Config, error) {
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config JSON: %w", err)
	}
	cfg.normalize()
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (cfg *Config) normalize() {
	cfg.Protocol = strings.TrimSpace(strings.ToLower(cfg.Protocol))
	if cfg.Protocol == "" {
		cfg.Protocol = "http"
	}
}

func (cfg *Config) validate() error {
	if cfg.ServiceName == "" {
		return fmt.Errorf("ServiceName is required")
	}
	if cfg.Host == "" {
		return fmt.Errorf("Host is required")
	}
	if cfg.Port == 0 {
		return fmt.Errorf("Port is required")
	}
	if cfg.Protocol != "http" && cfg.Protocol != "grpc" {
		return fmt.Errorf("Protocol must be either \"http\" or \"grpc\", got %q", cfg.Protocol)
	}
	return nil
}

func (cfg *Config) Endpoint() string {
	return fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
}

func (cfg *Config) ResourceAttributes() map[string]string {
	return cfg.ExtraResourceAttributes
}

func (cfg *Config) ServiceNameResourceAttribute() string {
	return cfg.ServiceName
}

func (cfg *Config) Context() context.Context {
	return context.Background()
}
