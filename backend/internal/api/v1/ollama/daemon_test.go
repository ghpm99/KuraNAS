package ollama

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"sync"
	"testing"
	"time"
)

// daemonStub is a tiny in-memory stand-in for the real Ollama daemon: it only
// answers /api/version, and only once `up` is true. Tests flip `up` to simulate
// the daemon coming online after `ollama serve` is spawned.
type daemonStub struct {
	mu sync.Mutex
	up bool
}

func (d *daemonStub) setUp(up bool) {
	d.mu.Lock()
	d.up = up
	d.mu.Unlock()
}

func (d *daemonStub) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/version", func(w http.ResponseWriter, r *http.Request) {
		d.mu.Lock()
		up := d.up
		d.mu.Unlock()
		if !up {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"version":"0.1.0"}`))
	})
	return mux
}

// newTestDaemonManager wires a manager against a live test server with the
// process-spawn and clock seams overridden so nothing real is launched.
func newTestDaemonManager(t *testing.T, cfg DaemonConfig, enabled bool, stub *daemonStub) (*DaemonManager, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(stub.handler())
	t.Cleanup(server.Close)

	m := NewDaemonManager(cfg, func() string { return server.URL }, func() bool { return enabled })
	m.sleep = func(time.Duration) {} // don't actually wait between polls
	return m, server
}

func TestEnsureRunningDisabledDoesNothing(t *testing.T) {
	stub := &daemonStub{up: false}
	m, _ := newTestDaemonManager(t, DaemonConfig{Autostart: true}, false, stub)

	spawned := false
	m.startServe = func() error { spawned = true; return nil }

	outcome, err := m.EnsureRunning(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if outcome != OutcomeDisabled {
		t.Fatalf("expected %q, got %q", OutcomeDisabled, outcome)
	}
	if spawned {
		t.Fatal("disabled provider must not spawn the daemon")
	}
}

func TestEnsureRunningAlreadyRunning(t *testing.T) {
	stub := &daemonStub{up: true}
	m, _ := newTestDaemonManager(t, DaemonConfig{Autostart: true}, true, stub)

	spawned := false
	m.startServe = func() error { spawned = true; return nil }

	outcome, err := m.EnsureRunning(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if outcome != OutcomeAlreadyRunning {
		t.Fatalf("expected %q, got %q", OutcomeAlreadyRunning, outcome)
	}
	if spawned {
		t.Fatal("a reachable daemon must not be spawned again")
	}
}

func TestEnsureRunningAutostartDisabled(t *testing.T) {
	stub := &daemonStub{up: false}
	m, _ := newTestDaemonManager(t, DaemonConfig{Autostart: false}, true, stub)

	spawned := false
	m.startServe = func() error { spawned = true; return nil }

	outcome, err := m.EnsureRunning(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if outcome != OutcomeUnreachable {
		t.Fatalf("expected %q, got %q", OutcomeUnreachable, outcome)
	}
	if spawned {
		t.Fatal("autostart off must not spawn the daemon")
	}
}

func TestEnsureRunningBinaryMissing(t *testing.T) {
	stub := &daemonStub{up: false}
	m, _ := newTestDaemonManager(t, DaemonConfig{Autostart: true}, true, stub)
	m.lookPath = func(string) (string, error) { return "", errors.New("not found") }

	spawned := false
	m.startServe = func() error { spawned = true; return nil }

	outcome, err := m.EnsureRunning(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if outcome != OutcomeBinaryMissing {
		t.Fatalf("expected %q, got %q", OutcomeBinaryMissing, outcome)
	}
	if spawned {
		t.Fatal("missing binary must not spawn the daemon")
	}
}

func TestEnsureRunningStartsDaemon(t *testing.T) {
	stub := &daemonStub{up: false}
	m, _ := newTestDaemonManager(t, DaemonConfig{
		Autostart:    true,
		StartTimeout: 5 * time.Second,
		PollInterval: time.Millisecond,
	}, true, stub)
	m.lookPath = func(string) (string, error) { return "/usr/bin/ollama", nil }
	m.startServe = func() error {
		stub.setUp(true) // the spawned daemon comes online
		return nil
	}

	outcome, err := m.EnsureRunning(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if outcome != OutcomeStarted {
		t.Fatalf("expected %q, got %q", OutcomeStarted, outcome)
	}
}

func TestEnsureRunningSpawnFails(t *testing.T) {
	stub := &daemonStub{up: false}
	m, _ := newTestDaemonManager(t, DaemonConfig{Autostart: true}, true, stub)
	m.lookPath = func(string) (string, error) { return "/usr/bin/ollama", nil }
	m.startServe = func() error { return errors.New("spawn boom") }

	outcome, err := m.EnsureRunning(context.Background())
	if err == nil {
		t.Fatal("expected spawn error to propagate")
	}
	if outcome != OutcomeUnreachable {
		t.Fatalf("expected %q, got %q", OutcomeUnreachable, outcome)
	}
}

func TestEnsureRunningNeverComesUp(t *testing.T) {
	stub := &daemonStub{up: false}
	m, _ := newTestDaemonManager(t, DaemonConfig{
		Autostart:    true,
		StartTimeout: 20 * time.Millisecond,
		PollInterval: time.Millisecond,
	}, true, stub)
	m.lookPath = func(string) (string, error) { return "/usr/bin/ollama", nil }
	m.startServe = func() error { return nil } // spawn "succeeds" but daemon stays down

	// Use the real clock so the deadline actually elapses; the timeout is tiny.
	outcome, err := m.EnsureRunning(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if outcome != OutcomeUnreachable {
		t.Fatalf("expected %q, got %q", OutcomeUnreachable, outcome)
	}
}

func TestSpawnServeStartsAndReapsProcess(t *testing.T) {
	binary, err := exec.LookPath("echo")
	if err != nil {
		t.Skip("echo binary not available")
	}

	m := NewDaemonManager(DaemonConfig{Binary: binary}, func() string { return "http://localhost:11434" }, func() bool { return true })
	if err := m.spawnServe(); err != nil {
		t.Fatalf("spawnServe should start a runnable binary: %v", err)
	}
	// Give the reaper goroutine a moment to Wait on the short-lived process.
	time.Sleep(50 * time.Millisecond)
}

func TestSpawnServeFailsForMissingBinary(t *testing.T) {
	m := NewDaemonManager(DaemonConfig{Binary: "definitely-not-a-real-binary-xyz"}, func() string { return "http://localhost:11434" }, func() bool { return true })
	if err := m.spawnServe(); err == nil {
		t.Fatal("expected spawnServe to fail for a missing binary")
	}
}

func TestServeEnvDerivesHost(t *testing.T) {
	m := NewDaemonManager(DaemonConfig{}, func() string { return "http://192.168.0.5:11500" }, func() bool { return true })
	env := m.serveEnv()
	if len(env) != 1 || env[0] != "OLLAMA_HOST=192.168.0.5:11500" {
		t.Fatalf("expected OLLAMA_HOST env, got %v", env)
	}
}

func TestServeEnvEmptyForUnparseableURL(t *testing.T) {
	m := NewDaemonManager(DaemonConfig{}, func() string { return "://bad" }, func() bool { return true })
	if env := m.serveEnv(); env != nil {
		t.Fatalf("expected nil env for unparseable URL, got %v", env)
	}
}
