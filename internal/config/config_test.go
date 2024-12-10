package config

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConfig(t *testing.T) {
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
		{
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
		{
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
		{
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
			// Create a new flag set for the test to avoid modifying the global state.
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			runAddr := fs.String("a", ":8080", "address to run HTTP server")
			baseURL := fs.String("b", "http://localhost:8080", "base URL for shortened URLs")

			// Set the command-line arguments for the test.
			err := fs.Parse(tt.args[1:])
			if err != nil {
				t.Fatalf("Не удалось разобрать флаги: %v", err)
			}

			// Set environment variables for the test.
			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			// Load configuration with test values.
			config := HTTPConfig{
				RunAddr: *runAddr,
				BaseURL: *baseURL,
			}
			if envRunAddr, ok := os.LookupEnv("SERVER_ADDRESS"); ok && envRunAddr != "" {
				config.RunAddr = envRunAddr
			}
			if envBaseURL, ok := os.LookupEnv("BASE_URL"); ok && envBaseURL != "" {
				config.BaseURL = envBaseURL
			}

			// Assert the expected values.
			assert.Equal(t, tt.expected, config)
		})
	}
}
