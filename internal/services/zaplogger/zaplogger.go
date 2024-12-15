package zaplogger

import (
	"fmt"

	"github.com/Dreeedy/shorturl/internal/config"
	"go.uber.org/zap"
)

type Logger interface {
	Info(msg string, fields ...zap.Field)
}

type zapLogger struct {
	config config.Config
	logger *zap.Logger
}

// NewZapLogger Initialize initializes the logger singleton with the required logging level.
func NewZapLogger(cfg config.Config) (Logger, error) {
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

	return &zapLogger{
		config: cfg, logger: zl,
	}, nil
}

// Info implementation of the Info method for ZapLogger.
func (z *zapLogger) Info(msg string, fields ...zap.Field) {
	z.logger.Info(msg, fields...)
}
