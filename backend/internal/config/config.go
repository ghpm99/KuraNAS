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
	}
}
