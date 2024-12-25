package gzip

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

type GzipMiddleware interface {
	CompressionHandler(next http.Handler) http.Handler
}

type gzipMiddleware struct {
}

func NewGzipMiddleware() *gzipMiddleware {
	return &gzipMiddleware{}
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
	size, err := c.zw.Write(p)
	if err != nil {
		return size, fmt.Errorf("gzip.Writer.Write: %w", err)
	}

	return size, nil
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
	reader, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("gzip.NewReader: %w", err)
	}

	return &compressReader{
		r:  r,
		zr: reader,
	}, nil
}

func (c compressReader) Read(p []byte) (n int, err error) {
	size, err := c.zr.Read(p)
	if err != nil {
		if errors.Is(err, io.EOF) {
			log.Printf("gzip.Reader.Read: EOF")
			return size, io.EOF
		}
		log.Printf("gzip.Reader.Read: %v", err)
		return size, fmt.Errorf("compressReader.Read: %w", err)
	}
	return size, nil
}

func (c *compressReader) Close() error {
	errIo := c.r.Close()
	if errIo != nil {
		return fmt.Errorf("io.ReadCloser.Close: %w", errIo)
	}
	errGzip := c.zr.Close()
	if errGzip != nil {
		return fmt.Errorf("gzip.Reader.Close: %w", errGzip)
	}

	return nil
}

func (ref *gzipMiddleware) CompressionHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w

		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		contentType := r.Header.Get("Content-Type")
		isCompressible := contentType == "application/json" || contentType == "text/html"

		if supportsGzip && isCompressible {
			cw := newCompressWriter(w)
			ow = cw
			defer func() {
				if err := cw.Close(); err != nil {
					log.Printf("Error closing compressWriter: %v", err)
				}
			}()
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				log.Printf("Error creating compressReader: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer func() {
				if err := cr.Close(); err != nil {
					log.Printf("Error closing compressReader: %v", err)
				}
			}()
		}

		next.ServeHTTP(ow, r)
	})
}
