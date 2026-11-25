package vacuum

import (
	"encoding/json"

	"nas-go/api/pkg/tuyacloud"
)

type cloudSettingsState struct {
	AccessID     string `json:"access_id"`
	AccessSecret string `json:"access_secret"`
	Region       string `json:"region"`
}

func (s cloudSettingsState) isConfigured() bool {
	return s.AccessID != "" && s.AccessSecret != "" && s.Region != ""
}

func (s cloudSettingsState) toDto() CloudSettingsDto {
	return CloudSettingsDto{
		AccessID:        s.AccessID,
		Region:          s.Region,
		HasAccessSecret: s.AccessSecret != "",
	}
}

func (s cloudSettingsState) toCredentials() tuyacloud.Credentials {
	return tuyacloud.Credentials{AccessID: s.AccessID, AccessSecret: s.AccessSecret, Region: s.Region}
}

func decodeCloudSettings(document string) (cloudSettingsState, error) {
	var settings cloudSettingsState
	if err := json.Unmarshal([]byte(document), &settings); err != nil {
		return cloudSettingsState{}, err
	}
	return settings, nil
}

func encodeCloudSettings(settings cloudSettingsState) (string, error) {
	document, err := json.Marshal(settings)
	if err != nil {
		return "", err
	}
	return string(document), nil
}
