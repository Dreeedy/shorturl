package main

import (
	"log"
	"net/http"

	"github.com/Dreeedy/shorturl/internal/config"
	"github.com/Dreeedy/shorturl/internal/db"
	"github.com/Dreeedy/shorturl/internal/handlers"
	"github.com/Dreeedy/shorturl/internal/middlewares/gzip"
	"github.com/Dreeedy/shorturl/internal/middlewares/httplogger"
	"github.com/Dreeedy/shorturl/internal/services/zaplogger"
	"github.com/Dreeedy/shorturl/internal/storages"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"
)

func main() {
	newConfig := config.NewConfig()
	httpConfig := newConfig.GetConfig()

	newZapLogger, zaploggerzErr := zaplogger.NewZapLogger(newConfig)
	if zaploggerzErr != nil {
		log.Fatal("zaplogger init failed:", zaploggerzErr)
	}

	var newDB *db.DB
	var errnewDB error
	storageType := storages.GetStorageType(newConfig, newZapLogger)
	if storageType == "db" {
		newDB, errnewDB = db.NewDB(newConfig, newZapLogger)
		if errnewDB != nil {
			log.Fatal("newDB init failed:", errnewDB)
		}
	}

	newStorageFactory := storages.NewStorageFactory(newConfig, newZapLogger, newDB)
	newStorage, newStoragezErr := newStorageFactory.CreateStorage(storageType)
	if newStoragezErr != nil {
		log.Fatal("newStorage init failed:", newStoragezErr)
	}

	if storageType == "db" {
		err := newDB.InitDB()
		if err != nil {
			log.Fatal("InitDB failed:", err)
		}
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

	newZapLogger.Info("Running server on %s\n", zap.String("RunAddr", httpConfig.RunAddr))
	newZapLogger.Info("Base URL for shortened URLs: %s\n", zap.String("BaseURL", httpConfig.BaseURL))

	err := http.ListenAndServe(httpConfig.RunAddr, router)
	if err != nil {
		log.Fatal("Server failed:", err)
	}

	defer newDB.GetConnPool().Close()
}
