package video

import (
	"database/sql"
	"errors"

	"nas-go/api/internal/api/v1/video/playlist"
	"nas-go/api/pkg/applog"
)

const (
	manualClassificationOther = "other"
	minimumManualSeason       = 0
	minimumManualEpisode      = 1
)

var manualClassifications = map[string]bool{
	string(playlist.ClassSeries):   true,
	string(playlist.ClassMovie):    true,
	string(playlist.ClassPersonal): true,
	string(playlist.ClassClip):     true,
	manualClassificationOther:      true,
}

type VideoFieldValidationError struct {
	Field      string
	MessageKey string
}

func (e *VideoFieldValidationError) Error() string {
	return "invalid video field: " + e.Field
}

// SetVideoClassification stores (or clears, when classification is nil) the
// owner's classification of one video, schedules the rebuild of the automatic
// playlists and returns the effective classification. It returns
// *VideoFieldValidationError for a value outside the accepted list and
// sql.ErrNoRows when the file is not a known video.
func (s *Service) SetVideoClassification(fileID int, classification *string) (VideoClassificationDto, error) {
	if classification != nil && !manualClassifications[*classification] {
		return VideoClassificationDto{}, &VideoFieldValidationError{Field: "classification", MessageKey: "ERROR_VIDEO_INVALID_CLASSIFICATION"}
	}
	if _, err := s.Repository.GetVideoFileByID(fileID); err != nil {
		return VideoClassificationDto{}, err
	}

	err := s.withTransaction(func(tx *sql.Tx) error {
		if err := s.Repository.UpsertVideoClassificationOverride(tx, fileID, nullStringFromPointer(classification)); err != nil {
			return err
		}
		return s.Repository.DeleteEmptyVideoManualOverride(tx, fileID)
	})
	if err != nil {
		return VideoClassificationDto{}, err
	}

	s.scheduleSmartPlaylistRebuild()
	return s.effectiveClassification(fileID)
}

func (s *Service) effectiveClassification(fileID int) (VideoClassificationDto, error) {
	state, err := s.Repository.GetVideoClassificationState(fileID)
	if err != nil {
		return VideoClassificationDto{}, err
	}
	if state.OverrideClassification.Valid {
		return VideoClassificationDto{FileID: fileID, Classification: state.OverrideClassification.String, IsOverride: true}, nil
	}
	return VideoClassificationDto{FileID: fileID, Classification: state.AutomaticClassification.String}, nil
}

// SetVideoEpisode stores the owner's season and episode of one video (both nil
// clear the correction), renumbers every automatic series playlist holding the
// video and returns the effective numbering. It returns
// *VideoFieldValidationError for a season below 0 or an episode below 1 and
// sql.ErrNoRows when the file is not a known video.
func (s *Service) SetVideoEpisode(fileID int, season *int, episode *int) (VideoEpisodeDto, error) {
	if season != nil && *season < minimumManualSeason {
		return VideoEpisodeDto{}, &VideoFieldValidationError{Field: "season", MessageKey: "ERROR_VIDEO_INVALID_SEASON"}
	}
	if episode != nil && *episode < minimumManualEpisode {
		return VideoEpisodeDto{}, &VideoFieldValidationError{Field: "episode", MessageKey: "ERROR_VIDEO_INVALID_EPISODE"}
	}
	if _, err := s.Repository.GetVideoFileByID(fileID); err != nil {
		return VideoEpisodeDto{}, err
	}

	err := s.withTransaction(func(tx *sql.Tx) error {
		if err := s.Repository.UpsertVideoEpisodeOverride(tx, fileID, nullInt64FromPointer(season), nullInt64FromPointer(episode)); err != nil {
			return err
		}
		return s.Repository.DeleteEmptyVideoManualOverride(tx, fileID)
	})
	if err != nil {
		return VideoEpisodeDto{}, err
	}

	if err := s.renumberSeriesPlaylistsOfVideo(fileID); err != nil {
		return VideoEpisodeDto{}, err
	}
	return s.effectiveEpisode(fileID)
}

func (s *Service) renumberSeriesPlaylistsOfVideo(videoID int) error {
	playlistIDs, err := s.Repository.GetAutoSeriesPlaylistIDsByVideo(videoID)
	if err != nil {
		return err
	}
	for _, playlistID := range playlistIDs {
		err := s.RenumberSeriesPlaylist(playlistID)
		if err != nil && !errors.Is(err, ErrPlaylistNotAutomaticSeries) {
			return err
		}
	}
	return nil
}

func (s *Service) effectiveEpisode(fileID int) (VideoEpisodeDto, error) {
	state, err := s.Repository.GetVideoEpisodeState(fileID)
	if err != nil {
		return VideoEpisodeDto{}, err
	}
	automatic := playlist.EpisodeNumbering{
		Season:  nullableInt(state.AutomaticSeason),
		Episode: nullableInt(state.AutomaticEpisode),
	}
	if automatic.Season == nil && automatic.Episode == nil {
		automatic = playlist.AutomaticEpisodeNumbering(playlist.VideoEntry{Name: state.Name})
	}
	override := playlist.EpisodeNumbering{
		Season:  nullableInt(state.OverrideSeason),
		Episode: nullableInt(state.OverrideEpisode),
	}
	effective := playlist.EffectiveNumbering(automatic, override)
	return VideoEpisodeDto{FileID: fileID, Season: effective.Season, Episode: effective.Episode}, nil
}

func (s *Service) scheduleSmartPlaylistRebuild() {
	if !s.isRebuildScheduled.CompareAndSwap(false, true) {
		return
	}
	s.pendingRebuilds.Add(1)
	go func() {
		defer s.pendingRebuilds.Done()
		s.isRebuildScheduled.Store(false)
		if err := s.RebuildSmartPlaylists(); err != nil {
			applog.ErrorWithStack("video: rebuild after classification change failed", err)
		}
	}()
}

func nullStringFromPointer(text *string) sql.NullString {
	if text == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *text, Valid: true}
}
