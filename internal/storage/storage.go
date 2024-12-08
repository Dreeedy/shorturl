package storage

import (
	"sync"
)

// структура для хранения URL и мьютекса
type Storage struct {
	urlMap     map[string]string
	urlMapLock *sync.Mutex
}

// создает новый экземпляр Storage
func NewStorage() *Storage {
	return &Storage{
		urlMap:     make(map[string]string),
		urlMapLock: &sync.Mutex{},
	}
}

// сохраняет URL в хранилище
func (s *Storage) SetURL(hash, originalURL string) {
	s.urlMapLock.Lock()
	defer s.urlMapLock.Unlock()
	s.urlMap[hash] = originalURL
}

// извлекает URL из хранилища
func (s *Storage) GetURL(hash string) (string, bool) {
	s.urlMapLock.Lock()
	defer s.urlMapLock.Unlock()
	originalURL, ok := s.urlMap[hash]
	return originalURL, ok
}

// проверяет есть ли URL в хранилище
func (s *Storage) Exists(hash string) bool {
	_, found := s.GetURL(hash)
	return found
}
