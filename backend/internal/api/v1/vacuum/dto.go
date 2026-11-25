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

type CloudSettingsDto struct {
	AccessID        string `json:"access_id"`
	Region          string `json:"region"`
	HasAccessSecret bool   `json:"has_access_secret"`
}

type UpdateCloudSettingsDto struct {
	AccessID     string `json:"access_id"`
	AccessSecret string `json:"access_secret"`
	Region       string `json:"region"`
}

type PointDto struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type ForbiddenZoneDto struct {
	Vertices [][2]int `json:"vertices"`
}

type MapDto struct {
	MapID          int64              `json:"map_id"`
	Width          int                `json:"width"`
	Height         int                `json:"height"`
	Resolution     int                `json:"resolution"`
	OriginX        int                `json:"origin_x"`
	OriginY        int                `json:"origin_y"`
	Cells          string             `json:"cells"`
	Dock           *PointDto          `json:"dock"`
	ForbiddenZones []ForbiddenZoneDto `json:"forbidden_zones"`
	ReadAt         time.Time          `json:"read_at"`
}

type RoomDto struct {
	ID   int     `json:"id"`
	Name *string `json:"name"`
	Area float64 `json:"area"`
}

type RoomsDto struct {
	MapID int64     `json:"map_id"`
	Items []RoomDto `json:"items"`
}

type UpdateRoomDto struct {
	MapID int64  `json:"map_id"`
	Name  string `json:"name"`
}

type PathDto struct {
	PathID   *int64    `json:"path_id"`
	Points   [][2]int  `json:"points"`
	Position *PointDto `json:"position"`
}

type RoomCleaningDto struct {
	MapID   int64 `json:"map_id"`
	RoomIDs []int `json:"room_ids"`
}

type SpotCleaningDto struct {
	MapID int64   `json:"map_id"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
}
