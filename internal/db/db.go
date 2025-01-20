package db

import (
	"fmt"

	"github.com/Dreeedy/shorturl/internal/config"
	"github.com/jackc/pgx"
	"go.uber.org/zap"
)

func InitDB(cfg config.Config, newLogger *zap.Logger) error {
	// Parse the connection string
	newLogger.Info("DBConnectionAdress", zap.String("DBConnectionAdress", cfg.GetConfig().DBConnectionAdress))
	connConfig, err := pgx.ParseConnectionString(cfg.GetConfig().DBConnectionAdress)
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

	newLogger.Info("Connection to remote database successfully established")

	createTableQuery := `
	CREATE TABLE IF NOT EXISTS url_mapping (
	uuid UUID PRIMARY KEY,
	hash VARCHAR(255) NOT NULL,
	original_url TEXT NOT NULL UNIQUE,
	last_operation_type VARCHAR(255) NOT NULL,
	correlation_id VARCHAR(255) NULL,
	short_url VARCHAR(255) NOT NULL
	);`

	_, err = conn.Exec(createTableQuery)
	if err != nil {
		newLogger.Error("Failed Exec sql", zap.Error(err))
		return fmt.Errorf("failed to execute SQL: %w", err)
	}

	return nil
}
