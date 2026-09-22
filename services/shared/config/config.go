package config

import (
	"os"
	"sync"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	Port              string
	PostgresURL       string
	MongoURL          string
	RedisURL          string
	JWTSecret         string
	MidtransServerKey string
	MidtransURL       string
	AIServiceURL      string
}

var (
	cfg  *Config
	once sync.Once
)

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	once.Do(func() {
		cfg = &Config{
			Port:              getEnv("PORT", "8080"),
			PostgresURL:       getEnv("POSTGRES_URL", "postgres://warungos:warungos@localhost:5432/warungos?sslmode=disable"),
			MongoURL:          getEnv("MONGO_URL", "mongodb://localhost:27017"),
			RedisURL:          getEnv("REDIS_URL", "redis://localhost:6379"),
			JWTSecret:         getEnv("JWT_SECRET", "dev-secret-change-in-production"),
			MidtransServerKey: getEnv("MIDTRANS_SERVER_KEY", ""),
			MidtransURL:       getEnv("MIDTRANS_URL", "https://api.sandbox.midtrans.com"),
			AIServiceURL:      getEnv("AI_SERVICE_URL", "http://localhost:5000"),
		}
	})
	return cfg
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
