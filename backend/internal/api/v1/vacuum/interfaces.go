package vacuum

import (
	"context"

	"nas-go/api/pkg/tuya"
	"nas-go/api/pkg/tuyacloud"
)

type RepositoryInterface interface {
	GetSettingsDocument() (document string, isFound bool, err error)
	UpsertSettingsDocument(document string) error
	GetCloudSettingsDocument() (document string, isFound bool, err error)
	UpsertCloudSettingsDocument(document string) error
	GetRoomNamesDocument() (document string, isFound bool, err error)
	UpsertRoomNamesDocument(document string) error
}

type CloudMapClient interface {
	ReadMapLayout(ctx context.Context, credentials tuyacloud.Credentials, deviceID string) (tuyacloud.MapLayout, error)
}

type PushListenerClient interface {
	Listen(ctx context.Context, device tuya.Device, onDataPoints func(dataPoints map[string]any)) error
}

type DeviceClient interface {
	ReadDataPoints(ctx context.Context, device tuya.Device) (map[string]any, error)
	WriteDataPoints(ctx context.Context, device tuya.Device, dataPoints map[string]any) error
	FindDevice(ctx context.Context, device tuya.Device, candidateAddresses []string) (string, error)
}

type CandidateAddressesFunc func() ([]string, error)

type ServiceInterface interface {
	GetSettings() (SettingsDto, error)
	UpdateSettings(request UpdateSettingsDto) (SettingsDto, error)
	TestConnection(ctx context.Context) error
	Discover(ctx context.Context) (DiscoveryDto, error)
	GetStatus(ctx context.Context) (StatusDto, error)
	SendCommand(ctx context.Context, command string) error
	SendRemoteControl(ctx context.Context, direction string) error
	GetPreferences(ctx context.Context) (PreferencesDto, error)
	UpdatePreferences(ctx context.Context, patch UpdatePreferencesDto) (PreferencesDto, error)
	GetConsumables(ctx context.Context) (ConsumablesDto, error)
	ResetConsumable(ctx context.Context, part string) error
	GetTotals(ctx context.Context) (TotalsDto, error)
	GetCloudSettings() (CloudSettingsDto, error)
	UpdateCloudSettings(request UpdateCloudSettingsDto) (CloudSettingsDto, error)
	TestCloudConnection(ctx context.Context) error
	GetMap(ctx context.Context) (MapDto, error)
	GetRooms(ctx context.Context) (RoomsDto, error)
	UpdateRoom(ctx context.Context, roomID int, request UpdateRoomDto) (RoomDto, error)
	GetPath() PathDto
	CleanRooms(ctx context.Context, request RoomCleaningDto) error
	CleanSpot(ctx context.Context, request SpotCleaningDto) error
}
