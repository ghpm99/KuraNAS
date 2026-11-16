package vacuum

import (
	"errors"
	"testing"
)

func TestGetSettingsNeverExposesTheKey(t *testing.T) {
	fixture := newServiceFixture(t)

	settings, err := fixture.service.GetSettings()
	want := SettingsDto{Address: testAddress, DeviceID: testDeviceID, HasLocalKey: true}
	if err != nil || settings != want {
		t.Fatalf("got %+v (%v), want %+v", settings, err, want)
	}
}

func TestGetSettingsWithNothingStored(t *testing.T) {
	fixture := newServiceFixture(t)
	fixture.repository.isFound = false

	settings, err := fixture.service.GetSettings()
	if err != nil || settings != (SettingsDto{}) {
		t.Fatalf("got %+v (%v)", settings, err)
	}
}

func TestGetSettingsRepositoryFailure(t *testing.T) {
	fixture := newServiceFixture(t)
	fixture.repository.getErr = errors.New("db down")

	if _, err := fixture.service.GetSettings(); err == nil {
		t.Fatal("expected an error")
	}
}

func TestGetSettingsCorruptDocument(t *testing.T) {
	fixture := newServiceFixture(t)
	fixture.repository.document = "{not json"

	if _, err := fixture.service.GetSettings(); err == nil {
		t.Fatal("expected an error")
	}
}

func TestUpdateKeepsStoredKeyWhenOmitted(t *testing.T) {
	fixture := newServiceFixture(t)

	settings, err := fixture.service.UpdateSettings(UpdateSettingsDto{Address: "10.0.0.9:6668", DeviceID: " new-id "})
	if err != nil {
		t.Fatal(err)
	}
	if settings != (SettingsDto{Address: "10.0.0.9:6668", DeviceID: "new-id", HasLocalKey: true}) {
		t.Fatalf("unexpected response %+v", settings)
	}
	if fixture.storedSettings(t).LocalKey != testLocalKey {
		t.Fatalf("stored key must be kept")
	}
}

func TestUpdateReplacesKeyWhenSent(t *testing.T) {
	fixture := newServiceFixture(t)

	_, err := fixture.service.UpdateSettings(UpdateSettingsDto{DeviceID: testDeviceID, LocalKey: "ponmlkjihgfedcba"})
	if err != nil || fixture.storedSettings(t).LocalKey != "ponmlkjihgfedcba" {
		t.Fatalf("key not replaced: %v", err)
	}
}

func TestUpdateAllowsEmptyAddress(t *testing.T) {
	fixture := newServiceFixture(t)

	if _, err := fixture.service.UpdateSettings(UpdateSettingsDto{DeviceID: testDeviceID}); err != nil {
		t.Fatal(err)
	}
	if fixture.storedSettings(t).Address != "" {
		t.Fatalf("address must be cleared")
	}
}

func TestUpdateWithoutAnyKeyIsRejected(t *testing.T) {
	fixture := newServiceFixture(t)
	fixture.repository.isFound = false

	_, err := fixture.service.UpdateSettings(UpdateSettingsDto{DeviceID: testDeviceID})
	if !errors.Is(err, ErrLocalKeyRequired) || fixture.repository.upsertCalls != 0 {
		t.Fatalf("got %v with %d saves", err, fixture.repository.upsertCalls)
	}
}

func TestUpdateValidation(t *testing.T) {
	tests := []struct {
		name    string
		request UpdateSettingsDto
		want    error
	}{
		{"empty device id", UpdateSettingsDto{DeviceID: "  "}, ErrInvalidDeviceID},
		{"short key", UpdateSettingsDto{DeviceID: "id", LocalKey: "short"}, ErrInvalidLocalKey},
		{"long key", UpdateSettingsDto{DeviceID: "id", LocalKey: "abcdefghijklmnopq"}, ErrInvalidLocalKey},
		{"address with spaces", UpdateSettingsDto{DeviceID: "id", Address: "not a host"}, ErrInvalidAddress},
		{"address bad port", UpdateSettingsDto{DeviceID: "id", Address: "10.0.0.1:99999"}, ErrInvalidAddress},
		{"address port zero", UpdateSettingsDto{DeviceID: "id", Address: "10.0.0.1:0"}, ErrInvalidAddress},
		{"address bad host with port", UpdateSettingsDto{DeviceID: "id", Address: "-bad-:6668"}, ErrInvalidAddress},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newServiceFixture(t)

			_, err := fixture.service.UpdateSettings(test.request)
			if !errors.Is(err, test.want) || fixture.repository.upsertCalls != 0 {
				t.Fatalf("got %v with %d saves", err, fixture.repository.upsertCalls)
			}
		})
	}
}

func TestUpdateAcceptsValidAddresses(t *testing.T) {
	for _, address := range []string{"10.0.0.5", "10.0.0.5:6668", "robot.local", "robot.local:6668", "::1", "[::1]:6668", ""} {
		t.Run(address, func(t *testing.T) {
			fixture := newServiceFixture(t)

			if _, err := fixture.service.UpdateSettings(UpdateSettingsDto{DeviceID: "id", Address: address}); err != nil {
				t.Fatalf("address %q rejected: %v", address, err)
			}
		})
	}
}

func TestUpdateRepositoryFailures(t *testing.T) {
	fixture := newServiceFixture(t)
	fixture.repository.upsertErr = errors.New("db down")
	if _, err := fixture.service.UpdateSettings(UpdateSettingsDto{DeviceID: "id"}); err == nil {
		t.Fatal("expected a save failure")
	}

	fixture.repository.getErr = errors.New("db down")
	if _, err := fixture.service.UpdateSettings(UpdateSettingsDto{DeviceID: "id"}); err == nil {
		t.Fatal("expected a load failure")
	}
}
