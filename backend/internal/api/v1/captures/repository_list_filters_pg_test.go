package captures

import (
	"database/sql"
	"testing"
	"time"

	"nas-go/api/internal/testutil"
	"nas-go/api/pkg/utils"
)

func TestGetCapturesFiltersByStatusAndPlatform_Postgres(t *testing.T) {
	repository := NewRepository(testutil.NewPostgresDB(t, "kuranas_captures_it"))

	createdIDs := map[string]int{}
	seed := []struct {
		name     string
		status   CaptureStatus
		platform string
	}{
		{"filter-a", CaptureStatusFailed, "filterplat-x"},
		{"filter-b", CaptureStatusUploaded, "filterplat-x"},
		{"filter-c", CaptureStatusFailed, "filterplat-y"},
	}
	for _, entry := range seed {
		err := repository.GetDbContext().ExecTx(func(tx *sql.Tx) error {
			created, createErr := repository.CreateCapture(tx, CaptureModel{
				Name: entry.name, FileName: entry.name + ".mp4", FilePath: "/tmp/" + entry.name,
				MediaType: "recording", MimeType: "video/mp4", CreatedAt: time.Now(), Status: entry.status,
			})
			if createErr != nil {
				return createErr
			}
			createdIDs[entry.name] = created.ID
			_, execErr := tx.Exec(`UPDATE captures SET platform = $1 WHERE id = $2`, entry.platform, created.ID)
			return execErr
		})
		if err != nil {
			t.Fatalf("seed %s: %v", entry.name, err)
		}
	}
	t.Cleanup(func() {
		_ = repository.GetDbContext().ExecTx(func(tx *sql.Tx) error {
			for _, id := range createdIDs {
				_ = repository.DeleteCapture(tx, id)
			}
			return nil
		})
	})

	filter := CaptureFilter{
		Status:   utils.Optional[string]{HasValue: true, Value: "failed"},
		Platform: utils.Optional[string]{HasValue: true, Value: "filterplat-x"},
	}
	page, err := repository.GetCaptures(filter, 1, 10)
	if err != nil {
		t.Fatalf("GetCaptures: %v", err)
	}
	if len(page.Items) != 1 || page.Items[0].ID != createdIDs["filter-a"] {
		t.Fatalf("expected only filter-a, got %+v", page.Items)
	}

	platformOnly := CaptureFilter{Platform: utils.Optional[string]{HasValue: true, Value: "filterplat-x"}}
	page, err = repository.GetCaptures(platformOnly, 1, 10)
	if err != nil {
		t.Fatalf("GetCaptures: %v", err)
	}
	if len(page.Items) != 2 {
		t.Fatalf("expected 2 items for platform, got %d", len(page.Items))
	}
}

func TestGetCapturePlatformsReturnsDistinctNonEmptySorted_Postgres(t *testing.T) {
	repository := NewRepository(testutil.NewPostgresDB(t, "kuranas_captures_it"))

	err := repository.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		_, deleteErr := tx.Exec(`DELETE FROM captures WHERE name LIKE 'platforms-seed-%'`)
		return deleteErr
	})
	if err != nil {
		t.Fatalf("clean: %v", err)
	}

	seededPlatforms := []string{"youtube", "crunchyroll", "crunchyroll", ""}
	createdIDs := []int{}
	for index, platform := range seededPlatforms {
		name := "platforms-seed-" + string(rune('a'+index))
		err := repository.GetDbContext().ExecTx(func(tx *sql.Tx) error {
			created, createErr := repository.CreateCapture(tx, CaptureModel{
				Name: name, FileName: name + ".mp4", FilePath: "/tmp/" + name,
				MediaType: "recording", MimeType: "video/mp4", CreatedAt: time.Now(), Status: CaptureStatusUploaded,
			})
			if createErr != nil {
				return createErr
			}
			createdIDs = append(createdIDs, created.ID)
			_, execErr := tx.Exec(`UPDATE captures SET platform = $1 WHERE id = $2`, platform, created.ID)
			return execErr
		})
		if err != nil {
			t.Fatalf("seed %s: %v", name, err)
		}
	}
	t.Cleanup(func() {
		_ = repository.GetDbContext().ExecTx(func(tx *sql.Tx) error {
			for _, id := range createdIDs {
				_ = repository.DeleteCapture(tx, id)
			}
			return nil
		})
	})

	platforms, err := repository.GetCapturePlatforms()
	if err != nil {
		t.Fatalf("GetCapturePlatforms: %v", err)
	}

	seededSubset := []string{}
	for _, platform := range platforms {
		if platform == "crunchyroll" || platform == "youtube" {
			seededSubset = append(seededSubset, platform)
		}
		if platform == "" {
			t.Fatalf("empty platform must not be returned: %v", platforms)
		}
	}
	if len(seededSubset) != 2 || seededSubset[0] != "crunchyroll" || seededSubset[1] != "youtube" {
		t.Fatalf("expected [crunchyroll youtube] distinct and sorted, got %v", platforms)
	}
}
