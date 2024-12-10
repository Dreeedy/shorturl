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

type HTTPHandler struct {
	Config  config.Config
	Storage storage.Storage
}

func NewHandler(cfg config.Config, stg storage.Storage) Handler {
	handler := &HTTPHandler{
		Config:  cfg,
		Storage: stg,
	}

	return handler
}

func (ref *HTTPHandler) ShortenedURL(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Invalid request method", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "Unable to read request body", http.StatusBadRequest)
		return
	}
	defer func() {
		if err := req.Body.Close(); err != nil {
			http.Error(res, "Unable to close request body", http.StatusInternalServerError)
		}
	}()

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
	if _, err := res.Write([]byte(shortenedURL)); err != nil {
		http.Error(res, "Unable to write response", http.StatusInternalServerError)
	}
}

// Function to generate an abbreviated URL.
func (ref *HTTPHandler) generateShortenedURL(originalURL string) (string, error) {
	const maxAttempts int = 10
	var hash string

	for range [maxAttempts]struct{}{} {
		hash = ref.generateRandomHash()
		if !ref.Storage.Exists(hash) {
			break
		}
	}

	if ref.Storage.Exists(hash) {
		return "", fmt.Errorf("failed to generate unique hash after %d attempts", maxAttempts)
	}

	ref.Storage.SetURL(hash, originalURL)

	cfg := ref.Config.GetConfig()

	parts := []string{cfg.BaseURL, "/", hash}
	shortenedURL := strings.Join(parts, "")

	return shortenedURL, nil
}

// Function for generating a random hash of fixed length.
func (ref *HTTPHandler) generateRandomHash() string {
	tUnixNano := time.Now().UnixNano()
	tUnixUint64 := uint64(tUnixNano)

	const size int = 4

	rand.Seed(tUnixUint64)
	b := make([]byte, size) // 8 hex characters.
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

func (ref *HTTPHandler) OriginalURL(res http.ResponseWriter, req *http.Request) {
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
