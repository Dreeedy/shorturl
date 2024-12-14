package zaplogger

import (
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

// Initialize инициализирует синглтон логера с необходимым уровнем логирования.
func NewZapLogger(cfg config.Config) (Logger, error) {
	// преобразуем текстовый уровень логирования в zap.AtomicLevel
	lvl, err := zap.ParseAtomicLevel(cfg.GetConfig().FlagLogLevel)
	if err != nil {
		return nil, err
	}
	// создаём новую конфигурацию логера
	zapCfg := zap.NewProductionConfig()
	// устанавливаем уровень
	zapCfg.Level = lvl
	// создаём логер на основе конфигурации
	zl, err := zapCfg.Build()
	if err != nil {
		return nil, err
	}

	return &zapLogger{
		config: cfg, logger: zl,
	}, nil
}

// Info реализация метода Info для ZapLogger
func (z *zapLogger) Info(msg string, fields ...zap.Field) {
	z.logger.Info(msg, fields...)
}
