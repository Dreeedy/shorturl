package dbstorage

import (
	"github.com/Dreeedy/shorturl/internal/config"
	"github.com/Dreeedy/shorturl/internal/storages/common"
	"github.com/jackc/pgx"
	"go.uber.org/zap"
)

type DBStorage struct {
	cfg config.Config
	log *zap.Logger
}

func NewDBStorage(newConfig config.Config, newLogger *zap.Logger) *DBStorage {
	return &DBStorage{
		log: newLogger,
		cfg: newConfig,
	}
}

// SetURL saves a URL in the storage.
func (ref *DBStorage) SetURL(data common.SetURLData) error {
	// Parse the connection string
	connConfig, err := pgx.ParseConnectionString(ref.cfg.GetConfig().DBConnectionAdress)
	if err != nil {
		ref.log.Error("Failed to parse connection string", zap.Error(err))
		return err
	}

	// Establish the connection
	conn, err := pgx.Connect(connConfig)
	if err != nil {
		ref.log.Error("Failed to connect to remote database", zap.Error(err))
		return err
	}
	defer func() {
		if conn != nil {
			if err := conn.Close(); err != nil {
				ref.log.Error("Failed to close connection to remote database", zap.Error(err))
			}
		}
	}()

	ref.log.Info("Connection to remote database successfully established")

	// Begin a transaction
	tx, err := conn.Begin()
	if err != nil {
		ref.log.Error("Failed to begin transaction", zap.Error(err))
		return err
	}
	defer func() {
		if tx != nil {
			if err := tx.Rollback(); err != nil {
				ref.log.Error("Failed to rollback transaction", zap.Error(err))
			}
		}
	}()

	query := `INSERT INTO url_mapping (uuid, short_url, original_url) VALUES ($1, $2, $3)`
	for _, item := range data {
		_, errExec := tx.Exec(query, item.UUID, item.ShortURL, item.OriginalURL)
		if errExec != nil {
			ref.log.Error("Failed to save URL", zap.Error(errExec))
			return errExec
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		ref.log.Error("Failed to commit transaction", zap.Error(err))
		return err
	}

	return nil
}

// GetURL retrieves a URL from the storage.
func (ref *DBStorage) GetURL(shortURL string) (string, bool) {

	// Parse the connection string
	connConfig, err := pgx.ParseConnectionString(ref.cfg.GetConfig().DBConnectionAdress)
	if err != nil {
		ref.log.Error("Failed to parse connection string", zap.Error(err))
	}
	// Establish the connection
	conn, err := pgx.Connect(connConfig)
	if err != nil {
		ref.log.Error("Failed to connect to remote database", zap.Error(err))
	}
	defer func() {
		if conn != nil {
			if err := conn.Close(); err != nil {
				ref.log.Error("Failed to close connection to remote database", zap.Error(err))
			}
		}
	}()

	ref.log.Info("Connection to remote database successfully established")

	var originalURL string
	query := `SELECT original_url FROM url_mapping WHERE short_url = $1`
	errQueryRow := conn.QueryRow(query, shortURL).Scan(&originalURL)
	if errQueryRow == pgx.ErrNoRows {
		return "", false
	} else if errQueryRow != nil {
		ref.log.Error("Failed to retrieve URL", zap.Error(errQueryRow))
		return "", false
	}
	return originalURL, true
}
