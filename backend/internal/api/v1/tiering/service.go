package tiering

import (
	"errors"
	"log"
	"time"

	"nas-go/api/internal/roots"
	tieringengine "nas-go/api/internal/worker/tiering"
)

// ErrInvalidColdDir flags a cold directory that is empty, relative or inside an
// indexed root (which would make the scanner index the cold copies).
var ErrInvalidColdDir = errors.New("tiering: invalid cold directory")

type Service struct {
	repository RepositoryInterface
	// listRoots is injectable for tests; production reads the live registry.
	listRoots func() []roots.Root
}

func NewService(repository RepositoryInterface) *Service {
	return &Service{
		repository: repository,
		listRoots:  roots.Enabled,
	}
}

func (s *Service) rootPaths() []string {
	registered := s.listRoots()
	paths := make([]string, 0, len(registered))
	for _, root := range registered {
		paths = append(paths, root.Path)
	}
	return paths
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

	// A cold directory is validated whenever present — an invalid path must not
	// sit dormant in the document waiting for the toggle to flip.
	if settings.Enabled || settings.ColdDirPath != "" {
		if err := tieringengine.ValidateColdDir(settings.ColdDirPath, s.rootPaths()); err != nil {
			return SettingsDto{}, ErrInvalidColdDir
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

func (s *Service) Usage() (TierUsageDto, error) {
	counts, err := s.repository.GetTierCounts()
	if err != nil {
		return TierUsageDto{}, err
	}
	return counts.toDto(), nil
}

func (s *Service) SetPhysicalPath(fileID int, physicalPath string) error {
	return s.repository.SetPhysicalPath(fileID, physicalPath)
}

// MigrationPlan resolves the persisted settings into one pass of work. The
// cutoff is symmetric: hot files idle since the cutoff are demoted, cold files
// used again since the cutoff are promoted, so a file can never be in both.
func (s *Service) MigrationPlan(now time.Time) (bool, string, []tieringengine.Promotion, []tieringengine.Demotion, error) {
	settings, err := s.loadSettings()
	if err != nil {
		return false, "", nil, nil, err
	}
	if !settings.Enabled || settings.ColdDirPath == "" {
		return false, "", nil, nil, nil
	}
	if err := tieringengine.ValidateColdDir(settings.ColdDirPath, s.rootPaths()); err != nil {
		return false, "", nil, nil, err
	}

	cutoff := now.Add(-time.Duration(settings.MinAgeDays) * 24 * time.Hour)

	promotionCandidates, err := s.repository.ListPromotionCandidates(cutoff)
	if err != nil {
		return false, "", nil, nil, err
	}
	promotions := make([]tieringengine.Promotion, 0, len(promotionCandidates))
	for _, candidate := range promotionCandidates {
		promotions = append(promotions, tieringengine.Promotion{
			FileID:   candidate.FileID,
			HotPath:  candidate.LogicalPath,
			ColdPath: candidate.PhysicalPath,
		})
	}

	demotionCandidates, err := s.repository.ListDemotionCandidates(settings.MinSizeBytes, cutoff)
	if err != nil {
		return false, "", nil, nil, err
	}
	demotions := make([]tieringengine.Demotion, 0, len(demotionCandidates))
	for _, candidate := range demotionCandidates {
		coldPath, ok := s.coldPathFor(settings.ColdDirPath, candidate.LogicalPath)
		if !ok {
			continue
		}
		demotions = append(demotions, tieringengine.Demotion{
			FileID:   candidate.FileID,
			HotPath:  candidate.LogicalPath,
			ColdPath: coldPath,
		})
	}

	return true, settings.ColdDirPath, promotions, demotions, nil
}

// coldPathFor maps a logical path to its cold location, skipping (with a log)
// any file that no longer belongs to an enabled root — we never migrate bytes
// we cannot map back deterministically.
func (s *Service) coldPathFor(coldDir string, logicalPath string) (string, bool) {
	owner, found := roots.OwnerOf(logicalPath)
	if !found {
		log.Printf("[tiering] skipping %q: not under any enabled root\n", logicalPath)
		return "", false
	}
	coldPath, err := tieringengine.ColdPathFor(coldDir, owner.Label, owner.Path, logicalPath)
	if err != nil {
		log.Printf("[tiering] skipping %q: %v\n", logicalPath, err)
		return "", false
	}
	return coldPath, true
}

// NextRunDue tells the scheduler whether to enqueue a tier_migration now: the
// feature is on and configured, no run is in flight, and the last run is older
// than the configured interval.
func (s *Service) NextRunDue(now time.Time) (bool, error) {
	settings, err := s.loadSettings()
	if err != nil {
		return false, err
	}
	if !settings.Enabled || settings.ColdDirPath == "" {
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
