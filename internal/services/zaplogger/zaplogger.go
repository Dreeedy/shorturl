package zaplogger

import (
	"fmt"

	"github.com/Dreeedy/shorturl/internal/config"
	"go.uber.org/zap"
)

type ZapLogger interface {
	NewZapLogger(cfg config.Config) (*zap.Logger, error)
}

// NewZapLogger Initialize initializes the logger singleton with the required logging level.
func NewZapLogger(cfg config.Config) (*zap.Logger, error) {
	// convert the text logging level to zap.AtomicLevel.
	lvl, err := zap.ParseAtomicLevel(cfg.GetConfig().FlagLogLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to parse atomic level: %w", err)
	}
	// create a new logger configuration.
	zapCfg := zap.NewProductionConfig()
	// set the level.
	zapCfg.Level = lvl
	// create a logger based on the configuration.
	zl, err := zapCfg.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return zl, nil
}
