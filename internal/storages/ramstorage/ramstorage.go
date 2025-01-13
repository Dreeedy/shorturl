package ramstorage

import (
	"fmt"
	"sync"

	"github.com/Dreeedy/shorturl/internal/storages/common"
)

// RAMStorage is a structure for storing URLs and a mutex.
type RAMStorage struct {
	urlMap    map[string]string
	urlMapMux *sync.Mutex
}

// NewRAMStorage creates a new instance of Storage.
func NewRAMStorage() *RAMStorage {
	return &RAMStorage{
		urlMap:    make(map[string]string),
		urlMapMux: &sync.Mutex{},
	}
}

// SetURL saves a URL in the storage.
func (s *RAMStorage) SetURL(data common.SetURLData) error {
	s.urlMapMux.Lock()
	defer s.urlMapMux.Unlock()

	for _, item := range data {
		if _, exists := s.urlMap[item.ShortURL]; exists {
			return fmt.Errorf("hash already exists for shortURL: %s", item.ShortURL)
		}
		s.urlMap[item.ShortURL] = item.OriginalURL
	}

	return nil
}

// GetURL retrieves a URL from the storage.
func (s *RAMStorage) GetURL(shortURL string) (string, bool) {
	s.urlMapMux.Lock()
	defer s.urlMapMux.Unlock()

	originalURL, ok := s.urlMap[shortURL]
	return originalURL, ok
}
