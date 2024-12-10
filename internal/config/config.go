package config

import (
	"flag"
	"os"
)

type Config interface {
	GetConfig() MyConfig
}

// MyConfig структура для хранения конфигурации.
type MyConfig struct {
	RunAddr string
	BaseURL string
}

func NewMyConfig() Config {
	config := &MyConfig{}

	flag.StringVar(&config.RunAddr, "a", ":8080", "address to run HTTP server")
	flag.StringVar(&config.BaseURL, "b", "http://localhost:8080", "base URL for shortened URLs")
	flag.Parse()

	// Переопределение значений из переменных окружения, если они установлены.
	if envRunAddr, ok := os.LookupEnv("SERVER_ADDRESS"); ok && envRunAddr != "" {
		config.RunAddr = envRunAddr
	}
	if envBaseURL, ok := os.LookupEnv("BASE_URL"); ok && envBaseURL != "" {
		config.BaseURL = envBaseURL
	}

	return config
}

func (ref *MyConfig) GetConfig() MyConfig {
	return *ref
}
