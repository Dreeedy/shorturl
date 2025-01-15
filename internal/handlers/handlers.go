package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Dreeedy/shorturl/internal/apperrors"
	"github.com/Dreeedy/shorturl/internal/config"
	"github.com/Dreeedy/shorturl/internal/storages"
	"github.com/Dreeedy/shorturl/internal/storages/common"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/jackc/pgx"
	"go.uber.org/zap"
)

const (
	errorKey                   = "err"
	unableToReadRqBody         = "Unable to read request body"
	contentType                = "Content-Type"
	contentTypeApplicationJSON = "application/json"
	invalidReqMethod           = "Invalid request method"
	urlIsEmpty                 = "URL is empty"
	unableToWriteResp          = "Unable to write response"
	unableToMarshalResp        = "Unable to marshal response"
	unableToCloseRqBody        = "Unable to close request body"
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
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchAPIRs []ShortURLItem

type ShortURLItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

func (ref *HandlerHTTP) ShortenedURL(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, invalidReqMethod, http.StatusBadRequest)
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
			ref.log.Error(unableToCloseRqBody, zap.String(errorKey, err.Error()))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}()

	originalURL := strings.TrimSpace(string(body))
	if originalURL == "" {
		http.Error(w, urlIsEmpty, http.StatusBadRequest)
		return
	}

	// Convert
	batchAPIRq := BatchAPIRq{
		{OriginalURL: originalURL},
	}
	setURLData := ref.generateShortenedURL(batchAPIRq)

	existingRecords, errSetURL := ref.stg.SetURL(setURLData)
	var errInsertConflict *apperrors.InsertConflictError
	if errSetURL != nil {
		if errors.As(errSetURL, &errInsertConflict) {
			fmt.Printf("Error Code: %d, Message: %s\n", errInsertConflict.Code, errInsertConflict.Message)

			w.Header().Set(contentType, contentTypeApplicationJSON)
			w.WriteHeader(http.StatusConflict)
			if _, err := w.Write([]byte(existingRecords[0].ShortURL)); err != nil {
				ref.log.Error(unableToWriteResp, zap.String(errorKey, err.Error()))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		} else {
			ref.log.Error("Internal Server Error", zap.String(errorKey, errSetURL.Error()))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set(contentType, "text/plain")
	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write([]byte(setURLData[0].ShortURL)); err != nil {
		ref.log.Error(unableToWriteResp, zap.String(errorKey, err.Error()))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (ref *HandlerHTTP) Shorten(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, invalidReqMethod, http.StatusBadRequest)
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
			ref.log.Error(unableToCloseRqBody, zap.String(errorKey, err.Error()))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}()

	originalURL := strings.TrimSpace(shortenAPIRq.URL)
	if originalURL == "" {
		http.Error(w, urlIsEmpty, http.StatusBadRequest)
		return
	}

	// Convert
	batchAPIRq := BatchAPIRq{
		{OriginalURL: originalURL},
	}
	setURLData := ref.generateShortenedURL(batchAPIRq)

	existingRecords, errSetURL := ref.stg.SetURL(setURLData)
	var errInsertConflict *apperrors.InsertConflictError
	if errSetURL != nil {
		if errors.As(errSetURL, &errInsertConflict) {
			fmt.Printf("Error Code: %d, Message: %s\n", errInsertConflict.Code, errInsertConflict.Message)

			// If there are existing records, return 409 Conflict and the existing short URLs
			conflictResponse := ShortenAPIRs{
				Result: existingRecords[0].ShortURL,
			}
			resp, err := json.Marshal(conflictResponse)
			if err != nil {
				ref.log.Error(unableToMarshalResp, zap.String(errorKey, err.Error()))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			w.Header().Set(contentType, contentTypeApplicationJSON)
			w.WriteHeader(http.StatusConflict)
			if _, err := w.Write(resp); err != nil {
				ref.log.Error(unableToWriteResp, zap.String(errorKey, err.Error()))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		} else {
			ref.log.Error("Internal Server Error", zap.String(errorKey, errSetURL.Error()))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	shortenAPIRs := ShortenAPIRs{
		Result: setURLData[0].ShortURL,
	}

	resp, err := json.Marshal(shortenAPIRs)
	if err != nil {
		ref.log.Error(unableToMarshalResp, zap.String(errorKey, err.Error()))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set(contentType, contentTypeApplicationJSON)
	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write(resp); err != nil {
		ref.log.Error(unableToWriteResp, zap.String(errorKey, err.Error()))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (ref *HandlerHTTP) generateShortenedURL(data BatchAPIRq) common.URLData {
	var result common.URLData
	cfg := ref.cfg.GetConfig()

	for _, item := range data {
		var hash = ref.generateRandomHash()
		shortenedURL := fmt.Sprintf("%s/%s", cfg.BaseURL, hash)

		resultItem := common.URLItem{
			UUID:          uuid.NewString(),
			Hash:          hash,
			OriginalURL:   item.OriginalURL,
			OperationType: "INSERT",
			CorrelationID: item.CorrelationID,
			ShortURL:      shortenedURL,
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
		http.Error(w, invalidReqMethod, http.StatusBadRequest)
		return
	}

	shortURL := chi.URLParam(req, "id")

	originalURL, found := ref.stg.GetURL(shortURL)

	if !found {
		http.Error(w, "URL not found", http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (ref *HandlerHTTP) Ping(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, invalidReqMethod, http.StatusBadRequest)
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
		http.Error(w, invalidReqMethod, http.StatusBadRequest)
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
			ref.log.Error(unableToCloseRqBody, zap.String(errorKey, err.Error()))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}()

	for _, item := range batchAPIRq {
		originalURL := strings.TrimSpace(item.OriginalURL)
		if originalURL == "" {
			http.Error(w, urlIsEmpty, http.StatusBadRequest)
			return
		}
	}

	initialCapacity := len(batchAPIRq)
	var batchAPIRs = make(BatchAPIRs, 0, initialCapacity)

	// Convert
	setURLData := ref.generateShortenedURL(batchAPIRq)

	existingRecords, errSetURL := ref.stg.SetURL(setURLData)
	var errInsertConflict *apperrors.InsertConflictError
	if errSetURL != nil {
		if errors.As(errSetURL, &errInsertConflict) {
			fmt.Printf("Error Code: %d, Message: %s\n", errInsertConflict.Code, errInsertConflict.Message)

			conflictResponses := make([]ShortURLItem, len(existingRecords))
			for i, record := range existingRecords {
				conflictResponses[i] = ShortURLItem{
					CorrelationID: record.CorrelationID,
					ShortURL:      record.Hash,
				}
			}
			resp, err := json.Marshal(conflictResponses)
			if err != nil {
				ref.log.Error(unableToMarshalResp, zap.String(errorKey, err.Error()))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			w.Header().Set(contentType, contentTypeApplicationJSON)
			w.WriteHeader(http.StatusConflict)
			if _, err := w.Write(resp); err != nil {
				ref.log.Error(unableToWriteResp, zap.String(errorKey, err.Error()))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		} else {
			ref.log.Error("Internal Server Error", zap.String(errorKey, errSetURL.Error()))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	for _, item := range setURLData {
		resultItem := ShortURLItem{
			CorrelationID: item.CorrelationID,
			ShortURL:      item.ShortURL,
		}
		batchAPIRs = append(batchAPIRs, resultItem)
	}

	ref.log.Sugar().Infow("Batch.batchAPIRs", "batchAPIRs", batchAPIRs)
	ref.log.Sugar().Infow("Batch.existingRecords", "existingRecords", existingRecords)

	resp, err := json.Marshal(batchAPIRs)
	if err != nil {
		ref.log.Error(unableToMarshalResp, zap.String(errorKey, err.Error()))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set(contentType, contentTypeApplicationJSON)
	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write(resp); err != nil {
		ref.log.Error(unableToWriteResp, zap.String(errorKey, err.Error()))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
