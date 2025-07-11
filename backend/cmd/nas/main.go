//go:build dev
// +build dev

package main

import (
	"log"
	"nas-go/api/internal/app"
)

func main() {
	log.Println("[MAIN][DEV] Iniciando Kuranas")
	application, err := app.InitializeApp()
	if err != nil {
		log.Printf("[MAIN][DEV] Erro ao iniciar: %s", err.Error())
		panic(err)
	}

	application.Run(":8000", false)
}
