package vacuum

import "time"

type SettingsDto struct {
	Address     string `json:"address"`
	DeviceID    string `json:"device_id"`
	HasLocalKey bool   `json:"has_local_key"`
}

type UpdateSettingsDto struct {
	Address  string `json:"address"`
	DeviceID string `json:"device_id"`
	LocalKey string `json:"local_key"`
}

type DiscoveryDto struct {
	Address string `json:"address"`
}

type StatusDto struct {
	State           string    `json:"state"`
	BatteryPercent  int       `json:"battery_percent"`
	CleaningSeconds int       `json:"cleaning_seconds"`
	CleaningArea    int       `json:"cleaning_area"`
	FaultCodes      []int     `json:"fault_codes"`
	ReadAt          time.Time `json:"read_at"`
}

type CommandDto struct {
	Command string `json:"command"`
}

type RemoteControlDto struct {
	Direction string `json:"direction"`
}

type PreferencesDto struct {
	Suction             *string `json:"suction"`
	WaterLevel          *string `json:"water_level"`
	Volume              *int    `json:"volume"`
	IsDoNotDisturb      *bool   `json:"is_do_not_disturb"`
	IsCarpetBoost       *bool   `json:"is_carpet_boost"`
	IsResumeAfterCharge *bool   `json:"is_resume_after_charge"`
	IsAutoEmpty         *bool   `json:"is_auto_empty"`
	AutoEmptyFrequency  *string `json:"auto_empty_frequency"`
}

type UpdatePreferencesDto = PreferencesDto

type ConsumableDto struct {
	Part       string `json:"part"`
	Seconds    int    `json:"seconds"`
	MaxSeconds int    `json:"max_seconds"`
}

type ConsumablesDto struct {
	Items []ConsumableDto `json:"items"`
}

type TotalsDto struct {
	CleaningSeconds int `json:"cleaning_seconds"`
	CleaningArea    int `json:"cleaning_area"`
	CleaningCount   int `json:"cleaning_count"`
}
