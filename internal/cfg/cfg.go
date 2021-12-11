package cfg

import (
	"github.com/caarlos0/env/v6"
	"log"
)

type Cfg struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
}

func Get() Cfg {
	var cfg Cfg
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	return cfg
}
