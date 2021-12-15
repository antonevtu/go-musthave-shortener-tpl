package cfg

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
	"log"
	"os"
)

type Cfg struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:":8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:"./storage.txt"`
}

func Get() Cfg {
	var cfg Cfg

	// Заполнение cfg значениями из переменных окружения, в том числе дефолтными значениями
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(os.Args)

	// Если заданы аргументы командной строки - перетираем значения переменных окружения
	flag.Func("a", "server address for shorten", func(flagValue string) error {
		cfg.ServerAddress = flagValue
		return nil
	})
	flag.Func("b", "base url for expand", func(flagValue string) error {
		cfg.BaseURL = flagValue
		return nil
	})
	flag.Func("f", "base url for expand", func(flagValue string) error {
		cfg.FileStoragePath = flagValue
		return nil
	})

	flag.Parse()
	return cfg
}
