package app

import (
	"context"
	"fmt"
	"nas-go/api/internal/api/v1/notifications"
	ollamamgmt "nas-go/api/internal/api/v1/ollama"
	"nas-go/api/internal/config"
	"nas-go/api/internal/discovery"
	"nas-go/api/internal/watcher"
	"nas-go/api/internal/worker/engine"
	"nas-go/api/pkg/applog"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/systemevent"

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
	startWorkersFn     = engine.StartWorkers
	newFolderWatcherFn = func(context *AppContext) FolderWatcherInterface {
		if context == nil ||
			context.WatchFolders == nil || context.WatchFolders.Service == nil ||
			context.Libraries == nil || context.Libraries.Service == nil ||
			context.Files == nil || context.Files.Service == nil {
			return nil
		}

		var notificationService notifications.ServiceInterface
		if context.Notifications != nil {
			notificationService = context.Notifications.Service
		}
		return watcher.NewFolderWatcher(
			context.WatchFolders.Service,
			context.Libraries.Service,
			context.Files.Service,
			notificationService,
			60*time.Second,
		)
	}
	newSystemEventFn = func(dbContext *database.DbContext) systemevent.ServiceInterface {
		return systemevent.NewService(dbContext)
	}
)

type FolderWatcherInterface interface {
	Start()
	Stop()
}

type Application struct {
	Router        *gin.Engine
	Context       *AppContext
	Server        *http.Server
	UDPListener   *discovery.UDPListener
	MdnsRegistrar *discovery.MdnsRegistrar
	FolderWatcher FolderWatcherInterface
	SystemEvents  systemevent.ServiceInterface
}

func InitializeApp() (*Application, error) {
	if err := loadConfigFn(); err != nil {
		return nil, err
	}
	initializeConfigFn()

	// Re-apply the log level now that the .env LOG_LEVEL is loaded (the early
	// file logger was installed before config, using only OS env vars).
	applog.SetLevel(applog.ParseLevel(config.AppConfig.LogLevel))

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
	systemEvents := newSystemEventFn(appContext.DB)
	if appContext.Libraries != nil && appContext.Libraries.Service != nil {
		if err := appContext.Libraries.Service.ResolveLibraries(); err != nil {
			log.Printf("[APP] Failed to resolve libraries: %v", err)
		}
	}

	router := newRouterFn()

	registerRoutesFn(router, appContext)

	var librariesService = appContext.Libraries
	workerFileContext := &engine.WorkerContext{
		FilesService:        appContext.Files.Service,
		VideoService:        appContext.Video.Service,
		MetadataService:     appContext.Files.MetadataRepository,
		Tasks:               *appContext.Tasks,
		Logger:              appContext.Logger,
		NotificationService: appContext.Notifications.Service,
		AIService:           appContext.AI,
		SystemEvents:        systemEvents,
	}
	if appContext.Image != nil {
		workerFileContext.ImageRepository = appContext.Image.Repository
	}
	if librariesService != nil {
		workerFileContext.LibrariesService = librariesService.Service
	}
	if appContext.Music != nil {
		workerFileContext.MusicService = appContext.Music.Service
		workerFileContext.AudioMetadataRepository = appContext.Music.AudioMetadataRepository
	}
	if appContext.Jobs != nil {
		workerFileContext.JobsRepository = appContext.Jobs.Repository
	}
	if appContext.Configuration != nil {
		workerFileContext.AISettings = appContext.Configuration.Service
	}

	startWorkersFn(workerFileContext, 200)

	folderWatcher := newFolderWatcherFn(appContext)
	if folderWatcher != nil {
		folderWatcher.Start()
	}

	udpListener := discovery.NewUDPListener(discovery.DefaultUDPPort, 8000)
	if err := udpListener.Start(); err != nil {
		log.Printf("[APP] Failed to start UDP discovery listener: %v", err)
	}

	mdnsRegistrar := discovery.NewMdnsRegistrar(discovery.DefaultServiceName, 8000)
	if err := mdnsRegistrar.Start(); err != nil {
		log.Printf("[APP] Failed to start mDNS registrar: %v", err)
	}

	if err := systemEvents.RecordStartup(); err != nil {
		log.Printf("[APP] Failed to record startup system event: %v", err)
	}

	startOllamaDaemon(appContext, systemEvents)

	return &Application{
		Router:        router,
		Context:       appContext,
		UDPListener:   udpListener,
		MdnsRegistrar: mdnsRegistrar,
		FolderWatcher: folderWatcher,
		SystemEvents:  systemEvents,
	}, nil
}

// startOllamaDaemon verifies (and, when needed and possible, spawns) the local
// Ollama daemon at boot. It runs in a background goroutine so it never delays
// ListenAndServe, and degrades gracefully — exactly like the UDP/mDNS startup:
// any failure is logged/recorded and the server keeps booting.
func startOllamaDaemon(appContext *AppContext, systemEvents systemevent.ServiceInterface) {
	if appContext == nil || appContext.Ollama == nil || appContext.Ollama.Daemon == nil {
		return
	}
	daemon := appContext.Ollama.Daemon

	applog.Go("ollama-autostart", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		outcome, err := daemon.EnsureRunning(ctx)
		if err != nil {
			applog.Warn("ollama autostart failed", "outcome", string(outcome), "error", err.Error())
		}
		applyOllamaOutcome(systemEvents, outcome)
	})
}

// applyOllamaOutcome maps a daemon lifecycle outcome to logs and observability
// events. Split out from the goroutine so the mapping is deterministically
// testable without spawning processes.
func applyOllamaOutcome(systemEvents systemevent.ServiceInterface, outcome ollamamgmt.EnsureOutcome) {
	switch outcome {
	case ollamamgmt.OutcomeStarted:
		applog.Info("ollama daemon started automatically")
		recordEvent(systemEvents, systemevent.EventTypeOllamaDaemonStarted,
			i18n.GetMessage("SYSTEM_EVENT_OLLAMA_DAEMON_STARTED"))
	case ollamamgmt.OutcomeAlreadyRunning:
		applog.Info("ollama daemon already running")
	case ollamamgmt.OutcomeBinaryMissing:
		applog.Info("ollama autostart skipped: binary not found (assuming remote daemon)")
	case ollamamgmt.OutcomeUnreachable:
		applog.Warn("ollama daemon unreachable and could not be started")
		recordEvent(systemEvents, systemevent.EventTypeOllamaDaemonDown,
			i18n.GetMessage("SYSTEM_EVENT_OLLAMA_DAEMON_UNREACHABLE"))
	case ollamamgmt.OutcomeDisabled:
		// Provider off — nothing to do.
	}
}

func recordEvent(systemEvents systemevent.ServiceInterface, eventType systemevent.EventType, description string) {
	if systemEvents == nil {
		return
	}
	if err := systemEvents.RecordEvent(eventType, description); err != nil {
		applog.Warn("failed to record system event", "event", string(eventType), "error", err.Error())
	}
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

	if app.UDPListener != nil {
		app.UDPListener.Stop()
	}

	if app.MdnsRegistrar != nil {
		app.MdnsRegistrar.Stop()
	}

	if app.SystemEvents != nil {
		if err := app.SystemEvents.RecordShutdown(); err != nil {
			log.Printf("[APP] Failed to record shutdown system event: %v", err)
		}
	}

	if app.FolderWatcher != nil {
		app.FolderWatcher.Stop()
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
