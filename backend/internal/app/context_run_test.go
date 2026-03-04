package app

import (
	"database/sql"
	"errors"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"nas-go/api/internal/config"
	"nas-go/api/internal/worker"
	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

func TestNewContextBuildsAllDependencies(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	defer db.Close()

	ctx := NewContext(db)
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
	if ctx.ConfigurationHandler == nil || ctx.UpdateHandler == nil {
		t.Fatalf("expected configuration and update handlers initialized")
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
	t.Cleanup(func() {
		loadConfigFn = origLoadConfig
		initializeConfigFn = origInitializeConfig
		loadTranslationsFn = origLoadTranslations
		configDatabaseFn = origConfigDatabase
		newContextFn = origNewContext
		newRouterFn = origNewRouter
		registerRoutesFn = origRegisterRoutes
		startWorkersFn = origStartWorkers
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
	t.Cleanup(func() {
		loadConfigFn = origLoadConfig
		initializeConfigFn = origInitializeConfig
		loadTranslationsFn = origLoadTranslations
		configDatabaseFn = origConfigDatabase
		newContextFn = origNewContext
		newRouterFn = origNewRouter
		registerRoutesFn = origRegisterRoutes
		startWorkersFn = origStartWorkers
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
			Tasks: &tasks,
			Files: &FileContext{},
			Video: &VideoContext{},
		}
	}
	newRouterFn = func(opts ...gin.OptionFunc) *gin.Engine { return gin.New(opts...) }
	registerCalled := false
	registerRoutesFn = func(router *gin.Engine, context *AppContext) { registerCalled = true }
	workersCalled := false
	startWorkersFn = func(context *worker.WorkerContext, numWorkers int) { workersCalled = true }

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

	config.AppConfig.Env = "dev"
	if _, err := InitializeApp(); err != nil {
		t.Fatalf("expected InitializeApp success in debug mode, err=%v", err)
	}
	if gin.Mode() != gin.DebugMode {
		t.Fatalf("expected debug mode for non-production env")
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
