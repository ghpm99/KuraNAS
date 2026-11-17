package autoshutdown

import (
	"time"

	"nas-go/api/pkg/applog"
)

// DefaultCheckInterval is the schedule-evaluation tick. One minute is enough
// because the configured time has minute granularity; the end of the warning
// period is handled by a dedicated timer so the countdown stays second-precise
// without querying the settings every second.
const DefaultCheckInterval = time.Minute

// Scheduler powers the machine off at the configured local time. At the
// configured minute it opens a warning period (shared through Countdown) and
// notifies; only when that period ends uncancelled does it issue the OS
// shutdown. A per-day guard keeps it from firing twice on the same day, which
// also makes a cancellation last until the next day.
type Scheduler struct {
	service        SchedulerInterface
	countdown      *Countdown
	shutdownFn     func() error
	notifyFn       func(graceSeconds int)
	nowFn          func() time.Time
	interval       time.Duration
	lastFired      string
	warningStarted chan struct{}
	stop           chan struct{}
	done           chan struct{}
}

// NewScheduler builds the scheduler. shutdownFn performs the immediate OS
// shutdown and is injectable so tests never power anything off; a nil
// shutdownFn falls back to the platform default. notifyFn may be nil.
func NewScheduler(service SchedulerInterface, countdown *Countdown, shutdownFn func() error, notifyFn func(graceSeconds int), interval time.Duration) *Scheduler {
	if shutdownFn == nil {
		shutdownFn = ExecuteShutdown
	}
	if interval <= 0 {
		interval = DefaultCheckInterval
	}
	return &Scheduler{
		service:        service,
		countdown:      countdown,
		shutdownFn:     shutdownFn,
		notifyFn:       notifyFn,
		nowFn:          time.Now,
		interval:       interval,
		warningStarted: make(chan struct{}, 1),
		stop:           make(chan struct{}),
		done:           make(chan struct{}),
	}
}

func (s *Scheduler) Start() {
	go s.run()
}

func (s *Scheduler) Stop() {
	close(s.stop)
	<-s.done
}

func (s *Scheduler) run() {
	defer close(s.done)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	var warningEnd <-chan time.Time
	for {
		select {
		case <-ticker.C:
			s.tick()
		case <-s.warningStarted:
			warningEnd = time.After(s.countdown.remaining(s.nowFn()))
		case <-warningEnd:
			warningEnd = nil
			s.shutdownIfWarningElapsed()
		case <-s.stop:
			return
		}
	}
}

func (s *Scheduler) tick() {
	s.shutdownIfWarningElapsed()

	now := s.nowFn()
	due, graceSeconds, err := s.service.DueNow(now)
	if err != nil {
		applog.ErrorWithStack("autoshutdown: failed to evaluate schedule", err)
		return
	}
	if !due {
		return
	}

	today := now.Format("2006-01-02")
	if s.lastFired == today {
		return
	}
	s.lastFired = today

	s.countdown.begin(now.Add(time.Duration(graceSeconds) * time.Second))
	s.signalWarningStarted()
	applog.Info("autoshutdown: warning period started", "grace_seconds", graceSeconds)

	if s.notifyFn != nil {
		s.notifyFn(graceSeconds)
	}
	s.shutdownIfWarningElapsed()
}

func (s *Scheduler) signalWarningStarted() {
	select {
	case s.warningStarted <- struct{}{}:
	default:
	}
}

func (s *Scheduler) shutdownIfWarningElapsed() {
	if !s.countdown.finishIfElapsed(s.nowFn()) {
		return
	}
	applog.Info("autoshutdown: warning period ended, shutting down")
	if err := s.shutdownFn(); err != nil {
		applog.ErrorWithStack("autoshutdown: failed to trigger shutdown", err)
	}
}
