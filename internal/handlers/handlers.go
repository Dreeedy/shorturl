package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Dreeedy/shorturl/internal/config"
	"github.com/Dreeedy/shorturl/internal/storage"
	"github.com/go-chi/chi/v5"
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

	shortenedURL := generateShortenedURL(originalURL)

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(shortenedURL))
}

// Функция для генерации сокращённого URL.
func generateShortenedURL(originalURL string) string {
	hash := generateHash(originalURL)

	// Обработка коллизий.
	for storageInstance.Exists(hash) {
		hash = generateRandomHash()
	}

	storageInstance.SetURL(hash, originalURL)

	cfg := config.GetConfig()
	parts := []string{cfg.BaseURL, "/", hash}
	shortenedURL := strings.Join(parts, "")

	return shortenedURL
}

// Функция для генерации хеша с использованием SHA-256.
func generateHash(s string) string {
	hash := sha256.Sum256([]byte(s))
	return hex.EncodeToString(hash[:])
}

// Функция для генерации случайного хеша в случае нахождения коллизии.
func generateRandomHash() string {
	tUnixNano := time.Now().UnixNano()
	tUnixUint64 := uint64(tUnixNano)

	rand.Seed(tUnixUint64)
	b := make([]byte, 16)
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
