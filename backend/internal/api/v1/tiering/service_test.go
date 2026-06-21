package tiering

import (
	"errors"
	"testing"
	"time"

	"nas-go/api/internal/roots"
	"nas-go/api/pkg/database"
)

type fakeRepo struct {
	document   string
	hasDoc     bool
	upserted   []string
	demotions  []CandidateModel
	promotions []CandidateModel
	setCalls   []struct {
		id   int
		path string
	}
	lastRun    LastRunModel
	hasLastRun bool
	counts     TierCountsModel
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
func (f *fakeRepo) ListDemotionCandidates(minSizeBytes int64, idleBefore time.Time) ([]CandidateModel, error) {
	return f.demotions, nil
}
func (f *fakeRepo) ListPromotionCandidates(usedAfter time.Time) ([]CandidateModel, error) {
	return f.promotions, nil
}
func (f *fakeRepo) SetPhysicalPath(fileID int, physicalPath string) error {
	f.setCalls = append(f.setCalls, struct {
		id   int
		path string
	}{fileID, physicalPath})
	return nil
}
func (f *fakeRepo) GetLastRun() (LastRunModel, bool, error) { return f.lastRun, f.hasLastRun, nil }
func (f *fakeRepo) GetTierCounts() (TierCountsModel, error) { return f.counts, nil }

func newTestService(repo RepositoryInterface) *Service {
	service := NewService(repo)
	service.listRoots = func() []roots.Root {
		return []roots.Root{{Label: "Casa", Path: "/mnt/dados", Enabled: true}}
	}
	return service
}

func TestGetSettingsDefaultsWhenAbsent(t *testing.T) {
	service := newTestService(&fakeRepo{})

	settings, err := service.GetSettings()
	if err != nil {
		t.Fatalf("GetSettings: %v", err)
	}
	if settings.Enabled || settings.MinAgeDays != 90 || settings.IntervalHours != 24 || settings.MinSizeBytes != 1<<20 {
		t.Fatalf("unexpected defaults: %+v", settings)
	}
}

func TestUpdateSettingsValidatesColdDir(t *testing.T) {
	repo := &fakeRepo{}
	service := newTestService(repo)

	cases := []SettingsDto{
		{Enabled: true, ColdDirPath: ""},
		{Enabled: true, ColdDirPath: "relativo"},
		{Enabled: true, ColdDirPath: "/mnt/dados/cold"},
		{Enabled: false, ColdDirPath: "/mnt/dados/cold"},
	}
	for _, dto := range cases {
		if _, err := service.UpdateSettings(dto); !errors.Is(err, ErrInvalidColdDir) {
			t.Fatalf("dto %+v: expected ErrInvalidColdDir, got %v", dto, err)
		}
	}
	if len(repo.upserted) != 0 {
		t.Fatal("invalid settings must not be persisted")
	}
}

func TestUpdateSettingsPersistsValidConfig(t *testing.T) {
	repo := &fakeRepo{}
	service := newTestService(repo)

	settings, err := service.UpdateSettings(SettingsDto{Enabled: true, ColdDirPath: "/mnt/cold"})
	if err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	if settings.MinAgeDays != 90 || settings.IntervalHours != 24 {
		t.Fatalf("defaults not applied: %+v", settings)
	}
	if len(repo.upserted) != 1 {
		t.Fatalf("expected 1 upsert, got %d", len(repo.upserted))
	}
	if _, err := service.UpdateSettings(SettingsDto{Enabled: false}); err != nil {
		t.Fatalf("disable: %v", err)
	}
}

func TestStatusAndUsage(t *testing.T) {
	started := time.Now().Add(-time.Hour)
	repo := &fakeRepo{
		document:   `{"enabled":true,"cold_dir_path":"/mnt/cold"}`,
		hasDoc:     true,
		hasLastRun: true,
		lastRun:    LastRunModel{JobID: 3, Status: "completed", CreatedAt: started, StartedAt: &started},
		counts:     TierCountsModel{HotFiles: 10, HotBytes: 1000, ColdFiles: 4, ColdBytes: 400},
	}
	service := newTestService(repo)

	status, err := service.Status()
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if !status.Enabled || !status.HasRun || status.Status != "completed" || status.StartedAt == nil {
		t.Fatalf("unexpected status: %+v", status)
	}

	usage, err := service.Usage()
	if err != nil || usage.HotFiles != 10 || usage.ColdBytes != 400 {
		t.Fatalf("unexpected usage: %+v %v", usage, err)
	}
}

func TestMigrationPlanDisabledWithoutConfig(t *testing.T) {
	service := newTestService(&fakeRepo{})

	enabled, _, _, _, err := service.MigrationPlan(time.Now())
	if err != nil || enabled {
		t.Fatalf("expected disabled, got enabled=%v err=%v", enabled, err)
	}
}

func TestMigrationPlanBuildsColdPaths(t *testing.T) {
	roots.Set([]roots.Root{{Label: "Casa", Path: "/mnt/dados", Enabled: true}})
	defer roots.Reset()

	repo := &fakeRepo{
		document: `{"enabled":true,"cold_dir_path":"/mnt/cold","min_age_days":30}`,
		hasDoc:   true,
		demotions: []CandidateModel{
			{FileID: 1, LogicalPath: "/mnt/dados/Documentos/a.txt", Size: 5 << 20},
			{FileID: 2, LogicalPath: "/elsewhere/b.txt", Size: 9 << 20}, // not under a root: skipped
		},
		promotions: []CandidateModel{
			{FileID: 7, LogicalPath: "/mnt/dados/c.txt", PhysicalPath: "/mnt/cold/Casa/c.txt"},
		},
	}
	service := newTestService(repo)

	enabled, coldDir, promotions, demotions, err := service.MigrationPlan(time.Now())
	if err != nil || !enabled {
		t.Fatalf("expected enabled plan, got %v %v", enabled, err)
	}
	if coldDir != "/mnt/cold" {
		t.Fatalf("unexpected coldDir %q", coldDir)
	}
	if len(demotions) != 1 || demotions[0].FileID != 1 {
		t.Fatalf("file outside a root must be skipped: %+v", demotions)
	}
	if demotions[0].ColdPath != "/mnt/cold/Casa/Documentos/a.txt" {
		t.Fatalf("unexpected cold path %q", demotions[0].ColdPath)
	}
	if len(promotions) != 1 || promotions[0].ColdPath != "/mnt/cold/Casa/c.txt" {
		t.Fatalf("unexpected promotions: %+v", promotions)
	}
}

func TestSetPhysicalPathDelegates(t *testing.T) {
	repo := &fakeRepo{}
	service := newTestService(repo)

	if err := service.SetPhysicalPath(5, "/mnt/cold/x"); err != nil {
		t.Fatalf("SetPhysicalPath: %v", err)
	}
	if len(repo.setCalls) != 1 || repo.setCalls[0].id != 5 || repo.setCalls[0].path != "/mnt/cold/x" {
		t.Fatalf("not delegated: %+v", repo.setCalls)
	}
}

func TestNextRunDue(t *testing.T) {
	now := time.Now()
	enabledDoc := `{"enabled":true,"cold_dir_path":"/mnt/cold","interval_hours":24}`

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
