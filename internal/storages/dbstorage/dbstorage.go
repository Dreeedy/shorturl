package dbstorage

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/Dreeedy/shorturl/internal/apperrors"
	"github.com/Dreeedy/shorturl/internal/config"
	"github.com/Dreeedy/shorturl/internal/db"
	"github.com/Dreeedy/shorturl/internal/storages/common"
	"github.com/jackc/pgx"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

const (
	maxArgCount  = 7
	argIDOffset1 = 1
	argIDOffset2 = 2
	argIDOffset3 = 3
	argIDOffset4 = 4
	argIDOffset5 = 5
	argIDOffset6 = 6
	argIDOffset7 = 7
)

type DBStorage interface {
	DeleteURLsByUser(hashes []string, userID int) error
	GetURLWithDeletedFlag(shortURL string) (string, bool, bool)
	GetURLsByUserID(userID int) (common.URLData, error)
}

type DBStorageImpl struct {
	db  db.DB
	log *zap.Logger
	cfg config.Config
}

func NewDBStorage(newConfig config.Config, newLogger *zap.Logger, newDB db.DB) *DBStorageImpl {
	return &DBStorageImpl{
		db:  newDB,
		log: newLogger,
		cfg: newConfig,
	}
}

func (ref *DBStorageImpl) SetURL(data common.URLData) (common.URLData, error) {
	tx, err := ref.db.GetConnPool().Begin()
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
        INSERT INTO url_mapping (uuid, hash, original_url, last_operation_type, correlation_id, short_url, user_id)
        VALUES `
	args := make([]interface{}, 0, len(data)*maxArgCount)
	var argCount int

	for i, item := range data {
		if i > 0 {
			query += ", "
		}
		query += `($` + strconv.Itoa(argCount+argIDOffset1) + `, $` + strconv.Itoa(argCount+argIDOffset2) + `, $` +
			strconv.Itoa(argCount+argIDOffset3) + `, $` + strconv.Itoa(argCount+argIDOffset4) + `, $` +
			strconv.Itoa(argCount+argIDOffset5) + `, $` + strconv.Itoa(argCount+argIDOffset6) + `, $` +
			strconv.Itoa(argCount+argIDOffset7) + `)`

		args = append(args, item.UUID, item.Hash, item.OriginalURL, "INSERT", item.CorrelationID, item.ShortURL, item.UsertID)
		argCount += maxArgCount

		ref.log.Info("SetURL()", zap.String("userID to DB", strconv.Itoa(item.UsertID)))
	}

	query += `
        ON CONFLICT (original_url, user_id) DO UPDATE
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
			&record.ShortURL, &record.UsertID, &record.IsDeleted); err != nil {
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
func (ref *DBStorageImpl) GetURL(shortURL string) (string, bool) {
	var originalURL string
	query := `SELECT original_url FROM url_mapping WHERE hash = $1`

	errQueryRow := ref.db.GetConnPool().QueryRow(query, shortURL).Scan(&originalURL)
	if errQueryRow != nil {
		if errors.Is(errQueryRow, pgx.ErrNoRows) {
			return "", false
		}
		ref.log.Error("Failed to retrieve URL", zap.Error(errQueryRow))
		return "", false
	}

	return originalURL, true
}

func Ping(newConfig config.Config, newLogger *zap.Logger) error {
	newLogger.Info("DBConnectionAdress", zap.String("DBConnectionAdress", newConfig.GetConfig().DBConnectionAdress))

	// Parse the connection string
	connConfig, err := pgx.ParseConnectionString(newConfig.GetConfig().DBConnectionAdress)
	if err != nil {
		newLogger.Error("Failed to parse connection string", zap.Error(err))
		return fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Establish the connection
	conn, err := pgx.Connect(connConfig)
	if err != nil {
		newLogger.Error("Failed to connect to remote database", zap.Error(err))
		return fmt.Errorf("failed to connect to remote database: %w", err)
	}
	defer func() {
		if conn != nil {
			if err := conn.Close(); err != nil {
				newLogger.Error("Failed to close connection to remote database", zap.Error(err))
			}
		}
	}()

	newLogger.Info("Connection to remote database successfully established")

	// Ping the database to ensure the connection is alive
	if err := conn.Ping(context.Background()); err != nil {
		newLogger.Error("Failed to ping the database", zap.Error(err))
		return fmt.Errorf("failed to ping the database: %w", err)
	}

	return nil
}

// GetURLsByUserID retrieves all URLs associated with a specific user ID.
func (ref *DBStorageImpl) GetURLsByUserID(userID int) (common.URLData, error) {
	query := `
	SELECT uuid, hash, original_url, last_operation_type, correlation_id, short_url, user_id
	FROM url_mapping
	WHERE user_id = $1
	;`
	rows, err := ref.db.GetConnPool().Query(query, userID)
	if err != nil {
		ref.log.Error("Failed to query URLs by user ID", zap.Error(err))
		return nil, fmt.Errorf("failed to query URLs by user ID: %w", err)
	}
	defer rows.Close()

	var results common.URLData
	for rows.Next() {
		var record common.URLItem
		if err := rows.Scan(&record.UUID, &record.Hash, &record.OriginalURL, &record.OperationType, &record.CorrelationID,
			&record.ShortURL, &record.UsertID); err != nil {
			ref.log.Error("Failed to scan row", zap.Error(err))
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, record)
	}

	if err := rows.Err(); err != nil {
		ref.log.Error("Row iteration error", zap.Error(err))
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	ref.log.Sugar().Infow("GetURLsByUserID results", "results", results)
	return results, nil
}

func (ref *DBStorageImpl) DeleteURLsByUser(hashes []string, userID int) error {
	tx, err := ref.db.GetConnPool().Begin()
	if err != nil {
		ref.log.Error("Failed to begin transaction", zap.Error(err))
		return fmt.Errorf("failed to begin transaction: %w", err)
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
    UPDATE url_mapping
    SET is_deleted = TRUE
    WHERE hash = ANY($1) AND user_id = $2;`
	_, err = tx.Exec(query, pq.Array(hashes), userID)
	if err != nil {
		ref.log.Error("Failed to delete URLs", zap.Error(err))
		return fmt.Errorf("failed to delete URLs: %w", err)
	}
	return nil
}

func (ref *DBStorageImpl) GetURLWithDeletedFlag(shortURL string) (string, bool, bool) {
	var originalURL string
	var isDeleted bool
	query := `SELECT original_url, is_deleted FROM url_mapping WHERE hash = $1`
	errQueryRow := ref.db.GetConnPool().QueryRow(query, shortURL).Scan(&originalURL, &isDeleted)
	if errQueryRow != nil {
		if errors.Is(errQueryRow, pgx.ErrNoRows) {
			return "", false, false
		}
		ref.log.Error("Failed to retrieve URL", zap.Error(errQueryRow))
		return "", false, false
	}
	return originalURL, true, isDeleted
}
