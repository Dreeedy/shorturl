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
	config := config.NewMyConfig()
	storage := storage.NewStorage()
	handler := handlers.NewMyHandler(config, storage)

	cfg := config.GetConfig()

	// Выводим конфигурацию.
	log.Printf("Running server on %s\n", cfg.RunAddr)
	log.Printf("Base URL for shortened URLs: %s\n", cfg.BaseURL)

	router := chi.NewRouter()
	router.Use(middleware.Logger) // Use the built-in logger middleware from chi.
	router.Use(middlewares.LoggingRQMiddleware)

	router.Post("/", handler.ShortenedURL)
	router.Get("/{id}", handler.OriginalURL)

	err := http.ListenAndServe(cfg.RunAddr, router)
	if err != nil {
		log.Fatal("Server failed:", err)
	}
}
