// Code increment #1 DONE
// Code increment #2 DONE
// Code increment #3 DONE
// Code increment #4 DONE
// Code increment #5 DONE

package main

import (
	"log"
	"net/http"

	"github.com/Dreeedy/shorturl/internal/config"
	"github.com/Dreeedy/shorturl/internal/handlers"
	"github.com/Dreeedy/shorturl/internal/middlewares"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg := config.GetConfig()

	// Выводим конфигурацию
	log.Printf("Running server on %s\n", cfg.RunAddr)
	log.Printf("Base URL for shortened URLs: %s\n", cfg.BaseURL)

	router := chi.NewRouter()
	router.Use(middleware.Logger) // Use the built-in logger middleware from chi
	router.Use(middlewares.LoggingRQMiddleware)

	router.Post("/", handlers.ShortenedURL)
	router.Get("/{id}", handlers.OriginalURL)

	err := http.ListenAndServe(cfg.RunAddr, router)
	if err != nil {
		log.Println("Server failed:", err)
		panic(err)
	}
}
