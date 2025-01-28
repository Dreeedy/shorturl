package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/Dreeedy/shorturl/internal/apperrors"
	"github.com/Dreeedy/shorturl/internal/config"
	"github.com/Dreeedy/shorturl/internal/db"
	"github.com/Dreeedy/shorturl/internal/storages"
	"github.com/Dreeedy/shorturl/internal/storages/common"
	"github.com/Dreeedy/shorturl/internal/storages/dbstorage"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
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
	userID := db.GetUsertIDFromContext(req, ref.log)
	setURLData := ref.generateShortenedURL(batchAPIRq, userID)

	existingRecords, errSetURL := ref.stg.SetURL(setURLData)
	var errInsertConflict *apperrors.InsertConflictError
	if errSetURL != nil {
		if errors.As(errSetURL, &errInsertConflict) {
			ref.log.Error("Error errInsertConflict:", zap.String(errorKey, strconv.Itoa(errInsertConflict.Code)),
				zap.String(errorKey, errInsertConflict.Message))

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

	// Convert
	batchAPIRq := BatchAPIRq{
		{OriginalURL: shortenAPIRq.URL},
	}
	userID := db.GetUsertIDFromContext(req, ref.log)
	setURLData := ref.generateShortenedURL(batchAPIRq, userID)

	existingRecords, errSetURL := ref.stg.SetURL(setURLData)
	var errInsertConflict *apperrors.InsertConflictError
	if errSetURL != nil {
		if errors.As(errSetURL, &errInsertConflict) {
			ref.log.Error("Error errInsertConflict:", zap.String(errorKey, strconv.Itoa(errInsertConflict.Code)),
				zap.String(errorKey, errInsertConflict.Message))

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

func (ref *HandlerHTTP) generateShortenedURL(data BatchAPIRq, userID int) common.URLData {
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
			UsertID:       userID,
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

	userID := db.GetUsertIDFromContext(req, ref.log)
	ref.log.Info("Ping()", zap.String("Read userID", strconv.Itoa(userID)))

	err := dbstorage.Ping(ref.cfg, ref.log)
	if err != nil {
		ref.log.Error("Failed Ping remote database", zap.Error(err))
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

	initialCapacity := len(batchAPIRq)
	var batchAPIRs = make(BatchAPIRs, 0, initialCapacity)

	userID := db.GetUsertIDFromContext(req, ref.log)
	setURLData := ref.generateShortenedURL(batchAPIRq, userID)

	existingRecords, errSetURL := ref.stg.SetURL(setURLData)
	var errInsertConflict *apperrors.InsertConflictError
	if errSetURL != nil {
		if errors.As(errSetURL, &errInsertConflict) {
			ref.log.Error("Error errInsertConflict:", zap.String(errorKey, strconv.Itoa(errInsertConflict.Code)),
				zap.String(errorKey, errInsertConflict.Message))

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
