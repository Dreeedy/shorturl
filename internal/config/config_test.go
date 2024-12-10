package config

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConfig(t *testing.T) {
	// Save the original command-line arguments.
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	tests := []struct {
		name     string
		args     []string
		envVars  map[string]string
		expected HTTPConfig
	}{
		{
			name: "default values",
			args: []string{"cmd"},
			expected: HTTPConfig{
				RunAddr: ":8080",
				BaseURL: "http://localhost:8080",
			},
		},
		{
			name: "flag custom values",
			args: []string{"cmd", "-a", ":8888", "-b", "http://127.0.0.1:8888"},
			expected: HTTPConfig{
				RunAddr: ":8888",
				BaseURL: "http://127.0.0.1:8888",
			},
		},
		{ // Проверяет, что конфигурация загружается из переменных окружения, если они установлены.
			name: "environment variables",
			args: []string{"cmd"},
			envVars: map[string]string{
				"SERVER_ADDRESS": ":8081",
				"BASE_URL":       "http://example.com:8081",
			},
			expected: HTTPConfig{
				RunAddr: ":8081",
				BaseURL: "http://example.com:8081",
			},
		},
		{ // Проверяет, что переменные окружения имеют приоритет над флагами командной строки.
			name: "environment variables override flags",
			args: []string{"cmd", "-a", ":8888", "-b", "http://127.0.0.1:8888"},
			envVars: map[string]string{
				"SERVER_ADDRESS": ":8081",
				"BASE_URL":       "http://example.com:8081",
			},
			expected: HTTPConfig{
				RunAddr: ":8081",
				BaseURL: "http://example.com:8081",
			},
		},
		{ // Проверяет, что конфигурация корректно загружается при использовании как переменных окружения,
			// так и флагов командной строки.
			name: "mixed environment variables and flags",
			args: []string{"cmd", "-a", ":8888"},
			envVars: map[string]string{
				"BASE_URL": "http://example.com:8081",
			},
			expected: HTTPConfig{
				RunAddr: ":8888",
				BaseURL: "http://example.com:8081",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the command-line arguments for the test.
			os.Args = tt.args
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			// Set the environment variables for the test.
			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			// Create a new instance of MyConfig and get the config.
			config := NewConfig()
			cfg := config.GetConfig()

			// Assert the expected values.
			assert.Equal(t, tt.expected, cfg)
		})
	}
}
