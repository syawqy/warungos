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
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/warungos/order-service/handler"
	"github.com/warungos/order-service/repository"
	"github.com/warungos/order-service/service"
	"github.com/warungos/shared/config"
	"github.com/warungos/shared/database"
	"github.com/warungos/shared/middleware"
	sharedredis "github.com/warungos/shared/redis"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	// PostgreSQL
	pool, err := database.NewPostgresPool(ctx, cfg.PostgresURL)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer pool.Close()

	// Redis
	rdb, err := sharedredis.NewClient(ctx, cfg.RedisURL)
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	defer rdb.Close()

	// PubSub
	pubsub := sharedredis.NewPubSub(rdb)

	// Layers
	orderRepo := repository.NewOrderRepository(pool)
	orderSvc := service.NewOrderService(orderRepo, pubsub)
	orderH := handler.NewOrderHandler(orderSvc)

	r := chi.NewRouter()

	// Global middleware
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(30 * time.Second))
	r.Use(middleware.NewCORS())

	r.Route("/orders", func(r chi.Router) {
		r.Use(middleware.Auth(cfg.JWTSecret))

		r.Post("/", orderH.CreateOrder)
		r.Get("/", orderH.ListOrders)
		r.Get("/count", orderH.CountByBranchAndDate)
		r.Get("/{id}", orderH.GetOrder)
		r.Patch("/{id}/status", orderH.UpdateOrderStatus)
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","service":"order-service"}`))
	})

	// Graceful shutdown
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		log.Printf("order-service starting on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}
	log.Println("order-service stopped")
}
