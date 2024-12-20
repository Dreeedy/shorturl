package storage

import (
	"errors"
	"sync"
)

type Storage interface {
	SetURL(hash, originalURL string) error
	GetURL(hash string) (string, bool)
}

// MyStorage is a structure for storing URLs and a mutex.
type MyStorage struct {
	urlMap    map[string]string
	urlMapMux *sync.Mutex
}

// NewStorage creates a new instance of Storage.
func NewStorage() Storage {
	storage := &MyStorage{
		urlMap:    make(map[string]string),
		urlMapMux: &sync.Mutex{},
	}

	return storage
}

// SetURL saves a URL in the storage.
func (s *MyStorage) SetURL(hash, originalURL string) error {
	s.urlMapMux.Lock()
	defer s.urlMapMux.Unlock()

	if _, exists := s.urlMap[hash]; exists {
		return errors.New("hash already exists")
	}

	s.urlMap[hash] = originalURL
	return nil
}

// GetURL retrieves a URL from the storage.
func (s *MyStorage) GetURL(hash string) (string, bool) {
	s.urlMapMux.Lock()
	defer s.urlMapMux.Unlock()
	originalURL, ok := s.urlMap[hash]
	return originalURL, ok
}
