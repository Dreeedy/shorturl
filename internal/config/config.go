package config

import (
	"flag"
	"os"
)

type Config interface {
	GetConfig() HTTPConfig
}

// HTTPConfig structure for storing the configuration.
type HTTPConfig struct {
	RunAddr      string
	BaseURL      string
	FlagLogLevel string
}

func NewConfig() Config {
	config := &HTTPConfig{}

	flag.StringVar(&config.RunAddr, "a", ":8080", "address to run HTTP server")
	flag.StringVar(&config.BaseURL, "b", "http://localhost:8080", "base URL for shortened URLs")
	flag.StringVar(&config.FlagLogLevel, "l", "info", "log level")
	flag.Parse()

	// Override values from environment variables if they are set.
	if envRunAddr, ok := os.LookupEnv("SERVER_ADDRESS"); ok && envRunAddr != "" {
		config.RunAddr = envRunAddr
	}
	if envBaseURL, ok := os.LookupEnv("BASE_URL"); ok && envBaseURL != "" {
		config.BaseURL = envBaseURL
	}
	if envFlagLogLevel, ok := os.LookupEnv("LOG_LEVEL"); ok && envFlagLogLevel != "" {
		config.FlagLogLevel = envFlagLogLevel
	}

	return config
}

func (ref *HTTPConfig) GetConfig() HTTPConfig {
	return *ref
}
