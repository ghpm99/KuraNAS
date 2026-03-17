package app

import (
	"context"
	"fmt"
	"nas-go/api/internal/config"
	"nas-go/api/internal/worker"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/i18n"

	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	loadConfigFn       = config.LoadConfig
	initializeConfigFn = config.InitializeConfig
	loadTranslationsFn = i18n.LoadTranslations
	configDatabaseFn   = database.ConfigDatabase
	newContextFn       = NewContext
	newRouterFn        = gin.Default
	registerRoutesFn   = RegisterRoutes
	startWorkersFn     = worker.StartWorkers
)

type Application struct {
	Router  *gin.Engine
	Context *AppContext
	Server  *http.Server
}

func InitializeApp() (*Application, error) {
	if err := loadConfigFn(); err != nil {
		return nil, err
	}
	initializeConfigFn()

	if err := loadTranslationsFn(); err != nil {
		return nil, err
	}

	if config.AppConfig.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	database, err := configDatabaseFn()
	if err != nil {
		return nil, err
	}

	appContext := newContextFn(database)

	router := newRouterFn()

	registerRoutesFn(router, appContext)

	workerFileContext := &worker.WorkerContext{
		FilesService:        appContext.Files.Service,
		VideoService:        appContext.Video.Service,
		MetadataService:     appContext.Files.MetadataRepository,
		Tasks:               *appContext.Tasks,
		Logger:              appContext.Logger,
		NotificationService: appContext.Notifications.Service,
	}
	if appContext.Jobs != nil {
		workerFileContext.JobsRepository = appContext.Jobs.Repository
	}

	startWorkersFn(workerFileContext, 200)

	return &Application{
		Router:  router,
		Context: appContext,
	}, nil
}

func (app *Application) Run(addr string, enableGraceFul bool) error {
	if app.Router == nil {
		return fmt.Errorf("router is nil")
	}

	server := &http.Server{
		Addr:    addr,
		Handler: app.Router.Handler(),
	}

	app.Server = server

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (app *Application) Stop() error {
	if app == nil {
		return nil
	}

	if app.Server == nil {
		return nil
	}

	log.Println("Parando servidor...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.Server.Shutdown(ctx); err != nil {
		log.Printf("Erro ao desligar servidor: %v\n", err)
		if closeErr := app.Server.Close(); closeErr != nil {
			log.Printf("Erro ao forcar fechamento do servidor: %v\n", closeErr)
			return closeErr
		}
		return err
	}

	log.Println("Servidor encerrado com sucesso.")
	return nil
}
