package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/warungos/menu-service/handler"
	"github.com/warungos/menu-service/repository"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}
	dbName := os.Getenv("MONGO_DB_NAME")
	if dbName == "" {
		dbName = "warungos_menu"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(clientOpts)
	if err != nil {
		log.Fatalf("failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("failed to ping MongoDB: %v", err)
	}
	log.Println("connected to MongoDB")

	db := client.Database(dbName)
	repo := repository.NewMongoMenuRepository(db)
	menuHandler := handler.NewMenuHandler(repo)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	r.Route("/menu", func(r chi.Router) {
		r.Mount("/", menuHandler.Routes())
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"menu-service"}`))
	})

	log.Printf("menu-service starting on port %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
