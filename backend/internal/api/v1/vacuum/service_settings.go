package vacuum

import (
	"fmt"
	"strings"
)

func (s *Service) loadSettings() (settingsState, error) {
	document, isFound, err := s.repository.GetSettingsDocument()
	if err != nil {
		return settingsState{}, err
	}
	if !isFound {
		return settingsState{}, nil
	}
	return decodeSettings(document)
}

func (s *Service) saveSettings(settings settingsState) error {
	document, err := encodeSettings(settings)
	if err != nil {
		return err
	}
	return s.repository.UpsertSettingsDocument(document)
}

func (s *Service) storeAddress(address string) error {
	s.settingsMutex.Lock()
	defer s.settingsMutex.Unlock()

	settings, err := s.loadSettings()
	if err != nil {
		return err
	}
	settings.Address = address
	return s.saveSettings(settings)
}

func (s *Service) GetSettings() (SettingsDto, error) {
	settings, err := s.loadSettings()
	if err != nil {
		return SettingsDto{}, err
	}
	return settings.toDto(), nil
}

func (s *Service) UpdateSettings(request UpdateSettingsDto) (SettingsDto, error) {
	address := strings.TrimSpace(request.Address)
	deviceID := strings.TrimSpace(request.DeviceID)
	if err := validateSettingsFields(address, deviceID, request.LocalKey); err != nil {
		return SettingsDto{}, err
	}

	previous, updated, err := s.replaceSettings(address, deviceID, request.LocalKey)
	if err != nil {
		return SettingsDto{}, err
	}
	s.invalidateSnapshot()
	if updated != previous {
		s.handleRobotChanged()
	}
	return updated.toDto(), nil
}

func (s *Service) SetRobotChangedHandler(handler func()) {
	s.robotChangedHandler = handler
}

func (s *Service) handleRobotChanged() {
	s.pathTracker.reset()
	s.invalidateMap()
	if s.robotChangedHandler != nil {
		s.robotChangedHandler()
	}
}

func (s *Service) replaceSettings(address string, deviceID string, newLocalKey string) (previous settingsState, updated settingsState, err error) {
	s.settingsMutex.Lock()
	defer s.settingsMutex.Unlock()

	current, err := s.loadSettings()
	if err != nil {
		return settingsState{}, settingsState{}, err
	}

	localKey := newLocalKey
	if localKey == "" {
		localKey = current.LocalKey
	}
	if localKey == "" {
		return settingsState{}, settingsState{}, fmt.Errorf("%w: no stored key and none sent", ErrLocalKeyRequired)
	}

	updated = settingsState{Address: address, DeviceID: deviceID, LocalKey: localKey}
	if err := s.saveSettings(updated); err != nil {
		return settingsState{}, settingsState{}, err
	}
	return current, updated, nil
}
