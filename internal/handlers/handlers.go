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

type Handler interface {
	ShortenedURL(res http.ResponseWriter, req *http.Request)
	OriginalURL(res http.ResponseWriter, req *http.Request)
	generateShortenedURL(originalURL string) (string, error)
	generateRandomHash() string
}

type MyHandler struct {
	Config  config.Config
	Storage storage.Storage
}

func NewMyHandler(config config.Config, storage storage.Storage) Handler {
	handler := &MyHandler{
		Config:  config,
		Storage: storage,
	}

	return handler
}

func (ref *MyHandler) ShortenedURL(res http.ResponseWriter, req *http.Request) {
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

	shortenedURL, err := ref.generateShortenedURL(originalURL)
	if err != nil {
		http.Error(res, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(shortenedURL))
}

// Функция для генерации сокращённого URL.
func (ref *MyHandler) generateShortenedURL(originalURL string) (string, error) {
	const maxAttempts = 10
	var hash string

	for attempts := 0; attempts < maxAttempts; attempts++ {
		hash = ref.generateRandomHash()
		if !ref.Storage.Exists(hash) {
			break
		}
	}

	if ref.Storage.Exists(hash) {
		return "", fmt.Errorf("failed to generate unique hash after %d attempts", maxAttempts)
	}

	ref.Storage.SetURL(hash, originalURL)

	config := ref.Config.GetConfig()

	parts := []string{config.BaseURL, "/", hash}
	shortenedURL := strings.Join(parts, "")

	return shortenedURL, nil
}

// Функция для генерации случайного хеша фиксированной длины.
func (ref *MyHandler) generateRandomHash() string {
	tUnixNano := time.Now().UnixNano()
	tUnixUint64 := uint64(tUnixNano)

	rand.Seed(tUnixUint64)
	b := make([]byte, 4) // 8 hex characters.
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (ref *MyHandler) OriginalURL(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Invalid request method", http.StatusBadRequest)
		return
	}

	id := chi.URLParam(req, "id")

	originalURL, found := ref.Storage.GetURL(id)

	if !found {
		http.Error(res, "URL not found", http.StatusBadRequest)
		return // Exit after handling this error.
	}

	// Set the Location header and send a redirect response.
	res.Header().Set("Location", originalURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}
