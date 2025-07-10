//go:build windows && !dev
// +build windows,!dev

package main

import (
	"log"
	"nas-go/api/internal/app"
	"os"

	"github.com/kardianos/service"
)

type program struct {
	app *app.Application
}

func (p *program) Start(s service.Service) error {
	go func() {
		application, err := app.InitializeApp()
		if err != nil {
			log.Fatalf("Erro ao inicializar app: %v", err)
		}
		p.app = application

		if err := application.Run(":8000", false); err != nil {
			log.Printf("Erro ao rodar servidor: %v", err)
		}
	}()
	return nil
}

func (p *program) Stop(s service.Service) error {
	if p.app != nil {
		return p.app.Stop()
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

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "install":
			err := s.Install()
			if err != nil {
				log.Printf("Erro ao instalar serviço: %v", err)
			} else {
				log.Printf("Serviço instalado com sucesso.")
			}
			return
		case "uninstall":
			s.Uninstall()
			return
		case "start":
			s.Start()
			return
		case "stop":
			s.Stop()
			return
		}
	}

	if err := s.Run(); err != nil {
		log.Fatal(err)
	}
}
