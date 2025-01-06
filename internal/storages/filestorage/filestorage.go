package filestorage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/Dreeedy/shorturl/internal/config"
	"go.uber.org/zap"
)

const (
	filePermission = 0o600
	errorKey       = "err"
)

type Storage interface {
	SetURL(uuid, shortURL, originalURL string) error
	GetURL(shortURL string) (string, bool)
	LoadFromFile() error
	AppendToFile(data URLData) error
}

type filestorage struct {
	urlMap    map[string]URLData
	urlMapMux *sync.Mutex
	cfg       config.Config
	log       *zap.Logger
}

type URLData struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func NewFilestorage(newConfig config.Config, newLogger *zap.Logger) *filestorage {
	newFilestorage := filestorage{
		urlMap:    make(map[string]URLData),
		urlMapMux: &sync.Mutex{},
		cfg:       newConfig,
		log:       newLogger,
	}

	if err := newFilestorage.LoadFromFile(); err != nil {
		newLogger.Error("Failed to load URLs from file: %v", zap.String(errorKey, err.Error()))
	}

	return &newFilestorage
}

// SetURL sets a new URL in the storage.
func (ref *filestorage) SetURL(uuid, shortURL, originalURL string) error {
	ref.urlMapMux.Lock()
	defer ref.urlMapMux.Unlock()

	if _, exists := ref.urlMap[shortURL]; exists {
		return errors.New("short URL already exists")
	}

	ref.urlMap[shortURL] = URLData{
		UUID:        uuid,
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}

	return ref.AppendToFile(URLData{
		UUID:        uuid,
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	})
}

// GetURL retrieves the original URL for a given short URL.
func (ref *filestorage) GetURL(shortURL string) (string, bool) {
	ref.urlMapMux.Lock()
	defer ref.urlMapMux.Unlock()

	data, ok := ref.urlMap[shortURL]
	if !ok {
		return "", false
	}
	return data.OriginalURL, true
}

// LoadFromFile loads URL data from the file.
func (ref *filestorage) LoadFromFile() error {
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
		ref.urlMap[data.ShortURL] = data
	}

	return nil
}

// AppendToFile appends URL data to the file.
func (ref *filestorage) AppendToFile(data URLData) error {
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
