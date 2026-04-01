package app

import (
	"database/sql"
	"errors"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/systemevent"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"nas-go/api/internal/config"
	"nas-go/api/internal/worker"
	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

func TestNewContextBuildsAllDependencies(t *testing.T) {
	ctx := NewContext(nil)
	if ctx == nil || ctx.DB == nil || ctx.Tasks == nil || ctx.Logger == nil {
		t.Fatalf("expected fully initialized app context")
	}
	if ctx.Files == nil || ctx.Files.Handler == nil || ctx.Files.Service == nil {
		t.Fatalf("expected files context initialized")
	}
	if ctx.Diary == nil || ctx.Diary.Handler == nil || ctx.Diary.Service == nil {
		t.Fatalf("expected diary context initialized")
	}
	if ctx.Music == nil || ctx.Music.Handler == nil || ctx.Music.Service == nil {
		t.Fatalf("expected music context initialized")
	}
	if ctx.Video == nil || ctx.Video.Handler == nil || ctx.Video.Service == nil {
		t.Fatalf("expected video context initialized")
	}
	if ctx.Configuration == nil || ctx.Configuration.Handler == nil || ctx.Configuration.Service == nil || ctx.UpdateHandler == nil || ctx.UpdateService == nil {
		t.Fatalf("expected configuration context, update handler and update service initialized")
	}
	if ctx.WatchFolders == nil || ctx.WatchFolders.Handler == nil || ctx.WatchFolders.Service == nil {
		t.Fatalf("expected watch folders context initialized")
	}
}

func TestApplicationRunAndStopGuards(t *testing.T) {
	app := &Application{}

	if err := app.Stop(); err != nil {
		t.Fatalf("expected Stop to succeed with nil server, got %v", err)
	}

	if err := app.Run(":8000", false); err == nil {
		t.Fatalf("expected Run error when router is nil")
	}

	app.Router = gin.New()
	if err := app.Run("invalid-addr", false); err == nil {
		t.Fatalf("expected Run to return listen error for invalid addr")
	}
}

func TestApplicationStop_NilReceiverDoesNotPanic(t *testing.T) {
	var app *Application

	if err := app.Stop(); err != nil {
		t.Fatalf("expected Stop to succeed for nil receiver, got %v", err)
	}
}

type failingListener struct{}

func (f *failingListener) Accept() (net.Conn, error) { return nil, errors.New("accept failed") }
func (f *failingListener) Close() error              { return nil }
func (f *failingListener) Addr() net.Addr            { return &net.TCPAddr{} }

func TestApplicationStop_IdleServerIsNoop(t *testing.T) {
	app := &Application{
		Router: gin.New(),
		Server: &http.Server{},
	}
	errCh := make(chan error, 1)
	go func() {
		errCh <- app.Server.Serve(&failingListener{})
	}()

	time.Sleep(20 * time.Millisecond)
	if err := app.Stop(); err != nil {
		t.Fatalf("expected Stop to be a no-op on idle server, got %v", err)
	}
	<-errCh
}

func TestInitializeAppReturnsErrorWhenTranslationsAreMissing(t *testing.T) {
	t.Setenv("LANGUAGE", "zz-ZZ")
	t.Setenv("ENV", "test")

	app, err := InitializeApp()
	if err == nil {
		t.Fatalf("expected InitializeApp to fail when translation file is missing")
	}
	if app != nil {
		t.Fatalf("expected nil app on InitializeApp failure")
	}
}

func TestInitializeAppLoadConfigAndDatabaseErrors(t *testing.T) {
	origLoadConfig := loadConfigFn
	origInitializeConfig := initializeConfigFn
	origLoadTranslations := loadTranslationsFn
	origConfigDatabase := configDatabaseFn
	origNewContext := newContextFn
	origNewRouter := newRouterFn
	origRegisterRoutes := registerRoutesFn
	origStartWorkers := startWorkersFn
	origNewFolderWatcher := newFolderWatcherFn
	origNewSystemEvent := newSystemEventFn
	t.Cleanup(func() {
		loadConfigFn = origLoadConfig
		initializeConfigFn = origInitializeConfig
		loadTranslationsFn = origLoadTranslations
		configDatabaseFn = origConfigDatabase
		newContextFn = origNewContext
		newRouterFn = origNewRouter
		registerRoutesFn = origRegisterRoutes
		startWorkersFn = origStartWorkers
		newFolderWatcherFn = origNewFolderWatcher
		newSystemEventFn = origNewSystemEvent
	})

	loadConfigFn = func() error { return errors.New("load failed") }
	if app, err := InitializeApp(); err == nil || app != nil {
		t.Fatalf("expected load config error")
	}

	loadConfigFn = func() error { return nil }
	initializeConfigFn = func() {}
	loadTranslationsFn = func() error { return nil }
	configDatabaseFn = func() (*sql.DB, error) { return nil, errors.New("db failed") }
	if app, err := InitializeApp(); err == nil || app != nil {
		t.Fatalf("expected database error")
	}
}

func TestInitializeAppSuccessAndModeSelection(t *testing.T) {
	origLoadConfig := loadConfigFn
	origInitializeConfig := initializeConfigFn
	origLoadTranslations := loadTranslationsFn
	origConfigDatabase := configDatabaseFn
	origNewContext := newContextFn
	origNewRouter := newRouterFn
	origRegisterRoutes := registerRoutesFn
	origStartWorkers := startWorkersFn
	origNewFolderWatcher := newFolderWatcherFn
	origNewSystemEvent := newSystemEventFn
	t.Cleanup(func() {
		loadConfigFn = origLoadConfig
		initializeConfigFn = origInitializeConfig
		loadTranslationsFn = origLoadTranslations
		configDatabaseFn = origConfigDatabase
		newContextFn = origNewContext
		newRouterFn = origNewRouter
		registerRoutesFn = origRegisterRoutes
		startWorkersFn = origStartWorkers
		newFolderWatcherFn = origNewFolderWatcher
		newSystemEventFn = origNewSystemEvent
	})

	loadConfigFn = func() error { return nil }
	initializeConfigFn = func() {}
	loadTranslationsFn = func() error { return nil }
	configDatabaseFn = func() (*sql.DB, error) {
		return &sql.DB{}, nil
	}
	newContextFn = func(db *sql.DB) *AppContext {
		tasks := make(chan utils.Task, 1)
		return &AppContext{
			Tasks:         &tasks,
			Files:         &FileContext{},
			Video:         &VideoContext{},
			Notifications: &NotificationContext{},
		}
	}
	newRouterFn = func(opts ...gin.OptionFunc) *gin.Engine { return gin.New(opts...) }
	registerCalled := false
	registerRoutesFn = func(router *gin.Engine, context *AppContext) { registerCalled = true }
	workersCalled := false
	startWorkersFn = func(context *worker.WorkerContext, numWorkers int) { workersCalled = true }
	folderWatcher := &folderWatcherSpy{}
	newFolderWatcherFn = func(context *AppContext) FolderWatcherInterface { return folderWatcher }
	systemEvents := &systemEventServiceSpy{}
	newSystemEventFn = func(*database.DbContext) systemevent.ServiceInterface { return systemEvents }

	config.AppConfig.Env = "production"
	app, err := InitializeApp()
	if err != nil || app == nil {
		t.Fatalf("expected InitializeApp success, err=%v", err)
	}
	if gin.Mode() != gin.ReleaseMode {
		t.Fatalf("expected release mode for production env")
	}
	if !registerCalled || !workersCalled {
		t.Fatalf("expected routes and workers to be started")
	}
	if folderWatcher.startCalls != 1 {
		t.Fatalf("expected folder watcher start once, got %d", folderWatcher.startCalls)
	}
	if systemEvents.startupCalls != 1 {
		t.Fatalf("expected startup event to be recorded once")
	}

	config.AppConfig.Env = "dev"
	if _, err := InitializeApp(); err != nil {
		t.Fatalf("expected InitializeApp success in debug mode, err=%v", err)
	}
	if gin.Mode() != gin.DebugMode {
		t.Fatalf("expected debug mode for non-production env")
	}
	if folderWatcher.startCalls != 2 {
		t.Fatalf("expected folder watcher start twice, got %d", folderWatcher.startCalls)
	}
}

func TestApplicationRunAndStopSuccess(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		if strings.Contains(err.Error(), "operation not permitted") {
			t.Skip("socket operations are not permitted in this test environment")
		}
		t.Fatalf("failed to reserve ephemeral port: %v", err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()

	app := &Application{Router: gin.New()}
	errCh := make(chan error, 1)
	go func() {
		errCh <- app.Run(addr, false)
	}()

	deadline := time.Now().Add(2 * time.Second)
	for app.Server == nil && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if app.Server == nil {
		t.Fatalf("server was not initialized by Run")
	}

	if err := app.Stop(); err != nil {
		t.Fatalf("expected Stop success for running server, got %v", err)
	}

	select {
	case runErr := <-errCh:
		if runErr != nil {
			t.Fatalf("expected Run to return nil after graceful stop, got %v", runErr)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("Run did not return after Stop")
	}
}

type systemEventServiceSpy struct {
	startupCalls  int
	shutdownCalls int
	startupErr    error
	shutdownErr   error
}

func (s *systemEventServiceSpy) RecordStartup() error {
	s.startupCalls++
	return s.startupErr
}

func (s *systemEventServiceSpy) RecordShutdown() error {
	s.shutdownCalls++
	return s.shutdownErr
}

type folderWatcherSpy struct {
	startCalls int
	stopCalls  int
}

func (s *folderWatcherSpy) Start() {
	s.startCalls++
}

func (s *folderWatcherSpy) Stop() {
	s.stopCalls++
}

func TestApplicationStopRecordsShutdownEvent(t *testing.T) {
	spy := &systemEventServiceSpy{}
	watcherSpy := &folderWatcherSpy{}
	app := &Application{
		Router:        gin.New(),
		FolderWatcher: watcherSpy,
		SystemEvents:  spy,
	}

	if err := app.Stop(); err != nil {
		t.Fatalf("expected Stop to succeed, got %v", err)
	}

	if spy.shutdownCalls != 1 {
		t.Fatalf("expected shutdown event to be recorded once")
	}
	if watcherSpy.stopCalls != 1 {
		t.Fatalf("expected folder watcher stop once")
	}
}

func TestInitializeAppStartupEventFailureDoesNotBreakInitialization(t *testing.T) {
	origLoadConfig := loadConfigFn
	origInitializeConfig := initializeConfigFn
	origLoadTranslations := loadTranslationsFn
	origConfigDatabase := configDatabaseFn
	origNewContext := newContextFn
	origNewRouter := newRouterFn
	origRegisterRoutes := registerRoutesFn
	origStartWorkers := startWorkersFn
	origNewFolderWatcher := newFolderWatcherFn
	origNewSystemEvent := newSystemEventFn
	t.Cleanup(func() {
		loadConfigFn = origLoadConfig
		initializeConfigFn = origInitializeConfig
		loadTranslationsFn = origLoadTranslations
		configDatabaseFn = origConfigDatabase
		newContextFn = origNewContext
		newRouterFn = origNewRouter
		registerRoutesFn = origRegisterRoutes
		startWorkersFn = origStartWorkers
		newFolderWatcherFn = origNewFolderWatcher
		newSystemEventFn = origNewSystemEvent
	})

	loadConfigFn = func() error { return nil }
	initializeConfigFn = func() {}
	loadTranslationsFn = func() error { return nil }
	configDatabaseFn = func() (*sql.DB, error) { return &sql.DB{}, nil }
	newContextFn = func(db *sql.DB) *AppContext {
		tasks := make(chan utils.Task, 1)
		return &AppContext{
			DB:            database.NewDbContext(db),
			Tasks:         &tasks,
			Files:         &FileContext{},
			Video:         &VideoContext{},
			Notifications: &NotificationContext{},
		}
	}
	newRouterFn = func(opts ...gin.OptionFunc) *gin.Engine { return gin.New(opts...) }
	registerRoutesFn = func(router *gin.Engine, context *AppContext) {}
	startWorkersFn = func(context *worker.WorkerContext, numWorkers int) {}
	newFolderWatcherFn = func(context *AppContext) FolderWatcherInterface { return nil }
	spy := &systemEventServiceSpy{startupErr: errors.New("startup log failed")}
	newSystemEventFn = func(*database.DbContext) systemevent.ServiceInterface { return spy }

	application, err := InitializeApp()
	if err != nil || application == nil {
		t.Fatalf("expected InitializeApp to succeed even if startup event logging fails, err=%v", err)
	}
	if spy.startupCalls != 1 {
		t.Fatalf("expected startup event call once")
	}
}

func TestApplicationStopIgnoresShutdownEventFailure(t *testing.T) {
	spy := &systemEventServiceSpy{shutdownErr: errors.New("shutdown log failed")}
	app := &Application{
		Router:       gin.New(),
		SystemEvents: spy,
	}

	if err := app.Stop(); err != nil {
		t.Fatalf("expected Stop to succeed even when shutdown event fails, got %v", err)
	}
	if spy.shutdownCalls != 1 {
		t.Fatalf("expected shutdown event call once")
	}
}
