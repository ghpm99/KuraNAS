package vacuum

import "encoding/json"

type settingsState struct {
	Address  string `json:"address"`
	DeviceID string `json:"device_id"`
	LocalKey string `json:"local_key"`
}

func (s settingsState) isConfigured() bool {
	return s.DeviceID != "" && s.LocalKey != ""
}

func (s settingsState) toDto() SettingsDto {
	return SettingsDto{
		Address:     s.Address,
		DeviceID:    s.DeviceID,
		HasLocalKey: s.LocalKey != "",
	}
}

func decodeSettings(document string) (settingsState, error) {
	var settings settingsState
	if err := json.Unmarshal([]byte(document), &settings); err != nil {
		return settingsState{}, err
	}
	return settings, nil
}

func encodeSettings(settings settingsState) (string, error) {
	document, err := json.Marshal(settings)
	if err != nil {
		return "", err
	}
	return string(document), nil
}
