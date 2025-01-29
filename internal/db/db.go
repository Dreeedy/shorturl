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
	createUsertTableQuery := `
    CREATE TABLE IF NOT EXISTS usert (
        user_id SERIAL PRIMARY KEY,
        token_expiration_date TIMESTAMPTZ NOT NULL
    );`
	createURLMappingTableQuery := `
    CREATE TABLE IF NOT EXISTS url_mapping (
        uuid UUID PRIMARY KEY,
        hash VARCHAR(255) NOT NULL,
        original_url TEXT NOT NULL,
        last_operation_type VARCHAR(255) NOT NULL,
        correlation_id VARCHAR(255) NULL,
        short_url VARCHAR(255) NOT NULL,
        user_id INTEGER REFERENCES usert(user_id),
        UNIQUE (original_url, user_id)
    );`
	insertDefaultUserQuery := `
	INSERT INTO usert (user_id, token_expiration_date)
	SELECT 0, NOW()
	WHERE NOT EXISTS (SELECT 1 FROM usert WHERE user_id = 0
	);`

	_, err := ref.pool.Exec(createUsertTableQuery)
	if err != nil {
		ref.log.Error("Failed to create usert table", zap.Error(err))
		return fmt.Errorf("failed to create user table: %w", err)
	}

	// Создание таблицы "url_mapping"
	_, err = ref.pool.Exec(createURLMappingTableQuery)
	if err != nil {
		ref.log.Error("Failed to create url_mapping table", zap.Error(err))
		return fmt.Errorf("failed to create url_mapping table: %w", err)
	}

	_, err = ref.pool.Exec(insertDefaultUserQuery)
	if err != nil {
		ref.log.Error("Failed to insert default user", zap.Error(err))
		return fmt.Errorf("failed to insert default user: %w", err)
	}

	return nil
}

func (ref *DB) GetConnPool() *pgx.ConnPool {
	return ref.pool
}
