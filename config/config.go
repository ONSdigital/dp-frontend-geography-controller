package config

import "github.com/ian-kent/gofigure"

var cfg Config

//Config represents service configuration for dp-frontend-geography-controller
type Config struct {
	BindAddr       string `env:"BIND_ADDR"`
	RendererURL    string `env:"RENDERER_URL"`
	CodeListAPIURL string `env:"CODELIST_API_URL"`
	DatasetAPIURL  string `env:"DATASET_API_URL"`
	EnableLoop11   bool   `env:"ENABLE_LOOP11"`
}

func init() {
	cfg = Config{
		BindAddr:       ":23700",
		RendererURL:    "http://localhost:20010",
		CodeListAPIURL: "http://localhost:22400",
		DatasetAPIURL:  "http://localhost:22000",
		EnableLoop11:   false,
	}
	err := gofigure.Gofigure(&cfg)
	if err != nil {
		panic(err)
	}
}

//Get ...
func Get() Config {
	return cfg
}
