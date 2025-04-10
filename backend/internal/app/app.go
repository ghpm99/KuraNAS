package app

import (
	"context"
	"nas-go/api/internal/config"
	"nas-go/api/internal/worker"
	"nas-go/api/pkg/database"

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
}

func InitializeApp() (*Application, error) {
	if err := config.LoadConfig(); err != nil {
		return nil, err
	}
	config.InitializeConfig()

	database, err := database.ConfigDatabase()
	if err != nil {
		return nil, err
	}

	appContext := NewContext(database)

	router := gin.Default()

	RegisterRoutes(router, appContext)

	workerFileContext := &worker.WorkerContext{
		Service:    appContext.Files.Service,
		Repository: appContext.Files.Repository,
		Tasks:      *appContext.Tasks,
	}

	worker.StartWorkers(workerFileContext, 2)

	return &Application{
		Router:  router,
		Context: appContext,
	}, nil
}

func (app *Application) Run(addr string, enableGraceFul bool) error {
	if enableGraceFul {
		return configGracefulStop(app.Router)
	} else {
		return app.Router.Run(addr)
	}
}

func configGracefulStop(router *gin.Engine) error {
	server := &http.Server{
		Addr:    ":8000",
		Handler: router.Handler(),
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		log.Println("timeout of 5 seconds.")
	}
	log.Println("Server exiting")

	return nil
}
