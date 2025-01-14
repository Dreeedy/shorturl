package main

import (
	"log"
	"net/http"

	"github.com/Dreeedy/shorturl/internal/config"
	"github.com/Dreeedy/shorturl/internal/handlers"
	"github.com/Dreeedy/shorturl/internal/middlewares/gzip"
	"github.com/Dreeedy/shorturl/internal/middlewares/httplogger"
	"github.com/Dreeedy/shorturl/internal/services/zaplogger"
	"github.com/Dreeedy/shorturl/internal/storages"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/jackc/pgx"
	"go.uber.org/zap"
)

func main() {
	newConfig := config.NewConfig()
	httpConfig := newConfig.GetConfig()
	newZapLogger, zaploggerzErr := zaplogger.NewZapLogger(newConfig)
	if zaploggerzErr != nil {
		log.Fatal("zaplogger init failed:", zaploggerzErr)
	}
	newStorageFactory := storages.NewStorageFactory(newConfig, newZapLogger)
	newStorage, storageType, newStoragezErr := newStorageFactory.CreateStorage()
	if newStoragezErr != nil {
		log.Fatal("newStorage init failed:", newStoragezErr)
	}
	newHandlerHTTP := handlers.NewhandlerHTTP(newConfig, newStorage, newZapLogger)
	newHTTPLogger := httplogger.NewHTTPLogger(newConfig, newZapLogger)
	newGzipMiddleware := gzip.NewGzipMiddleware()

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(newGzipMiddleware.CompressionHandler)
	router.Use(newHTTPLogger.RqRsLogger)

	router.Post("/", newHandlerHTTP.ShortenedURL)
	router.Get("/{id}", newHandlerHTTP.OriginalURL)
	router.Post("/api/shorten", newHandlerHTTP.Shorten)
	router.Post("/api/shorten/batch", newHandlerHTTP.Batch)
	router.Get("/ping", newHandlerHTTP.Ping)

	if storageType == "db" {
		initDBErr := initDB(newConfig, newZapLogger)
		if initDBErr != nil {
			log.Fatal("initDB failed:", initDBErr)
		}
	}

	newZapLogger.Info("Running server on %s\n", zap.String("RunAddr", httpConfig.RunAddr))
	newZapLogger.Info("Base URL for shortened URLs: %s\n", zap.String("BaseURL", httpConfig.BaseURL))

	err := http.ListenAndServe(httpConfig.RunAddr, router)
	if err != nil {
		log.Fatal("Server failed:", err)
	}
}

func initDB(cfg config.Config, log *zap.Logger) error {
	// Parse the connection string
	log.Info("DBConnectionAdress", zap.String("DBConnectionAdress", cfg.GetConfig().DBConnectionAdress))
	connConfig, err := pgx.ParseConnectionString(cfg.GetConfig().DBConnectionAdress)
	if err != nil {
		log.Error("Failed to parse connection string", zap.Error(err))
		return err
	}

	// Establish the connection
	conn, err := pgx.Connect(connConfig)
	if err != nil {
		log.Error("Failed to connect to remote database", zap.Error(err))
		return err
	}
	defer func() {
		if conn != nil {
			if err := conn.Close(); err != nil {
				log.Error("Failed to close connection to remote database", zap.Error(err))
			}
		}
	}()

	log.Info("Connection to remote database successfully established")

	createTableQuery := `
	CREATE TABLE IF NOT EXISTS url_mapping (
	uuid UUID PRIMARY KEY,
	short_url VARCHAR(255) NOT NULL,
	original_url TEXT NOT NULL UNIQUE
	);`

	_, err = conn.Exec(createTableQuery)
	if err != nil {
		log.Error("Failed Exec sql", zap.Error(err))
		return err
	}

	return nil
}
