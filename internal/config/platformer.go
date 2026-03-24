package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// PlatformerConfig represents a platformer.yaml file.
type PlatformerConfig struct {
	Name    string            `yaml:"name"`
	Type    string            `yaml:"type"`
	Runtime string            `yaml:"runtime"`
	Handler string            `yaml:"handler"`
	Memory  int               `yaml:"memory"`
	Timeout int               `yaml:"timeout"`
	Env     map[string]string `yaml:"environment"`
	API     *APIConfig        `yaml:"api"`
	DB      *DBConfig         `yaml:"database"`
}

// APIConfig controls HTTP API Gateway provisioning.
type APIConfig struct {
	Enabled bool   `yaml:"enabled"`
	Stage   string `yaml:"stage"`
}

// DBConfig lists DynamoDB tables for the app.
type DBConfig struct {
	Tables []TableConfig `yaml:"tables"`
}

// TableConfig names a single DynamoDB table.
type TableConfig struct {
	Name string `yaml:"name"`
}

// Load reads and parses a platformer.yaml file at path.
func Load(path string) (*PlatformerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	var cfg PlatformerConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}

	if cfg.Name == "" {
		return nil, fmt.Errorf("%s: name is required", path)
	}
	if cfg.Type == "" {
		return nil, fmt.Errorf("%s: type is required", path)
	}
	if cfg.Runtime == "" {
		return nil, fmt.Errorf("%s: runtime is required", path)
	}
	if cfg.Handler == "" {
		return nil, fmt.Errorf("%s: handler is required", path)
	}

	return &cfg, nil
}
