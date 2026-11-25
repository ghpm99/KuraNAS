package vacuum

import (
	"context"
	"errors"
	"strings"
	"testing"

	"nas-go/api/pkg/tuyacloud"
)

func TestGetCloudSettingsNeverReturnsTheSecret(t *testing.T) {
	fixture := newServiceFixture(t)

	settings, err := fixture.service.GetCloudSettings()
	if err != nil {
		t.Fatal(err)
	}
	want := CloudSettingsDto{AccessID: testCloudAccessID, Region: "eu", HasAccessSecret: true}
	if settings != want {
		t.Fatalf("got %+v, want %+v", settings, want)
	}
}

func TestGetCloudSettingsWithoutStoredDocument(t *testing.T) {
	fixture := newServiceFixture(t)
	fixture.repository.isCloudFound = false

	settings, err := fixture.service.GetCloudSettings()
	if err != nil || settings != (CloudSettingsDto{}) {
		t.Fatalf("got %+v, %v", settings, err)
	}
}

func TestUpdateCloudSettingsKeepsTheStoredSecretWhenNoneIsSent(t *testing.T) {
	fixture := newServiceFixture(t)

	updated, err := fixture.service.UpdateCloudSettings(UpdateCloudSettingsDto{AccessID: " new-id ", Region: "us"})
	if err != nil {
		t.Fatal(err)
	}
	if updated != (CloudSettingsDto{AccessID: "new-id", Region: "us", HasAccessSecret: true}) {
		t.Fatalf("unexpected dto: %+v", updated)
	}
	stored, err := decodeCloudSettings(fixture.repository.cloudDocument)
	if err != nil || stored.AccessSecret != testCloudAccessSecret || stored.AccessID != "new-id" {
		t.Fatalf("unexpected stored settings: %+v %v", stored, err)
	}
}

func TestUpdateCloudSettingsReplacesTheSecretWhenSent(t *testing.T) {
	fixture := newServiceFixture(t)
	newSecret := strings.Repeat("z", cloudAccessSecretLength)

	if _, err := fixture.service.UpdateCloudSettings(UpdateCloudSettingsDto{AccessID: "id", AccessSecret: newSecret, Region: "in"}); err != nil {
		t.Fatal(err)
	}
	stored, _ := decodeCloudSettings(fixture.repository.cloudDocument)
	if stored.AccessSecret != newSecret || stored.Region != "in" {
		t.Fatalf("unexpected stored settings: %+v", stored)
	}
}

func TestUpdateCloudSettingsValidation(t *testing.T) {
	tests := []struct {
		name    string
		request UpdateCloudSettingsDto
		want    error
	}{
		{"empty access id", UpdateCloudSettingsDto{AccessID: "  ", Region: "eu"}, ErrInvalidCloudAccessID},
		{"unknown region", UpdateCloudSettingsDto{AccessID: "id", Region: "xx"}, ErrInvalidCloudRegion},
		{"empty region", UpdateCloudSettingsDto{AccessID: "id"}, ErrInvalidCloudRegion},
		{"short secret", UpdateCloudSettingsDto{AccessID: "id", AccessSecret: "short", Region: "eu"}, ErrInvalidCloudAccessSecret},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newServiceFixture(t)
			before := fixture.repository.cloudDocument

			_, err := fixture.service.UpdateCloudSettings(test.request)
			if !errors.Is(err, test.want) {
				t.Fatalf("got %v, want %v", err, test.want)
			}
			if fixture.repository.cloudDocument != before {
				t.Fatal("invalid settings must not be stored")
			}
		})
	}
}

func TestUpdateCloudSettingsRequiresASecretWhenNoneIsStored(t *testing.T) {
	fixture := newServiceFixture(t)
	fixture.repository.isCloudFound = false

	_, err := fixture.service.UpdateCloudSettings(UpdateCloudSettingsDto{AccessID: "id", Region: "eu"})
	if !errors.Is(err, ErrCloudAccessSecretRequired) {
		t.Fatalf("got %v", err)
	}
}

func TestUpdateCloudSettingsPropagatesStorageFailures(t *testing.T) {
	fixture := newServiceFixture(t)
	fixture.repository.upsertErr = errors.New("db down")

	if _, err := fixture.service.UpdateCloudSettings(UpdateCloudSettingsDto{AccessID: "id", Region: "eu"}); err == nil {
		t.Fatal("expected a storage failure")
	}
	fixture.repository.getErr = errors.New("db down")
	if _, err := fixture.service.GetCloudSettings(); err == nil {
		t.Fatal("expected a read failure")
	}
}

func TestUpdateCloudSettingsDropsTheCachedMap(t *testing.T) {
	fixture := newServiceFixture(t)
	if _, err := fixture.service.GetMap(context.Background()); err != nil {
		t.Fatal(err)
	}

	if _, err := fixture.service.UpdateCloudSettings(UpdateCloudSettingsDto{AccessID: "other", Region: "eu"}); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.service.GetMap(context.Background()); err != nil {
		t.Fatal(err)
	}
	if fixture.cloud.calls != 2 {
		t.Fatalf("map must be read again after the credentials changed, got %d reads", fixture.cloud.calls)
	}
}

func TestTestCloudConnectionReadsTheMapIgnoringTheCache(t *testing.T) {
	fixture := newServiceFixture(t)
	if _, err := fixture.service.GetMap(context.Background()); err != nil {
		t.Fatal(err)
	}

	if err := fixture.service.TestCloudConnection(context.Background()); err != nil {
		t.Fatal(err)
	}
	if fixture.cloud.calls != 2 {
		t.Fatalf("connection test must reach the cloud, got %d reads", fixture.cloud.calls)
	}
}

func TestTestCloudConnectionReportsEachCloudFailure(t *testing.T) {
	tests := []struct {
		cloudFailure error
		want         error
	}{
		{tuyacloud.ErrUnreachable, ErrCloudUnreachable},
		{tuyacloud.ErrCredentialsRejected, ErrCloudCredentialsRejected},
		{tuyacloud.ErrUnknownRegion, ErrCloudCredentialsRejected},
		{tuyacloud.ErrServiceNotAuthorized, ErrCloudServiceNotAuthorized},
		{tuyacloud.ErrInvalidResponse, ErrCloudInvalidResponse},
		{context.DeadlineExceeded, ErrCloudUnreachable},
	}
	for _, test := range tests {
		t.Run(test.want.Error(), func(t *testing.T) {
			fixture := newServiceFixture(t)
			fixture.cloud.failure = test.cloudFailure

			if err := fixture.service.TestCloudConnection(context.Background()); !errors.Is(err, test.want) {
				t.Fatalf("got %v, want %v", err, test.want)
			}
		})
	}
}

func TestTestCloudConnectionPassesThroughUnexpectedFailures(t *testing.T) {
	fixture := newServiceFixture(t)
	unexpected := errors.New("boom")
	fixture.cloud.failure = unexpected

	if err := fixture.service.TestCloudConnection(context.Background()); !errors.Is(err, unexpected) {
		t.Fatalf("got %v", err)
	}
}
