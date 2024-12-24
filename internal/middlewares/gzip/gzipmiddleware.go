package gzip

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Dreeedy/shorturl/internal/services/zaplogger"
)

type GzipMiddleware interface {
	CompressionHandler(next http.Handler) http.Handler
}

type gzipMiddleware struct {
	log zaplogger.Logger
}

func NewGzipMiddleware(newLogger zaplogger.Logger) *gzipMiddleware {
	return &gzipMiddleware{
		log: newLogger,
	}
}

// allows the server to transparently compress the transmitted data and set the correct HTTP headers.
type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	n, err := c.zw.Write(p)
	if err != nil {
		return n, fmt.Errorf("gzip.Writer.Write: %w", err)
	}
	return n, nil
}

func (c *compressWriter) WriteHeader(statusCode int) {
	maxStatusCode := 300
	if statusCode < maxStatusCode {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

// Close closes the gzip.Writer and flushes all data from the buffer.
func (c *compressWriter) Close() error {
	err := c.zw.Close()
	if err != nil {
		return fmt.Errorf("gzip.Writer.Close: %w", err)
	}
	return nil
}

// allows the server to transparently decompress the data received from the client.
type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("gzip.NewReader: %w", err)
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c *compressReader) Read(p []byte) (n int, err error) {
	n, err = c.zr.Read(p)
	if err != nil {
		return n, fmt.Errorf("gzip.Reader.Read: %w", err)
	}
	return n, nil
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return fmt.Errorf("io.ReadCloser.Close: %w", err)
	}
	if err := c.zr.Close(); err != nil {
		return fmt.Errorf("gzip.Reader.Close: %w", err)
	}
	return nil
}

// CompressionHandler is a middleware function for handling data compression and decompression.
func (ref *gzipMiddleware) CompressionHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// by default, set the original http.ResponseWriter as the one that will be passed to the next function.
		ow := w

		// check if the client can receive compressed data from the server in gzip format.
		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			// wrap the original http.ResponseWriter with a new one that supports compression.
			cw := newCompressWriter(w)
			// change the original http.ResponseWriter to the new one.
			ow = cw
			// do not forget to send all compressed data to the client after the middleware is finished.
			defer func() {
				if err := cw.Close(); err != nil {
					ref.log.Info(fmt.Sprintf("Error closing compressWriter: %v", err))
				}
			}()
		}

		// check if the client sent compressed data to the server in gzip format.
		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			// wrap the request body in an io.Reader that supports decompression.
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				ref.log.Info(fmt.Sprintf("Error creating compressReader: %v", err))
				return
			}
			// change the request body to the new one.
			r.Body = cr
			defer func() {
				if err := cr.Close(); err != nil {
					ref.log.Info(fmt.Sprintf("Error closing compressReader: %v", err))
				}
			}()
		}

		// pass control to the handler.
		next.ServeHTTP(ow, r)
	})
}
