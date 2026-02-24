package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all application configuration.
type Config struct {
	PollInterval    time.Duration `yaml:"poll_interval"`
	OrphanThreshold time.Duration `yaml:"orphan_threshold"`
	Connections     []Connection  `yaml:"connections"`
}

// Connection represents a named database connection profile.
// Each connection has its own default queue.
type Connection struct {
	Name         string `yaml:"name"`
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	Database     string `yaml:"database"`
	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	SSLMode      string `yaml:"sslmode"`
	DefaultQueue string `yaml:"default_queue"`
}

// Load reads and parses a YAML config file from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	cfg := &Config{
		PollInterval:    5 * time.Second,
		OrphanThreshold: 30 * time.Minute,
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks that the config has all required fields.
func (c *Config) Validate() error {
	if len(c.Connections) == 0 {
		return fmt.Errorf("config: at least one connection is required")
	}

	for i, conn := range c.Connections {
		if conn.Name == "" {
			return fmt.Errorf("config: connections[%d].name is required", i)
		}
		if conn.Host == "" {
			return fmt.Errorf("config: connections[%d].host is required", i)
		}
		if conn.Database == "" {
			return fmt.Errorf("config: connections[%d].database is required", i)
		}
		if conn.Username == "" {
			return fmt.Errorf("config: connections[%d].username is required", i)
		}
		if conn.Port == 0 {
			c.Connections[i].Port = 5432
		}
		if conn.SSLMode == "" {
			c.Connections[i].SSLMode = "prefer"
		}
		if conn.DefaultQueue == "" {
			c.Connections[i].DefaultQueue = "default"
		}
	}

	if c.PollInterval < 1*time.Second {
		c.PollInterval = 1 * time.Second
	}

	if c.OrphanThreshold < 1*time.Minute {
		c.OrphanThreshold = 1 * time.Minute
	}

	return nil
}

// GetConnection finds a connection by name.
func (c *Config) GetConnection(name string) (*Connection, error) {
	for i := range c.Connections {
		if c.Connections[i].Name == name {
			return &c.Connections[i], nil
		}
	}
	return nil, fmt.Errorf("connection %q not found", name)
}

// ConnString builds a PostgreSQL connection string for a connection.
func ConnString(conn *Connection) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		conn.Username, conn.Password, conn.Host, conn.Port, conn.Database, conn.SSLMode,
	)
}

// FindConfigPath returns the first config file found in the search order:
// 1. Explicit path (if non-empty)
// 2. $PROCRASTINATE_CONFIG env var
// 3. ~/.config/procrastinate-cli/config.yaml
// 4. ./config.yaml
func FindConfigPath(explicit string) (string, error) {
	if explicit != "" {
		if _, err := os.Stat(explicit); err != nil {
			return "", fmt.Errorf("config file not found: %s", explicit)
		}
		return explicit, nil
	}

	if envPath := os.Getenv("PROCRASTINATE_CONFIG"); envPath != "" {
		if _, err := os.Stat(envPath); err == nil {
			return envPath, nil
		}
	}

	home, err := os.UserHomeDir()
	if err == nil {
		xdgPath := filepath.Join(home, ".config", "procrastinate-cli", "config.yaml")
		if _, err := os.Stat(xdgPath); err == nil {
			return xdgPath, nil
		}
	}

	if _, err := os.Stat("config.yaml"); err == nil {
		return "config.yaml", nil
	}

	return "", fmt.Errorf("no config file found; create one at ~/.config/procrastinate-cli/config.yaml or use --config")
}
