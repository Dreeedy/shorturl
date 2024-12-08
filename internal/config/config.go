package config

import (
	"flag"
	"os"
	"sync"
)

// Config структура для хранения конфигурации.
type Config struct {
	RunAddr string
	BaseURL string
}

var (
	cfg Config
	// Использование sync.Once для гарантии однократной инициализации конфигурации.
	cfgOnce sync.Once
)

// parseFlags функция для инициализации полей структуры Config на основе аргументов командной строки.
func parseFlags() {
	flag.StringVar(&cfg.RunAddr, "a", ":8080", "address to run HTTP server")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "base URL for shortened URLs")
	flag.Parse()

	// Переопределение значений из переменных окружения, если они установлены.
	if envRunAddr, ok := os.LookupEnv("SERVER_ADDRESS"); ok && envRunAddr != "" {
		cfg.RunAddr = envRunAddr
	}
	if envBaseURL, ok := os.LookupEnv("BASE_URL"); ok && envBaseURL != "" {
		cfg.BaseURL = envBaseURL
	}
}

// GetConfig функция для получения конфигурации.
func GetConfig() Config {
	cfgOnce.Do(parseFlags)
	return cfg
}
