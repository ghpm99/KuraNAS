package backup

import (
	"errors"
	"testing"
	"time"

	"nas-go/api/internal/api/v1/trash"
	backupengine "nas-go/api/internal/worker/backup"
	"nas-go/api/pkg/database"
)

type fakeRepo struct {
	document   string
	hasDoc     bool
	upserted   []string
	pending    int
	lastRun    LastRunModel
	hasLastRun bool
	stamps     []string
	loadErr    error
}

func (f *fakeRepo) GetDbContext() *database.DbContext { return nil }
func (f *fakeRepo) GetSettingsDocument() (string, bool, error) {
	return f.document, f.hasDoc, f.loadErr
}
func (f *fakeRepo) UpsertSettingsDocument(document string) error {
	f.upserted = append(f.upserted, document)
	return nil
}
func (f *fakeRepo) CountPendingFiles() (int, error) { return f.pending, nil }
func (f *fakeRepo) GetLastRun() (LastRunModel, bool, error) {
	return f.lastRun, f.hasLastRun, nil
}
func (f *fakeRepo) StampLastBackup(path string, at time.Time) error {
	f.stamps = append(f.stamps, path)
	return nil
}

func newTestService(repo RepositoryInterface) *Service {
	service := NewService(repo)
	service.listRoots = func() []backupengine.Root {
		return []backupengine.Root{{Label: "Casa", Path: "/mnt/dados"}}
	}
	return service
}

func TestGetSettingsDefaultsWhenAbsent(t *testing.T) {
	service := newTestService(&fakeRepo{})

	settings, err := service.GetSettings()
	if err != nil {
		t.Fatalf("GetSettings: %v", err)
	}
	if settings.Enabled || settings.RetentionDays != 30 || settings.IntervalHours != 24 {
		t.Fatalf("unexpected defaults: %+v", settings)
	}
}

func TestUpdateSettingsValidatesDestination(t *testing.T) {
	repo := &fakeRepo{}
	service := newTestService(repo)

	cases := []SettingsDto{
		{Enabled: true, DestinationPath: ""},
		{Enabled: true, DestinationPath: "relativo"},
		{Enabled: true, DestinationPath: "/mnt/dados/backup"},
		{Enabled: false, DestinationPath: "/mnt/dados/backup"},
	}
	for _, dto := range cases {
		if _, err := service.UpdateSettings(dto); !errors.Is(err, ErrInvalidDestination) {
			t.Fatalf("dto %+v: expected ErrInvalidDestination, got %v", dto, err)
		}
	}
	if len(repo.upserted) != 0 {
		t.Fatal("invalid settings must not be persisted")
	}
}

func TestUpdateSettingsPersistsValidConfig(t *testing.T) {
	repo := &fakeRepo{}
	service := newTestService(repo)

	settings, err := service.UpdateSettings(SettingsDto{
		Enabled:         true,
		DestinationPath: "/mnt/backup",
	})
	if err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	if settings.RetentionDays != 30 || settings.IntervalHours != 24 {
		t.Fatalf("defaults not applied: %+v", settings)
	}
	if len(repo.upserted) != 1 {
		t.Fatalf("expected 1 upsert, got %d", len(repo.upserted))
	}

	// Disabled with empty destination is a valid way to turn the feature off.
	if _, err := service.UpdateSettings(SettingsDto{Enabled: false}); err != nil {
		t.Fatalf("disable: %v", err)
	}
}

func TestStatusAndPending(t *testing.T) {
	started := time.Now().Add(-time.Hour)
	repo := &fakeRepo{
		document:   `{"enabled":true,"destination_path":"/mnt/backup"}`,
		hasDoc:     true,
		pending:    12,
		hasLastRun: true,
		lastRun:    LastRunModel{JobID: 3, Status: "completed", CreatedAt: started, StartedAt: &started},
	}
	service := newTestService(repo)

	status, err := service.Status()
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if !status.Enabled || !status.HasRun || status.Status != "completed" || status.StartedAt == nil {
		t.Fatalf("unexpected status: %+v", status)
	}

	pending, err := service.Pending()
	if err != nil || pending.PendingFiles != 12 {
		t.Fatalf("unexpected pending: %+v %v", pending, err)
	}
}

func TestRunOptionsDisabledWithoutConfig(t *testing.T) {
	service := newTestService(&fakeRepo{})

	enabled, _, err := service.RunOptions()
	if err != nil || enabled {
		t.Fatalf("expected disabled, got enabled=%v err=%v", enabled, err)
	}
}

func TestRunOptionsBuildsEngineOptions(t *testing.T) {
	repo := &fakeRepo{
		document: `{"enabled":true,"destination_path":"/mnt/backup","retention_days":7}`,
		hasDoc:   true,
	}
	service := newTestService(repo)

	enabled, opts, err := service.RunOptions()
	if err != nil || !enabled {
		t.Fatalf("expected enabled, got %v %v", enabled, err)
	}
	if opts.Destination != "/mnt/backup" || opts.RetentionDays != 7 || len(opts.Roots) != 1 {
		t.Fatalf("unexpected options: %+v", opts)
	}
	if len(opts.SkipDirNames) != 1 || opts.SkipDirNames[0] != trash.DirName {
		t.Fatalf("trash dir must be excluded: %v", opts.SkipDirNames)
	}

	if err := opts.Stamp("/mnt/dados/a.txt", time.Now()); err != nil {
		t.Fatalf("stamp: %v", err)
	}
	if len(repo.stamps) != 1 || repo.stamps[0] != "/mnt/dados/a.txt" {
		t.Fatalf("stamp not wired to repository: %v", repo.stamps)
	}
}

func TestNextRunDue(t *testing.T) {
	now := time.Now()
	enabledDoc := `{"enabled":true,"destination_path":"/mnt/backup","interval_hours":24}`

	cases := []struct {
		name string
		repo *fakeRepo
		want bool
	}{
		{"disabled feature", &fakeRepo{}, false},
		{"never ran", &fakeRepo{document: enabledDoc, hasDoc: true}, true},
		{"run in flight", &fakeRepo{document: enabledDoc, hasDoc: true, hasLastRun: true,
			lastRun: LastRunModel{Status: "running", CreatedAt: now.Add(-48 * time.Hour)}}, false},
		{"recent run", &fakeRepo{document: enabledDoc, hasDoc: true, hasLastRun: true,
			lastRun: LastRunModel{Status: "completed", CreatedAt: now.Add(-time.Hour)}}, false},
		{"stale run", &fakeRepo{document: enabledDoc, hasDoc: true, hasLastRun: true,
			lastRun: LastRunModel{Status: "completed", CreatedAt: now.Add(-25 * time.Hour)}}, true},
	}

	for _, testCase := range cases {
		service := newTestService(testCase.repo)
		due, err := service.NextRunDue(now)
		if err != nil {
			t.Fatalf("%s: %v", testCase.name, err)
		}
		if due != testCase.want {
			t.Fatalf("%s: expected due=%v, got %v", testCase.name, testCase.want, due)
		}
	}
}
