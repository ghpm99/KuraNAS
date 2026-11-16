package vacuum

import (
	"context"
	"errors"
	"fmt"

	"nas-go/api/pkg/tuya"
)

func (s *Service) Discover(ctx context.Context) (DiscoveryDto, error) {
	s.deviceMutex.Lock()
	defer s.deviceMutex.Unlock()

	settings, err := s.loadSettings()
	if err != nil {
		return DiscoveryDto{}, err
	}
	if !settings.isConfigured() {
		return DiscoveryDto{}, ErrNotConfigured
	}

	address, err := s.searchAndStoreAddress(ctx, settings)
	if err != nil {
		return DiscoveryDto{}, err
	}
	return DiscoveryDto{Address: address}, nil
}

func (s *Service) TestConnection(ctx context.Context) error {
	s.deviceMutex.Lock()
	defer s.deviceMutex.Unlock()

	return s.runOnDevice(ctx, func(device tuya.Device) error {
		_, readErr := s.deviceClient.ReadDataPoints(ctx, device)
		return readErr
	})
}

func (s *Service) canSearchAgain() bool {
	s.sharedStateMutex.Lock()
	defer s.sharedStateMutex.Unlock()
	return s.lastDiscoveryAt.IsZero() || s.now().Sub(s.lastDiscoveryAt) > minDiscoveryInterval
}

func (s *Service) markSearchStarted() {
	s.sharedStateMutex.Lock()
	defer s.sharedStateMutex.Unlock()
	s.lastDiscoveryAt = s.now()
}

func (s *Service) searchAndStoreAddress(ctx context.Context, settings settingsState) (string, error) {
	s.markSearchStarted()

	searchContext, cancelSearch := context.WithTimeout(ctx, discoveryBudget)
	defer cancelSearch()

	candidates, err := s.candidateAddresses()
	if err != nil {
		return "", fmt.Errorf("listing candidate addresses: %w", err)
	}

	address, err := s.deviceClient.FindDevice(searchContext, settings.toDevice(), candidates)
	switch {
	case errors.Is(err, tuya.ErrDeviceNotFound):
		return "", fmt.Errorf("%w: %w", ErrDiscoveryNotFound, err)
	case errors.Is(err, tuya.ErrInvalidDevice):
		return "", fmt.Errorf("%w: %w", ErrNotConfigured, err)
	case err != nil:
		return "", err
	}

	if err := s.storeAddress(address); err != nil {
		return "", err
	}
	return address, nil
}
