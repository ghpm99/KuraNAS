package video

import (
	"database/sql"
	"errors"

	"nas-go/api/internal/api/v1/video/playlist"
	"nas-go/api/pkg/ai"
	"nas-go/api/pkg/database"
)

type Service struct {
	Repository     RepositoryInterface
	PlaylistEngine *playlist.PlaylistEngine
	AIService      ai.ServiceInterface
}

type videoItemProgress struct {
	Status      string
	ProgressPct float64
}

var (
	ErrVideoNotInPlaylist      = errors.New("video not in selected playlist")
	ErrPlaybackStateNotFound   = errors.New("playback state not found")
	ErrInvalidBehaviorEvent    = errors.New("invalid behavior event")
	ErrPlaylistNameRequired    = errors.New("playlist name is required")
	ErrPlaylistReorderRequired = errors.New("playlist reorder items are required")
	ErrNoVideosForContext      = errors.New("no videos found for context")
	ErrPlaybackNavigation      = errors.New("playback navigation unavailable")
	ErrPlaylistWithoutItems    = errors.New("playlist has no items")
)

func NewService(repository RepositoryInterface, aiService ai.ServiceInterface) ServiceInterface {
	return &Service{
		Repository:     repository,
		PlaylistEngine: playlist.NewPlaylistEngine(),
		AIService:      aiService,
	}
}

func (s *Service) withTransaction(fn func(tx *sql.Tx) error) error {
	return database.ExecOptionalTx(s.Repository.GetDbContext(), fn)
}
