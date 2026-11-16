package vacuum

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"nas-go/api/pkg/tuya"
)

func unreachableFailure() error {
	return fmt.Errorf("%w: dial failed", tuya.ErrUnreachable)
}

func TestSnapshotIsReusedWithinTwoSeconds(t *testing.T) {
	fixture := newServiceFixture(t)
	fixture.client.readFn = func(tuya.Device) (map[string]any, error) {
		return map[string]any{dpBatteryPercent: float64(80)}, nil
	}

	for range 5 {
		if _, err := fixture.service.GetStatus(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := fixture.service.GetTotals(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(fixture.client.readDevices) != 1 {
		t.Fatalf("expected a single device read, got %d", len(fixture.client.readDevices))
	}

	fixture.clock.advance(2 * time.Second)
	if _, err := fixture.service.GetStatus(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(fixture.client.readDevices) != 2 {
		t.Fatalf("expected a new read after the snapshot expired, got %d", len(fixture.client.readDevices))
	}
}

func TestSnapshotIsInvalidatedByWrite(t *testing.T) {
	fixture := newServiceFixture(t)

	if _, err := fixture.service.GetStatus(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := fixture.service.SendCommand(context.Background(), "locate"); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.service.GetStatus(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(fixture.client.readDevices) != 2 {
		t.Fatalf("expected a fresh read after a write, got %d reads", len(fixture.client.readDevices))
	}
}

func TestSettingsUpdateInvalidatesSnapshot(t *testing.T) {
	fixture := newServiceFixture(t)

	if _, err := fixture.service.GetStatus(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.service.UpdateSettings(UpdateSettingsDto{Address: testAddress, DeviceID: "other"}); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.service.GetStatus(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(fixture.client.readDevices) != 2 {
		t.Fatalf("expected a fresh read after settings changed, got %d", len(fixture.client.readDevices))
	}
}

func TestDeviceAccessIsSerialized(t *testing.T) {
	fixture := newServiceFixture(t)
	var activeReads, maxActiveReads int
	var counterMutex sync.Mutex
	fixture.client.readFn = func(tuya.Device) (map[string]any, error) {
		counterMutex.Lock()
		activeReads++
		maxActiveReads = max(maxActiveReads, activeReads)
		counterMutex.Unlock()
		time.Sleep(5 * time.Millisecond)
		counterMutex.Lock()
		activeReads--
		counterMutex.Unlock()
		return map[string]any{}, nil
	}

	var group sync.WaitGroup
	for range 8 {
		group.Add(1)
		go func() {
			defer group.Done()
			_ = fixture.service.TestConnection(context.Background())
		}()
	}
	group.Wait()

	if maxActiveReads != 1 {
		t.Fatalf("device was accessed concurrently: %d", maxActiveReads)
	}
}

func TestUnreachableTriggersSearchStoresAddressAndRetriesOnce(t *testing.T) {
	fixture := newServiceFixture(t)
	fixture.client.readFn = func(device tuya.Device) (map[string]any, error) {
		if device.Address == testAddress {
			return nil, unreachableFailure()
		}
		return map[string]any{dpBatteryPercent: float64(55)}, nil
	}
	fixture.client.findFn = func(device tuya.Device, _ []string) (string, error) {
		if device.DeviceID != testDeviceID || device.LocalKey != testLocalKey {
			t.Errorf("search used wrong credentials: %+v", device)
		}
		return "10.1.1.3", nil
	}

	status, err := fixture.service.GetStatus(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if status.BatteryPercent != 55 {
		t.Fatalf("retry did not return the new reading: %+v", status)
	}
	if fixture.storedSettings(t).Address != "10.1.1.3" {
		t.Fatalf("new address not stored: %+v", fixture.storedSettings(t))
	}
	if fixture.client.findCalls != 1 || len(fixture.client.readDevices) != 2 {
		t.Fatalf("expected one search and two reads, got %d and %d", fixture.client.findCalls, len(fixture.client.readDevices))
	}
	if len(fixture.client.findCandidates[0]) != 2 {
		t.Fatalf("search did not receive the network candidates: %v", fixture.client.findCandidates)
	}
}

func TestSecondUnreachableWithinOneMinuteDoesNotSearchAgain(t *testing.T) {
	fixture := newServiceFixture(t)
	fixture.client.readFn = func(tuya.Device) (map[string]any, error) { return nil, unreachableFailure() }

	if err := fixture.service.TestConnection(context.Background()); !errors.Is(err, ErrDeviceUnreachable) {
		t.Fatalf("expected unreachable, got %v", err)
	}
	fixture.clock.advance(time.Minute)
	if err := fixture.service.TestConnection(context.Background()); !errors.Is(err, ErrDeviceUnreachable) {
		t.Fatalf("expected unreachable, got %v", err)
	}
	if fixture.client.findCalls != 1 {
		t.Fatalf("searched again within one minute: %d searches", fixture.client.findCalls)
	}

	fixture.clock.advance(time.Second)
	if err := fixture.service.TestConnection(context.Background()); !errors.Is(err, ErrDeviceUnreachable) {
		t.Fatalf("expected unreachable, got %v", err)
	}
	if fixture.client.findCalls != 2 {
		t.Fatalf("expected a new search after more than one minute, got %d", fixture.client.findCalls)
	}
}

func TestRetryAfterSearchFailingAgainIsUnreachable(t *testing.T) {
	fixture := newServiceFixture(t)
	fixture.client.readFn = func(tuya.Device) (map[string]any, error) { return nil, unreachableFailure() }
	fixture.client.findFn = func(tuya.Device, []string) (string, error) { return "10.1.1.3", nil }

	err := fixture.service.TestConnection(context.Background())
	if !errors.Is(err, ErrDeviceUnreachable) {
		t.Fatalf("expected unreachable, got %v", err)
	}
	if len(fixture.client.readDevices) != 2 {
		t.Fatalf("operation must be retried exactly once, got %d reads", len(fixture.client.readDevices))
	}
}

func TestWriteAlsoSearchesWhenUnreachable(t *testing.T) {
	fixture := newServiceFixture(t)
	fixture.client.writeFn = func(device tuya.Device, _ map[string]any) error {
		if device.Address == testAddress {
			return unreachableFailure()
		}
		return nil
	}
	fixture.client.findFn = func(tuya.Device, []string) (string, error) { return "10.1.1.3", nil }

	if err := fixture.service.SendCommand(context.Background(), "start"); err != nil {
		t.Fatal(err)
	}
	if len(fixture.client.writes) != 2 || fixture.client.writes[1].device.Address != "10.1.1.3" {
		t.Fatalf("write was not retried on the new address: %+v", fixture.client.writes)
	}
}

func TestSearchWithoutResultReturnsUnreachable(t *testing.T) {
	fixture := newServiceFixture(t)
	fixture.client.readFn = func(tuya.Device) (map[string]any, error) { return nil, unreachableFailure() }

	err := fixture.service.TestConnection(context.Background())
	if !errors.Is(err, ErrDeviceUnreachable) || errors.Is(err, ErrDiscoveryNotFound) {
		t.Fatalf("expected plain unreachable, got %v", err)
	}
	if fixture.storedSettings(t).Address != testAddress {
		t.Fatalf("address must not change without a result")
	}
}

func TestEmptyAddressSearchesBeforeFirstOperation(t *testing.T) {
	fixture := newServiceFixture(t)
	fixture.repository.document, _ = encodeSettings(settingsState{DeviceID: testDeviceID, LocalKey: testLocalKey})
	fixture.client.findFn = func(tuya.Device, []string) (string, error) { return "10.1.1.2", nil }

	if err := fixture.service.TestConnection(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(fixture.client.readDevices) != 1 || fixture.client.readDevices[0].Address != "10.1.1.2" {
		t.Fatalf("read must happen only on the found address: %+v", fixture.client.readDevices)
	}
}

func TestEmptyAddressWithoutSearchResultIsUnreachable(t *testing.T) {
	fixture := newServiceFixture(t)
	fixture.repository.document, _ = encodeSettings(settingsState{DeviceID: testDeviceID, LocalKey: testLocalKey})

	err := fixture.service.TestConnection(context.Background())
	if !errors.Is(err, ErrDeviceUnreachable) || len(fixture.client.readDevices) != 0 {
		t.Fatalf("expected unreachable without reading, got %v (%d reads)", err, len(fixture.client.readDevices))
	}
}

func TestCandidateListingFailureIsNotMaskedAsUnreachable(t *testing.T) {
	fixture := newServiceFixture(t)
	listingFailure := errors.New("no interfaces")
	fixture.service.candidateAddresses = func() ([]string, error) { return nil, listingFailure }
	fixture.client.readFn = func(tuya.Device) (map[string]any, error) { return nil, unreachableFailure() }

	err := fixture.service.TestConnection(context.Background())
	if !errors.Is(err, listingFailure) {
		t.Fatalf("expected the listing failure, got %v", err)
	}
}

func TestStoreAddressFailureAfterSearchIsReturned(t *testing.T) {
	fixture := newServiceFixture(t)
	fixture.repository.upsertErr = errors.New("db down")
	fixture.client.readFn = func(tuya.Device) (map[string]any, error) { return nil, unreachableFailure() }
	fixture.client.findFn = func(tuya.Device, []string) (string, error) { return "10.1.1.3", nil }

	if err := fixture.service.TestConnection(context.Background()); !errors.Is(err, fixture.repository.upsertErr) {
		t.Fatalf("expected the repository failure, got %v", err)
	}
}

func TestDeviceFailuresAreTranslated(t *testing.T) {
	tests := []struct {
		name    string
		failure error
		want    error
	}{
		{"key rejected", fmt.Errorf("%w: bad", tuya.ErrKeyRejected), ErrLocalKeyRejected},
		{"invalid device", fmt.Errorf("%w: bad", tuya.ErrInvalidDevice), ErrNotConfigured},
		{"deadline", context.DeadlineExceeded, ErrDeviceUnreachable},
		{"canceled", context.Canceled, ErrDeviceUnreachable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newServiceFixture(t)
			fixture.client.readFn = func(tuya.Device) (map[string]any, error) { return nil, test.failure }

			if err := fixture.service.TestConnection(context.Background()); !errors.Is(err, test.want) {
				t.Fatalf("expected %v, got %v", test.want, err)
			}
			if fixture.client.findCalls != 0 {
				t.Fatalf("only unreachable failures may trigger a search")
			}
		})
	}
}

func TestUnknownDeviceFailureIsReturnedUntouched(t *testing.T) {
	fixture := newServiceFixture(t)
	unknownFailure := errors.New("boom")
	fixture.client.readFn = func(tuya.Device) (map[string]any, error) { return nil, unknownFailure }

	if err := fixture.service.TestConnection(context.Background()); err != unknownFailure {
		t.Fatalf("expected the original error, got %v", err)
	}
}

func TestNotConfiguredAndSettingsLoadFailure(t *testing.T) {
	fixture := newServiceFixture(t)
	fixture.repository.isFound = false
	if err := fixture.service.TestConnection(context.Background()); !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("expected not configured, got %v", err)
	}
	if _, err := fixture.service.Discover(context.Background()); !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("expected not configured, got %v", err)
	}

	fixture.repository.getErr = errors.New("db down")
	if err := fixture.service.TestConnection(context.Background()); !errors.Is(err, fixture.repository.getErr) {
		t.Fatalf("expected the repository failure, got %v", err)
	}
	if _, err := fixture.service.Discover(context.Background()); !errors.Is(err, fixture.repository.getErr) {
		t.Fatalf("expected the repository failure, got %v", err)
	}
}

func TestTestConnectionDoesNotUseSnapshot(t *testing.T) {
	fixture := newServiceFixture(t)
	if _, err := fixture.service.GetStatus(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := fixture.service.TestConnection(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(fixture.client.readDevices) != 2 {
		t.Fatalf("connection test must perform a real read, got %d reads", len(fixture.client.readDevices))
	}
}

func TestDiscoverStoresAddress(t *testing.T) {
	fixture := newServiceFixture(t)
	fixture.client.findFn = func(tuya.Device, []string) (string, error) { return "10.1.1.2", nil }

	discovery, err := fixture.service.Discover(context.Background())
	if err != nil || discovery.Address != "10.1.1.2" {
		t.Fatalf("unexpected discovery: %+v %v", discovery, err)
	}
	if fixture.storedSettings(t).Address != "10.1.1.2" {
		t.Fatalf("address not stored")
	}
}

func TestDiscoverOutcomes(t *testing.T) {
	tests := []struct {
		name string
		find func(tuya.Device, []string) (string, error)
		want error
	}{
		{"nothing found", func(tuya.Device, []string) (string, error) { return "", tuya.ErrDeviceNotFound }, ErrDiscoveryNotFound},
		{"invalid credentials", func(tuya.Device, []string) (string, error) { return "", tuya.ErrInvalidDevice }, ErrNotConfigured},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newServiceFixture(t)
			fixture.client.findFn = test.find

			_, err := fixture.service.Discover(context.Background())
			if !errors.Is(err, test.want) {
				t.Fatalf("expected %v, got %v", test.want, err)
			}
			if fixture.repository.upsertCalls != 0 {
				t.Fatalf("no address may be stored without a result")
			}
		})
	}
}

func TestDiscoverIgnoresTheOneMinuteLimit(t *testing.T) {
	fixture := newServiceFixture(t)
	fixture.client.findFn = func(tuya.Device, []string) (string, error) { return "10.1.1.2", nil }

	for range 3 {
		if _, err := fixture.service.Discover(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	if fixture.client.findCalls != 3 {
		t.Fatalf("manual discovery must always search, got %d", fixture.client.findCalls)
	}
}
