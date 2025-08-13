package config

import (
	"os"
	"time"
)

type AppConfigStruct struct {
	EntryPoint      string
	Lang            string
	EnableWorkers   bool
	StartupTime     time.Time
	RecentFilesKeep int
	Env             string
	DbHost          string
	DbPort          string
	DbUser          string
	DbPassword      string
	DbName          string
}

var AppConfig AppConfigStruct

func InitializeConfig() {
	AppConfig = AppConfigStruct{
		EntryPoint:      os.Getenv("ENTRY_POINT"),
		Lang:            os.Getenv("LANGUAGE"),
		EnableWorkers:   os.Getenv("ENABLE_WORKERS") == "true",
		StartupTime:     time.Now(),
		RecentFilesKeep: 10,
		Env:             os.Getenv("ENV"),
		DbHost:          os.Getenv("DB_HOST"),
		DbPort:          os.Getenv("DB_PORT"),
		DbUser:          os.Getenv("DB_USER"),
		DbPassword:      os.Getenv("DB_PASSWORD"),
		DbName:          os.Getenv("DB_NAME"),
	}
}
