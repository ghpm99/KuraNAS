package video

import (
	"database/sql"
	"testing"

	"nas-go/api/internal/testutil"
)

func TestRecentBehaviorSources_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_video_it")
	repository := NewRepository(dbContext)

	seedErr := dbContext.ExecTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(`TRUNCATE video_watch_progress, video_behavior_event, home_file RESTART IDENTITY CASCADE`); err != nil {
			return err
		}
		if _, err := tx.Exec(`INSERT INTO home_file (name, path, parent_path, format, size, updated_at, created_at, type, checksum, deleted_at)
			VALUES ('a.mp4', '/v/a.mp4', '/v', '.mp4', 1, now(), now(), 2, '', NULL),
			       ('b.mp4', '/v/b.mp4', '/v', '.mp4', 1, now(), now(), 2, '', NULL),
			       ('c.mp4', '/v/c.mp4', '/v', '.mp4', 1, now(), now(), 2, '', NULL)`); err != nil {
			return err
		}
		if _, err := tx.Exec(`INSERT INTO video_watch_progress (client_id, video_id, position_seconds, duration_seconds, completed, updated_at) VALUES
			('tv', 1, 10, 100, false, '2026-01-01 10:00:00'),
			('tv', 2, 90, 100, true,  '2026-01-03 10:00:00'),
			('phone', 3, 5, 50, false, '2026-01-02 10:00:00')`); err != nil {
			return err
		}
		_, err := tx.Exec(`INSERT INTO video_behavior_event (client_id, video_id, playlist_id, event_type, position, duration, watched_pct, created_at) VALUES
			('tv', 1, NULL, 'skipped', 5, 50, 10, '2026-01-01 10:00:00'),
			('tv', 2, NULL, 'completed', 50, 50, 100, '2026-01-05 10:00:00'),
			('phone', 3, NULL, 'skipped', 7, 70, 10, '2026-01-04 10:00:00')`)
		return err
	})
	if seedErr != nil {
		t.Fatalf("seed: %v", seedErr)
	}

	progress, err := repository.ListRecentVideoWatchProgress(10)
	if err != nil {
		t.Fatalf("list progress: %v", err)
	}
	if len(progress) != 3 || progress[0].VideoID != 2 || progress[1].VideoID != 3 || progress[2].VideoID != 1 {
		t.Fatalf("expected progress ordered by updated_at desc (2,3,1), got %+v", progress)
	}
	if progress[0].ClientID != "tv" || progress[0].PositionSeconds != 90 || progress[0].DurationSeconds != 100 || !progress[0].Completed || progress[0].UpdatedAt.IsZero() {
		t.Fatalf("unexpected progress columns %+v", progress[0])
	}

	limited, err := repository.ListRecentVideoWatchProgress(1)
	if err != nil || len(limited) != 1 || limited[0].VideoID != 2 {
		t.Fatalf("expected limit to keep newest row, got %+v err=%v", limited, err)
	}

	skips, err := repository.ListRecentBehaviorEventsByType("skipped", 10)
	if err != nil {
		t.Fatalf("list skips: %v", err)
	}
	if len(skips) != 2 || skips[0].VideoID != 3 || skips[1].VideoID != 1 {
		t.Fatalf("expected only skipped ordered by created_at desc (3,1), got %+v", skips)
	}
	if skips[0].ClientID != "phone" || skips[0].PlaylistID != 0 || skips[0].EventType != "skipped" || skips[0].Position != 7 || skips[0].Duration != 70 || skips[0].WatchedPct != 10 || skips[0].CreatedAt.IsZero() {
		t.Fatalf("unexpected skip columns %+v", skips[0])
	}

	limitedSkips, err := repository.ListRecentBehaviorEventsByType("skipped", 1)
	if err != nil || len(limitedSkips) != 1 || limitedSkips[0].VideoID != 3 {
		t.Fatalf("expected limit to keep newest skip, got %+v err=%v", limitedSkips, err)
	}
}
