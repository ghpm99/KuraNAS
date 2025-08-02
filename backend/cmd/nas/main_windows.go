//go:build windows && !dev
// +build windows,!dev

package main

import (
	"log"
	"nas-go/api/internal/app"
	"os"
	"path/filepath"
	"time"

	"github.com/kardianos/service"
)

type program struct {
	app    *app.Application
	logger service.Logger
	quit   chan struct{}
}

func (p *program) Start(s service.Service) error {
	if p.logger != nil {
		p.logger.Info("Iniciando serviço KuraNAS...")
	}

	setupWorkingDirectory()
	setupFileLogger()

	application, err := app.InitializeApp()
	if err != nil {
		if p.logger != nil {
			p.logger.Errorf("Erro ao inicializar app: %v", err)
		} else {
			log.Printf("Erro ao inicializar app: %v", err)
		}
		return err
	}
	if p.logger != nil {
		p.logger.Info("Serviço KuraNAS iniciado com sucesso.")
	}

	p.app = application

	go func() {

		if err := application.Run(":8000", true); err != nil {
			if p.logger != nil {
				p.logger.Errorf("Erro ao executar servidor: %v", err)
			} else {
				log.Printf("Erro ao executar servidor: %v", err)
			}
		}
		if p.logger != nil {
			p.logger.Info("Serviço KuraNAS finalizado.")
		}
	}()

	if p.logger != nil {
		p.logger.Info("KuraNAS iniciado na porta 8000")
	}

	return nil
}

func (p *program) Stop(s service.Service) error {
	if p.logger != nil {
		p.logger.Info("Parando serviço KuraNAS")
	}

	if p.app != nil && p.app.Context != nil {
		p.app.Stop()
		close(p.quit)
	}

	if p.logger != nil {
		p.logger.Info("Serviço KuraNAS parado.")
	}

	return nil
}

func main() {
	svcConfig := &service.Config{
		Name:        "KuraNAS",
		DisplayName: "KuraNAS Service",
		Description: "Serviço do sistema KuraNAS.",
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}

	logger, err := s.Logger(nil)
	if err == nil {
		prg.logger = logger
	}

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "install":
			err := s.Install()
			if err != nil {
				logger.Errorf("Erro ao instalar serviço: %v", err)
			} else {
				logger.Info("Serviço instalado com sucesso.")
			}
			return
		case "uninstall":
			err := s.Uninstall()
			if err != nil {
				logger.Errorf("Erro ao remover: %v", err)
			} else {
				logger.Info("Serviço removido com sucesso.")
			}
			return
		case "start":
			err := s.Start()
			if err != nil {
				logger.Errorf("Erro ao iniciar: %v", err)
			} else {
				logger.Info("Serviço iniciado.")
			}
			return
		case "stop":
			err := s.Stop()
			if err != nil {
				logger.Errorf("Erro ao parar: %v", err)
			} else {
				logger.Info("Serviço parado.")
			}
			return
		}
	}

	if err := s.Run(); err != nil {
		log.Fatal(err)
	}
}

func setupWorkingDirectory() {
	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		_ = os.Chdir(exeDir)
	}
}

func setupFileLogger() {
	exePath, err := os.Executable()
	if err != nil {
		log.Println("Erro ao obter caminho do executável:", err)
		return
	}

	exeDir := filepath.Dir(exePath)
	logDir := filepath.Join(exeDir, "log")

	err = os.MkdirAll(logDir, os.ModePerm)
	if err != nil {
		log.Println("Erro ao criar diretório de log:", err)
		return
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logFileName := "kuranas-" + timestamp + ".log"
	logFile := filepath.Join(logDir, logFileName)

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Println("Erro ao abrir arquivo de log:", err)
		return
	}

	log.SetOutput(file)
	log.SetFlags(log.LstdFlags | log.LUTC)

	log.Println("Arquivo de log iniciado:", logFile)
}
