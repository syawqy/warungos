package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"

	sharedConfig "github.com/warungos/shared/config"
	sharedMiddleware "github.com/warungos/shared/middleware"
	sharedRedis "github.com/warungos/shared/redis"
)

func main() {
	cfg := sharedConfig.Load()
	ctx := context.Background()

	// Redis for rate limiting
	rdb, err := sharedRedis.NewClient(ctx, cfg.RedisURL)
	if err != nil {
		log.Printf("Warning: Redis unavailable, rate limiting disabled: %v", err)
	}

	// Service URLs from environment
	services := map[string]string{
		"auth":      getEnv("AUTH_SERVICE_URL", "http://localhost:8081"),
		"menu":      getEnv("MENU_SERVICE_URL", "http://localhost:8082"),
		"order":     getEnv("ORDER_SERVICE_URL", "http://localhost:8083"),
		"payment":   getEnv("PAYMENT_SERVICE_URL", "http://localhost:8084"),
		"inventory": getEnv("INVENTORY_SERVICE_URL", "http://localhost:8085"),
	}

	// Auth middleware
	authMiddleware := sharedMiddleware.Auth(cfg.JWTSecret)

	// Rate limiter
	var rateLimiter *sharedRedis.RateLimiter
	if rdb != nil {
		rateLimiter = sharedRedis.NewRateLimiter(rdb)
	}

	r := chi.NewRouter()

	// Global middleware
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.Timeout(30 * time.Second))
	r.Use(sharedMiddleware.NewCORS())

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","service":"api-gateway"}`))
	})

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Public routes (no auth required)
		r.Post("/auth/login", proxy(services["auth"]))
		r.Post("/auth/register", proxy(services["auth"]))
		r.Post("/auth/refresh", proxy(services["auth"]))
		r.Post("/payments/webhook", proxy(services["payment"]))

		// Protected routes (auth required)
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware)

			if rateLimiter != nil {
				r.Use(rateLimiter.Middleware)
			}

			r.Get("/auth/me", proxy(services["auth"]))

			r.Route("/menu", func(r chi.Router) {
				r.Get("/", proxy(services["menu"]))
				r.Get("/stats", proxy(services["menu"]))
				r.Get("/{id}", proxy(services["menu"]))
				r.Post("/", proxy(services["menu"]))
				r.Put("/{id}", proxy(services["menu"]))
				r.Delete("/{id}", proxy(services["menu"]))
			})

			r.Route("/orders", func(r chi.Router) {
				r.Get("/", proxy(services["order"]))
				r.Post("/", proxy(services["order"]))
				r.Get("/{id}", proxy(services["order"]))
				r.Patch("/{id}/status", proxy(services["order"]))
				r.Post("/{id}/cancel", proxy(services["order"]))
			})

			r.Route("/payments", func(r chi.Router) {
				r.Post("/create", proxy(services["payment"]))
				r.Get("/{order_id}", proxy(services["payment"]))
			})

			r.Route("/inventory", func(r chi.Router) {
				r.Get("/", proxy(services["inventory"]))
				r.Post("/", proxy(services["inventory"]))
				r.Get("/alerts", proxy(services["inventory"]))
				r.Post("/reserve", proxy(services["inventory"]))
				r.Get("/{id}", proxy(services["inventory"]))
				r.Patch("/{id}", proxy(services["inventory"]))
			})
		})
	})

	addr := ":" + cfg.Port
	log.Printf("API Gateway starting on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}

func proxy(serviceURL string) http.HandlerFunc {
	target, _ := url.Parse(serviceURL)
	p := httputil.NewSingleHostReverseProxy(target)

	p.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Proxy error to %s: %v", serviceURL, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprintf(w, `{"error":"service unavailable"}`)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// Chi Route groups don't strip URL.Path; reverse proxy needs clean path
		if len(r.URL.Path) > 7 && r.URL.Path[:7] == "/api/v1" {
			r.URL.Path = r.URL.Path[7:]
			if r.URL.Path == "" {
				r.URL.Path = "/"
			}
		}
		log.Printf("PROXY: %s %s -> %s", r.Method, r.URL.Path, serviceURL)
		r.Header.Set("X-Forwarded-Host", r.Host)
		p.ServeHTTP(w, r)
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
