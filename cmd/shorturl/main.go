package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
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
func shortenedUrl(res http.ResponseWriter, req *http.Request) {
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
	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(shortenedURL)) // http://localhost:8080/3152b10a
}

// Функция для генерации сокращённого URL
func generateShortenedURL(originalURL string) string {
	hash := fmt.Sprintf("%x", hash(originalURL))
	shortenedURL := "http://localhost:8080/" + hash

	urlMapLock.Lock()
	defer urlMapLock.Unlock()
	urlMap[hash] = originalURL

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
func originaUrl(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Invalid request method", http.StatusBadRequest)
		return
	}

	path := req.URL.Path
	parts := strings.Split(path, "/")

	var id string

	if len(parts) > 2 {
		id = parts[2]
	} else {
		http.Error(res, "ID not provided", http.StatusBadRequest)
		return
	}

	fmt.Println("id =>", id)

	// Locking for concurrent access to urlMap
	urlMapLock.Lock()
	originalURL, found := urlMap[id]
	urlMapLock.Unlock() // Unlock after reading

	if !found {
		http.Error(res, "URL not found", http.StatusBadRequest)
		return // Exit after handling this error
	}

	// Set the Location header and send a redirect response
	res.Header().Set("Location", originalURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func main() {
	log.Println("Server started at :8080")

	mux := http.NewServeMux()
	mux.HandleFunc("/url", shortenedUrl)
	mux.HandleFunc("/url/{id}", originaUrl)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		fmt.Println("Server failed:", err)
		panic(err)
	}
}
