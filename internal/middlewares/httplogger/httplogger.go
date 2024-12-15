package httplogger

import (
	"net/http"
	"time"

	"github.com/Dreeedy/shorturl/internal/config"
	"github.com/Dreeedy/shorturl/internal/services/zaplogger"
	"go.uber.org/zap"
)

type RqRsLogger interface {
	RequestLogger(msg string, fields ...zap.Field)
}

type httpLogger struct {
	cfg config.Config
	log zaplogger.Logger
}

func NewHTTPLogger(config config.Config, logger zaplogger.Logger) *httpLogger {
	return &httpLogger{
		cfg: config,
		log: logger,
	}
}

func (ref *httpLogger) RqRsLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rec := responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(&rec, r)

		duration := time.Since(start)
		ref.log.Info("HTTP request and response",
			zap.String("uri", r.RequestURI),
			zap.String("method", r.Method),
			zap.Duration("duration", duration),
			zap.Int("status", rec.statusCode),
			zap.Int("size", rec.size),
		)
	})
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(bytes []byte) (int, error) {
	size, err := r.ResponseWriter.Write(bytes)
	r.size += size
	return size, err
}
