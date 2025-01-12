package storages

import (
	"fmt"
	"strings"

	"github.com/Dreeedy/shorturl/internal/config"
	"github.com/Dreeedy/shorturl/internal/storages/dbstorage"
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

func (ref *StorageFactory) CreateStorage() (Storage, string, error) {
	cfg := ref.cfg.GetConfig()

	storageType := cfg.StorageType

	if len(strings.TrimSpace(cfg.DBConnectionAdress)) > 0 {
		storageType = "db"
	}

	ref.log.Info("Selected storage type", zap.String("storageType", storageType))

	switch storageType {
	case "ram":
		return ramstorage.NewRAMStorage(), storageType, nil
	case "file":
		return filestorage.NewFilestorage(ref.cfg, ref.log), storageType, nil
	case "db":
		return dbstorage.NewDBStorage(ref.cfg, ref.log), storageType, nil
	default:
		return nil, storageType, fmt.Errorf("unknown storage type: %s", storageType)
	}
}
