package httplogger

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Dreeedy/shorturl/internal/config"
	"go.uber.org/zap"
)

type RqRsLogger interface {
	RqRsLogger(next http.Handler) http.Handler
}

type httpLogger struct {
	cfg config.Config
	log *zap.Logger
}

func NewHTTPLogger(newConfig config.Config, newLogger *zap.Logger) *httpLogger {
	return &httpLogger{
		cfg: newConfig,
		log: newLogger,
	}
}

func (ref *httpLogger) RqRsLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		log.Println("run httploggermiddleware")

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
	if err != nil {
		return 0, fmt.Errorf("failed to write response: %w", err)
	}
	r.size += size

	return size, nil
}
