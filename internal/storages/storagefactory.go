package storages

import (
	"fmt"
	"strings"

	"github.com/Dreeedy/shorturl/internal/config"
	"github.com/Dreeedy/shorturl/internal/db"
	"github.com/Dreeedy/shorturl/internal/storages/common"
	"github.com/Dreeedy/shorturl/internal/storages/dbstorage"
	"github.com/Dreeedy/shorturl/internal/storages/filestorage"
	"github.com/Dreeedy/shorturl/internal/storages/ramstorage"
	"go.uber.org/zap"
)

type Storage interface {
	SetURL(data common.URLData) (common.URLData, error)
	GetURL(shortURL string) (string, bool)
}

type StorageFactory struct {
	cfg config.Config
	log *zap.Logger
	db  db.DB
}

func NewStorageFactory(newConfig config.Config, newLogger *zap.Logger, newDB db.DB) *StorageFactory {
	return &StorageFactory{
		cfg: newConfig,
		log: newLogger,
		db:  newDB,
	}
}

func (ref *StorageFactory) CreateStorage(storageType string) (Storage, error) {
	switch storageType {
	case "ram":
		return ramstorage.NewRAMStorage(), nil
	case "file":
		return filestorage.NewFilestorage(ref.cfg, ref.log), nil
	case "db":
		newDBStorage := dbstorage.NewDBStorage(ref.cfg, ref.log, ref.db)
		return newDBStorage, nil
	default:
		return nil, fmt.Errorf("unknown storage type: %s", storageType)
	}
}

func GetStorageType(newConfig config.Config, newLogger *zap.Logger) string {
	cfg := newConfig.GetConfig()

	storageType := cfg.StorageType

	if len(strings.TrimSpace(cfg.DBConnectionAdress)) > 0 {
		storageType = "db"
	}

	newLogger.Info("Selected storage type", zap.String("storageType", storageType))

	return storageType
}
