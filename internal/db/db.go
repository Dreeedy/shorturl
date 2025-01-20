package db

import (
	"fmt"

	"github.com/Dreeedy/shorturl/internal/config"
	"github.com/jackc/pgx"
	"go.uber.org/zap"
)

type DB struct {
	cfg  config.Config
	log  *zap.Logger
	pool *pgx.ConnPool
}

const (
	maxConnections = 10
)

func NewDB(newConfig config.Config, newLogger *zap.Logger) (*DB, error) {
	// Parse the connection string
	DBConnectionAdress := newConfig.GetConfig().DBConnectionAdress
	newLogger.Info("DBConnectionAdress", zap.String("DBConnectionAdress", DBConnectionAdress))
	newConnConfig, err := pgx.ParseConnectionString(DBConnectionAdress)
	if err != nil {
		newLogger.Error("Failed to parse connection string", zap.Error(err))
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	poolConfig := pgx.ConnPoolConfig{
		ConnConfig:     newConnConfig,
		MaxConnections: maxConnections,
	}

	// Create a connection pool
	newConnPool, err := pgx.NewConnPool(poolConfig)
	if err != nil {
		newLogger.Error("Failed to create connection pool", zap.Error(err))
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	var newDB = &DB{
		cfg:  newConfig,
		log:  newLogger,
		pool: newConnPool,
	}

	return newDB, nil
}

func (ref *DB) InitDB() error {
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS url_mapping (
	uuid UUID PRIMARY KEY,
	hash VARCHAR(255) NOT NULL,
	original_url TEXT NOT NULL UNIQUE,
	last_operation_type VARCHAR(255) NOT NULL,
	correlation_id VARCHAR(255) NULL,
	short_url VARCHAR(255) NOT NULL
	);`

	_, err := ref.pool.Exec(createTableQuery)
	if err != nil {
		ref.log.Error("Failed Exec sql", zap.Error(err))
		return fmt.Errorf("failed to execute SQL: %w", err)
	}

	return nil
}

func (ref *DB) GetConnPool() *pgx.ConnPool {
	return ref.pool
}
