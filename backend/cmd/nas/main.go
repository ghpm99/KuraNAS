//go:build dev
// +build dev

package main

import (
	"context"
	"io"
	"log"
	"nas-go/api/internal/app"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/applog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

func main() {
	setupDevFileLogger()

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

// setupDevFileLogger points the forensic sink at <projectRoot>/log so that
// log/kuranas-*.log exists and survives `go run` restarts during development —
// unlike os.Executable()-derived paths, which land in the ephemeral build temp
// dir under `go run`. It still echoes to stdout so `make run` shows logs live.
func setupDevFileLogger() {
	opts := applog.Options{
		Writer:    os.Stdout,
		Level:     applog.ParseLevel(os.Getenv("LOG_LEVEL")),
		AddSource: true,
	}

	root := config.FindProjectRoot()
	if root == "" {
		applog.Setup(opts)
		return
	}

	logDir := filepath.Join(root, "log")
	rotating, err := applog.NewRotatingFile(applog.RotateConfig{
		Dir:    logDir,
		Prefix: "kuranas-",
	})
	if err != nil {
		log.Printf("[MAIN][DEV] Falha ao iniciar log forense em arquivo: %v", err)
		applog.Setup(opts)
		return
	}

	opts.Writer = io.MultiWriter(os.Stdout, rotating)
	applog.Setup(opts)
	applog.Info("forensic log started (dev)", "dir", logDir)
}
