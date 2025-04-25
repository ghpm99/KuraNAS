package config

import "os"

type AppConfigStruct struct {
	EntryPoint string
	Lang       string
}

var AppConfig AppConfigStruct

func InitializeConfig() {
	AppConfig.EntryPoint = os.Getenv("ENTRY_POINT")
	AppConfig.Lang = os.Getenv("lang")
}
