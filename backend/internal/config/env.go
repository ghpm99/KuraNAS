package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadConfig() error {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment")
	}

	return nil
}

func Get(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return fallback
}
