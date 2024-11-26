// Code increment #1 DONE

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
func originaURL(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Invalid request method", http.StatusBadRequest)
		return
	}

	path := req.URL.Path
	parts := strings.Split(path, "/")

	fmt.Println("path =>", path)
	fmt.Println("parts =>", parts)
	fmt.Println("parts 0 =>", parts[0])
	fmt.Println("parts 1 =>", parts[1])

	var id string

	if len(parts) >= 1 {
		id = parts[1]
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
	mux.HandleFunc("/", shortenedURL)
	mux.HandleFunc("/{id}", originaURL)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		fmt.Println("Server failed:", err)
		panic(err)
	}
}
