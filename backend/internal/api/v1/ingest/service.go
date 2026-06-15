package ingest

import (
	"database/sql"
	"encoding/json"
	"net/url"
	"path/filepath"
	"strings"

	jobs "nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/config"
	"nas-go/api/internal/roots"
	"nas-go/api/internal/worker/job"
	"nas-go/api/pkg/database"
)

// remoteFetchStepType mirrors job.StepTypeRemoteFetch as a literal so the
// service does not import the worker engine (which would create a cycle).
const (
	remoteFetchJobType  = "remote_fetch"
	remoteFetchStepType = "remote_fetch"
)

type Service struct {
	jobsRepo jobEnqueuer
}

func NewService(jobsRepo jobEnqueuer) *Service {
	return &Service{jobsRepo: jobsRepo}
}

// Fetch validates the request and enqueues a background yt-dlp download. It
// returns the job id so the caller can track progress through the jobs API.
func (s *Service) Fetch(request FetchRequestDto) (int, error) {
	rawURL := strings.TrimSpace(request.URL)
	if !isFetchableURL(rawURL) {
		return 0, ErrInvalidURL
	}
	if _, ok := ResolvePreset(request.Preset); !ok {
		return 0, ErrInvalidPreset
	}
	outputDir, err := resolveTarget(roots.Enabled(), request.TargetRoot, request.Subfolder)
	if err != nil {
		return 0, err
	}
	if s.jobsRepo == nil {
		return 0, ErrJobsUnavailable
	}

	binary := strings.TrimSpace(config.AppConfig.YtDlpPath)
	if binary == "" {
		binary = "yt-dlp"
	}

	stepPayload, err := json.Marshal(RemoteFetchStepPayload{
		URL:       rawURL,
		Preset:    request.Preset,
		OutputDir: outputDir,
		Binary:    binary,
	})
	if err != nil {
		return 0, err
	}
	scope, err := json.Marshal(map[string]any{"url": rawURL, "preset": request.Preset, "target": outputDir})
	if err != nil {
		return 0, err
	}

	var jobID int
	err = database.ExecOptionalTx(s.jobsRepo.GetDbContext(), func(tx *sql.Tx) error {
		created, createErr := s.jobsRepo.CreateJob(tx, jobs.JobModel{
			Type:            remoteFetchJobType,
			Priority:        string(job.JobPriorityNormal),
			Scope:           scope,
			Status:          string(job.JobStatusQueued),
			CancelRequested: false,
		})
		if createErr != nil {
			return createErr
		}
		_, stepErr := s.jobsRepo.CreateStep(tx, jobs.StepModel{
			JobID:       created.ID,
			Type:        remoteFetchStepType,
			Status:      string(job.StepStatusQueued),
			DependsOn:   []byte("[]"),
			Attempts:    0,
			MaxAttempts: 1,
			Progress:    0,
			Payload:     stepPayload,
		})
		if stepErr != nil {
			return stepErr
		}
		jobID = created.ID
		return nil
	})
	if err != nil {
		return 0, err
	}
	return jobID, nil
}

// ListTargets exposes the enabled storage roots as save destinations.
func (s *Service) ListTargets() []TargetDto {
	return targetsFrom(roots.Enabled())
}

// ListPresets exposes the selectable download presets.
func (s *Service) ListPresets() []PresetDto {
	return availablePresets()
}

// isFetchableURL accepts only absolute http(s) URLs. Rejecting other schemes
// keeps file://, ftp:// and shell-looking input from ever reaching yt-dlp.
func isFetchableURL(raw string) bool {
	if raw == "" {
		return false
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return false
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false
	}
	return parsed.Host != ""
}

// resolveTarget validates that targetRoot is one of the enabled roots and joins
// the (sanitized) subfolder under it, refusing any path that escapes the root.
func resolveTarget(enabled []roots.Root, targetRoot, subfolder string) (string, error) {
	cleanRoot := filepath.Clean(strings.TrimSpace(targetRoot))
	var matched string
	for _, root := range enabled {
		if root.Path == cleanRoot {
			matched = root.Path
			break
		}
	}
	if matched == "" {
		return "", ErrInvalidTarget
	}

	sub := strings.TrimSpace(subfolder)
	if sub == "" {
		return matched, nil
	}
	// Clean and confine: the joined path must stay inside the root.
	joined := filepath.Clean(filepath.Join(matched, sub))
	if joined != matched && !strings.HasPrefix(joined, matched+string(filepath.Separator)) {
		return "", ErrInvalidSubfolder
	}
	return joined, nil
}

func targetsFrom(enabled []roots.Root) []TargetDto {
	out := make([]TargetDto, 0, len(enabled))
	for _, root := range enabled {
		out = append(out, TargetDto{Label: root.Label, Path: root.Path})
	}
	return out
}
