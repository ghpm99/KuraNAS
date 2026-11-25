package vacuum

import (
	"context"
	"errors"
	"fmt"

	"nas-go/api/pkg/tuyacloud"
)

func (s *Service) GetMap(ctx context.Context) (MapDto, error) {
	snapshot, err := s.currentMap(ctx)
	if err != nil {
		return MapDto{}, err
	}
	return newMapDto(snapshot), nil
}

func (s *Service) currentMap(ctx context.Context) (mapSnapshot, error) {
	s.mapMutex.Lock()
	defer s.mapMutex.Unlock()

	if cached := s.cachedMap(); cached.isFreshAt(s.now()) {
		return cached, nil
	}
	return s.fetchMap(ctx)
}

func (s *Service) mapWithID(ctx context.Context, mapID int64) (mapSnapshot, error) {
	snapshot, err := s.currentMap(ctx)
	if err != nil {
		return mapSnapshot{}, err
	}
	if snapshot.layout.MapID != mapID {
		return mapSnapshot{}, fmt.Errorf("%w: sent %d, robot has %d", ErrMapChanged, mapID, snapshot.layout.MapID)
	}
	return snapshot, nil
}

func (s *Service) fetchMap(ctx context.Context) (mapSnapshot, error) {
	cloudSettings, err := s.loadCloudSettings()
	if err != nil {
		return mapSnapshot{}, err
	}
	if !cloudSettings.isConfigured() {
		return mapSnapshot{}, ErrCloudNotConfigured
	}
	robotSettings, err := s.loadSettings()
	if err != nil {
		return mapSnapshot{}, err
	}
	if robotSettings.DeviceID == "" {
		return mapSnapshot{}, ErrNotConfigured
	}

	layout, err := s.cloudClient.ReadMapLayout(ctx, cloudSettings.toCredentials(), robotSettings.DeviceID)
	if err != nil {
		return mapSnapshot{}, translateCloudError(err)
	}

	snapshot := mapSnapshot{layout: layout, readAt: s.now(), isRead: true}
	s.storeMap(snapshot)
	return snapshot, nil
}

func (s *Service) cachedMap() mapSnapshot {
	s.sharedStateMutex.Lock()
	defer s.sharedStateMutex.Unlock()
	return s.mapCache
}

func (s *Service) storeMap(snapshot mapSnapshot) {
	s.sharedStateMutex.Lock()
	defer s.sharedStateMutex.Unlock()
	s.mapCache = snapshot
}

func (s *Service) invalidateMap() {
	s.storeMap(mapSnapshot{})
}

func translateCloudError(cloudErr error) error {
	switch {
	case errors.Is(cloudErr, tuyacloud.ErrServiceNotAuthorized):
		return fmt.Errorf("%w: %w", ErrCloudServiceNotAuthorized, cloudErr)
	case errors.Is(cloudErr, tuyacloud.ErrCredentialsRejected), errors.Is(cloudErr, tuyacloud.ErrUnknownRegion):
		return fmt.Errorf("%w: %w", ErrCloudCredentialsRejected, cloudErr)
	case errors.Is(cloudErr, tuyacloud.ErrInvalidResponse):
		return fmt.Errorf("%w: %w", ErrCloudInvalidResponse, cloudErr)
	case errors.Is(cloudErr, tuyacloud.ErrUnreachable),
		errors.Is(cloudErr, context.DeadlineExceeded),
		errors.Is(cloudErr, context.Canceled):
		return fmt.Errorf("%w: %w", ErrCloudUnreachable, cloudErr)
	default:
		return cloudErr
	}
}
