package app

import (
	"testing"

	ollamamgmt "nas-go/api/internal/api/v1/ollama"
	"nas-go/api/pkg/systemevent"
)

type recordingEventSpy struct {
	events  []systemevent.EventType
	failOn  systemevent.EventType
	failErr error
}

func (s *recordingEventSpy) RecordStartup() error  { return nil }
func (s *recordingEventSpy) RecordShutdown() error { return nil }
func (s *recordingEventSpy) RecordEvent(eventType systemevent.EventType, _ string) error {
	s.events = append(s.events, eventType)
	if eventType == s.failOn {
		return s.failErr
	}
	return nil
}

func TestApplyOllamaOutcomeRecordsEvents(t *testing.T) {
	cases := []struct {
		name    string
		outcome ollamamgmt.EnsureOutcome
		want    []systemevent.EventType
	}{
		{"started records started event", ollamamgmt.OutcomeStarted, []systemevent.EventType{systemevent.EventTypeOllamaDaemonStarted}},
		{"unreachable records down event", ollamamgmt.OutcomeUnreachable, []systemevent.EventType{systemevent.EventTypeOllamaDaemonDown}},
		{"already running records nothing", ollamamgmt.OutcomeAlreadyRunning, nil},
		{"binary missing records nothing", ollamamgmt.OutcomeBinaryMissing, nil},
		{"disabled records nothing", ollamamgmt.OutcomeDisabled, nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			spy := &recordingEventSpy{}
			applyOllamaOutcome(spy, tc.outcome)
			if len(spy.events) != len(tc.want) {
				t.Fatalf("expected events %v, got %v", tc.want, spy.events)
			}
			for i, want := range tc.want {
				if spy.events[i] != want {
					t.Fatalf("event %d: expected %q, got %q", i, want, spy.events[i])
				}
			}
		})
	}
}

func TestApplyOllamaOutcomeSurvivesRecordError(t *testing.T) {
	spy := &recordingEventSpy{failOn: systemevent.EventTypeOllamaDaemonStarted, failErr: errFake}
	// Must not panic even though RecordEvent fails.
	applyOllamaOutcome(spy, ollamamgmt.OutcomeStarted)
}

func TestRecordEventNilServiceNoPanic(t *testing.T) {
	recordEvent(nil, systemevent.EventTypeOllamaDaemonStarted, "x")
}

func TestStartOllamaDaemonNilGuards(t *testing.T) {
	// None of these should panic or block.
	startOllamaDaemon(nil, &recordingEventSpy{})
	startOllamaDaemon(&AppContext{}, &recordingEventSpy{})
	startOllamaDaemon(&AppContext{Ollama: &OllamaContext{}}, &recordingEventSpy{})
}

var errFake = fakeError("boom")

type fakeError string

func (e fakeError) Error() string { return string(e) }
