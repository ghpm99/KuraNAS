package music

import (
	"database/sql"
	"time"

	"nas-go/api/pkg/utils"
)

type playHistoryRepositoryRecorder struct {
	musicRepoMock
	insertedClientID      string
	insertedFileID        int
	insertedPlayedSeconds int
	insertFailure         error
	receivedEarliest      *time.Time
	playedTracks          []PlayedTrackModel
}

func (m *playHistoryRepositoryRecorder) InsertPlayEvent(tx *sql.Tx, clientID string, fileID int, playedSeconds int) error {
	m.insertedClientID = clientID
	m.insertedFileID = fileID
	m.insertedPlayedSeconds = playedSeconds
	return m.insertFailure
}

func (m *playHistoryRepositoryRecorder) GetMostPlayedTracks(earliestPlayedAt *time.Time, page int, pageSize int) (utils.PaginationResponse[PlayedTrackModel], error) {
	m.receivedEarliest = earliestPlayedAt
	return utils.PaginationResponse[PlayedTrackModel]{Items: m.playedTracks}, nil
}

func (m *playHistoryRepositoryRecorder) GetRecentlyPlayedTracks(page int, pageSize int) (utils.PaginationResponse[PlayedTrackModel], error) {
	return utils.PaginationResponse[PlayedTrackModel]{Items: m.playedTracks}, nil
}

func (m *musicRepoMock) InsertPlayEvent(tx *sql.Tx, clientID string, fileID int, playedSeconds int) error {
	return nil
}

func (m *musicRepoMock) GetMostPlayedTracks(earliestPlayedAt *time.Time, page int, pageSize int) (utils.PaginationResponse[PlayedTrackModel], error) {
	return utils.PaginationResponse[PlayedTrackModel]{}, nil
}

func (m *musicRepoMock) GetRecentlyPlayedTracks(page int, pageSize int) (utils.PaginationResponse[PlayedTrackModel], error) {
	return utils.PaginationResponse[PlayedTrackModel]{}, nil
}

func (m *musicHandlerServiceMock) RecordPlay(clientID string, request RecordPlayRequest) error {
	return nil
}

func (m *musicHandlerServiceMock) GetMostPlayedTracks(period PlayPeriod, page int, pageSize int) (utils.PaginationResponse[MusicPlayedTrackDto], error) {
	return utils.PaginationResponse[MusicPlayedTrackDto]{Items: []MusicPlayedTrackDto{{PlayCount: 3}}}, nil
}

func (m *musicHandlerServiceMock) GetRecentlyPlayedTracks(page int, pageSize int) (utils.PaginationResponse[MusicPlayedTrackDto], error) {
	return utils.PaginationResponse[MusicPlayedTrackDto]{Items: []MusicPlayedTrackDto{{PlayCount: 1}}}, nil
}

func (m *musicHandlerErrServiceMock) RecordPlay(clientID string, request RecordPlayRequest) error {
	return errQueueFailure
}

func (m *musicHandlerErrServiceMock) GetMostPlayedTracks(period PlayPeriod, page int, pageSize int) (utils.PaginationResponse[MusicPlayedTrackDto], error) {
	return utils.PaginationResponse[MusicPlayedTrackDto]{}, errQueueFailure
}

func (m *musicHandlerErrServiceMock) GetRecentlyPlayedTracks(page int, pageSize int) (utils.PaginationResponse[MusicPlayedTrackDto], error) {
	return utils.PaginationResponse[MusicPlayedTrackDto]{}, errQueueFailure
}
