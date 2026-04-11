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

// defaults returns a Config with all features enabled and sensible monitoring values.
// Used when config.yaml is missing so the service starts without a file on first run.
func defaults() *Config {
	return &Config{
		Features: Features{
			EmailAlerts:       true,
			SlackAlerts:       true,
			DiscordAlerts:     true,
			AvatarUploads:     true,
			PublicStatusPages: true,
			EmailVerification: true,
			PasswordReset:     true,
		},
		Monitoring: Monitoring{
			WorkerTickSeconds:           10,
			DefaultCheckIntervalSeconds: 60,
			MaxResponseTimeMs:           30000,
		},
	}
}

// Load reads config.yaml. Path defaults to "./config.yaml" and can be overridden
// via CONFIG_PATH. If the file does not exist the service starts with defaults
// (all features on) and logs a warning — it never hard-crashes on a missing file.
func Load() (*Config, error) {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = "config.yaml"
	}

	f, err := os.Open(path)
	if os.IsNotExist(err) {
		// Not fatal — fall back to defaults so the service still starts.
		fmt.Printf("WARN: %s not found, using built-in defaults (all features enabled)\n", path)
		return defaults(), nil
	}
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	cfg := defaults() // start from defaults so missing yaml keys don't zero out
	if err := yaml.NewDecoder(f).Decode(cfg); err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}

	// Re-apply monitoring floor values in case yaml had zeros
	if cfg.Monitoring.WorkerTickSeconds == 0 {
		cfg.Monitoring.WorkerTickSeconds = 10
	}
	if cfg.Monitoring.DefaultCheckIntervalSeconds == 0 {
		cfg.Monitoring.DefaultCheckIntervalSeconds = 60
	}
	if cfg.Monitoring.MaxResponseTimeMs == 0 {
		cfg.Monitoring.MaxResponseTimeMs = 30000
	}

	return cfg, nil
}
