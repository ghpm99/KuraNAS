package music

import (
	"errors"
	"testing"
	"time"

	"nas-go/api/internal/api/v1/files"
)

func TestRecordPlayRejectsInvalidRequestsBeforeTheRepository(t *testing.T) {
	cases := []RecordPlayRequest{
		{FileID: 0, PlayedSeconds: 30},
		{FileID: 1, PlayedSeconds: -1},
		{FileID: 1, PlayedSeconds: maxPlayedSeconds + 1},
	}
	for _, request := range cases {
		recorder := &playHistoryRepositoryRecorder{}
		service := newMusicServiceForTest(t, &recorder.musicRepoMock)
		service.Repository = recorder

		if err := service.RecordPlay("device-12345", request); !errors.Is(err, ErrInvalidPlayRequest) {
			t.Fatalf("request %+v error = %v", request, err)
		}
		if recorder.insertedFileID != 0 {
			t.Fatalf("invalid request %+v reached the repository", request)
		}
	}
}

func TestRecordPlayStoresClientFileAndSecondsAndWrapsFailure(t *testing.T) {
	recorder := &playHistoryRepositoryRecorder{}
	service := newMusicServiceForTest(t, &recorder.musicRepoMock)
	service.Repository = recorder

	if err := service.RecordPlay("device-12345", RecordPlayRequest{FileID: 9, PlayedSeconds: 42}); err != nil {
		t.Fatalf("record: %v", err)
	}
	if recorder.insertedClientID != "device-12345" || recorder.insertedFileID != 9 || recorder.insertedPlayedSeconds != 42 {
		t.Fatalf("stored = %+v", recorder)
	}

	failure := errors.New("boom")
	recorder.insertFailure = failure
	if err := service.RecordPlay("device-12345", RecordPlayRequest{FileID: 9}); !errors.Is(err, failure) {
		t.Fatalf("error = %v, want wrapped failure", err)
	}
}

func TestMostPlayedTracksResolvesPeriodAndRejectsUnknownOne(t *testing.T) {
	recorder := &playHistoryRepositoryRecorder{}
	service := newMusicServiceForTest(t, &recorder.musicRepoMock)
	service.Repository = recorder

	if _, err := service.GetMostPlayedTracks(PlayPeriodAll, 1, 10); err != nil || recorder.receivedEarliest != nil {
		t.Fatalf("all period: err=%v earliest=%v", err, recorder.receivedEarliest)
	}
	if _, err := service.GetMostPlayedTracks(PlayPeriodLast30Days, 1, 10); err != nil || recorder.receivedEarliest == nil {
		t.Fatalf("30d period: err=%v earliest=%v", err, recorder.receivedEarliest)
	}
	expectedEarliest := time.Now().Add(-mostPlayedRecentPeriod)
	if difference := recorder.receivedEarliest.Sub(expectedEarliest); difference > time.Minute || difference < -time.Minute {
		t.Fatalf("earliest = %v, want about %v", recorder.receivedEarliest, expectedEarliest)
	}
	if _, err := service.GetMostPlayedTracks(PlayPeriod("7d"), 1, 10); !errors.Is(err, ErrInvalidPlayRequest) {
		t.Fatalf("unknown period error = %v", err)
	}
}

func TestPlayedTracksKeepRepositoryOrderAndSkipMissingFiles(t *testing.T) {
	lastPlayedAt := time.Date(2026, time.April, 1, 10, 0, 0, 0, time.UTC)
	repository := &playedTracksFilesRepository{
		playHistoryRepositoryRecorder: playHistoryRepositoryRecorder{playedTracks: []PlayedTrackModel{
			{FileID: 5, PlayCount: 4, LastPlayedAt: lastPlayedAt},
			{FileID: 99, PlayCount: 3, LastPlayedAt: lastPlayedAt},
			{FileID: 2, PlayCount: 1, LastPlayedAt: lastPlayedAt},
		}},
		fileModels: []files.FileModel{{ID: 2, Name: "two.mp3"}, {ID: 5, Name: "five.mp3"}},
	}
	service := &Service{Repository: repository}

	recentlyPlayed, err := service.GetRecentlyPlayedTracks(1, 10)
	if err != nil {
		t.Fatalf("recent plays: %v", err)
	}
	if len(recentlyPlayed.Items) != 2 || recentlyPlayed.Items[0].Track.ID != 5 || recentlyPlayed.Items[0].PlayCount != 4 || recentlyPlayed.Items[1].Track.ID != 2 {
		t.Fatalf("items = %+v", recentlyPlayed.Items)
	}

	mostPlayed, err := service.GetMostPlayedTracks(PlayPeriodAll, 1, 10)
	if err != nil || len(mostPlayed.Items) != 2 {
		t.Fatalf("most played: err=%v items=%+v", err, mostPlayed.Items)
	}
}

type playedTracksFilesRepository struct {
	playHistoryRepositoryRecorder
	fileModels []files.FileModel
}

func (r *playedTracksFilesRepository) GetLibraryFilesByIDs(fileIDs []int) ([]files.FileModel, error) {
	return r.fileModels, nil
}
