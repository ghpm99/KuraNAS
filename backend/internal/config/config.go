package config

import "os"

type AppConfigStruct struct {
	EntryPoint string
}

var AppConfig AppConfigStruct

func InitializeConfig() {
	AppConfig.EntryPoint = os.Getenv("ENTRY_POINT")
}
