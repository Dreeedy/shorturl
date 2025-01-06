package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Dreeedy/shorturl/internal/config"
	"github.com/Dreeedy/shorturl/internal/storages/filestorage"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	errorKey           = "err"
	unableToReadRqBody = "Unable to read request body"
)

type Handler interface {
	ShortenedURL(w http.ResponseWriter, req *http.Request)
	OriginalURL(w http.ResponseWriter, req *http.Request)
	Shorten(w http.ResponseWriter, req *http.Request)
	generateShortenedURL(originalURL string) (string, error)
	generateRandomHash() string
}

type handlerHTTP struct {
	cfg config.Config
	stg filestorage.Storage
	log *zap.Logger
}

func NewhandlerHTTP(newConfig config.Config, newStorage filestorage.Storage, newLogger *zap.Logger) *handlerHTTP {
	return &handlerHTTP{
		cfg: newConfig,
		stg: newStorage,
		log: newLogger,
	}
}

type ShortenAPIRq struct {
	URL string `json:"url"`
}

type ShortenAPIRs struct {
	Result string `json:"result"`
}

func (ref *handlerHTTP) ShortenedURL(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		ref.log.Error(unableToReadRqBody, zap.String(errorKey, err.Error()))
		http.Error(w, unableToReadRqBody, http.StatusBadRequest)
		return
	}
	defer func() {
		if err := req.Body.Close(); err != nil {
			ref.log.Error("Unable to close request body", zap.String(errorKey, err.Error()))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}()

	originalURL := strings.TrimSpace(string(body))
	if originalURL == "" {
		http.Error(w, "URL is empty", http.StatusBadRequest)
		return
	}

	shortenedURL, err := ref.generateShortenedURL(originalURL)
	if err != nil {
		ref.log.Error("Internal Server Error", zap.String(errorKey, err.Error()))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write([]byte(shortenedURL)); err != nil {
		ref.log.Error("Unable to write response", zap.String(errorKey, err.Error()))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (ref *handlerHTTP) Shorten(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
		return
	}

	var shortenAPIRq ShortenAPIRq

	if err := json.NewDecoder(req.Body).Decode(&shortenAPIRq); err != nil {
		ref.log.Error(unableToReadRqBody, zap.String(errorKey, err.Error()))
		http.Error(w, unableToReadRqBody, http.StatusBadRequest)
		return
	}
	defer func() {
		if err := req.Body.Close(); err != nil {
			ref.log.Error("Unable to close request body", zap.String(errorKey, err.Error()))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}()

	originalURL := strings.TrimSpace(shortenAPIRq.URL)
	if originalURL == "" {
		http.Error(w, "URL is empty", http.StatusBadRequest)
		return
	}

	shortenedURL, err := ref.generateShortenedURL(originalURL)
	if err != nil {
		ref.log.Error("Internal Server Error", zap.String(errorKey, err.Error()))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	shortenAPIRs := ShortenAPIRs{
		Result: shortenedURL,
	}

	resp, err := json.Marshal(shortenAPIRs)
	if err != nil {
		ref.log.Error("Unable to marshal response", zap.String(errorKey, err.Error()))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write(resp); err != nil {
		ref.log.Error("Unable to write response", zap.String(errorKey, err.Error()))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (ref *handlerHTTP) generateShortenedURL(originalURL string) (string, error) {
	const maxAttempts int = 10
	var attempts = 0
	var hash string

	for range [maxAttempts]struct{}{} {
		hash = ref.generateRandomHash()

		if err := ref.stg.SetURL(uuid.NewString(), hash, originalURL); err == nil {
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

func (ref *handlerHTTP) generateRandomHash() string {
	const size int = 4

	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		ref.log.Fatal("Unable to generate random hash", zap.String(errorKey, err.Error()))
	}
	return hex.EncodeToString(b)
}

func (ref *handlerHTTP) OriginalURL(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
		return
	}

	id := chi.URLParam(req, "id")

	originalURL, found := ref.stg.GetURL(id)

	if !found {
		http.Error(w, "URL not found", http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
