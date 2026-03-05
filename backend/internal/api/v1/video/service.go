package video

import (
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
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

func (s *Service) StartPlayback(clientID string, videoID int, playlistID *int) (PlaybackSessionDto, error) {
	videoFile, err := s.Repository.GetVideoFileByID(videoID)
	if err != nil {
		return PlaybackSessionDto{}, err
	}

	var playlist VideoPlaylistModel
	if playlistID != nil {
		playlist, err = s.Repository.GetVideoPlaylistByID(*playlistID)
		if err != nil {
			return PlaybackSessionDto{}, err
		}
		inPlaylist, checkErr := s.Repository.CheckVideoInPlaylist(playlist.ID, videoID)
		if checkErr != nil {
			return PlaybackSessionDto{}, checkErr
		}
		if !inPlaylist {
			return PlaybackSessionDto{}, errors.New("video nao pertence a playlist selecionada")
		}
	} else {
		playlist, err = s.ensureContextPlaylist(string(ContextFolder), videoFile.ParentPath)
		if err != nil {
			return PlaybackSessionDto{}, err
		}
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
	playlist, err := s.Repository.GetVideoPlaylistByID(playlistID)
	if err != nil {
		return VideoPlaylistModel{}, err
	}
	return playlist, nil
}

func (s *Service) GetHomeCatalog(clientID string, limit int) (VideoHomeCatalogDto, error) {
	const defaultLimit = 24
	const maxLimit = 100

	normalizedLimit := limit
	if normalizedLimit <= 0 {
		normalizedLimit = defaultLimit
	} else if normalizedLimit > maxLimit {
		normalizedLimit = maxLimit
	}

	fetchLimit := normalizedLimit * 4

	allVideos, err := s.Repository.GetCatalogVideos(fetchLimit)
	if err != nil {
		return VideoHomeCatalogDto{}, err
	}
	recentVideos, err := s.Repository.GetRecentVideos(normalizedLimit)
	if err != nil {
		return VideoHomeCatalogDto{}, err
	}

	state, _ := s.Repository.GetPlaybackState(clientID)

	series := make([]VideoCatalogItemDto, 0, normalizedLimit)
	movies := make([]VideoCatalogItemDto, 0, normalizedLimit)
	personal := make([]VideoCatalogItemDto, 0, normalizedLimit)

	for _, video := range allVideos {
		item := s.toCatalogItem(video, state)
		classification := classifyVideo(video)
		if classification == "series" && len(series) < normalizedLimit {
			series = append(series, item)
			continue
		}
		if classification == "movie" && len(movies) < normalizedLimit {
			movies = append(movies, item)
			continue
		}
		if len(personal) < normalizedLimit {
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

func (s *Service) RebuildSmartPlaylists() error {
	videos, err := s.Repository.GetAllVideosForGrouping()
	if err != nil {
		return err
	}

	groups := buildSmartGroups(videos)

	return s.withTransaction(func(tx *sql.Tx) error {
		for _, group := range groups {
			playlist, upsertErr := s.Repository.UpsertAutoPlaylist(
				tx,
				group.PlaylistType,
				group.SourceKey,
				group.Name,
				group.GroupMode,
				group.Classification,
			)
			if upsertErr != nil {
				return upsertErr
			}

			exclusions, exclusionsErr := s.Repository.GetPlaylistExclusions(playlist.ID)
			if exclusionsErr != nil {
				return exclusionsErr
			}

			filtered := make([]int, 0, len(group.VideoIDs))
			for _, id := range group.VideoIDs {
				if !exclusions[id] {
					filtered = append(filtered, id)
				}
			}

			if err := s.Repository.DeleteAutoPlaylistItems(tx, playlist.ID); err != nil {
				return err
			}
			if err := s.Repository.InsertPlaylistItemsWithSource(tx, playlist.ID, filtered, "auto"); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Service) GetPlaylists(includeHidden bool) ([]VideoPlaylistDto, error) {
	models, err := s.Repository.GetVideoPlaylists(includeHidden)
	if err != nil {
		return nil, err
	}

	result := make([]VideoPlaylistDto, 0, len(models))
	for _, model := range models {
		result = append(result, model.ToDto(nil))
	}
	return result, nil
}

func (s *Service) GetPlaylistByID(id int) (VideoPlaylistDto, error) {
	playlist, err := s.Repository.GetVideoPlaylistByID(id)
	if err != nil {
		return VideoPlaylistDto{}, err
	}

	items, err := s.Repository.GetVideoPlaylistItemsDetailed(id)
	if err != nil {
		return VideoPlaylistDto{}, err
	}

	itemDtos := make([]VideoPlaylistItemDto, 0, len(items))
	for _, item := range items {
		itemDtos = append(itemDtos, item.ToDto("not_started"))
	}
	playlist.ItemCount = len(itemDtos)
	return playlist.ToDto(itemDtos), nil
}

func (s *Service) SetPlaylistHidden(playlistID int, hidden bool) error {
	return s.withTransaction(func(tx *sql.Tx) error {
		return s.Repository.SetPlaylistHidden(tx, playlistID, hidden)
	})
}

func (s *Service) AddVideoToPlaylist(playlistID int, videoID int) error {
	return s.withTransaction(func(tx *sql.Tx) error {
		if err := s.Repository.AddPlaylistVideoManual(tx, playlistID, videoID); err != nil {
			return err
		}
		return s.Repository.DeletePlaylistExclusion(tx, playlistID, videoID)
	})
}

func (s *Service) RemoveVideoFromPlaylist(playlistID int, videoID int) error {
	playlist, err := s.Repository.GetVideoPlaylistByID(playlistID)
	if err != nil {
		return err
	}

	return s.withTransaction(func(tx *sql.Tx) error {
		if err := s.Repository.RemovePlaylistVideo(tx, playlistID, videoID); err != nil {
			return err
		}
		if playlist.IsAuto {
			return s.Repository.UpsertPlaylistExclusion(tx, playlistID, videoID)
		}
		return nil
	})
}

func (s *Service) GetUnassignedVideos(limit int) ([]VideoFileDto, error) {
	if limit <= 0 {
		limit = 2000
	}
	models, err := s.Repository.GetUnassignedVideos(limit)
	if err != nil {
		return nil, err
	}
	result := make([]VideoFileDto, 0, len(models))
	for _, model := range models {
		result = append(result, model.ToDto())
	}
	return result, nil
}

func (s *Service) UpdatePlaylistName(playlistID int, name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return errors.New("nome da playlist nao pode ser vazio")
	}

	return s.withTransaction(func(tx *sql.Tx) error {
		return s.Repository.UpdatePlaylistName(tx, playlistID, trimmed)
	})
}

func (s *Service) ReorderPlaylistItems(playlistID int, items []ReorderPlaylistItemRequest) error {
	if len(items) == 0 {
		return errors.New("nenhum item enviado para reordenacao")
	}

	seenVideo := map[int]bool{}
	seenOrder := map[int]bool{}
	for _, item := range items {
		if seenVideo[item.VideoID] {
			return fmt.Errorf("video_id duplicado na reordenacao: %d", item.VideoID)
		}
		if seenOrder[item.OrderIndex] {
			return fmt.Errorf("order_index duplicado na reordenacao: %d", item.OrderIndex)
		}
		seenVideo[item.VideoID] = true
		seenOrder[item.OrderIndex] = true
	}

	return s.withTransaction(func(tx *sql.Tx) error {
		for _, item := range items {
			if err := s.Repository.ReorderPlaylistItem(tx, playlistID, item.VideoID, item.OrderIndex); err != nil {
				return err
			}
		}
		return nil
	})
}

type smartGroup struct {
	SourceKey      string
	Name           string
	PlaylistType   string
	GroupMode      string
	Classification string
	VideoIDs       []int
}

func buildSmartGroups(videos []VideoFileModel) []smartGroup {
	folderCount := map[string]int{}
	prefixCount := map[string]int{}

	for _, video := range videos {
		folderCount[video.ParentPath]++
		prefix := inferTitlePrefix(video.Name)
		if prefix != "" {
			prefixCount[prefix]++
		}
	}

	groupMap := map[string]*smartGroup{}
	order := []string{}

	addToGroup := func(key string, group smartGroup, videoID int) {
		existing, ok := groupMap[key]
		if !ok {
			copyGroup := group
			copyGroup.VideoIDs = []int{}
			groupMap[key] = &copyGroup
			order = append(order, key)
			existing = groupMap[key]
		}
		existing.VideoIDs = append(existing.VideoIDs, videoID)
	}

	for _, video := range videos {
		classification := classifySmartVideo(video)
		folderBase := strings.TrimSpace(filepath.Base(video.ParentPath))
		folderIsStrong := folderCount[video.ParentPath] >= 2 && !isGenericFolderName(folderBase)
		prefix := inferTitlePrefix(video.Name)
		prefixIsStrong := prefix != "" && prefixCount[prefix] >= 2

		switch {
		case folderIsStrong:
			addToGroup(
				"folder:"+video.ParentPath,
				smartGroup{
					SourceKey:      "folder:" + video.ParentPath,
					Name:           folderBase,
					PlaylistType:   "folder",
					GroupMode:      "folder",
					Classification: classification,
				},
				video.ID,
			)
		case prefixIsStrong:
			addToGroup(
				"prefix:"+prefix,
				smartGroup{
					SourceKey:      "prefix:" + prefix,
					Name:           strings.Title(prefix),
					PlaylistType:   "series",
					GroupMode:      "prefix",
					Classification: classification,
				},
				video.ID,
			)
		default:
			playlistType := "custom"
			if classification == "movie" {
				playlistType = "movie"
			}
			singletonName := strings.TrimSpace(strings.TrimSuffix(video.Name, filepath.Ext(video.Name)))
			addToGroup(
				fmt.Sprintf("single:%d", video.ID),
				smartGroup{
					SourceKey:      fmt.Sprintf("single:%d", video.ID),
					Name:           singletonName,
					PlaylistType:   playlistType,
					GroupMode:      "single",
					Classification: classification,
				},
				video.ID,
			)
		}
	}

	result := make([]smartGroup, 0, len(order))
	for _, key := range order {
		group := groupMap[key]
		sort.Ints(group.VideoIDs)
		result = append(result, *group)
	}
	return result
}

func inferTitlePrefix(name string) string {
	noExt := strings.TrimSpace(strings.TrimSuffix(name, filepath.Ext(name)))
	if noExt == "" {
		return ""
	}

	bracketCleanup := regexp.MustCompile(`\\[[^\\]]+\\]|\\([^\\)]+\\)`)
	episodeSuffix := regexp.MustCompile(`(?i)[\\s._-]*(s\\d{1,2}e\\d{1,2}|ep\\.?\\s?\\d+|epis[oó]dio\\s?\\d+|part\\s?\\d+|parte\\s?\\d+|\\d{1,3})$`)
	spaceCollapse := regexp.MustCompile(`[\\s._-]+`)

	value := strings.ToLower(noExt)
	value = bracketCleanup.ReplaceAllString(value, "")
	value = episodeSuffix.ReplaceAllString(value, "")
	value = spaceCollapse.ReplaceAllString(value, " ")
	value = strings.TrimSpace(value)
	return value
}

func isGenericFolderName(name string) bool {
	value := strings.ToLower(strings.TrimSpace(name))
	if value == "" {
		return true
	}
	generic := map[string]bool{
		"videos": true, "video": true, "movies": true, "filmes": true, "downloads": true, "clips": true,
		"desktop": true, "documentos": true, "documents": true, "media": true,
	}
	return generic[value]
}

func classifySmartVideo(video VideoFileModel) string {
	path := strings.ToLower(video.Path + " " + video.ParentPath + " " + video.Name)
	if strings.Contains(path, "steam") || strings.Contains(path, "tutorial") || strings.Contains(path, "sample") {
		return "program"
	}
	if strings.Contains(path, "anime") || strings.Contains(path, "animes") {
		return "anime"
	}
	if strings.Contains(path, "series") || strings.Contains(path, "season") || strings.Contains(path, "temporada") {
		return "series"
	}
	if strings.Contains(path, "movie") || strings.Contains(path, "filme") || strings.Contains(path, "cinema") {
		return "movie"
	}
	return "personal"
}
