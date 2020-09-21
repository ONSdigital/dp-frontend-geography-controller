package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

var cfg *Config

//Config represents service configuration for dp-frontend-geography-controller
type Config struct {
	BindAddr                   string        `envconfig:"BIND_ADDR"`
	APIRouterURL               string        `envconfig:"API_ROUTER_URL"`
	RendererURL                string        `envconfig:"RENDERER_URL"`
	GracefulShutdownTimeout    time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	HealthCheckInterval        time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
}

// Get returns the default config with any modifications through environment
// variables
func Get() (cfg *Config, err error) {

	cfg = &Config{
		BindAddr:                   ":23700",
		APIRouterURL:               "http://localhost:23200/v1",
		RendererURL:                "http://localhost:20010",
		GracefulShutdownTimeout:    5 * time.Second,
		HealthCheckInterval:        30 * time.Second,
		HealthCheckCriticalTimeout: 90 * time.Second,
	}

	return cfg, envconfig.Process("", cfg)
}
