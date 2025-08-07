package config

import (
	"log"
	"os"
	"strings"
	"time"
)

var validJournalModes = map[string]bool{
	"DELETE":   true,
	"TRUNCATE": true,
	"PERSIST":  true,
	"MEMORY":   true,
	"WAL":      true,
	"OFF":      true,
}

type AppConfigStruct struct {
	EntryPoint      string
	Lang            string
	EnableWorkers   bool
	StartupTime     time.Time
	RecentFilesKeep int
	Env             string
	DbJournalMode   string
	DbBuzyTimeout   int
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
		DbJournalMode:   getJournalMode(os.Getenv("JOURNAL_MODE")),
		DbBuzyTimeout:   600000,
	}
}

func getJournalMode(mode string) string {
	if !validJournalModes[strings.ToUpper(mode)] {
		log.Printf("Modo de journal inválido: %s. Usando WAL como padrão.", mode)
		return "WAL"
	}
	return mode
}
