package ramstorage

import (
	"errors"
	"sync"
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
func (s *RAMStorage) SetURL(uuid, shortURL, originalURL string) error {
	s.urlMapMux.Lock()
	defer s.urlMapMux.Unlock()

	if _, exists := s.urlMap[shortURL]; exists {
		return errors.New("hash already exists")
	}

	s.urlMap[shortURL] = originalURL
	return nil
}

// GetURL retrieves a URL from the storage.
func (s *RAMStorage) GetURL(shortURL string) (string, bool) {
	s.urlMapMux.Lock()
	defer s.urlMapMux.Unlock()
	originalURL, ok := s.urlMap[shortURL]
	return originalURL, ok
}
