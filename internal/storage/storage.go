package storage

import (
	"sync"
)

type Storage interface {
	SetURL(hash, originalURL string)
	GetURL(hash string) (string, bool)
	Exists(hash string) bool
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
func (s *MyStorage) SetURL(hash, originalURL string) {
	s.urlMapMux.Lock()
	defer s.urlMapMux.Unlock()
	s.urlMap[hash] = originalURL
}

// GetURL retrieves a URL from the storage.
func (s *MyStorage) GetURL(hash string) (string, bool) {
	s.urlMapMux.Lock()
	defer s.urlMapMux.Unlock()
	originalURL, ok := s.urlMap[hash]
	return originalURL, ok
}

// Exists checks if a URL exists in the storage.
func (s *MyStorage) Exists(hash string) bool {
	_, found := s.GetURL(hash)
	return found
}
