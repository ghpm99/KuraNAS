package takeout

import (
	"testing"

	"nas-go/api/internal/testutil"
)

func newSidecarRepositoryForTest(t *testing.T) *SidecarRepository {
	t.Helper()
	repository := NewSidecarRepository(testutil.NewPostgresDB(t, "kuranas_takeout_import_it"))
	if _, err := repository.DbContext.GetDatabase().Exec(`TRUNCATE takeout_photo_sidecar`); err != nil {
		t.Fatalf("truncate sidecars: %v", err)
	}
	return repository
}

func TestSidecarRepositorySavesAndFindsByPath_Postgres(t *testing.T) {
	repository := newSidecarRepositoryForTest(t)
	latitude, longitude := -23.5, -46.6

	err := repository.SaveSidecars([]PhotoSidecar{
		{Path: "/lib/takeout/a.jpg", Description: "praia", Latitude: &latitude, Longitude: &longitude},
		{Path: "/lib/takeout/b.jpg", Description: "só descrição"},
	})
	if err != nil {
		t.Fatalf("SaveSidecars: %v", err)
	}

	full, err := repository.FindSidecarByPath("/lib/takeout/a.jpg")
	if err != nil || full == nil || full.Description != "praia" || !full.HasLocation() || *full.Latitude != latitude || *full.Longitude != longitude {
		t.Fatalf("unexpected full sidecar %+v err=%v", full, err)
	}
	descriptionOnly, err := repository.FindSidecarByPath("/lib/takeout/b.jpg")
	if err != nil || descriptionOnly == nil || descriptionOnly.Description != "só descrição" || descriptionOnly.HasLocation() {
		t.Fatalf("unexpected description-only sidecar %+v err=%v", descriptionOnly, err)
	}
}

func TestSidecarRepositoryUpsertsAndMissesUnknownPath_Postgres(t *testing.T) {
	repository := newSidecarRepositoryForTest(t)

	if err := repository.SaveSidecars([]PhotoSidecar{{Path: "/lib/takeout/a.jpg", Description: "antiga"}}); err != nil {
		t.Fatalf("first save: %v", err)
	}
	if err := repository.SaveSidecars([]PhotoSidecar{{Path: "/lib/takeout/a.jpg", Description: "nova"}}); err != nil {
		t.Fatalf("second save: %v", err)
	}

	updated, err := repository.FindSidecarByPath("/lib/takeout/a.jpg")
	if err != nil || updated == nil || updated.Description != "nova" {
		t.Fatalf("expected the upsert to keep the latest description, got %+v err=%v", updated, err)
	}
	missing, err := repository.FindSidecarByPath("/lib/takeout/unknown.jpg")
	if err != nil || missing != nil {
		t.Fatalf("unknown path must return nil without error, got %+v err=%v", missing, err)
	}
}
