package captures

import (
	"database/sql"
	"testing"
	"time"

	"nas-go/api/internal/testutil"
	"nas-go/api/pkg/utils"
)

func TestGetCapturesBreaksCreatedAtTiesByIDDescending_Postgres(t *testing.T) {
	repository := NewRepository(testutil.NewPostgresDB(t, "kuranas_captures_it"))
	sharedCreatedAt := time.Now().Truncate(time.Second)
	tiePlatform := "order-tie-platform"

	var createdIDs []int
	for _, captureName := range []string{"order-tie-a", "order-tie-b", "order-tie-c"} {
		err := repository.GetDbContext().ExecTx(func(tx *sql.Tx) error {
			created, createErr := repository.CreateCapture(tx, CaptureModel{
				Name: captureName, FileName: captureName + ".mp4", FilePath: "/tmp/" + captureName,
				MediaType: "recording", MimeType: "video/mp4", CreatedAt: sharedCreatedAt, Status: CaptureStatusUploaded,
			})
			if createErr != nil {
				return createErr
			}
			createdIDs = append(createdIDs, created.ID)
			_, execErr := tx.Exec(`UPDATE captures SET platform = $1, created_at = $2 WHERE id = $3`, tiePlatform, sharedCreatedAt, created.ID)
			return execErr
		})
		if err != nil {
			t.Fatalf("seed %s: %v", captureName, err)
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

	page, err := repository.GetCaptures(CaptureFilter{Platform: utils.Optional[string]{HasValue: true, Value: tiePlatform}}, 1, 10)
	if err != nil {
		t.Fatalf("GetCaptures: %v", err)
	}
	if len(page.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(page.Items))
	}
	for position, expectedID := range []int{createdIDs[2], createdIDs[1], createdIDs[0]} {
		if page.Items[position].ID != expectedID {
			t.Fatalf("position %d: expected id %d, got %d", position, expectedID, page.Items[position].ID)
		}
	}
}
