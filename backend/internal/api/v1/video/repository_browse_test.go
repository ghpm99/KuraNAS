package video

import (
	"database/sql/driver"
	"fmt"
	"regexp"
	"testing"
	"time"

	queries "nas-go/api/pkg/database/queries/video"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func browseCols(total int) []string {
	cols := make([]string, 0, total)
	for i := 1; i <= total; i++ {
		cols = append(cols, fmt.Sprintf("c%d", i))
	}
	return cols
}

func TestVideoRepositoryGetVideos(t *testing.T) {
	repo, mock, db := newVideoRepoWithMock(t)
	defer db.Close()
	now := time.Now()
	fileType := 1

	videoValues := []driver.Value{
		5, "video", "/tmp/video.mp4", "/tmp", ".mp4", int64(30), now, now, nil, nil, fileType, "sum3", nil, false,
		6, 5, "/tmp/video.mp4", "mp4", "30", "60.0", 1920, 1080, 30.0, 1800, "1000000", "h264", "H.264", "yuv420p",
		40, "High", "16:9", "aac", 2, "48000", "192000", now,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetVideosQuery)).
		WillReturnRows(sqlmock.NewRows(browseCols(len(videoValues))).AddRow(videoValues...))
	mock.ExpectRollback()
	videos, err := repo.GetVideos(1, 10)
	if err != nil || len(videos.Items) != 1 {
		t.Fatalf("GetVideos failed len=%d err=%v", len(videos.Items), err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetVideosQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectRollback()
	if _, err := repo.GetVideos(1, 10); err == nil {
		t.Fatalf("expected GetVideos scan error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}
