package video

import (
	"database/sql"
	"errors"
	"sort"

	"nas-go/api/internal/api/v1/video/playlist"
)

var ErrPlaylistNotAutomaticSeries = errors.New("playlist is not an automatic series playlist")

func (s *Service) loadEpisodeOverrides(groups []playlist.SmartGroup) (map[int]playlist.EpisodeNumbering, error) {
	seriesVideoIDs := []int{}
	for _, group := range groups {
		if playlist.IsSeriesLikePlaylist(group.PlaylistType, group.Classification) {
			seriesVideoIDs = append(seriesVideoIDs, group.VideoIDs...)
		}
	}

	overrides, err := s.Repository.GetVideoEpisodeOverrides(seriesVideoIDs)
	if err != nil {
		return nil, err
	}

	overrideNumberingByVideoID := make(map[int]playlist.EpisodeNumbering, len(overrides))
	for videoID, override := range overrides {
		overrideNumberingByVideoID[videoID] = playlist.EpisodeNumbering{
			Season:  nullableInt(override.Season),
			Episode: nullableInt(override.Episode),
		}
	}
	return overrideNumberingByVideoID, nil
}

func buildSeriesPlaylistItems(
	videoIDs []int,
	entriesByVideoID map[int]playlist.VideoEntry,
	overrideNumberingByVideoID map[int]playlist.EpisodeNumbering,
) []SeriesPlaylistItemInput {
	automaticByVideoID := make(map[int]playlist.EpisodeNumbering, len(videoIDs))
	episodes := make([]playlist.NumberedEpisode, 0, len(videoIDs))
	for _, videoID := range videoIDs {
		entry := entriesByVideoID[videoID]
		automatic := playlist.AutomaticEpisodeNumbering(entry)
		automaticByVideoID[videoID] = automatic
		episodes = append(episodes, playlist.NumberedEpisode{
			VideoID:   videoID,
			Name:      entry.Name,
			Numbering: playlist.EffectiveNumbering(automatic, overrideNumberingByVideoID[videoID]),
		})
	}
	playlist.SortEpisodesByNumbering(episodes)

	items := make([]SeriesPlaylistItemInput, 0, len(episodes))
	for _, episode := range episodes {
		automatic := automaticByVideoID[episode.VideoID]
		items = append(items, SeriesPlaylistItemInput{
			VideoID: episode.VideoID,
			Season:  nullInt64FromPointer(automatic.Season),
			Episode: nullInt64FromPointer(automatic.Episode),
		})
	}
	return items
}

func nullInt64FromPointer(number *int) sql.NullInt64 {
	if number == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(*number), Valid: true}
}

// RenumberSeriesPlaylist re-sorts the automatic items of one automatic series
// playlist by effective season and episode (manual override first, automatic
// value otherwise), reusing the order positions the automatic items already
// occupy so manual items keep theirs. It returns ErrPlaylistNotAutomaticSeries
// for any other playlist.
func (s *Service) RenumberSeriesPlaylist(playlistID int) error {
	seriesPlaylist, err := s.Repository.GetVideoPlaylistByID(playlistID)
	if err != nil {
		return err
	}
	if !isAutomaticSeries(seriesPlaylist) {
		return ErrPlaylistNotAutomaticSeries
	}

	return s.withTransaction(func(tx *sql.Tx) error {
		autoItems, err := s.Repository.GetAutoPlaylistItemsForRenumbering(tx, playlistID)
		if err != nil {
			return err
		}

		itemIDByVideoID := make(map[int]int, len(autoItems))
		occupiedOrderIndexes := make([]int, 0, len(autoItems))
		episodes := make([]playlist.NumberedEpisode, 0, len(autoItems))
		for _, item := range autoItems {
			itemIDByVideoID[item.VideoID] = item.ItemID
			occupiedOrderIndexes = append(occupiedOrderIndexes, item.OrderIndex)
			automatic := playlist.EpisodeNumbering{
				Season:  nullableInt(item.AutomaticSeason),
				Episode: nullableInt(item.AutomaticEpisode),
			}
			override := playlist.EpisodeNumbering{
				Season:  nullableInt(item.OverrideSeason),
				Episode: nullableInt(item.OverrideEpisode),
			}
			episodes = append(episodes, playlist.NumberedEpisode{
				VideoID:   item.VideoID,
				Name:      item.Name,
				Numbering: playlist.EffectiveNumbering(automatic, override),
			})
		}
		sort.Ints(occupiedOrderIndexes)
		playlist.SortEpisodesByNumbering(episodes)

		sortedItemIDs := make([]int, 0, len(episodes))
		for _, episode := range episodes {
			sortedItemIDs = append(sortedItemIDs, itemIDByVideoID[episode.VideoID])
		}
		return s.Repository.UpdatePlaylistItemOrder(tx, playlistID, sortedItemIDs, occupiedOrderIndexes)
	})
}
