package autoshutdown

import (
	"log"
	"time"
)

// DefaultCheckInterval is the scheduler tick. One minute is fine because the
// configured time has minute granularity.
const DefaultCheckInterval = time.Minute

// Scheduler powers the machine off at the configured local time. It is the
// "agendador" of auto-shutdown — same background shape as trash.Purger: a single
// goroutine that, on each tick, asks the service whether a shutdown is due and,
// when it is, warns the user and triggers the OS shutdown. A per-day guard keeps
// it from firing twice in the same minute window.
type Scheduler struct {
	service    SchedulerInterface
	shutdownFn func(graceSeconds int) error
	notifyFn   func(graceSeconds int)
	nowFn      func() time.Time
	interval   time.Duration
	lastFired  string // "2006-01-02" of the last successful trigger
	stop       chan struct{}
	done       chan struct{}
}

// NewScheduler builds the scheduler. shutdownFn performs the OS shutdown and is
// injectable so tests never power anything off; a nil shutdownFn falls back to
// the platform default. notifyFn may be nil (no warning is emitted).
func NewScheduler(service SchedulerInterface, shutdownFn func(graceSeconds int) error, notifyFn func(graceSeconds int), interval time.Duration) *Scheduler {
	if shutdownFn == nil {
		shutdownFn = ExecuteShutdown
	}
	if interval <= 0 {
		interval = DefaultCheckInterval
	}
	return &Scheduler{
		service:    service,
		shutdownFn: shutdownFn,
		notifyFn:   notifyFn,
		nowFn:      time.Now,
		interval:   interval,
		stop:       make(chan struct{}),
		done:       make(chan struct{}),
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

	for {
		select {
		case <-ticker.C:
			s.tick()
		case <-s.stop:
			return
		}
	}
}

func (s *Scheduler) tick() {
	now := s.nowFn()
	due, graceSeconds, err := s.service.DueNow(now)
	if err != nil {
		log.Printf("autoshutdown: failed to evaluate schedule: %v", err)
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

	if s.notifyFn != nil {
		s.notifyFn(graceSeconds)
	}

	log.Printf("autoshutdown: scheduled shutdown triggered (grace %ds)", graceSeconds)
	if err := s.shutdownFn(graceSeconds); err != nil {
		log.Printf("autoshutdown: failed to trigger shutdown: %v", err)
	}
}
