package config

import (
	"os"
)

type AppConfigStruct struct {
	EntryPoint    string
	Lang          string
	EnableWorkers bool
}

var AppConfig AppConfigStruct

func InitializeConfig() {
	AppConfig = AppConfigStruct{
		EntryPoint:    os.Getenv("ENTRY_POINT"),
		Lang:          os.Getenv("LANGUAGE"),
		EnableWorkers: os.Getenv("ENABLE_WORKERS") == "true",
	}
}
