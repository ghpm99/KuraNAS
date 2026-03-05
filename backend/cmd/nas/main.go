//go:build dev
// +build dev

package main

import (
	"context"
	"log"
	"nas-go/api/internal/app"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	log.Println("[MAIN][DEV] Iniciando Kuranas")

	application, err := app.InitializeApp()
	if err != nil {
		log.Printf("[MAIN][DEV] Erro ao iniciar: %s", err.Error())
		panic(err)
	}

	ctx, stopSignal := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignal()

	runErr := make(chan error, 1)
	go func() {
		runErr <- application.Run(":8000", false)
	}()

	select {
	case err := <-runErr:
		if err != nil {
			log.Printf("[MAIN][DEV] Servidor finalizado com erro: %v", err)
		}
	case <-ctx.Done():
		log.Println("[MAIN][DEV] Sinal de interrupcao recebido, iniciando shutdown...")
		if err := application.Stop(); err != nil {
			log.Printf("[MAIN][DEV] Erro ao encerrar aplicacao: %v", err)
		}

		select {
		case err := <-runErr:
			if err != nil {
				log.Printf("[MAIN][DEV] Servidor finalizado com erro apos shutdown: %v", err)
			}
		case <-time.After(6 * time.Second):
			log.Println("[MAIN][DEV] Timeout aguardando encerramento do servidor")
		}
	}
}
