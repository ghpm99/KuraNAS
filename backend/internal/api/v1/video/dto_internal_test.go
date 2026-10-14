package video

import (
	"database/sql"
	"encoding/json"
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

func TestVideoFileModelToDtoMapsTechnicalMetadata(t *testing.T) {
	populated := VideoFileModel{
		DurationSeconds: sql.NullFloat64{Float64: 5400.25, Valid: true},
		Width:           sql.NullInt64{Int64: 1920, Valid: true},
		Height:          sql.NullInt64{Int64: 1080, Valid: true},
		VideoCodec:      sql.NullString{String: "h264", Valid: true},
	}
	populatedDto := populated.ToDto()
	if populatedDto.DurationSeconds == nil || *populatedDto.DurationSeconds != 5400.25 ||
		populatedDto.Width == nil || *populatedDto.Width != 1920 ||
		populatedDto.Height == nil || *populatedDto.Height != 1080 ||
		populatedDto.VideoCodec == nil || *populatedDto.VideoCodec != "h264" {
		t.Fatalf("populated technical metadata not mapped: %+v", populatedDto)
	}

	empty := VideoFileModel{}
	emptyJSON, err := json.Marshal(empty.ToDto())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(emptyJSON, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, field := range []string{"duration_seconds", "width", "height", "video_codec"} {
		value, isPresent := decoded[field]
		if !isPresent || value != nil {
			t.Fatalf("%s must be present and null, got %v (present=%v)", field, value, isPresent)
		}
	}
}
