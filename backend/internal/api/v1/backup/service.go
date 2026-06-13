package backup

import (
	"errors"
	"time"

	"nas-go/api/internal/api/v1/trash"
	"nas-go/api/internal/roots"
	backupengine "nas-go/api/internal/worker/backup"
)

// ErrInvalidDestination flags a backup destination that is empty, relative or
// inside an indexed root (which would make the backup index itself).
var ErrInvalidDestination = errors.New("backup: invalid destination")

type Service struct {
	repository RepositoryInterface
	// listRoots is injectable for tests; production reads the live registry.
	listRoots func() []backupengine.Root
}

func NewService(repository RepositoryInterface) *Service {
	return &Service{
		repository: repository,
		listRoots:  enabledRoots,
	}
}

func enabledRoots() []backupengine.Root {
	registered := roots.Enabled()
	list := make([]backupengine.Root, 0, len(registered))
	for _, root := range registered {
		list = append(list, backupengine.Root{Label: root.Label, Path: root.Path})
	}
	return list
}

func (s *Service) loadSettings() (SettingsModel, error) {
	document, found, err := s.repository.GetSettingsDocument()
	if err != nil {
		return SettingsModel{}, err
	}
	if !found {
		return defaultSettings(), nil
	}
	return decodeSettings(document)
}

func (s *Service) GetSettings() (SettingsDto, error) {
	settings, err := s.loadSettings()
	if err != nil {
		return SettingsDto{}, err
	}
	return settings.toDto(), nil
}

func (s *Service) UpdateSettings(dto SettingsDto) (SettingsDto, error) {
	settings := dto.toModel()

	// A destination is validated whenever present — an invalid path must not
	// sit dormant in the document waiting for the toggle to flip.
	if settings.Enabled || settings.DestinationPath != "" {
		if err := backupengine.ValidateDestination(settings.DestinationPath, s.listRoots()); err != nil {
			return SettingsDto{}, ErrInvalidDestination
		}
	}

	document, err := encodeSettings(settings)
	if err != nil {
		return SettingsDto{}, err
	}
	if err := s.repository.UpsertSettingsDocument(document); err != nil {
		return SettingsDto{}, err
	}
	return settings.toDto(), nil
}

func (s *Service) Status() (StatusDto, error) {
	settings, err := s.loadSettings()
	if err != nil {
		return StatusDto{}, err
	}

	status := StatusDto{Enabled: settings.Enabled}

	lastRun, found, err := s.repository.GetLastRun()
	if err != nil {
		return StatusDto{}, err
	}
	if found {
		status.HasRun = true
		status.Status = lastRun.Status
		status.StartedAt = lastRun.StartedAt
		status.EndedAt = lastRun.EndedAt
		status.LastError = lastRun.LastError
	}
	return status, nil
}

func (s *Service) Pending() (PendingDto, error) {
	pending, err := s.repository.CountPendingFiles()
	if err != nil {
		return PendingDto{}, err
	}
	return PendingDto{PendingFiles: pending}, nil
}

// RunOptions resolves the persisted settings into engine options for the
// backup_run step. The trash dir is always excluded from the copy.
func (s *Service) RunOptions() (bool, backupengine.Options, error) {
	settings, err := s.loadSettings()
	if err != nil {
		return false, backupengine.Options{}, err
	}
	if !settings.Enabled || settings.DestinationPath == "" {
		return false, backupengine.Options{}, nil
	}

	rootList := s.listRoots()
	if err := backupengine.ValidateDestination(settings.DestinationPath, rootList); err != nil {
		return false, backupengine.Options{}, err
	}

	return true, backupengine.Options{
		Roots:         rootList,
		Destination:   settings.DestinationPath,
		RetentionDays: settings.RetentionDays,
		SkipDirNames:  []string{trash.DirName},
		Stamp:         s.repository.StampLastBackup,
	}, nil
}

// NextRunDue tells the scheduler whether to enqueue a backup_run now: the
// feature is on, no run is in flight, and the last run is older than the
// configured interval.
func (s *Service) NextRunDue(now time.Time) (bool, error) {
	settings, err := s.loadSettings()
	if err != nil {
		return false, err
	}
	if !settings.Enabled || settings.DestinationPath == "" {
		return false, nil
	}

	lastRun, found, err := s.repository.GetLastRun()
	if err != nil {
		return false, err
	}
	if !found {
		return true, nil
	}
	if lastRun.Status == "queued" || lastRun.Status == "running" {
		return false, nil
	}

	interval := time.Duration(settings.IntervalHours) * time.Hour
	return now.Sub(lastRun.CreatedAt) >= interval, nil
}
