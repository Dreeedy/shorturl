// # Сборка приложения
// go build -o shortener

// # Запуск с переменными окружения
// $env:SERVER_ADDRESS=":8081"; $env:BASE_URL="http://localhost:8081"; ./shortener.exe

// # Запуск с флагами командной строки
// ./shortener -a :8888 -b http://localhost:8888

// # Запуск с переменными окружения и флагами командной строки (переменные окружения имеют приоритет)
// $env:SERVER_ADDRESS=":8081"; $env:BASE_URL="http://localhost:8081"; ./shortener -a :8888 -b http://localhost:8888

package config

import (
	"flag"
	"os"
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

	// Переопределение значений из переменных окружения, если они установлены
	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		cfg.RunAddr = envRunAddr
	}
	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		cfg.BaseURL = envBaseURL
	}
}

// GetConfig функция для получения конфигурации
func GetConfig() Config {
	cfgOnce.Do(parseFlags)
	return cfg
}
