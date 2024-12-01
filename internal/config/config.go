// Использование флаго конфигурации:
// go buid
// F:\shorturl> .\shortener -a :8888 -b http://localhost:8888
package config

import (
	"flag"
	"sync"
)

// Config структура для хранения конфигурации
type Config struct {
	RunAddr string
	BaseURL string
}

var (
	cfg Config
	// Использование sync.Once для гарантии однократной инициализации конфигурации
	cfgOnce sync.Once
)

// parseFlags функция для инициализации полей структуры Config на основе аргументов командной строки
func parseFlags() {
	flag.StringVar(&cfg.RunAddr, "a", ":8080", "address to run HTTP server")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "base URL for shortened URLs")
	flag.Parse()
}

// GetConfig функция для получения конфигурации
func GetConfig() Config {
	cfgOnce.Do(parseFlags)
	return cfg
}
