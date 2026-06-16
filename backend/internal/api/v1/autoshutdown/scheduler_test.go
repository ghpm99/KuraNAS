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

func newTestScheduler(service SchedulerInterface) (*Scheduler, *int, *int) {
	shutdownCalls := 0
	notifyCalls := 0
	scheduler := NewScheduler(service,
		func(graceSeconds int) error { shutdownCalls++; return nil },
		func(graceSeconds int) { notifyCalls++ },
		time.Minute,
	)
	return scheduler, &shutdownCalls, &notifyCalls
}

func TestSchedulerFiresOncePerDay(t *testing.T) {
	service := &stubScheduleService{due: true, grace: 30}
	scheduler, shutdownCalls, notifyCalls := newTestScheduler(service)

	scheduler.nowFn = func() time.Time { return mustTime(t, "2026-06-15 03:00") }
	scheduler.tick()
	scheduler.tick() // same minute / day: must not fire again

	if *shutdownCalls != 1 || *notifyCalls != 1 {
		t.Fatalf("expected exactly one shutdown+notify, got shutdown=%d notify=%d", *shutdownCalls, *notifyCalls)
	}
}

func TestSchedulerFiresAgainNextDay(t *testing.T) {
	service := &stubScheduleService{due: true, grace: 30}
	scheduler, shutdownCalls, _ := newTestScheduler(service)

	scheduler.nowFn = func() time.Time { return mustTime(t, "2026-06-15 03:00") }
	scheduler.tick()
	scheduler.nowFn = func() time.Time { return mustTime(t, "2026-06-16 03:00") }
	scheduler.tick()

	if *shutdownCalls != 2 {
		t.Fatalf("expected two shutdowns across two days, got %d", *shutdownCalls)
	}
}

func TestSchedulerSkipsWhenNotDue(t *testing.T) {
	service := &stubScheduleService{due: false}
	scheduler, shutdownCalls, _ := newTestScheduler(service)

	scheduler.nowFn = func() time.Time { return mustTime(t, "2026-06-15 12:00") }
	scheduler.tick()

	if *shutdownCalls != 0 {
		t.Fatalf("expected no shutdown when not due, got %d", *shutdownCalls)
	}
}

func TestSchedulerHandlesServiceError(t *testing.T) {
	service := &stubScheduleService{err: errors.New("boom")}
	scheduler, shutdownCalls, _ := newTestScheduler(service)

	scheduler.nowFn = func() time.Time { return mustTime(t, "2026-06-15 03:00") }
	scheduler.tick()

	if *shutdownCalls != 0 {
		t.Fatalf("expected no shutdown on service error, got %d", *shutdownCalls)
	}
}

func TestSchedulerStartStop(t *testing.T) {
	service := &stubScheduleService{due: false}
	scheduler := NewScheduler(service, func(int) error { return nil }, nil, 10*time.Millisecond)
	scheduler.Start()
	time.Sleep(25 * time.Millisecond)
	scheduler.Stop()

	if service.calls == 0 {
		t.Fatal("expected the running scheduler to evaluate at least once")
	}
}

func TestNewSchedulerDefaults(t *testing.T) {
	scheduler := NewScheduler(&stubScheduleService{}, nil, nil, 0)
	if scheduler.shutdownFn == nil {
		t.Fatal("expected a default shutdownFn")
	}
	if scheduler.interval != DefaultCheckInterval {
		t.Fatalf("expected default interval, got %v", scheduler.interval)
	}
}
