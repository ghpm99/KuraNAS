package video

import (
	"database/sql"
	"errors"
	"testing"

	"nas-go/api/internal/testutil"
)

func TestVideoWatchProgress_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_video_it")
	repository := NewRepository(dbContext)

	seedErr := dbContext.ExecTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(`TRUNCATE video_watch_progress, home_file RESTART IDENTITY CASCADE`); err != nil {
			return err
		}
		_, err := tx.Exec(`INSERT INTO home_file (name, path, parent_path, format, size, updated_at, created_at, type, checksum, deleted_at)
			VALUES ('a.mp4', '/v/a.mp4', '/v', '.mp4', 1, now(), now(), 2, '', NULL),
			       ('b.mp4', '/v/b.mp4', '/v', '.mp4', 1, now(), now(), 2, '', NULL)`)
		return err
	})
	if seedErr != nil {
		t.Fatalf("seed: %v", seedErr)
	}

	save := func(progress VideoWatchProgressModel) {
		t.Helper()
		err := dbContext.ExecTx(func(tx *sql.Tx) error {
			_, upsertErr := repository.UpsertVideoWatchProgress(tx, progress)
			return upsertErr
		})
		if err != nil {
			t.Fatalf("upsert: %v", err)
		}
	}

	save(VideoWatchProgressModel{ClientID: "tv", VideoID: 1, PositionSeconds: 10, DurationSeconds: 100})
	save(VideoWatchProgressModel{ClientID: "tv", VideoID: 1, PositionSeconds: 55, DurationSeconds: 100, Completed: true})
	save(VideoWatchProgressModel{ClientID: "tv", VideoID: 2, PositionSeconds: 3, DurationSeconds: 30})
	save(VideoWatchProgressModel{ClientID: "phone", VideoID: 1, PositionSeconds: 99, DurationSeconds: 100})

	overwritten, err := repository.GetVideoWatchProgress("tv", 1)
	if err != nil {
		t.Fatalf("get tv/1: %v", err)
	}
	if overwritten.PositionSeconds != 55 || !overwritten.Completed || overwritten.UpdatedAt.IsZero() {
		t.Fatalf("expected upsert to overwrite the row, got %+v", overwritten)
	}

	otherVideo, err := repository.GetVideoWatchProgress("tv", 2)
	if err != nil || otherVideo.PositionSeconds != 3 || otherVideo.Completed {
		t.Fatalf("expected isolated progress for video 2, got %+v err=%v", otherVideo, err)
	}

	otherClient, err := repository.GetVideoWatchProgress("phone", 1)
	if err != nil || otherClient.PositionSeconds != 99 {
		t.Fatalf("expected isolated progress for other client, got %+v err=%v", otherClient, err)
	}

	if _, err := repository.GetVideoWatchProgress("tv", 99); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows for missing progress, got %v", err)
	}

	if err := dbContext.ExecTx(func(tx *sql.Tx) error {
		_, deleteErr := tx.Exec(`DELETE FROM home_file WHERE id = 1`)
		return deleteErr
	}); err != nil {
		t.Fatalf("delete file: %v", err)
	}
	if _, err := repository.GetVideoWatchProgress("tv", 1); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected cascade to remove progress, got %v", err)
	}
	if _, err := repository.GetVideoWatchProgress("tv", 2); err != nil {
		t.Fatalf("progress of surviving video must remain: %v", err)
	}
}
