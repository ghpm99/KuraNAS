package video

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type Service struct {
	Repository RepositoryInterface
}

func NewService(repository RepositoryInterface) ServiceInterface {
	return &Service{Repository: repository}
}

func (s *Service) withTransaction(fn func(tx *sql.Tx) error) error {
	return s.Repository.GetDbContext().ExecTx(fn)
}

func (s *Service) StartPlayback(clientID string, videoID int) (PlaybackSessionDto, error) {
	videoFile, err := s.Repository.GetVideoFileByID(videoID)
	if err != nil {
		return PlaybackSessionDto{}, err
	}

	playlist, err := s.ensureContextPlaylist(string(ContextFolder), videoFile.ParentPath)
	if err != nil {
		return PlaybackSessionDto{}, err
	}

	state, _ := s.Repository.GetPlaybackState(clientID)
	if !state.PlaylistID.Valid || int(state.PlaylistID.Int64) != playlist.ID || !state.VideoID.Valid || int(state.VideoID.Int64) != videoID {
		state = VideoPlaybackStateModel{
			ClientID:    clientID,
			CurrentTime: 0,
			Duration:    0,
			IsPaused:    false,
			Completed:   false,
		}
	}
	state.PlaylistID = sql.NullInt64{Int64: int64(playlist.ID), Valid: true}
	state.VideoID = sql.NullInt64{Int64: int64(videoID), Valid: true}

	if err := s.withTransaction(func(tx *sql.Tx) error {
		updatedState, upsertErr := s.Repository.UpsertPlaybackState(tx, state)
		if upsertErr != nil {
			return upsertErr
		}
		state = updatedState
		return s.Repository.TouchPlaylist(tx, playlist.ID)
	}); err != nil {
		return PlaybackSessionDto{}, err
	}

	return s.buildSession(clientID, playlist, state)
}

func (s *Service) GetPlaybackState(clientID string) (PlaybackSessionDto, error) {
	state, err := s.Repository.GetPlaybackState(clientID)
	if err != nil {
		return PlaybackSessionDto{}, err
	}
	if !state.PlaylistID.Valid {
		return PlaybackSessionDto{}, errors.New("estado sem playlist ativa")
	}

	playlist, err := s.playlistByIDAndClient(state)
	if err != nil {
		return PlaybackSessionDto{}, err
	}

	return s.buildSession(clientID, playlist, state)
}

func (s *Service) UpdatePlaybackState(clientID string, req UpdatePlaybackStateRequest) (VideoPlaybackStateDto, error) {
	state, err := s.Repository.GetPlaybackState(clientID)
	if err != nil {
		state = VideoPlaybackStateModel{ClientID: clientID, IsPaused: true}
	}

	state.ClientID = clientID
	if req.PlaylistID != nil {
		state.PlaylistID = sql.NullInt64{Int64: int64(*req.PlaylistID), Valid: true}
	}
	if req.VideoID != nil {
		state.VideoID = sql.NullInt64{Int64: int64(*req.VideoID), Valid: true}
	}
	if req.CurrentTime != nil {
		state.CurrentTime = *req.CurrentTime
	}
	if req.Duration != nil {
		state.Duration = *req.Duration
	}
	if req.IsPaused != nil {
		state.IsPaused = *req.IsPaused
	}
	if req.Completed != nil {
		state.Completed = *req.Completed
	}

	if err := s.withTransaction(func(tx *sql.Tx) error {
		updatedState, upsertErr := s.Repository.UpsertPlaybackState(tx, state)
		if upsertErr != nil {
			return upsertErr
		}
		state = updatedState
		if state.PlaylistID.Valid {
			return s.Repository.TouchPlaylist(tx, int(state.PlaylistID.Int64))
		}
		return nil
	}); err != nil {
		return VideoPlaybackStateDto{}, fmt.Errorf("erro ao atualizar estado do player de video: %w", err)
	}

	state.ClientID = clientID
	return state.ToDto(), nil
}

func (s *Service) NextVideo(clientID string) (PlaybackSessionDto, error) {
	return s.shiftPlayback(clientID, 1)
}

func (s *Service) PreviousVideo(clientID string) (PlaybackSessionDto, error) {
	return s.shiftPlayback(clientID, -1)
}

func (s *Service) ensureContextPlaylist(contextType string, sourcePath string) (VideoPlaylistModel, error) {
	playlist, err := s.Repository.GetPlaylistByContext(contextType, sourcePath)
	if err == nil {
		items, itemsErr := s.Repository.GetPlaylistItems(playlist.ID)
		if itemsErr == nil && len(items) > 0 {
			return playlist, nil
		}
	}

	videos, err := s.Repository.GetVideosByParentPath(sourcePath)
	if err != nil {
		return VideoPlaylistModel{}, err
	}
	if len(videos) == 0 {
		return VideoPlaylistModel{}, errors.New("nenhum video encontrado para o contexto")
	}

	if playlist.ID == 0 {
		err = s.withTransaction(func(tx *sql.Tx) error {
			created, createErr := s.Repository.CreatePlaylist(tx, contextType, sourcePath)
			if createErr != nil {
				return createErr
			}
			playlist = created

			videoIDs := make([]int, 0, len(videos))
			for _, video := range videos {
				videoIDs = append(videoIDs, video.ID)
			}
			return s.Repository.ReplacePlaylistItems(tx, playlist.ID, videoIDs)
		})
		if err != nil {
			return VideoPlaylistModel{}, err
		}
		return playlist, nil
	}

	err = s.withTransaction(func(tx *sql.Tx) error {
		videoIDs := make([]int, 0, len(videos))
		for _, video := range videos {
			videoIDs = append(videoIDs, video.ID)
		}
		return s.Repository.ReplacePlaylistItems(tx, playlist.ID, videoIDs)
	})
	if err != nil {
		return VideoPlaylistModel{}, err
	}
	return playlist, nil
}

func (s *Service) buildSession(clientID string, playlist VideoPlaylistModel, state VideoPlaybackStateModel) (PlaybackSessionDto, error) {
	items, err := s.Repository.GetPlaylistItems(playlist.ID)
	if err != nil {
		return PlaybackSessionDto{}, err
	}

	itemDtos := make([]VideoPlaylistItemDto, 0, len(items))
	currentVideoID := -1
	if state.VideoID.Valid {
		currentVideoID = int(state.VideoID.Int64)
	}

	for _, item := range items {
		status := "not_started"
		if item.VideoID == currentVideoID {
			if state.Completed {
				status = "completed"
			} else if state.CurrentTime > 0 {
				status = "in_progress"
			}
		}
		itemDtos = append(itemDtos, item.ToDto(status))
	}

	state.ClientID = clientID
	return PlaybackSessionDto{
		Playlist:      playlist.ToDto(itemDtos),
		PlaybackState: state.ToDto(),
	}, nil
}

func (s *Service) shiftPlayback(clientID string, direction int) (PlaybackSessionDto, error) {
	state, err := s.Repository.GetPlaybackState(clientID)
	if err != nil {
		return PlaybackSessionDto{}, err
	}
	if !state.PlaylistID.Valid || !state.VideoID.Valid {
		return PlaybackSessionDto{}, errors.New("sem video ativo para avancar ou retroceder")
	}

	playlist, err := s.playlistByIDAndClient(state)
	if err != nil {
		return PlaybackSessionDto{}, err
	}

	items, err := s.Repository.GetPlaylistItems(playlist.ID)
	if err != nil {
		return PlaybackSessionDto{}, err
	}
	if len(items) == 0 {
		return PlaybackSessionDto{}, errors.New("playlist sem itens")
	}

	currentIndex := -1
	for i, item := range items {
		if item.VideoID == int(state.VideoID.Int64) {
			currentIndex = i
			break
		}
	}
	if currentIndex < 0 {
		currentIndex = 0
	}

	nextIndex := currentIndex + direction
	if nextIndex < 0 {
		nextIndex = 0
	}
	if nextIndex >= len(items) {
		nextIndex = len(items) - 1
	}

	state.VideoID = sql.NullInt64{Int64: int64(items[nextIndex].VideoID), Valid: true}
	state.CurrentTime = 0
	state.Duration = 0
	state.IsPaused = false
	state.Completed = false

	if err := s.withTransaction(func(tx *sql.Tx) error {
		updatedState, upsertErr := s.Repository.UpsertPlaybackState(tx, state)
		if upsertErr != nil {
			return upsertErr
		}
		state = updatedState
		return s.Repository.TouchPlaylist(tx, playlist.ID)
	}); err != nil {
		return PlaybackSessionDto{}, err
	}

	return s.buildSession(clientID, playlist, state)
}

func (s *Service) playlistByIDAndClient(state VideoPlaybackStateModel) (VideoPlaylistModel, error) {
	playlistID := int(state.PlaylistID.Int64)
	if !state.VideoID.Valid {
		return VideoPlaylistModel{}, errors.New("estado sem video atual")
	}

	videoFile, err := s.Repository.GetVideoFileByID(int(state.VideoID.Int64))
	if err != nil {
		return VideoPlaylistModel{}, err
	}

	playlist, err := s.Repository.GetPlaylistByContext(string(ContextFolder), videoFile.ParentPath)
	if err != nil {
		return VideoPlaylistModel{}, err
	}
	if playlist.ID != playlistID {
		return VideoPlaylistModel{}, errors.New("playlist ativa inconsistente com o contexto")
	}
	return playlist, nil
}

func (s *Service) GetHomeCatalog(clientID string, limit int) (VideoHomeCatalogDto, error) {
	if limit <= 0 {
		limit = 24
	}

	allVideos, err := s.Repository.GetCatalogVideos(limit * 4)
	if err != nil {
		return VideoHomeCatalogDto{}, err
	}
	recentVideos, err := s.Repository.GetRecentVideos(limit)
	if err != nil {
		return VideoHomeCatalogDto{}, err
	}

	state, _ := s.Repository.GetPlaybackState(clientID)

	series := make([]VideoCatalogItemDto, 0, limit)
	movies := make([]VideoCatalogItemDto, 0, limit)
	personal := make([]VideoCatalogItemDto, 0, limit)

	for _, video := range allVideos {
		item := s.toCatalogItem(video, state)
		classification := classifyVideo(video)
		if classification == "series" && len(series) < limit {
			series = append(series, item)
			continue
		}
		if classification == "movie" && len(movies) < limit {
			movies = append(movies, item)
			continue
		}
		if len(personal) < limit {
			personal = append(personal, item)
		}
	}

	recent := make([]VideoCatalogItemDto, 0, len(recentVideos))
	for _, video := range recentVideos {
		recent = append(recent, s.toCatalogItem(video, state))
	}

	continueWatching := []VideoCatalogItemDto{}
	if state.VideoID.Valid {
		for _, video := range allVideos {
			if video.ID == int(state.VideoID.Int64) {
				item := s.toCatalogItem(video, state)
				if item.Status == "in_progress" {
					continueWatching = append(continueWatching, item)
				}
				break
			}
		}
	}

	return VideoHomeCatalogDto{
		Sections: []VideoCatalogSectionDto{
			{Key: "continue", Title: "Continue assistindo", Items: continueWatching},
			{Key: "series", Title: "Series", Items: series},
			{Key: "movies", Title: "Filmes", Items: movies},
			{Key: "personal", Title: "Videos pessoais", Items: personal},
			{Key: "recent", Title: "Adicionados recentemente", Items: recent},
		},
	}, nil
}

func (s *Service) toCatalogItem(video VideoFileModel, state VideoPlaybackStateModel) VideoCatalogItemDto {
	status := "not_started"
	progressPct := 0.0
	if state.VideoID.Valid && int(state.VideoID.Int64) == video.ID {
		if state.Completed {
			status = "completed"
			progressPct = 100
		} else if state.CurrentTime > 0 {
			status = "in_progress"
			if state.Duration > 0 {
				progressPct = (state.CurrentTime / state.Duration) * 100
			}
		}
	}

	if progressPct < 0 {
		progressPct = 0
	}
	if progressPct > 100 {
		progressPct = 100
	}

	return VideoCatalogItemDto{
		Video:       video.ToDto(),
		Status:      status,
		ProgressPct: progressPct,
	}
}

func classifyVideo(video VideoFileModel) string {
	path := strings.ToLower(video.Path + " " + video.ParentPath + " " + video.Name)
	episodePattern := regexp.MustCompile(`s\\d{1,2}e\\d{1,2}|ep\\.?\\s?\\d+|epis[oó]dio\\s?\\d+`)

	if strings.Contains(path, "/series") || strings.Contains(path, "/anime") || strings.Contains(path, "season") || strings.Contains(path, "temporada") || episodePattern.MatchString(path) {
		return "series"
	}
	if strings.Contains(path, "/movies") || strings.Contains(path, "/filmes") || strings.Contains(path, "/movie") {
		return "movie"
	}
	return "personal"
}
