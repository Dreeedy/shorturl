package storages

import (
	"fmt"

	"github.com/Dreeedy/shorturl/internal/config"
	"github.com/Dreeedy/shorturl/internal/storages/filestorage"
	"github.com/Dreeedy/shorturl/internal/storages/ramstorage"
	"go.uber.org/zap"
)

type Storage interface {
	SetURL(uuid, shortURL, originalURL string) error
	GetURL(shortURL string) (string, bool)
}

type StorageFactory struct {
	cfg config.Config
	log *zap.Logger
}

func NewStorageFactory(newConfig config.Config, newLogger *zap.Logger) *StorageFactory {
	return &StorageFactory{
		cfg: newConfig,
		log: newLogger,
	}
}

func (ref *StorageFactory) CreateStorage(storageType string) (Storage, error) {
	switch storageType {
	case "ram":
		return ramstorage.NewRAMStorage(), nil
	case "file":
		return filestorage.NewFilestorage(ref.cfg, ref.log), nil
	default:
		return nil, fmt.Errorf("unknown storage type: %s", storageType)
	}
}
