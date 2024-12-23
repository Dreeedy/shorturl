package filestorage

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"sync"

	"github.com/Dreeedy/shorturl/internal/config"
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

func (ref *filestorage) GetURL(shortURL string) (string, bool) {
	ref.urlMapMux.Lock()
	defer ref.urlMapMux.Unlock()

	data, ok := ref.urlMap[shortURL]
	if !ok {
		return "", false
	}
	return data.OriginalURL, true
}

func (ref *filestorage) LoadFromFile() error {
	ref.urlMapMux.Lock()
	defer ref.urlMapMux.Unlock()

	filePath := ref.cfg.GetConfig().FileStoragePath
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		log.Printf("File does not exist, creating: %s", filePath)
		file, err := os.Create(filePath)
		if err != nil {
			log.Printf("err: %s", err)
			return err
		}
		file.Close()
		return nil // File created, nothing to load.
	} else if err != nil {
		log.Printf("err: %s", err)
		return err
	}

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("LoadFromFile err: %s", err)
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	for {
		var data URLData
		if err := decoder.Decode(&data); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		ref.urlMap[data.ShortURL] = data
	}

	return nil
}

func (ref *filestorage) AppendToFile(data URLData) error {
	file, err := os.OpenFile(ref.cfg.GetConfig().FileStoragePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Printf("err: %s", err)
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(data); err != nil {
		log.Printf("err: %s", err)
		return err
	}

	return nil
}
