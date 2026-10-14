package video

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

const (
	folderPlaylistSourcePrefix = "folder:"
	seriesPlaylistSourcePrefix = "series:"
)

var ErrPosterNotFound = errors.New("local poster not found")

func (s *Service) FindPlaylistPosterFileID(playlistID int) (int, error) {
	playlist, err := s.Repository.GetVideoPlaylistByID(playlistID)
	if err != nil {
		return 0, err
	}

	folderPath, err := s.resolvePlaylistPosterFolder(playlist)
	if err != nil {
		return 0, err
	}

	posterFileID, isFound, err := s.findSeriesPosterFileID(folderPath)
	if err != nil {
		return 0, err
	}
	if !isFound {
		return 0, ErrPosterNotFound
	}
	return posterFileID, nil
}

func (s *Service) resolvePlaylistPosterFolder(playlist VideoPlaylistModel) (string, error) {
	if strings.HasPrefix(playlist.SourcePath, folderPlaylistSourcePrefix) {
		return strings.TrimPrefix(playlist.SourcePath, folderPlaylistSourcePrefix), nil
	}
	if !strings.HasPrefix(playlist.SourcePath, seriesPlaylistSourcePrefix) {
		return "", ErrPosterNotFound
	}

	dominantFolder, err := s.Repository.GetPlaylistDominantFolder(playlist.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrPosterNotFound
	}
	if err != nil {
		return "", err
	}
	return dominantFolder, nil
}

func (s *Service) findSeriesPosterFileID(folderPath string) (int, bool, error) {
	posterFileID, isFound, err := s.findSeriesPosterInFolder(folderPath)
	if err != nil || isFound {
		return posterFileID, isFound, err
	}

	parentFolder, isSeasonFolder := seasonParentFolder(folderPath)
	if !isSeasonFolder {
		return 0, false, nil
	}
	return s.findSeriesPosterInFolder(parentFolder)
}

func (s *Service) findSeriesPosterInFolder(folderPath string) (int, bool, error) {
	candidates, err := s.Repository.ListPosterCandidatesInFolder(folderPath)
	if err != nil {
		return 0, false, err
	}
	poster, isFound := pickSeriesPoster(candidates)
	return poster.ID, isFound, nil
}

func (s *Service) FindMoviePosterFileID(videoID int) (int, error) {
	video, err := s.Repository.GetVideoFileByID(videoID)
	if err != nil {
		return 0, err
	}

	candidates, err := s.Repository.ListPosterCandidatesInFolder(video.ParentPath)
	if err != nil {
		return 0, err
	}
	if len(candidates) == 0 {
		return 0, ErrPosterNotFound
	}

	videosInFolder, err := s.Repository.CountVideosInFolder(video.ParentPath)
	if err != nil {
		return 0, fmt.Errorf("FindMoviePosterFileID: %w", err)
	}

	poster, isFound := pickMoviePoster(video.Name, videosInFolder == 1, candidates)
	if !isFound {
		return 0, ErrPosterNotFound
	}
	return poster.ID, nil
}
