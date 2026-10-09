package music

import (
	"database/sql"
	"fmt"
	"time"

	"nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/utils"
)

func (s *Service) RecordPlay(clientID string, request RecordPlayRequest) error {
	if !request.isValid() {
		return ErrInvalidPlayRequest
	}
	err := s.withTransaction(func(tx *sql.Tx) error {
		return s.Repository.InsertPlayEvent(tx, clientID, request.FileID, request.PlayedSeconds)
	})
	if err != nil {
		return fmt.Errorf("erro ao registrar a reproducao: %w", err)
	}
	return nil
}

func (s *Service) GetMostPlayedTracks(period PlayPeriod, page int, pageSize int) (utils.PaginationResponse[MusicPlayedTrackDto], error) {
	if !period.isValid() {
		return utils.PaginationResponse[MusicPlayedTrackDto]{}, ErrInvalidPlayRequest
	}
	page, pageSize = normalizePagination(page, pageSize)
	mostPlayed, err := s.Repository.GetMostPlayedTracks(period.earliestPlayedAt(time.Now()), page, pageSize)
	if err != nil {
		return utils.PaginationResponse[MusicPlayedTrackDto]{}, err
	}
	return s.loadPlayedTracks(mostPlayed)
}

func (s *Service) GetRecentlyPlayedTracks(page int, pageSize int) (utils.PaginationResponse[MusicPlayedTrackDto], error) {
	page, pageSize = normalizePagination(page, pageSize)
	recentlyPlayed, err := s.Repository.GetRecentlyPlayedTracks(page, pageSize)
	if err != nil {
		return utils.PaginationResponse[MusicPlayedTrackDto]{}, err
	}
	return s.loadPlayedTracks(recentlyPlayed)
}

func (s *Service) loadPlayedTracks(playedTracks utils.PaginationResponse[PlayedTrackModel]) (utils.PaginationResponse[MusicPlayedTrackDto], error) {
	fileIDs := make([]int, 0, len(playedTracks.Items))
	for _, playedTrack := range playedTracks.Items {
		fileIDs = append(fileIDs, playedTrack.FileID)
	}

	fileModels, err := s.Repository.GetLibraryFilesByIDs(fileIDs)
	if err != nil {
		return utils.PaginationResponse[MusicPlayedTrackDto]{}, err
	}

	filesByID := map[int]files.FileDto{}
	for _, fileModel := range fileModels {
		fileDto, dtoErr := fileModel.ToDto()
		if dtoErr != nil {
			return utils.PaginationResponse[MusicPlayedTrackDto]{}, dtoErr
		}
		fileDto.Metadata = fileModel.Metadata
		filesByID[fileModel.ID] = fileDto
	}

	items := make([]MusicPlayedTrackDto, 0, len(playedTracks.Items))
	for _, playedTrack := range playedTracks.Items {
		fileDto, exists := filesByID[playedTrack.FileID]
		if !exists {
			continue
		}
		items = append(items, MusicPlayedTrackDto{Track: fileDto, PlayCount: playedTrack.PlayCount, LastPlayedAt: playedTrack.LastPlayedAt})
	}

	return utils.PaginationResponse[MusicPlayedTrackDto]{Items: items, Pagination: playedTracks.Pagination}, nil
}
