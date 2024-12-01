package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/Dreeedy/shorturl/internal/config"
	"github.com/go-chi/chi/v5"
)

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
func ShortenedURL(res http.ResponseWriter, req *http.Request) {
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
func OriginalURL(res http.ResponseWriter, req *http.Request) {
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
