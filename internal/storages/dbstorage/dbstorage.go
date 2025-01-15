package dbstorage

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/Dreeedy/shorturl/internal/apperrors"
	"github.com/Dreeedy/shorturl/internal/config"
	"github.com/Dreeedy/shorturl/internal/storages/common"
	"github.com/jackc/pgx"
	"go.uber.org/zap"
)

const (
	maxConnections = 10
	maxArgCount    = 6
	argIDOffset1   = 1
	argIDOffset2   = 2
	argIDOffset3   = 3
	argIDOffset4   = 4
	argIDOffset5   = 5
	argIDOffset6   = 6
)

type DBStorage struct {
	pool *pgx.ConnPool
	log  *zap.Logger
}

func NewDBStorage(newConfig config.Config, newLogger *zap.Logger) (*DBStorage, error) {
	cfg, err := pgx.ParseConnectionString(newConfig.GetConfig().DBConnectionAdress)
	if err != nil {
		newLogger.Error("Failed to parse connection string", zap.Error(err))
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	poolConfig := pgx.ConnPoolConfig{
		ConnConfig:     cfg,
		MaxConnections: maxConnections,
	}

	pool, err := pgx.NewConnPool(poolConfig)
	if err != nil {
		newLogger.Error("Failed to create connection pool", zap.Error(err))
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	return &DBStorage{
		pool: pool,
		log:  newLogger,
	}, nil
}

func (ref *DBStorage) SetURL(data common.URLData) (common.URLData, error) {
	tx, err := ref.pool.Begin()
	if err != nil {
		ref.log.Error("Failed to begin transaction", zap.Error(err))
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
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

	query := `
        INSERT INTO url_mapping (uuid, hash, original_url, last_operation_type, correlation_id, short_url)
        VALUES `
	args := make([]interface{}, 0, len(data)*maxArgCount)
	var argCount int

	for i, item := range data {
		if i > 0 {
			query += ", "
		}
		query += `($` + strconv.Itoa(argCount+argIDOffset1) + `, $` + strconv.Itoa(argCount+argIDOffset2) + `, $` +
			strconv.Itoa(argCount+argIDOffset3) + `, $` + strconv.Itoa(argCount+argIDOffset4) + `, $` +
			strconv.Itoa(argCount+argIDOffset5) + `, $` + strconv.Itoa(argCount+argIDOffset6) + `)`
		args = append(args, item.UUID, item.Hash, item.OriginalURL, "INSERT", item.CorrelationID, item.ShortURL)
		argCount += maxArgCount
	}

	query += `
        ON CONFLICT (original_url) DO UPDATE
        SET original_url = EXCLUDED.original_url, last_operation_type = 'UPDATE'
        RETURNING *;`

	ref.log.Sugar().Infow("query", "query", query)
	ref.log.Sugar().Infow("args", "args", args)

	rows, errExec := tx.Query(query, args...)
	if errExec != nil {
		ref.log.Error("Failed to save URL", zap.Error(errExec))
		return nil, fmt.Errorf("failed to save URL: %w", errExec)
	}
	defer rows.Close()

	var existingRecords common.URLData
	for rows.Next() {
		var record common.URLItem
		var operationType string
		if err := rows.Scan(&record.UUID, &record.Hash, &record.OriginalURL, &operationType, &record.CorrelationID,
			&record.ShortURL); err != nil {
			ref.log.Error("Failed to scan row", zap.Error(err))
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		record.OperationType = operationType
		if record.OperationType == "UPDATE" {
			existingRecords = append(existingRecords, record)
		}
	}

	if err := rows.Err(); err != nil {
		ref.log.Error("Row iteration error", zap.Error(err))
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	ref.log.Sugar().Infow("existingRecords", "existingRecords", existingRecords)

	if len(existingRecords) > 0 {
		errorCode := 409
		return existingRecords, fmt.Errorf("insert conflict: %w",
			apperrors.NewInsertConflict(errorCode, "Insert conflict"))
	} else {
		return nil, nil
	}
}

// GetURL retrieves a URL from the storage.
func (ref *DBStorage) GetURL(shortURL string) (string, bool) {
	var originalURL string
	query := `SELECT original_url FROM url_mapping WHERE hash = $1`
	errQueryRow := ref.pool.QueryRow(query, shortURL).Scan(&originalURL)
	if errors.Is(errQueryRow, pgx.ErrNoRows) {
		return "", false
	} else if errQueryRow != nil {
		ref.log.Error("Failed to retrieve URL", zap.Error(errQueryRow))
		return "", false
	}
	return originalURL, true
}
