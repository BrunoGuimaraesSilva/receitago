package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	PostgresDSN string
	MongoURI    string
	DataDir     string
	ServerPort  string
	Timeout     time.Duration
}

func Load() *Config {
	// Load from .env if exists
	_ = godotenv.Load()

	return &Config{
		PostgresDSN: getenv("POSTGRES_DSN", "postgres://receitago:receitago@192.168.0.200:5432/receitago?sslmode=disable"),
		MongoURI:    getenv("MONGO_URI", "mongodb://192.168.0.200:27017"),
		DataDir:     getenv("DATA_DIR", "../../resources"),
		ServerPort:  getenv("SERVER_PORT", "8080"),
		Timeout:     getDuration("REQUEST_TIMEOUT", 120*time.Minute),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		d, err := strconv.Atoi(v)
		if err != nil {
			log.Printf("⚠️ invalid duration for %s, using fallback: %s\n", key, fallback)
			return fallback
		}
		return time.Duration(d) * time.Second
	}
	return fallback
}
