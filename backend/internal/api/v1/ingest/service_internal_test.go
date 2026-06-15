package ingest

import (
	"database/sql"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"

	jobs "nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/roots"
	"nas-go/api/pkg/database"
)

type fakeJobsRepo struct {
	createdJob  jobs.JobModel
	createdStep jobs.StepModel
	jobErr      error
	stepErr     error
}

func (r *fakeJobsRepo) GetDbContext() *database.DbContext { return database.NewDbContext(nil) }

func (r *fakeJobsRepo) CreateJob(tx *sql.Tx, job jobs.JobModel) (jobs.JobModel, error) {
	if r.jobErr != nil {
		return jobs.JobModel{}, r.jobErr
	}
	job.ID = 42
	r.createdJob = job
	return job, nil
}

func (r *fakeJobsRepo) CreateStep(tx *sql.Tx, step jobs.StepModel) (jobs.StepModel, error) {
	if r.stepErr != nil {
		return jobs.StepModel{}, r.stepErr
	}
	r.createdStep = step
	return step, nil
}

func seedRoot(t *testing.T) roots.Root {
	t.Helper()
	root := roots.Root{ID: 1, Path: filepath.Clean(t.TempDir()), Label: "Midia", Enabled: true}
	roots.Set([]roots.Root{root})
	t.Cleanup(roots.Reset)
	return root
}

func TestFetchSuccessEnqueuesJobAndStep(t *testing.T) {
	root := seedRoot(t)
	repo := &fakeJobsRepo{}
	service := NewService(repo)

	jobID, err := service.Fetch(FetchRequestDto{
		URL:        "https://youtu.be/abc123",
		Preset:     "audio_mp3",
		TargetRoot: root.Path,
		Subfolder:  "musicas",
	})
	if err != nil {
		t.Fatalf("Fetch returned error: %v", err)
	}
	if jobID != 42 {
		t.Fatalf("expected job id 42, got %d", jobID)
	}
	if repo.createdJob.Type != remoteFetchJobType {
		t.Fatalf("expected job type %q, got %q", remoteFetchJobType, repo.createdJob.Type)
	}
	if repo.createdStep.Type != remoteFetchStepType {
		t.Fatalf("expected step type %q, got %q", remoteFetchStepType, repo.createdStep.Type)
	}

	var payload RemoteFetchStepPayload
	if err := json.Unmarshal(repo.createdStep.Payload, &payload); err != nil {
		t.Fatalf("decode step payload: %v", err)
	}
	wantDir := filepath.Join(root.Path, "musicas")
	if payload.OutputDir != wantDir {
		t.Fatalf("expected output dir %q, got %q", wantDir, payload.OutputDir)
	}
	if payload.Binary == "" {
		t.Fatalf("expected a binary in the payload")
	}
}

func TestFetchValidationErrors(t *testing.T) {
	root := seedRoot(t)
	service := NewService(&fakeJobsRepo{})

	cases := []struct {
		name    string
		request FetchRequestDto
		wantErr error
	}{
		{"empty url", FetchRequestDto{URL: "", Preset: "audio_mp3", TargetRoot: root.Path}, ErrInvalidURL},
		{"non http url", FetchRequestDto{URL: "file:///etc/passwd", Preset: "audio_mp3", TargetRoot: root.Path}, ErrInvalidURL},
		{"unknown preset", FetchRequestDto{URL: "https://x.test/v", Preset: "nope", TargetRoot: root.Path}, ErrInvalidPreset},
		{"unknown target", FetchRequestDto{URL: "https://x.test/v", Preset: "audio_mp3", TargetRoot: "/not/a/root"}, ErrInvalidTarget},
		{"escaping subfolder", FetchRequestDto{URL: "https://x.test/v", Preset: "audio_mp3", TargetRoot: root.Path, Subfolder: "../../etc"}, ErrInvalidSubfolder},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := service.Fetch(tc.request); !errors.Is(err, tc.wantErr) {
				t.Fatalf("expected %v, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestFetchJobsUnavailable(t *testing.T) {
	root := seedRoot(t)
	service := NewService(nil)
	_, err := service.Fetch(FetchRequestDto{URL: "https://x.test/v", Preset: "audio_mp3", TargetRoot: root.Path})
	if !errors.Is(err, ErrJobsUnavailable) {
		t.Fatalf("expected ErrJobsUnavailable, got %v", err)
	}
}

func TestFetchPropagatesRepoError(t *testing.T) {
	root := seedRoot(t)
	service := NewService(&fakeJobsRepo{jobErr: errors.New("db down")})
	if _, err := service.Fetch(FetchRequestDto{URL: "https://x.test/v", Preset: "audio_mp3", TargetRoot: root.Path}); err == nil {
		t.Fatal("expected an error when CreateJob fails")
	}
}

func TestIsFetchableURL(t *testing.T) {
	cases := map[string]bool{
		"https://youtu.be/abc": true,
		"http://example.com/v": true,
		"ftp://host/file":      false,
		"file:///etc/passwd":   false,
		"not a url":            false,
		"":                     false,
		"https://":             false,
	}
	for raw, want := range cases {
		if got := isFetchableURL(raw); got != want {
			t.Errorf("isFetchableURL(%q) = %v, want %v", raw, got, want)
		}
	}
}

func TestResolveTarget(t *testing.T) {
	enabled := []roots.Root{{Path: "/srv/media", Label: "Media", Enabled: true}}

	if dir, err := resolveTarget(enabled, "/srv/media", ""); err != nil || dir != "/srv/media" {
		t.Fatalf("root only: got %q, %v", dir, err)
	}
	if dir, err := resolveTarget(enabled, "/srv/media", "anime/winter"); err != nil || dir != "/srv/media/anime/winter" {
		t.Fatalf("with subfolder: got %q, %v", dir, err)
	}
	if _, err := resolveTarget(enabled, "/other", ""); !errors.Is(err, ErrInvalidTarget) {
		t.Fatalf("expected ErrInvalidTarget, got %v", err)
	}
	if _, err := resolveTarget(enabled, "/srv/media", "../escape"); !errors.Is(err, ErrInvalidSubfolder) {
		t.Fatalf("expected ErrInvalidSubfolder, got %v", err)
	}
}

func TestListTargetsAndPresets(t *testing.T) {
	root := seedRoot(t)
	service := NewService(&fakeJobsRepo{})

	targets := service.ListTargets()
	if len(targets) != 1 || targets[0].Path != root.Path {
		t.Fatalf("unexpected targets: %+v", targets)
	}

	presets := service.ListPresets()
	if len(presets) == 0 {
		t.Fatal("expected at least one preset")
	}
	for _, p := range presets {
		if p.Key == "" || p.Label == "" {
			t.Fatalf("preset missing key/label: %+v", p)
		}
		if _, ok := ResolvePreset(p.Key); !ok {
			t.Fatalf("listed preset %q does not resolve", p.Key)
		}
	}
}

func TestResolvePresetUnknown(t *testing.T) {
	if _, ok := ResolvePreset("does_not_exist"); ok {
		t.Fatal("expected unknown preset to not resolve")
	}
}
