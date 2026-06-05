//go:build windows && !dev
// +build windows,!dev

package main

import (
	"log"
	"nas-go/api/internal/app"
	"nas-go/api/pkg/applog"
	"os"
	"path/filepath"
	"strconv"

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

	application.Context.UpdateService.SetShutdownFn(func() {
		if p.logger != nil {
			p.logger.Info("Update applied, stopping service for restart...")
		}
		application.Stop()
		os.Exit(0)
	})

	applog.Go("http-server", func() {
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
	})

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
		Option: service.KeyValue{
			"OnFailure":              "restart",
			"OnFailureDelayDuration": "5s",
			"OnFailureResetPeriod":   10,
		},
	}

	prg := &program{quit: make(chan struct{})}
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

	// The file logger is installed before the .env is loaded, so logging knobs
	// come from the real OS environment here; the level is re-applied from the
	// loaded config once InitializeApp runs.
	rotating, err := applog.NewRotatingFile(applog.RotateConfig{
		Dir:        logDir,
		Prefix:     "kuranas-",
		MaxSizeMB:  envIntOrDefault("LOG_MAX_SIZE_MB", 50),
		MaxBackups: envIntOrDefault("LOG_MAX_BACKUPS", 10),
		MaxAgeDays: envIntOrDefault("LOG_MAX_AGE_DAYS", 30),
	})
	if err != nil {
		log.Println("Erro ao iniciar arquivo de log:", err)
		return
	}

	applog.Setup(applog.Options{
		Writer:    rotating,
		Level:     applog.ParseLevel(os.Getenv("LOG_LEVEL")),
		AddSource: true,
	})

	// Capture runtime panic/fatal stack traces (written straight to the OS
	// stderr handle) into the forensic file too.
	if err := applog.RedirectStderr(rotating.File()); err != nil {
		applog.Warn("could not redirect stderr to log file", "error", err.Error())
	}

	applog.Info("forensic log started", "dir", logDir)
}

func envIntOrDefault(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
