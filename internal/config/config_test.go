// Выполнить все тесты в проекте:
// F:\shorturl> go test ./... -v

package config

import (
	"flag"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConfig(t *testing.T) {
	// Save the original command-line arguments
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	tests := []struct {
		name     string
		args     []string
		expected Config
	}{
		{
			name: "default values",
			args: []string{"cmd"},
			expected: Config{
				RunAddr: ":8080",
				BaseURL: "http://localhost:8080",
			},
		},
		{
			name: "custom values",
			args: []string{"cmd", "-a", ":8888", "-b", "http://127.0.0.1:8888"},
			expected: Config{
				RunAddr: ":8888",
				BaseURL: "http://127.0.0.1:8888",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the command-line arguments for the test
			os.Args = tt.args
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			// Reset the config singleton
			cfgOnce = sync.Once{}

			// Get the config
			cfg := GetConfig()

			// Assert the expected values
			assert.Equal(t, tt.expected, cfg)
		})
	}
}
