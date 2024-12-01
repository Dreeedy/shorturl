// Code increment #1 DONE
// Code increment #2 DONE
// Code increment #3 DONE
// Code increment #4 DONE
// Code increment #5 DONE

package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Dreeedy/shorturl/internal/config"
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

var (
	urlMap     = make(map[string]string)
	urlMapLock sync.Mutex
)

// @Description Endpoint to shorten a given URL
// @ID /url
// @Accept text/plain
// @Produce text/plain
// @Param url body string true "URL to be shortened"
// @Success 201 {string} string "Shortened URL"
// @Router / [post]
func shortenedURL(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Invalid request method", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "Unable to read request body", http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	originalURL := strings.TrimSpace(string(body))
	if originalURL == "" {
		http.Error(res, "URL is empty", http.StatusBadRequest)
		return
	}

	shortenedURL := generateShortenedURL(originalURL)

	log.Printf("shortenedURL.originalURL: %s", originalURL)
	log.Printf("shortenedURL.shortenedURL: %s", shortenedURL)

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(shortenedURL))
}

// Функция для генерации сокращённого URL
func generateShortenedURL(originalURL string) string {
	hash := fmt.Sprintf("%x", hash(originalURL))

	urlMapLock.Lock()
	defer urlMapLock.Unlock()
	urlMap[hash] = originalURL

	cfg := config.GetConfig()
	parts := []string{cfg.BaseURL, "/", hash}
	shortenedURL := strings.Join(parts, "")

	return shortenedURL
}

// Пример хеш-функции
func hash(s string) uint32 {
	var hash uint32
	for _, c := range s {
		hash = hash*31 + uint32(c)
	}
	return hash
}

// @Description xxx
// @ID /url/{id}
// @Accept text/plain
// @Produce text/plain
// @Router / [get]
func originalURL(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Invalid request method", http.StatusBadRequest)
		return
	}

	id := chi.URLParam(req, "id")

	log.Println("id =>", id)

	// Locking for concurrent access to urlMap
	urlMapLock.Lock()
	originalURL, found := urlMap[id]
	urlMapLock.Unlock() // Unlock after reading

	log.Printf("urlMap: %s", urlMap)
	log.Printf("urlMap[id]: %s, %s", id, urlMap[id])

	if !found {
		http.Error(res, "URL not found", http.StatusBadRequest)
		return // Exit after handling this error
	}

	log.Println("URL found")

	// Set the Location header and send a redirect response
	res.Header().Set("Location", originalURL)

	log.Println("Header Set Location")

	res.WriteHeader(http.StatusTemporaryRedirect)

	log.Println("WriteHeader")
}

func main() {
	cfg := config.GetConfig()

	// Выводим конфигурацию
	log.Printf("Running server on %s\n", cfg.RunAddr)
	log.Printf("Base URL for shortened URLs: %s\n", cfg.BaseURL)

	r := chi.NewRouter()
	r.Use(middleware.Logger) // Use the built-in logger middleware from chi
	r.Use(LoggingRQMiddleware)

	r.Post("/", shortenedURL)
	r.Get("/{id}", originalURL)

	err := http.ListenAndServe(cfg.RunAddr, r)
	if err != nil {
		log.Println("Server failed:", err)
		panic(err)
	}
}
