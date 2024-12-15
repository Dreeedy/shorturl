package main

import (
	"log"
	"net/http"

	"github.com/Dreeedy/shorturl/internal/config"
	"github.com/Dreeedy/shorturl/internal/handlers"
	"github.com/Dreeedy/shorturl/internal/middlewares/httplogger"
	"github.com/Dreeedy/shorturl/internal/services/zaplogger"
	"github.com/Dreeedy/shorturl/internal/storages/ramstorage"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func main() {
	newConfig := config.NewConfig()
	httpConfig := newConfig.GetConfig()
	newStorage := ramstorage.NewStorage()
	newHandlerHTTP := handlers.NewhandlerHTTP(newConfig, newStorage)
	newZapLogger, _ := zaplogger.NewZapLogger(newConfig)
	newHTTPLogger := httplogger.NewHTTPLogger(newConfig, newZapLogger)

	// Выводим конфигурацию.
	log.Printf("Running server on %s\n", httpConfig.RunAddr)
	log.Printf("Base URL for shortened URLs: %s\n", httpConfig.BaseURL)

	router := chi.NewRouter()
	router.Use(middleware.Logger) // Use the built-in logger middleware from chi.
	router.Use(newHTTPLogger.RqRsLogger)

	router.Post("/", newHandlerHTTP.ShortenedURL)
	router.Get("/{id}", newHandlerHTTP.OriginalURL)
	router.Post("/api/shorten", newHandlerHTTP.Shorten)

	err := http.ListenAndServe(httpConfig.RunAddr, router)
	if err != nil {
		log.Fatal("Server failed:", err)
	}
}
