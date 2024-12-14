package main

import (
	"log"
	"net/http"

	"github.com/Dreeedy/shorturl/internal/config"
	"github.com/Dreeedy/shorturl/internal/handlers"
	"github.com/Dreeedy/shorturl/internal/middlewares"
	"github.com/Dreeedy/shorturl/internal/storage"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func main() {
	cfg := config.NewConfig()
	stg := storage.NewStorage()
	httpHandler := handlers.NewHandler(cfg, stg)
	httpConfig := cfg.GetConfig()

	// Выводим конфигурацию.
	log.Printf("Running server on %s\n", httpConfig.RunAddr)
	log.Printf("Base URL for shortened URLs: %s\n", httpConfig.BaseURL)

	router := chi.NewRouter()
	router.Use(middleware.Logger) // Use the built-in logger middleware from chi.
	router.Use(middlewares.LoggingRQMiddleware)

	router.Post("/", httpHandler.ShortenedURL)
	router.Get("/{id}", httpHandler.OriginalURL)

	err := http.ListenAndServe(httpConfig.RunAddr, router)
	if err != nil {
		log.Fatal("Server failed:", err)
	}
}
