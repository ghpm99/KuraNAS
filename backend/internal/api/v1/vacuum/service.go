package vacuum

import (
	"sync"
	"time"

	"nas-go/api/pkg/tuya"
)

const (
	snapshotLifetime     = 2 * time.Second
	discoveryBudget      = 5 * time.Second
	minDiscoveryInterval = time.Minute
)

type dataPointSnapshot struct {
	values map[string]any
	readAt time.Time
}

func (snapshot dataPointSnapshot) isFreshAt(instant time.Time) bool {
	return snapshot.values != nil && instant.Sub(snapshot.readAt) < snapshotLifetime
}

type Service struct {
	repository         RepositoryInterface
	deviceClient       DeviceClient
	candidateAddresses CandidateAddressesFunc
	now                func() time.Time

	settingsMutex sync.Mutex
	deviceMutex   sync.Mutex

	sharedStateMutex sync.Mutex
	snapshot         dataPointSnapshot
	lastDiscoveryAt  time.Time
}

func NewService(repository RepositoryInterface, deviceClient DeviceClient, candidateAddresses CandidateAddressesFunc) *Service {
	return &Service{
		repository:         repository,
		deviceClient:       deviceClient,
		candidateAddresses: candidateAddresses,
		now:                time.Now,
	}
}

func (s *Service) cachedSnapshot() dataPointSnapshot {
	s.sharedStateMutex.Lock()
	defer s.sharedStateMutex.Unlock()
	return s.snapshot
}

func (s *Service) storeSnapshot(snapshot dataPointSnapshot) {
	s.sharedStateMutex.Lock()
	defer s.sharedStateMutex.Unlock()
	s.snapshot = snapshot
}

func (s *Service) invalidateSnapshot() {
	s.storeSnapshot(dataPointSnapshot{})
}

func (settings settingsState) toDevice() tuya.Device {
	return tuya.Device{
		Address:  settings.Address,
		DeviceID: settings.DeviceID,
		LocalKey: settings.LocalKey,
	}
}
