// Package config loads config.yaml once at startup.
// All feature flags live here so services can check them without hitting the DB.
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Features struct {
	// Alert channel types
	EmailAlerts   bool `yaml:"email_alerts"`
	SlackAlerts   bool `yaml:"slack_alerts"`
	DiscordAlerts bool `yaml:"discord_alerts"`

	// User features
	AvatarUploads     bool `yaml:"avatar_uploads"`
	PublicStatusPages bool `yaml:"public_status_pages"`
	EmailVerification bool `yaml:"email_verification"`
	PasswordReset     bool `yaml:"password_reset"`

	// Coming soon
	SSLExpiryChecks   bool `yaml:"ssl_expiry_checks"`
	MultiRegionChecks bool `yaml:"multi_region_checks"`
	IncidentNotes     bool `yaml:"incident_notes"`
	OnCallSchedules   bool `yaml:"on_call_schedules"`
}

type Monitoring struct {
	WorkerTickSeconds           int `yaml:"worker_tick_seconds"`
	DefaultCheckIntervalSeconds int `yaml:"default_check_interval_seconds"`
	MaxResponseTimeMs           int `yaml:"max_response_time_ms"`
}

type Config struct {
	Features   Features   `yaml:"features"`
	Monitoring Monitoring `yaml:"monitoring"`
}

// Load reads config.yaml from path. The path defaults to "./config.yaml" and
// can be overridden via the CONFIG_PATH env var.
func Load() (*Config, error) {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = "config.yaml"
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	var cfg Config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}

	// Apply safe defaults so a partially-written config doesn't silently break things
	if cfg.Monitoring.WorkerTickSeconds == 0 {
		cfg.Monitoring.WorkerTickSeconds = 10
	}
	if cfg.Monitoring.DefaultCheckIntervalSeconds == 0 {
		cfg.Monitoring.DefaultCheckIntervalSeconds = 60
	}
	if cfg.Monitoring.MaxResponseTimeMs == 0 {
		cfg.Monitoring.MaxResponseTimeMs = 30000
	}

	return &cfg, nil
}
