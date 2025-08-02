//go:build dev
// +build dev

package main

import (
	"log"
	"nas-go/api/internal/app"
	"os"
	"os/signal"
)

type program struct {
	app  *app.Application
	quit chan struct{}
}

func main() {
	log.Println("[MAIN][DEV] Iniciando Kuranas")
	prg := &program{}

	prg.quit = make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		// We received an interrupt signal, shut down.
		prg.app.Stop()
		close(prg.quit)
	}()

	application, err := app.InitializeApp()
	if err != nil {
		log.Printf("[MAIN][DEV] Erro ao iniciar: %s", err.Error())
		panic(err)
	}

	application.Run(":8000", false)
	<-prg.quit
}
