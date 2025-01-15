package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Dreeedy/shorturl/internal/config"
	"github.com/Dreeedy/shorturl/internal/storages"
	"github.com/Dreeedy/shorturl/internal/storages/common"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/jackc/pgx"
	"go.uber.org/zap"
)

const (
	errorKey           = "err"
	unableToReadRqBody = "Unable to read request body"
)

type HandlerHTTP struct {
	cfg config.Config
	stg storages.Storage
	log *zap.Logger
}

func NewhandlerHTTP(newConfig config.Config, newStorage storages.Storage, newLogger *zap.Logger) *HandlerHTTP {
	return &HandlerHTTP{
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

type BatchAPIRq []OriginalURLItem

type OriginalURLItem struct {
	CorrelationId string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchAPIRs []ShortURLItem

type ShortURLItem struct {
	CorrelationId string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type BbbRs []BbbItem

type BbbItem struct {
	CorrelationId string
	OriginalURL   string
	Hash          string
	ShortURL      string
	UUID          string
}

func ConvertBbbRsToSetURLData(bbb BbbRs) common.SetURLData {
	var setURLData common.SetURLData
	for _, item := range bbb {
		setURLData = append(setURLData, common.SetURLItem{
			UUID:        item.UUID,
			Hash:        item.Hash,
			OriginalURL: item.OriginalURL,
		})
	}
	return setURLData
}

func (ref *HandlerHTTP) ShortenedURL(w http.ResponseWriter, req *http.Request) {
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

	// Convert
	aaa := BatchAPIRq{
		{OriginalURL: originalURL},
	}
	bbb := ref.generateShortenedURL(aaa)
	setURLData := ConvertBbbRsToSetURLData(bbb)

	_, errSetURL := ref.stg.SetURL(setURLData)
	if errSetURL != nil {
		ref.log.Error("Internal Server Error", zap.String(errorKey, errSetURL.Error()))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write([]byte(bbb[0].ShortURL)); err != nil {
		ref.log.Error("Unable to write response", zap.String(errorKey, err.Error()))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (ref *HandlerHTTP) Shorten(w http.ResponseWriter, req *http.Request) {
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

	// Convert
	aaa := BatchAPIRq{
		{OriginalURL: originalURL},
	}
	bbb := ref.generateShortenedURL(aaa)
	setURLData := ConvertBbbRsToSetURLData(bbb)

	_, errSetURL := ref.stg.SetURL(setURLData)
	if errSetURL != nil {
		ref.log.Error("Internal Server Error", zap.String(errorKey, errSetURL.Error()))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	shortenAPIRs := ShortenAPIRs{
		Result: bbb[0].ShortURL,
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

func (ref *HandlerHTTP) generateShortenedURL(data BatchAPIRq) BbbRs {
	var result BbbRs
	cfg := ref.cfg.GetConfig()

	for _, item := range data {
		var hash = ref.generateRandomHash()
		shortenedURL := fmt.Sprintf("%s/%s", cfg.BaseURL, hash)

		resultItem := BbbItem{
			item.CorrelationId,
			item.OriginalURL,
			hash,
			shortenedURL,
			uuid.NewString(),
		}
		result = append(result, resultItem)
	}

	return result
}

func (ref *HandlerHTTP) generateRandomHash() string {
	const size int = 4

	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		ref.log.Fatal("Unable to generate random hash", zap.String(errorKey, err.Error()))
	}
	return hex.EncodeToString(b)
}

func (ref *HandlerHTTP) OriginalURL(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
		return
	}

	shortUrl := chi.URLParam(req, "id")

	originalURL, found := ref.stg.GetURL(shortUrl)

	if !found {
		http.Error(w, "URL not found", http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (ref *HandlerHTTP) Ping(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
		return
	}

	ref.log.Info("DBConnectionAdress", zap.String("DBConnectionAdress", ref.cfg.GetConfig().DBConnectionAdress))

	// Parse the connection string
	connConfig, err := pgx.ParseConnectionString(ref.cfg.GetConfig().DBConnectionAdress)
	if err != nil {
		ref.log.Error("Failed to parse connection string", zap.Error(err))
		http.Error(w, "Failed to parse connection string", http.StatusInternalServerError)
		return
	}

	// Establish the connection
	conn, err := pgx.Connect(connConfig)
	if err != nil {
		ref.log.Error("Failed to connect to remote database", zap.Error(err))
		http.Error(w, "Failed to connect to remote database", http.StatusInternalServerError)
		return
	}
	defer func() {
		if conn != nil {
			if err := conn.Close(); err != nil {
				ref.log.Error("Failed to close connection to remote database", zap.Error(err))
			}
		}
	}()

	ref.log.Info("Connection to remote database successfully established")

	// Ping the database to ensure the connection is alive
	if err := conn.Ping(context.Background()); err != nil {
		ref.log.Error("Failed to ping the database", zap.Error(err))
		http.Error(w, "Failed to ping the database", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (ref *HandlerHTTP) Batch(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
		return
	}

	var batchAPIRq BatchAPIRq

	if err := json.NewDecoder(req.Body).Decode(&batchAPIRq); err != nil {
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

	for _, item := range batchAPIRq {
		originalURL := strings.TrimSpace(item.OriginalURL)
		if originalURL == "" {
			http.Error(w, "URL is empty", http.StatusBadRequest)
			return
		}
	}

	var batchAPIRs BatchAPIRs

	// Convert
	bbb := ref.generateShortenedURL(batchAPIRq)
	setURLData := ConvertBbbRsToSetURLData(bbb)

	existingRecords, errSetURL := ref.stg.SetURL(setURLData)
	if errSetURL != nil {
		ref.log.Error("Internal Server Error", zap.String(errorKey, errSetURL.Error()))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	for _, item := range bbb {
		resultItem := ShortURLItem{
			CorrelationId: item.CorrelationId,
			ShortURL:      item.ShortURL,
		}
		batchAPIRs = append(batchAPIRs, resultItem)
	}

	ref.log.Sugar().Infow("Batch.batchAPIRs", "batchAPIRs", batchAPIRs)
	ref.log.Sugar().Infow("Batch.existingRecords", "existingRecords", existingRecords)

	if len(existingRecords) > 0 {
		// If there are existing records, return 409 Conflict and the existing short URLs
		conflictResponses := make([]ShortURLItem, len(existingRecords))
		for i, record := range existingRecords {
			conflictResponses[i] = ShortURLItem{
				CorrelationId: "record.CorrelationId",
				ShortURL:      record.Hash,
			}
		}
		resp, err := json.Marshal(conflictResponses)
		if err != nil {
			ref.log.Error("Unable to marshal response", zap.String(errorKey, err.Error()))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		if _, err := w.Write(resp); err != nil {
			ref.log.Error("Unable to write response", zap.String(errorKey, err.Error()))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	resp, err := json.Marshal(batchAPIRs)
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
