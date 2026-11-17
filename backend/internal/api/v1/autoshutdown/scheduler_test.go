package autoshutdown

import (
	"errors"
	"testing"
	"time"
)

type stubScheduleService struct {
	due   bool
	grace int
	err   error
	calls int
}

func (s *stubScheduleService) DueNow(now time.Time) (bool, int, error) {
	s.calls++
	return s.due, s.grace, s.err
}

func mustTime(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse("2006-01-02 15:04", value)
	if err != nil {
		t.Fatalf("parse time: %v", err)
	}
	return parsed
}

type schedulerHarness struct {
	scheduler     *Scheduler
	countdown     *Countdown
	now           time.Time
	shutdownCalls int
	notifyCalls   int
}

func newHarness(t *testing.T, service SchedulerInterface, startAt string) *schedulerHarness {
	t.Helper()
	harness := &schedulerHarness{countdown: NewCountdown(), now: mustTime(t, startAt)}
	harness.scheduler = NewScheduler(service, harness.countdown,
		func() error { harness.shutdownCalls++; return nil },
		func(graceSeconds int) { harness.notifyCalls++ },
		time.Minute,
	)
	clock := func() time.Time { return harness.now }
	harness.scheduler.nowFn = clock
	harness.countdown.nowFn = clock
	return harness
}

func (h *schedulerHarness) advance(duration time.Duration) {
	h.now = h.now.Add(duration)
	h.scheduler.tick()
}

func TestSchedulerEntersWarningAtScheduledTimeWithoutShuttingDown(t *testing.T) {
	harness := newHarness(t, &stubScheduleService{due: true, grace: 60}, "2026-06-15 03:00")

	harness.scheduler.tick()

	status := harness.countdown.Status()
	if status.State != "warning" || status.SecondsRemaining == nil || *status.SecondsRemaining != 60 {
		t.Fatalf("expected warning with 60s remaining, got %+v", status)
	}
	if harness.shutdownCalls != 0 || harness.notifyCalls != 1 {
		t.Fatalf("expected notify only, got shutdown=%d notify=%d", harness.shutdownCalls, harness.notifyCalls)
	}
}

func TestSchedulerDoesNotShutDownBeforeWarningEnds(t *testing.T) {
	harness := newHarness(t, &stubScheduleService{due: true, grace: 60}, "2026-06-15 03:00")
	harness.scheduler.tick()

	harness.advance(59 * time.Second)

	if harness.shutdownCalls != 0 || harness.countdown.Status().State != "warning" {
		t.Fatalf("expected still warning without shutdown, got shutdown=%d", harness.shutdownCalls)
	}
}

func TestSchedulerShutsDownOnceWhenWarningEnds(t *testing.T) {
	harness := newHarness(t, &stubScheduleService{due: true, grace: 60}, "2026-06-15 03:00")
	harness.scheduler.tick()

	harness.advance(60 * time.Second)
	harness.advance(time.Second)

	if harness.shutdownCalls != 1 || harness.countdown.Status().State != "idle" {
		t.Fatalf("expected one shutdown and idle, got shutdown=%d", harness.shutdownCalls)
	}
}

func TestSchedulerDoesNotShutDownWhenCancelledDuringWarning(t *testing.T) {
	harness := newHarness(t, &stubScheduleService{due: true, grace: 60}, "2026-06-15 03:00")
	harness.scheduler.tick()

	status, err := harness.countdown.CancelToday()
	harness.advance(2 * time.Minute)

	if err != nil || status.State != "idle" {
		t.Fatalf("expected cancel to succeed and go idle, got %+v %v", status, err)
	}
	if harness.shutdownCalls != 0 || harness.notifyCalls != 1 {
		t.Fatalf("expected no shutdown and no new warning, got shutdown=%d notify=%d", harness.shutdownCalls, harness.notifyCalls)
	}
}

func TestSchedulerShutsDownNextDayAfterCancellation(t *testing.T) {
	harness := newHarness(t, &stubScheduleService{due: true, grace: 60}, "2026-06-15 03:00")
	harness.scheduler.tick()
	if _, err := harness.countdown.CancelToday(); err != nil {
		t.Fatalf("cancel failed: %v", err)
	}

	harness.now = mustTime(t, "2026-06-16 03:00")
	harness.scheduler.tick()
	if harness.countdown.Status().State != "warning" {
		t.Fatalf("expected warning on the next day, got %+v", harness.countdown.Status())
	}
	harness.advance(61 * time.Second)

	if harness.shutdownCalls != 1 {
		t.Fatalf("expected the next day to shut down, got %d", harness.shutdownCalls)
	}
}

func TestCancelTodayWhenIdleReturnsNotInWarningPeriod(t *testing.T) {
	_, err := NewCountdown().CancelToday()
	if !errors.Is(err, ErrNotInWarningPeriod) {
		t.Fatalf("expected ErrNotInWarningPeriod, got %v", err)
	}
}

func TestSchedulerWarnsOncePerDay(t *testing.T) {
	harness := newHarness(t, &stubScheduleService{due: true, grace: 30}, "2026-06-15 03:00")

	harness.scheduler.tick()
	harness.scheduler.tick()

	if harness.notifyCalls != 1 {
		t.Fatalf("expected one warning, got %d", harness.notifyCalls)
	}
}

func TestSchedulerSkipsWhenNotDue(t *testing.T) {
	harness := newHarness(t, &stubScheduleService{due: false}, "2026-06-15 12:00")

	harness.scheduler.tick()

	if harness.shutdownCalls != 0 || harness.countdown.Status().State != "idle" {
		t.Fatalf("expected idle when not due, got shutdown=%d", harness.shutdownCalls)
	}
}

func TestSchedulerHandlesServiceError(t *testing.T) {
	harness := newHarness(t, &stubScheduleService{err: errors.New("boom")}, "2026-06-15 03:00")

	harness.scheduler.tick()

	if harness.shutdownCalls != 0 || harness.countdown.Status().State != "idle" {
		t.Fatalf("expected idle on service error, got shutdown=%d", harness.shutdownCalls)
	}
}

func TestSchedulerStartStop(t *testing.T) {
	service := &stubScheduleService{due: false}
	scheduler := NewScheduler(service, NewCountdown(), func() error { return nil }, nil, 10*time.Millisecond)
	scheduler.Start()
	time.Sleep(25 * time.Millisecond)
	scheduler.Stop()

	if service.calls == 0 {
		t.Fatal("expected the running scheduler to evaluate at least once")
	}
}

func TestNewSchedulerDefaults(t *testing.T) {
	scheduler := NewScheduler(&stubScheduleService{}, NewCountdown(), nil, nil, 0)
	if scheduler.shutdownFn == nil {
		t.Fatal("expected a default shutdownFn")
	}
	if scheduler.interval != DefaultCheckInterval {
		t.Fatalf("expected default interval, got %v", scheduler.interval)
	}
}

func TestSchedulerRunShutsDownWhenWarningTimerEnds(t *testing.T) {
	shutdownDone := make(chan struct{})
	countdown := NewCountdown()
	scheduler := NewScheduler(&stubScheduleService{due: true, grace: 1}, countdown,
		func() error { close(shutdownDone); return nil }, nil, 10*time.Millisecond)
	scheduler.Start()
	defer scheduler.Stop()

	select {
	case <-shutdownDone:
	case <-time.After(3 * time.Second):
		t.Fatal("expected the dedicated warning timer to trigger the shutdown")
	}
}
