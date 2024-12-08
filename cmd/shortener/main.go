package main

import (
	"log"
	"net/http"

	"github.com/Dreeedy/shorturl/internal/config"
	"github.com/Dreeedy/shorturl/internal/handlers"
	"github.com/Dreeedy/shorturl/internal/middlewares"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func main() {
	cfg := config.GetConfig()

	// Выводим конфигурацию.
	log.Printf("Running server on %s\n", cfg.RunAddr)
	log.Printf("Base URL for shortened URLs: %s\n", cfg.BaseURL)

	router := chi.NewRouter()
	router.Use(middleware.Logger) // Use the built-in logger middleware from chi.
	router.Use(middlewares.LoggingRQMiddleware)

	router.Post("/", handlers.ShortenedURL)
	router.Get("/{id}", handlers.OriginalURL)

	err := http.ListenAndServe(cfg.RunAddr, router)
	if err != nil {
		log.Fatal("Server failed:", err)
	}
}
