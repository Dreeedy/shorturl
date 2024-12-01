// Code increment #1 DONE
// Code increment #2 DONE
// Code increment #3 DONE
// Code increment #4 DONE
// Code increment #5 DONE

package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/Dreeedy/shorturl/internal/config"
	"github.com/Dreeedy/shorturl/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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

func main() {
	cfg := config.GetConfig()

	// Выводим конфигурацию
	log.Printf("Running server on %s\n", cfg.RunAddr)
	log.Printf("Base URL for shortened URLs: %s\n", cfg.BaseURL)

	r := chi.NewRouter()
	r.Use(middleware.Logger) // Use the built-in logger middleware from chi
	r.Use(LoggingRQMiddleware)

	r.Post("/", handlers.ShortenedURL)
	r.Get("/{id}", handlers.OriginalURL)

	err := http.ListenAndServe(cfg.RunAddr, r)
	if err != nil {
		log.Println("Server failed:", err)
		panic(err)
	}
}
