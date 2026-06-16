package autoshutdown

import (
	"errors"
	"fmt"
	"time"
)

// ErrInvalidSettingsRequest flags a request with a malformed time or an
// out-of-range grace period.
var ErrInvalidSettingsRequest = errors.New("autoshutdown: invalid settings request")

type Service struct {
	repository RepositoryInterface
}

func NewService(repository RepositoryInterface) *Service {
	return &Service{repository: repository}
}

func (s *Service) loadSettings() (settingsState, error) {
	document, found, err := s.repository.GetSettingsDocument()
	if err != nil {
		return settingsState{}, err
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
	normalizedTime, err := normalizeTime(dto.Time)
	if err != nil {
		return SettingsDto{}, fmt.Errorf("%w: time", ErrInvalidSettingsRequest)
	}
	if dto.GracePeriodSeconds < 0 || dto.GracePeriodSeconds > maxGracePeriodSeconds {
		return SettingsDto{}, fmt.Errorf("%w: grace_period_seconds", ErrInvalidSettingsRequest)
	}

	settings := dto.toState()
	settings.Time = normalizedTime
	settings = settings.withDefaults()

	document, err := encodeSettings(settings)
	if err != nil {
		return SettingsDto{}, err
	}
	if err := s.repository.UpsertSettingsDocument(document); err != nil {
		return SettingsDto{}, err
	}
	return settings.toDto(), nil
}

func (s *Service) SuggestedTime() (SuggestedTimeDto, error) {
	medianSeconds, sampleSize, err := s.repository.GetShutdownTimeMedian()
	if err != nil {
		return SuggestedTimeDto{}, err
	}
	if sampleSize < minSampleSize {
		return SuggestedTimeDto{Available: false, SampleSize: sampleSize}, nil
	}
	return SuggestedTimeDto{
		Available:  true,
		Time:       secondsToTime(medianSeconds),
		SampleSize: sampleSize,
	}, nil
}

func (s *Service) DueNow(now time.Time) (bool, int, error) {
	settings, err := s.loadSettings()
	if err != nil {
		return false, 0, err
	}
	if !settings.Enabled {
		return false, 0, nil
	}
	if now.Format("15:04") != settings.Time {
		return false, 0, nil
	}
	return true, settings.GracePeriodSeconds, nil
}
