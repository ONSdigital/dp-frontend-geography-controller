package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

var cfg *Config

//Config represents service configuration for dp-frontend-geography-controller
type Config struct {
	BindAddr                   string        `envconfig:"BIND_ADDR"`
	RendererURL                string        `envconfig:"RENDERER_URL"`
	CodeListAPIURL             string        `envconfig:"CODELIST_API_URL"`
	DatasetAPIURL              string        `envconfig:"DATASET_API_URL"`
	EnableLoop11               bool          `envconfig:"ENABLE_LOOP11"`
	GracefulShutdownTimeout    time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	HealthCheckInterval        time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
}

// Get returns the default config with any modifications through environment
// variables
func Get() (cfg *Config, err error) {

	cfg = &Config{
		BindAddr:                   ":23700",
		RendererURL:                "http://localhost:20010",
		CodeListAPIURL:             "http://localhost:22400",
		DatasetAPIURL:              "http://localhost:22000",
		EnableLoop11:               false,
		GracefulShutdownTimeout:    5 * time.Second,
		HealthCheckInterval:        10 * time.Second,
		HealthCheckCriticalTimeout: time.Minute,
	}

	return cfg, envconfig.Process("", cfg)
}
