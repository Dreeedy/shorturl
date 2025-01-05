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
	RunAddr         string
	BaseURL         string
	FlagLogLevel    string
	FileStoragePath string
}

func NewConfig() Config {
	config := &HTTPConfig{}

	flag.StringVar(&config.RunAddr, "a", ":8080", "address to run HTTP server")
	flag.StringVar(&config.BaseURL, "b", "http://localhost:8080", "base URL for shortened URLs")
	flag.StringVar(&config.FlagLogLevel, "l", "info", "log level")
	flag.StringVar(&config.FileStoragePath, "f", "default_filestorage.json",
		"path to the file where data in JSON format is saved")
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
	if fileStoragePath, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok && fileStoragePath != "" {
		config.FileStoragePath = fileStoragePath
	}

	return config
}

func (ref *HTTPConfig) GetConfig() HTTPConfig {
	return *ref
}
