package vacuum

import (
	"context"
	"errors"
	"fmt"

	"nas-go/api/pkg/tuya"
)

func (s *Service) readSnapshot(ctx context.Context) (dataPointSnapshot, error) {
	s.deviceMutex.Lock()
	defer s.deviceMutex.Unlock()

	if cached := s.cachedSnapshot(); cached.isFreshAt(s.now()) {
		return cached, nil
	}

	var dataPoints map[string]any
	err := s.runOnDevice(ctx, func(device tuya.Device) error {
		var readErr error
		dataPoints, readErr = s.deviceClient.ReadDataPoints(ctx, device)
		return readErr
	})
	if err != nil {
		return dataPointSnapshot{}, err
	}

	snapshot := dataPointSnapshot{values: dataPoints, readAt: s.now()}
	s.storeSnapshot(snapshot)
	return snapshot, nil
}

func (s *Service) writeDataPoints(ctx context.Context, dataPoints map[string]any) error {
	s.deviceMutex.Lock()
	defer s.deviceMutex.Unlock()
	defer s.invalidateSnapshot()

	return s.runOnDevice(ctx, func(device tuya.Device) error {
		return s.deviceClient.WriteDataPoints(ctx, device, dataPoints)
	})
}

func (s *Service) runOnDevice(ctx context.Context, operation func(device tuya.Device) error) error {
	settings, err := s.loadSettings()
	if err != nil {
		return err
	}
	if !settings.isConfigured() {
		return ErrNotConfigured
	}
	if settings.Address == "" {
		emptyAddressFailure := fmt.Errorf("%w: no address stored", tuya.ErrUnreachable)
		return s.runAfterRediscovery(ctx, settings, emptyAddressFailure, operation)
	}

	operationErr := operation(settings.toDevice())
	if !errors.Is(operationErr, tuya.ErrUnreachable) {
		return translateDeviceError(operationErr)
	}
	return s.runAfterRediscovery(ctx, settings, operationErr, operation)
}

func (s *Service) runAfterRediscovery(
	ctx context.Context,
	settings settingsState,
	unreachableFailure error,
	operation func(device tuya.Device) error,
) error {
	if !s.canSearchAgain() {
		return translateDeviceError(unreachableFailure)
	}

	address, err := s.searchAndStoreAddress(ctx, settings)
	if errors.Is(err, ErrDiscoveryNotFound) {
		return translateDeviceError(unreachableFailure)
	}
	if err != nil {
		return err
	}

	settings.Address = address
	return translateDeviceError(operation(settings.toDevice()))
}

func translateDeviceError(deviceErr error) error {
	switch {
	case deviceErr == nil:
		return nil
	case errors.Is(deviceErr, tuya.ErrKeyRejected):
		return fmt.Errorf("%w: %w", ErrLocalKeyRejected, deviceErr)
	case errors.Is(deviceErr, tuya.ErrInvalidDevice):
		return fmt.Errorf("%w: %w", ErrNotConfigured, deviceErr)
	case errors.Is(deviceErr, tuya.ErrUnreachable),
		errors.Is(deviceErr, context.DeadlineExceeded),
		errors.Is(deviceErr, context.Canceled):
		return fmt.Errorf("%w: %w", ErrDeviceUnreachable, deviceErr)
	default:
		return deviceErr
	}
}
