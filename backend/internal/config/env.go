package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func LoadConfig() error {
	exePath, err := os.Executable()
	if err != nil {
		log.Printf("Erro ao obter caminho do executável: %v", err)
		return err
	}

	exeDir := filepath.Dir(exePath)
	envPath := filepath.Join(exeDir, ".env")

	if err := godotenv.Load(envPath); err != nil {
		log.Println("No .env file found at", envPath, ", using system environment")
	}

	return nil
}

func Get(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return fallback
}
