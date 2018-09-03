package config

import "github.com/ian-kent/gofigure"

var cfg Config

//Config represents service configuration for dp-frontend-geography-controller
type Config struct {
	BindAddr        string `env:"BIND_ADDR"`
	RendererURL     string `env:"RENDERER_URL"`
	CodeListsAPIURL string `env:"CODELISTS_API_URL"`
}

func init() {
	cfg = Config{
		BindAddr:        ":23700",
		RendererURL:     "http://localhost:20010",
		CodeListsAPIURL: "http://localhost:22400",
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
