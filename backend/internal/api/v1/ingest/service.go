package ingest

import (
	"database/sql"
	"encoding/json"
	"net/url"
	"path/filepath"
	"strings"

	jobs "nas-go/api/internal/api/v1/jobs"
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

// MaxURLLength caps the URL accepted for download, in bytes.
const MaxURLLength = 2048

type Service struct {
	jobsRepo      jobEnqueuer
	downloadsRepo downloadRecorder
	ytDlp         ytDlpInstallProbe
}

func NewService(jobsRepo jobEnqueuer, downloadsRepo downloadRecorder, ytDlp ytDlpInstallProbe) *Service {
	return &Service{jobsRepo: jobsRepo, downloadsRepo: downloadsRepo, ytDlp: ytDlp}
}

// Fetch validates the request, records it as a queued download and enqueues a
// background yt-dlp job for it, all in one transaction. It returns the job id
// to track progress and the id of the recorded download.
func (s *Service) Fetch(request FetchRequestDto) (FetchResponseDto, error) {
	rawURL := strings.TrimSpace(request.URL)
	if len(rawURL) > MaxURLLength {
		return FetchResponseDto{}, ErrURLTooLong
	}
	if !isFetchableURL(rawURL) {
		return FetchResponseDto{}, ErrInvalidURL
	}
	if _, ok := ResolvePreset(request.Preset); !ok {
		return FetchResponseDto{}, ErrInvalidPreset
	}
	enabledRoots := roots.Enabled()
	outputDir, err := resolveTarget(enabledRoots, request.TargetRoot, request.Subfolder)
	if err != nil {
		return FetchResponseDto{}, err
	}
	if s.jobsRepo == nil || s.downloadsRepo == nil {
		return FetchResponseDto{}, ErrJobsUnavailable
	}
	if !s.ytDlp.IsInstalled() {
		return FetchResponseDto{}, ErrYtDlpNotInstalled
	}
	targetRoot := findRootByPath(enabledRoots, request.TargetRoot)
	subfolder, err := relativeSubfolder(targetRoot.Path, outputDir)
	if err != nil {
		return FetchResponseDto{}, err
	}

	scope, err := json.Marshal(map[string]any{"url": rawURL, "preset": request.Preset, "target": outputDir})
	if err != nil {
		return FetchResponseDto{}, err
	}

	var response FetchResponseDto
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
		downloadID, recordErr := s.downloadsRepo.CreateQueuedDownload(tx, QueuedDownload{
			JobID:          created.ID,
			URL:            rawURL,
			Preset:         request.Preset,
			TargetRootID:   targetRoot.ID,
			TargetRootName: targetRoot.Label,
			Subfolder:      subfolder,
		})
		if recordErr != nil {
			return recordErr
		}
		stepPayload, payloadErr := json.Marshal(RemoteFetchStepPayload{
			DownloadID: downloadID,
			URL:        rawURL,
			Preset:     request.Preset,
			OutputDir:  outputDir,
			Binary:     resolveYtDlpBinary(),
		})
		if payloadErr != nil {
			return payloadErr
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
		response = FetchResponseDto{JobID: created.ID, DownloadID: downloadID}
		return nil
	})
	if err != nil {
		return FetchResponseDto{}, err
	}
	return response, nil
}

func findRootByPath(enabled []roots.Root, rawPath string) roots.Root {
	cleanPath := filepath.Clean(strings.TrimSpace(rawPath))
	for _, root := range enabled {
		if root.Path == cleanPath {
			return root
		}
	}
	return roots.Root{}
}

func relativeSubfolder(rootPath, outputDir string) (string, error) {
	relative, err := filepath.Rel(rootPath, outputDir)
	if err != nil {
		return "", ErrInvalidSubfolder
	}
	if relative == "." {
		return "", nil
	}
	return relative, nil
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
	confined := filepath.Clean(filepath.Join(matched, sub))
	if confined != matched && !strings.HasPrefix(confined, matched+string(filepath.Separator)) {
		return "", ErrInvalidSubfolder
	}
	return confined, nil
}

func targetsFrom(enabled []roots.Root) []TargetDto {
	out := make([]TargetDto, 0, len(enabled))
	for _, root := range enabled {
		out = append(out, TargetDto{Label: root.Label, Path: root.Path})
	}
	return out
}
