package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/warungos/inventory-service/handler"
	"github.com/warungos/inventory-service/repository"
	"github.com/warungos/shared/config"
	"github.com/warungos/shared/database"
	sharedMiddleware "github.com/warungos/shared/middleware"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	// Connect to PostgreSQL
	pgPool, err := database.NewPostgresPool(ctx, cfg.PostgresURL)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer pgPool.Close()

	// Initialize components
	repo := repository.NewInventoryRepo(pgPool)
	h := handler.NewInventoryHandler(repo)

	// Setup router
	r := chi.NewRouter()
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(sharedMiddleware.NewCORS())

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","service":"inventory-service"}`))
	})

	// API routes
	r.Route("/inventory", func(r chi.Router) {
		r.Get("/", h.List)
		r.Post("/", h.Create)
		r.Get("/alerts", h.GetAlerts)
		r.Post("/reserve", h.Reserve)
		r.Get("/{id}", h.GetByID)
		r.Patch("/{id}", h.Update)
	})

	// Start server
	addr := ":" + os.Getenv("PORT")
	if addr == ":" {
		addr = ":8085"
	}

	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down inventory service...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		srv.Shutdown(shutdownCtx)
	}()

	log.Printf("Inventory service starting on %s", addr)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}
