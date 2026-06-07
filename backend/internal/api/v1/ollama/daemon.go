package ollama

import (
	"context"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"nas-go/api/pkg/applog"
)

// EnsureOutcome reports what EnsureRunning did, so the caller can map it to
// observability events without the daemon manager depending on systemevent.
type EnsureOutcome string

const (
	// OutcomeDisabled means the Ollama provider is not enabled, so the daemon
	// lifecycle is left untouched (zero side effects on boot).
	OutcomeDisabled EnsureOutcome = "disabled"
	// OutcomeAlreadyRunning means the daemon answered the probe, so nothing was
	// spawned.
	OutcomeAlreadyRunning EnsureOutcome = "already_running"
	// OutcomeStarted means the daemon was unreachable and we spawned it; it
	// answered within the start timeout.
	OutcomeStarted EnsureOutcome = "started"
	// OutcomeBinaryMissing means the daemon was unreachable and there is no local
	// binary to spawn, so the host is assumed to talk to a remote daemon.
	OutcomeBinaryMissing EnsureOutcome = "binary_missing"
	// OutcomeUnreachable means the daemon is down and could not be brought up
	// (autostart disabled, spawn failed, or it never answered in time).
	OutcomeUnreachable EnsureOutcome = "unreachable"
)

// DaemonConfig holds the host-level lifecycle knobs for the local daemon. They
// come from the environment (not the ai_providers table) because they describe
// the machine running the backend, not provider tuning.
type DaemonConfig struct {
	Autostart    bool
	Binary       string
	StartTimeout time.Duration
	PollInterval time.Duration
}

// DaemonManager verifies the local Ollama daemon at boot and, when it is down
// and a binary is available, spawns `ollama serve` and waits for it to come up.
// Everything else (chat, pulls, model listing) still talks plain HTTP — this is
// the only piece that touches the daemon process.
type DaemonManager struct {
	cfg     DaemonConfig
	client  *Client
	baseURL func() string
	enabled func() bool

	// Seams kept injectable so the orchestration can be tested without a real
	// daemon or binary on the box.
	lookPath   func(string) (string, error)
	startServe func() error
	sleep      func(time.Duration)
	now        func() time.Time
}

// NewDaemonManager builds the manager from the shared base-URL resolver (so it
// honours DB config changes) and an `enabled` predicate reading the provider
// table. startServe defaults to spawning the real binary; tests override it.
func NewDaemonManager(cfg DaemonConfig, baseURL func() string, enabled func() bool) *DaemonManager {
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = 500 * time.Millisecond
	}
	if cfg.StartTimeout <= 0 {
		cfg.StartTimeout = 30 * time.Second
	}
	if strings.TrimSpace(cfg.Binary) == "" {
		cfg.Binary = "ollama"
	}

	m := &DaemonManager{
		cfg:      cfg,
		client:   NewClient(baseURL),
		baseURL:  baseURL,
		enabled:  enabled,
		lookPath: exec.LookPath,
		sleep:    time.Sleep,
		now:      time.Now,
	}
	m.startServe = m.spawnServe
	return m
}

// EnsureRunning is the boot hook. It is safe to call on every startup: it does
// nothing when the provider is disabled or the daemon is already up, and it
// never blocks longer than the configured start timeout.
func (m *DaemonManager) EnsureRunning(ctx context.Context) (EnsureOutcome, error) {
	if m == nil {
		return OutcomeDisabled, nil
	}
	if m.enabled != nil && !m.enabled() {
		return OutcomeDisabled, nil
	}

	if m.reachable(ctx) {
		return OutcomeAlreadyRunning, nil
	}

	if !m.cfg.Autostart {
		return OutcomeUnreachable, nil
	}

	if _, err := m.lookPath(m.cfg.Binary); err != nil {
		// No local binary: this host most likely points at a remote daemon, so
		// spawning here is meaningless. Not treated as a hard error.
		return OutcomeBinaryMissing, nil
	}

	if err := m.startServe(); err != nil {
		return OutcomeUnreachable, err
	}

	if m.waitReachable(ctx) {
		return OutcomeStarted, nil
	}
	return OutcomeUnreachable, nil
}

// reachable probes the daemon with a short, bounded request.
func (m *DaemonManager) reachable(ctx context.Context) bool {
	probeCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	_, err := m.client.Version(probeCtx)
	return err == nil
}

// waitReachable polls until the daemon answers or the start timeout elapses.
func (m *DaemonManager) waitReachable(ctx context.Context) bool {
	deadline := m.now().Add(m.cfg.StartTimeout)
	for {
		if m.reachable(ctx) {
			return true
		}
		if !m.now().Before(deadline) {
			return false
		}
		select {
		case <-ctx.Done():
			return false
		default:
		}
		m.sleep(m.cfg.PollInterval)
	}
}

// spawnServe launches `ollama serve` detached from the current request: the
// process outlives EnsureRunning and is reaped by a background Wait so a crash
// is logged instead of leaving a zombie. OLLAMA_HOST is pinned to the
// configured base URL so a custom host/port is honoured.
func (m *DaemonManager) spawnServe() error {
	cmd := exec.Command(m.cfg.Binary, "serve")
	cmd.Env = append(os.Environ(), m.serveEnv()...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return err
	}

	applog.Go("ollama-serve-wait", func() {
		if err := cmd.Wait(); err != nil {
			applog.Warn("ollama serve exited", "error", err.Error())
		} else {
			applog.Info("ollama serve exited")
		}
	})
	return nil
}

// serveEnv derives OLLAMA_HOST from the configured base URL so the daemon binds
// where the rest of the app expects to find it. Returns nothing for the default
// localhost:11434, which the daemon already uses out of the box.
func (m *DaemonManager) serveEnv() []string {
	parsed, err := url.Parse(strings.TrimSpace(m.baseURL()))
	if err != nil || parsed.Host == "" {
		return nil
	}
	return []string{"OLLAMA_HOST=" + parsed.Host}
}
