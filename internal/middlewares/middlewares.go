package middlewares

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"time"
)

// Middleware логгирует все запросы к серверу.
func LoggingRQMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log the request details
		log.Printf("")
		log.Printf("===== ===== RQ {")

		start := time.Now()
		log.Printf("Request: %s %s %s %s\n", r.Method, r.URL.Path, r.RemoteAddr, time.Since(start))
		log.Printf("Headers: %v", r.Header)
		log.Printf("Query Parameters: %v", r.URL.Query())

		// Log form values if the request method is POST or PUT
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			if err := r.ParseForm(); err != nil {
				log.Printf("Error parsing form: %v", err)
			} else {
				log.Printf("Form Values: %v", r.PostForm)
			}
		}

		// Log the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading body: %v", err)
		} else {
			log.Printf("Body: %s", body)
			// Restore the body for the next handler
			r.Body = io.NopCloser(bytes.NewBuffer(body))
		}

		log.Printf("===== ===== RQ }")

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
