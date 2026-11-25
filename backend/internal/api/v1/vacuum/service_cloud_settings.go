package vacuum

import (
	"context"
	"fmt"
	"strings"
)

func (s *Service) loadCloudSettings() (cloudSettingsState, error) {
	document, isFound, err := s.repository.GetCloudSettingsDocument()
	if err != nil {
		return cloudSettingsState{}, err
	}
	if !isFound {
		return cloudSettingsState{}, nil
	}
	return decodeCloudSettings(document)
}

func (s *Service) saveCloudSettings(settings cloudSettingsState) error {
	document, err := encodeCloudSettings(settings)
	if err != nil {
		return err
	}
	return s.repository.UpsertCloudSettingsDocument(document)
}

func (s *Service) GetCloudSettings() (CloudSettingsDto, error) {
	settings, err := s.loadCloudSettings()
	if err != nil {
		return CloudSettingsDto{}, err
	}
	return settings.toDto(), nil
}

func (s *Service) UpdateCloudSettings(request UpdateCloudSettingsDto) (CloudSettingsDto, error) {
	accessID := strings.TrimSpace(request.AccessID)
	region := strings.TrimSpace(request.Region)
	if err := validateCloudSettingsFields(accessID, request.AccessSecret, region); err != nil {
		return CloudSettingsDto{}, err
	}

	updated, err := s.replaceCloudSettings(accessID, request.AccessSecret, region)
	if err != nil {
		return CloudSettingsDto{}, err
	}
	s.invalidateMap()
	return updated.toDto(), nil
}

func (s *Service) replaceCloudSettings(accessID string, newAccessSecret string, region string) (cloudSettingsState, error) {
	s.cloudSettingsMutex.Lock()
	defer s.cloudSettingsMutex.Unlock()

	current, err := s.loadCloudSettings()
	if err != nil {
		return cloudSettingsState{}, err
	}

	accessSecret := newAccessSecret
	if accessSecret == "" {
		accessSecret = current.AccessSecret
	}
	if accessSecret == "" {
		return cloudSettingsState{}, fmt.Errorf("%w: no stored secret and none sent", ErrCloudAccessSecretRequired)
	}

	updated := cloudSettingsState{AccessID: accessID, AccessSecret: accessSecret, Region: region}
	if err := s.saveCloudSettings(updated); err != nil {
		return cloudSettingsState{}, err
	}
	return updated, nil
}

func (s *Service) TestCloudConnection(ctx context.Context) error {
	s.mapMutex.Lock()
	defer s.mapMutex.Unlock()

	_, err := s.fetchMap(ctx)
	return err
}
