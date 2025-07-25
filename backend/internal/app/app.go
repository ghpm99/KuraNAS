package app

import (
	"context"
	"nas-go/api/internal/config"
	"nas-go/api/internal/worker"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/i18n"

	"log"
	"net/http"
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

	if err := i18n.LoadTranslations(); err != nil {
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
		FilesService:    appContext.Files.Service,
		MetadataService: appContext.Files.MetadataRepository,
		Tasks:           *appContext.Tasks,
		Logger:          appContext.Logger,
	}

	worker.StartWorkers(workerFileContext, 200)

	return &Application{
		Router:  router,
		Context: appContext,
	}, nil
}

func (app *Application) Run(addr string, enableGraceFul bool) error {
	server := &http.Server{
		Addr:    addr,
		Handler: app.Router.Handler(),
	}

	app.Server = server

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %s\n", err)
	}

	return nil
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
