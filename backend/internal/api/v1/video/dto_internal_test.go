package video

import (
	"database/sql"
	"testing"
	"time"
)

func TestVideoPlaylistAndPlaybackToDtoBranches(t *testing.T) {
	now := time.Now()
	coverID := int64(12)

	withRefs := VideoPlaylistModel{
		ID:           1,
		Name:         "playlist",
		CreatedAt:    now,
		UpdatedAt:    now,
		LastPlayedAt: sql.NullTime{Valid: true, Time: now},
		CoverVideoID: sql.NullInt64{Valid: true, Int64: coverID},
	}
	dto := withRefs.ToDto(nil)
	if dto.LastPlayedAt == nil || dto.CoverVideoID == nil || *dto.CoverVideoID != int(coverID) {
		t.Fatalf("expected optional fields to be populated: %+v", dto)
	}

	withoutRefs := VideoPlaylistModel{
		ID:           2,
		Name:         "playlist-2",
		CreatedAt:    now,
		UpdatedAt:    now,
		LastPlayedAt: sql.NullTime{Valid: false},
		CoverVideoID: sql.NullInt64{Valid: false},
	}
	dto2 := withoutRefs.ToDto(nil)
	if dto2.LastPlayedAt != nil || dto2.CoverVideoID != nil {
		t.Fatalf("expected optional fields to be nil: %+v", dto2)
	}

	state := VideoPlaybackStateModel{
		ID:         1,
		ClientID:   "c1",
		PlaylistID: sql.NullInt64{Valid: true, Int64: 7},
		VideoID:    sql.NullInt64{Valid: true, Int64: 9},
	}
	stateDto := state.ToDto()
	if stateDto.PlaylistID == nil || stateDto.VideoID == nil {
		t.Fatalf("expected playback nullable ids to be filled: %+v", stateDto)
	}
}

func TestVideoNewService(t *testing.T) {
	repo := &videoRepoMock{}
	service := NewService(repo)
	_, ok := service.(*Service)
	if !ok {
		t.Fatalf("expected concrete service")
	}
}
