package filestorage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/Dreeedy/shorturl/internal/config"
	"github.com/Dreeedy/shorturl/internal/storages/common"
	"github.com/Dreeedy/shorturl/internal/storages/ramstorage"
	"go.uber.org/zap"
)

const (
	filePermission = 0o600
	errorKey       = "err"
)

type Filestorage struct {
	ramStorage *ramstorage.RAMStorage
	urlMapMux  *sync.Mutex
	cfg        config.Config
	log        *zap.Logger
}

type URLData struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func NewFilestorage(newConfig config.Config, newLogger *zap.Logger) *Filestorage {
	newFilestorage := Filestorage{
		ramStorage: ramstorage.NewRAMStorage(),
		urlMapMux:  &sync.Mutex{},
		cfg:        newConfig,
		log:        newLogger,
	}

	if err := newFilestorage.LoadFromFile(); err != nil {
		newLogger.Error("Failed to load URLs from file: %v", zap.String(errorKey, err.Error()))
	}

	return &newFilestorage
}

// SetURL sets a new URL in the storage.
func (ref *Filestorage) SetURL(data common.URLData) (common.URLData, error) {
	ref.urlMapMux.Lock()
	defer ref.urlMapMux.Unlock()

	if _, err := ref.ramStorage.SetURL(data); err != nil {
		return nil, fmt.Errorf("failed to set URL in memory store: %w", err)
	}

	for _, item := range data {
		if err := ref.AppendToFile(URLData{
			UUID:        item.UUID,
			ShortURL:    item.Hash,
			OriginalURL: item.OriginalURL,
		}); err != nil {
			return nil, fmt.Errorf("failed to append URL to file: %w", err)
		}
	}

	return nil, nil
}

// GetURL retrieves the original URL for a given short URL.
func (ref *Filestorage) GetURL(shortURL string) (string, bool) {
	ref.urlMapMux.Lock()
	defer ref.urlMapMux.Unlock()

	return ref.ramStorage.GetURL(shortURL)
}

// LoadFromFile loads URL data from the file.
func (ref *Filestorage) LoadFromFile() error {
	ref.urlMapMux.Lock()
	defer ref.urlMapMux.Unlock()

	filePath := ref.cfg.GetConfig().FileStoragePath

	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, filePermission)
	if err != nil {
		return fmt.Errorf("os.OpenFile: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			ref.log.Error("Error closing file: %v", zap.String(errorKey, err.Error()))
		}
	}()

	decoder := json.NewDecoder(file)
	for {
		var data URLData
		if err := decoder.Decode(&data); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("json.Decoder.Decode: %w", err)
		}
		var setURLData common.URLData
		item := common.URLItem{
			UUID:        data.UUID,
			Hash:        data.ShortURL,
			OriginalURL: data.OriginalURL,
		}
		setURLData = append(setURLData, item)
		if _, err := ref.ramStorage.SetURL(setURLData); err != nil {
			return fmt.Errorf("failed to set URL in memory store: %w", err)
		}
	}

	return nil
}

// AppendToFile appends URL data to the file.
func (ref *Filestorage) AppendToFile(data URLData) error {
	file, err := os.OpenFile(ref.cfg.GetConfig().FileStoragePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, filePermission)
	if err != nil {
		return fmt.Errorf("os.OpenFile: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			ref.log.Error("Error closing file: %v", zap.String(errorKey, err.Error()))
		}
	}()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("json.Encoder.Encode: %w", err)
	}

	return nil
}
