package dbstorage

import (
	"strconv"

	"github.com/Dreeedy/shorturl/internal/config"
	"github.com/Dreeedy/shorturl/internal/storages/common"
	"github.com/jackc/pgx"
	"go.uber.org/zap"
)

type DBStorage struct {
	pool *pgx.ConnPool
	log  *zap.Logger
}

func NewDBStorage(newConfig config.Config, newLogger *zap.Logger) (*DBStorage, error) {
	config, err := pgx.ParseConnectionString(newConfig.GetConfig().DBConnectionAdress)
	if err != nil {
		newLogger.Error("Failed to parse connection string", zap.Error(err))
		return nil, err
	}

	poolConfig := pgx.ConnPoolConfig{
		ConnConfig:     config,
		MaxConnections: 10,
	}

	pool, err := pgx.NewConnPool(poolConfig)
	if err != nil {
		newLogger.Error("Failed to create connection pool", zap.Error(err))
		return nil, err
	}

	return &DBStorage{
		pool: pool,
		log:  newLogger,
	}, nil
}

func (ref *DBStorage) SetURL(data common.SetURLData) (common.SetURLData, error) {
	tx, err := ref.pool.Begin()
	if err != nil {
		ref.log.Error("Failed to begin transaction", zap.Error(err))
		return nil, err
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				ref.log.Error("Failed to rollback transaction", zap.Error(rollbackErr))
			}
		} else {
			if commitErr := tx.Commit(); commitErr != nil {
				ref.log.Error("Failed to commit transaction", zap.Error(commitErr))
				err = commitErr
			}
		}
	}()

	var existingRecords common.SetURLData
	query := `INSERT INTO url_mapping (uuid, short_url, original_url) VALUES `
	var args []interface{}
	var argCount int

	for i, item := range data {
		if i > 0 {
			query += ", "
		}
		query += `($` + strconv.Itoa(argCount+1) + `, $` + strconv.Itoa(argCount+2) + `, $` + strconv.Itoa(argCount+3) + `)`
		args = append(args, item.UUID, item.Hash, item.OriginalURL)
		argCount += 3
	}

	_, errExec := tx.Exec(query, args...)
	if errExec != nil {
		ref.log.Error("Failed to save URL", zap.Error(errExec))
		return nil, errExec
	}

	ref.log.Sugar().Infow("existingRecords", "existingRecords", existingRecords)

	return existingRecords, nil
}

// GetURL retrieves a URL from the storage.
func (ref *DBStorage) GetURL(shortURL string) (string, bool) {
	var originalURL string
	query := `SELECT original_url FROM url_mapping WHERE short_url = $1`
	errQueryRow := ref.pool.QueryRow(query, shortURL).Scan(&originalURL)
	if errQueryRow == pgx.ErrNoRows {
		return "", false
	} else if errQueryRow != nil {
		ref.log.Error("Failed to retrieve URL", zap.Error(errQueryRow))
		return "", false
	}
	return originalURL, true
}
