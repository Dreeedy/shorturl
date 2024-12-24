package filestorage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"github.com/Dreeedy/shorturl/internal/config"
)

const (
	filePermission = 0o600
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
}

type URLData struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func NewFilestorage(newConfig config.Config) *filestorage {
	newFilestorage := filestorage{
		urlMap:    make(map[string]URLData),
		urlMapMux: &sync.Mutex{},
		cfg:       newConfig,
	}

	if err := newFilestorage.LoadFromFile(); err != nil {
		log.Fatalf("Failed to load URLs from file: %v", err)
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
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		log.Printf("File does not exist, creating: %s", filePath)
		file, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("os.Create: %w", err)
		}
		if err := file.Close(); err != nil {
			return fmt.Errorf("file.Close: %w", err)
		}
		return nil // File created, nothing to load.
	} else if err != nil {
		return fmt.Errorf("os.Stat: %w", err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("os.Open: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Error closing file: %v", err)
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
			log.Printf("Error closing file: %v", err)
		}
	}()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("json.Encoder.Encode: %w", err)
	}

	return nil
}
