package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

var cfg *Config

//Config represents service configuration for dp-frontend-geography-controller
type Config struct {
	BindAddr     string `envconfig:"BIND_ADDR"`
	APIRouterURL string `envconfig:"API_ROUTER_URL"`

	SiteDomain               string    `envconfig:"SITE_DOMAIN"`
	PatternLibraryAssetsPath string    `envconfig:"PATTERN_LIBRARY_ASSETS_PATH"`
	SupportedLanguages       [2]string `envconfig:"SUPPORTED_LANGUAGES"`

	GracefulShutdownTimeout    time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	HealthCheckInterval        time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
}

// Get returns the default config with any modifications through environment
// variables
func Get() (cfg *Config, err error) {

	cfg = &Config{
		BindAddr:                   ":23700",
		APIRouterURL:               "http://localhost:22400",
		SiteDomain:                 "ons.gov.uk",
		SupportedLanguages:         [2]string{"en", "cy"},
		PatternLibraryAssetsPath:   "http://localhost:9000/dist",
		GracefulShutdownTimeout:    5 * time.Second,
		HealthCheckInterval:        30 * time.Second,
		HealthCheckCriticalTimeout: 90 * time.Second,
	}

	return cfg, envconfig.Process("", cfg)
}
