package autoshutdown

import (
	"errors"
	"testing"
	"time"

	"nas-go/api/pkg/database"
)

type stubRepository struct {
	document   string
	found      bool
	getErr     error
	upsertErr  error
	upserted   string
	median     float64
	sampleSize int
	medianErr  error
}

func (s *stubRepository) GetDbContext() *database.DbContext { return nil }

func (s *stubRepository) GetSettingsDocument() (string, bool, error) {
	return s.document, s.found, s.getErr
}

func (s *stubRepository) UpsertSettingsDocument(document string) error {
	s.upserted = document
	return s.upsertErr
}

func (s *stubRepository) GetShutdownTimeMedian() (float64, int, error) {
	return s.median, s.sampleSize, s.medianErr
}

func TestGetSettingsDefaultsWhenMissing(t *testing.T) {
	svc := NewService(&stubRepository{found: false})
	settings, err := svc.GetSettings()
	if err != nil {
		t.Fatalf("GetSettings: %v", err)
	}
	if settings.Enabled || settings.Time != defaultTime || settings.GracePeriodSeconds != defaultGracePeriodSeconds {
		t.Fatalf("unexpected defaults: %+v", settings)
	}
}

func TestUpdateSettingsValidNormalizes(t *testing.T) {
	stub := &stubRepository{}
	svc := NewService(stub)

	out, err := svc.UpdateSettings(SettingsDto{Enabled: true, Time: "7:5", GracePeriodSeconds: 30})
	if err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	if out.Time != "07:05" {
		t.Fatalf("expected normalized time 07:05, got %q", out.Time)
	}
	if stub.upserted == "" {
		t.Fatal("expected document to be persisted")
	}
}

func TestUpdateSettingsInvalid(t *testing.T) {
	svc := NewService(&stubRepository{})

	cases := []SettingsDto{
		{Time: "25:00", GracePeriodSeconds: 10},
		{Time: "12:60", GracePeriodSeconds: 10},
		{Time: "noon", GracePeriodSeconds: 10},
		{Time: "12:00", GracePeriodSeconds: -1},
		{Time: "12:00", GracePeriodSeconds: maxGracePeriodSeconds + 1},
	}
	for _, dto := range cases {
		if _, err := svc.UpdateSettings(dto); !errors.Is(err, ErrInvalidSettingsRequest) {
			t.Fatalf("dto %+v: expected ErrInvalidSettingsRequest, got %v", dto, err)
		}
	}
}

func TestSuggestedTimeAvailable(t *testing.T) {
	svc := NewService(&stubRepository{median: 10800, sampleSize: 4})
	out, err := svc.SuggestedTime()
	if err != nil {
		t.Fatalf("SuggestedTime: %v", err)
	}
	if !out.Available || out.Time != "03:00" || out.SampleSize != 4 {
		t.Fatalf("unexpected suggestion: %+v", out)
	}
}

func TestSuggestedTimeTooFewSamples(t *testing.T) {
	svc := NewService(&stubRepository{median: 10800, sampleSize: 2})
	out, err := svc.SuggestedTime()
	if err != nil {
		t.Fatalf("SuggestedTime: %v", err)
	}
	if out.Available {
		t.Fatalf("expected unavailable suggestion, got %+v", out)
	}
}

func TestDueNow(t *testing.T) {
	at := func(hhmm string) time.Time {
		parsed, _ := time.Parse("15:04", hhmm)
		return parsed
	}

	enabledDoc := `{"enabled":true,"time":"03:00","grace_period_seconds":45}`
	disabledDoc := `{"enabled":false,"time":"03:00","grace_period_seconds":45}`

	t.Run("fires at the configured minute", func(t *testing.T) {
		svc := NewService(&stubRepository{document: enabledDoc, found: true})
		due, grace, err := svc.DueNow(at("03:00"))
		if err != nil || !due || grace != 45 {
			t.Fatalf("expected due with grace 45, got due=%v grace=%v err=%v", due, grace, err)
		}
	})

	t.Run("does not fire at another minute", func(t *testing.T) {
		svc := NewService(&stubRepository{document: enabledDoc, found: true})
		if due, _, _ := svc.DueNow(at("03:01")); due {
			t.Fatal("must not fire outside the configured minute")
		}
	})

	t.Run("never fires when disabled", func(t *testing.T) {
		svc := NewService(&stubRepository{document: disabledDoc, found: true})
		if due, _, _ := svc.DueNow(at("03:00")); due {
			t.Fatal("must not fire when disabled")
		}
	})
}
