package app

import (
	"context"
	"nas-go/api/internal/config"
	"nas-go/api/internal/worker"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/i18n"

	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

type Application struct {
	Router  *gin.Engine
	Context *AppContext
	Server  *http.Server
}

func InitializeApp() (*Application, error) {
	if err := config.LoadConfig(); err != nil {
		return nil, err
	}
	config.InitializeConfig()
	i18n.LoadTranslations()
	if err := config.LoadConfig(); err != nil {
		return nil, err
	}

	if config.AppConfig.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	database, err := database.ConfigDatabase()
	if err != nil {
		return nil, err
	}

	appContext := NewContext(database)

	router := gin.Default()

	RegisterRoutes(router, appContext)

	workerFileContext := &worker.WorkerContext{
		Service: appContext.Files.Service,
		Tasks:   *appContext.Tasks,
		Logger:  appContext.Logger,
	}

	worker.StartWorkers(workerFileContext, 16)

	return &Application{
		Router:  router,
		Context: appContext,
	}, nil
}

func (app *Application) Run(addr string, enableGraceFul bool) error {
	server := &http.Server{
		Addr:    ":8000",
		Handler: app.Router.Handler(),
	}

	app.Server = server

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	if enableGraceFul {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		return app.Stop()
	}

	select {}
}

func (app *Application) Stop() error {
	log.Println("Parando servidor...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.Server.Shutdown(ctx); err != nil {
		log.Printf("Erro ao desligar servidor: %v\n", err)
		return err
	}

	log.Println("Servidor encerrado com sucesso.")
	return nil
}
