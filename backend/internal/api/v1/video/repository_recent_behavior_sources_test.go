package video

import (
	"errors"
	"regexp"
	"testing"
	"time"

	queries "nas-go/api/pkg/database/queries/video"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestListRecentVideoWatchProgress(t *testing.T) {
	repo, mock, db := newVideoRepoWithMock(t)
	defer db.Close()
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.ListRecentVideoWatchProgressQuery)).
		WithArgs(500).
		WillReturnRows(sqlmock.NewRows([]string{"client_id", "video_id", "position_seconds", "duration_seconds", "completed", "updated_at"}).
			AddRow("tv", 1, 12.5, 100.0, true, now))
	mock.ExpectRollback()

	progress, err := repo.ListRecentVideoWatchProgress(500)
	if err != nil || len(progress) != 1 {
		t.Fatalf("expected one progress row, got %+v err=%v", progress, err)
	}
	if progress[0].ClientID != "tv" || progress[0].VideoID != 1 || progress[0].PositionSeconds != 12.5 || !progress[0].Completed {
		t.Fatalf("unexpected progress %+v", progress[0])
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.ListRecentVideoWatchProgressQuery)).
		WithArgs(500).
		WillReturnError(errors.New("boom"))
	mock.ExpectRollback()

	if _, err := repo.ListRecentVideoWatchProgress(500); err == nil {
		t.Fatalf("expected query error")
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.ListRecentVideoWatchProgressQuery)).
		WithArgs(500).
		WillReturnRows(sqlmock.NewRows([]string{"client_id"}).AddRow("tv"))
	mock.ExpectRollback()

	if _, err := repo.ListRecentVideoWatchProgress(500); err == nil {
		t.Fatalf("expected scan error")
	}
}

func TestListRecentBehaviorEventsByType(t *testing.T) {
	repo, mock, db := newVideoRepoWithMock(t)
	defer db.Close()
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.ListRecentBehaviorEventsByTypeQuery)).
		WithArgs("skipped", 500).
		WillReturnRows(sqlmock.NewRows([]string{"id", "client_id", "video_id", "playlist_id", "event_type", "position", "duration", "watched_pct", "created_at"}).
			AddRow(3, "tv", 1, 0, "skipped", 5.0, 50.0, 10.0, now))
	mock.ExpectRollback()

	events, err := repo.ListRecentBehaviorEventsByType("skipped", 500)
	if err != nil || len(events) != 1 {
		t.Fatalf("expected one event, got %+v err=%v", events, err)
	}
	if events[0].ID != 3 || events[0].EventType != "skipped" || events[0].WatchedPct != 10 {
		t.Fatalf("unexpected event %+v", events[0])
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.ListRecentBehaviorEventsByTypeQuery)).
		WithArgs("skipped", 500).
		WillReturnError(errors.New("boom"))
	mock.ExpectRollback()

	if _, err := repo.ListRecentBehaviorEventsByType("skipped", 500); err == nil {
		t.Fatalf("expected query error")
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.ListRecentBehaviorEventsByTypeQuery)).
		WithArgs("skipped", 500).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectRollback()

	if _, err := repo.ListRecentBehaviorEventsByType("skipped", 500); err == nil {
		t.Fatalf("expected scan error")
	}
}
