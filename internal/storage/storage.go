package storage

import (
	"sync"
)

// Storage is a structure for storing URLs and a mutex.
type Storage struct {
	urlMap    map[string]string
	urlMapMux *sync.Mutex
}

// NewStorage creates a new instance of Storage.
func NewStorage() *Storage {
	return &Storage{
		urlMap:    make(map[string]string),
		urlMapMux: &sync.Mutex{},
	}
}

// SetURL saves a URL in the storage.
func (s *Storage) SetURL(hash, originalURL string) {
	s.urlMapMux.Lock()
	defer s.urlMapMux.Unlock()
	s.urlMap[hash] = originalURL
}

// GetURL retrieves a URL from the storage.
func (s *Storage) GetURL(hash string) (string, bool) {
	s.urlMapMux.Lock()
	defer s.urlMapMux.Unlock()
	originalURL, ok := s.urlMap[hash]
	return originalURL, ok
}

// Exists checks if a URL exists in the storage.
func (s *Storage) Exists(hash string) bool {
	_, found := s.GetURL(hash)
	return found
}
