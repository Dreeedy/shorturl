package handlers

import (
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Dreeedy/shorturl/internal/config"
	"github.com/Dreeedy/shorturl/internal/storage"
	"github.com/go-chi/chi"
	"golang.org/x/exp/rand"
)

var storageInstance = storage.NewStorage()

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

	shortenedURL, err := generateShortenedURL(originalURL)
	if err != nil {
		http.Error(res, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(shortenedURL))
}

// Функция для генерации сокращённого URL.
func generateShortenedURL(originalURL string) (string, error) {
	const maxAttempts = 10
	var hash string

	for attempts := 0; attempts < maxAttempts; attempts++ {
		hash = generateRandomHash()
		if !storageInstance.Exists(hash) {
			break
		}
	}

	if storageInstance.Exists(hash) {
		return "", fmt.Errorf("failed to generate unique hash after %d attempts", maxAttempts)
	}

	storageInstance.SetURL(hash, originalURL)

	cfg := config.GetConfig()
	parts := []string{cfg.BaseURL, "/", hash}
	shortenedURL := strings.Join(parts, "")

	return shortenedURL, nil
}

// Функция для генерации случайного хеша фиксированной длины.
func generateRandomHash() string {
	tUnixNano := time.Now().UnixNano()
	tUnixUint64 := uint64(tUnixNano)

	rand.Seed(tUnixUint64)
	b := make([]byte, 4) // 8 hex characters.
	rand.Read(b)
	return hex.EncodeToString(b)
}

func OriginalURL(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Invalid request method", http.StatusBadRequest)
		return
	}

	id := chi.URLParam(req, "id")

	originalURL, found := storageInstance.GetURL(id)

	if !found {
		http.Error(res, "URL not found", http.StatusBadRequest)
		return // Exit after handling this error.
	}

	// Set the Location header and send a redirect response.
	res.Header().Set("Location", originalURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}
