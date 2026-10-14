package video

import (
	"database/sql"
	"testing"

	"nas-go/api/internal/testutil"
)

func TestGetVideoDurationSecondsByFileID_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_video_it")
	repository := NewRepository(dbContext)
	videoIDsByName, _ := seedTechnicalMetadataFixture(t, repository)

	testCases := []struct {
		name          string
		fileID        int
		expectedValid bool
		expected      float64
	}{
		{"latest metadata row wins", videoIDsByName["full.mkv"], true, 5400.25},
		{"non numeric duration is null", videoIDsByName["unparsable.mkv"], false, 0},
		{"file without metadata is null", videoIDsByName["bare.mkv"], false, 0},
		{"unknown file is null", 999999, false, 0},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			duration, err := repository.GetVideoDurationSecondsByFileID(testCase.fileID)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if duration != (sql.NullFloat64{Float64: testCase.expected, Valid: testCase.expectedValid}) {
				t.Fatalf("expected valid=%v %v, got %+v", testCase.expectedValid, testCase.expected, duration)
			}
		})
	}
}
