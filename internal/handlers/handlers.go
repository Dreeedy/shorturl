package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/Dreeedy/shorturl/internal/config"
	"github.com/Dreeedy/shorturl/internal/storages/ramstorage"
	"github.com/go-chi/chi"
)

type Handler interface {
	ShortenedURL(res http.ResponseWriter, req *http.Request)
	OriginalURL(res http.ResponseWriter, req *http.Request)
	generateShortenedURL(originalURL string) (string, error)
	generateRandomHash() string
}

type handlerHTTP struct {
	cfg config.Config
	stg ramstorage.Storage
}

func NewhandlerHTTP(config config.Config, storage ramstorage.Storage) *handlerHTTP {
	return &handlerHTTP{
		cfg: config,
		stg: storage,
	}
}

func (ref *handlerHTTP) ShortenedURL(res http.ResponseWriter, req *http.Request) {
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
			log.Printf("Unable to close request body: %v", err)
			http.Error(res, "Internal Server Error", http.StatusInternalServerError)
		}
	}()

	originalURL := strings.TrimSpace(string(body))
	if originalURL == "" {
		http.Error(res, "URL is empty", http.StatusBadRequest)
		return
	}

	shortenedURL, err := ref.generateShortenedURL(originalURL)
	if err != nil {
		log.Printf("Internal Server Error: %v", err)
		http.Error(res, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	if _, err := res.Write([]byte(shortenedURL)); err != nil {
		log.Printf("Unable to write response: %v", err)
		http.Error(res, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Function to generate an abbreviated URL.
func (ref *handlerHTTP) generateShortenedURL(originalURL string) (string, error) {
	const maxAttempts int = 10
	var attempts = 0
	var hash string

	for range [maxAttempts]struct{}{} {
		hash = ref.generateRandomHash()

		if err := ref.stg.SetURL(hash, originalURL); err == nil {
			break
		}

		attempts++
	}

	if attempts >= maxAttempts {
		return "", fmt.Errorf("failed to generate unique hash after %d attempts", attempts)
	}

	cfg := ref.cfg.GetConfig()

	parts := []string{cfg.BaseURL, "/", hash}
	shortenedURL := strings.Join(parts, "")

	return shortenedURL, nil
}

// Function for generating a random hash of fixed length.
func (ref *handlerHTTP) generateRandomHash() string {
	const size int = 4

	b := make([]byte, size) // 8 hex characters.
	if _, err := rand.Read(b); err != nil {
		log.Fatalf("Unable to generate random hash: %v", err)
	}
	return hex.EncodeToString(b)
}

func (ref *handlerHTTP) OriginalURL(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Invalid request method", http.StatusBadRequest)
		return
	}

	id := chi.URLParam(req, "id")

	originalURL, found := ref.stg.GetURL(id)

	if !found {
		http.Error(res, "URL not found", http.StatusBadRequest)
		return // Exit after handling this error.
	}

	// Set the Location header and send a redirect response.
	res.Header().Set("Location", originalURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}
