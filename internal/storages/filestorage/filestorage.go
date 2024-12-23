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
	SaveToFile() error
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

	log.Printf("SetURL 1")

	if _, exists := ref.urlMap[shortURL]; exists {
		return errors.New("short URL already exists")
	}

	log.Printf("SetURL 2")

	ref.urlMap[shortURL] = URLData{
		UUID:        uuid,
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}

	log.Printf("SetURL 3")

	return ref.SaveToFile()
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

	log.Printf("LoadFromFile 1")

	filePath := ref.cfg.GetConfig().FileStoragePath
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		log.Printf("File does not exist, creating: %s", filePath)
		file, err := os.Create(filePath)
		if err != nil {
			log.Printf("err 1: %s", err)
			return err
		}
		file.Close()
		return nil // File created, nothing to load.
	} else if err != nil {
		log.Printf("err 2: %s", err)
		return err
	}

	log.Printf("LoadFromFile 2")

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("LoadFromFile err: %s", err)
		return err
	}
	defer file.Close()

	log.Printf("LoadFromFile 3")

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

	log.Printf("LoadFromFile 4")

	return nil
}

func (ref *filestorage) SaveToFile() error {
	log.Printf("SaveToFile 1")

	file, err := os.Create(ref.cfg.GetConfig().FileStoragePath)
	if err != nil {
		log.Printf("SaveToFile err: %s", err)
		return err
	}
	defer file.Close()

	log.Printf("SaveToFile 2")

	encoder := json.NewEncoder(file)
	for _, data := range ref.urlMap {
		if err := encoder.Encode(data); err != nil {
			log.Printf("SaveToFile 2 err: %s", err)
			return err
		}
	}

	log.Printf("SaveToFile 3")

	return nil
}
