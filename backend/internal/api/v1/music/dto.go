package music

import (
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/utils"
	"time"
)

type PlaylistDto struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsSystem    bool      `json:"is_system"`
	IsAuto      bool      `json:"is_auto"`
	Kind        string    `json:"kind"`
	SourceKey   string    `json:"source_key"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	TrackCount  int       `json:"track_count"`
}

type PlaylistTrackDto struct {
	ID       int           `json:"id"`
	Position int           `json:"position"`
	AddedAt  time.Time     `json:"added_at"`
	File     files.FileDto `json:"file"`
}

type CreatePlaylistRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type UpdatePlaylistRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type AddTrackRequest struct {
	FileID int `json:"file_id" binding:"required"`
}

type ReorderTrackRequest struct {
	Tracks []ReorderTrackItem `json:"tracks" binding:"required"`
}

type ReorderTrackItem struct {
	FileID   int `json:"file_id"`
	Position int `json:"position"`
}

type PlayerStateDto struct {
	ID              int       `json:"id"`
	ClientID        string    `json:"client_id"`
	PlaylistID      *int      `json:"playlist_id"`
	CurrentFileID   *int      `json:"current_file_id"`
	CurrentPosition float64   `json:"current_position"`
	Volume          float64   `json:"volume"`
	Shuffle         bool      `json:"shuffle"`
	RepeatMode      string    `json:"repeat_mode"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type UpdatePlayerStateRequest struct {
	PlaylistID      *int    `json:"playlist_id"`
	CurrentFileID   *int    `json:"current_file_id"`
	CurrentPosition float64 `json:"current_position"`
	Volume          float64 `json:"volume"`
	Shuffle         bool    `json:"shuffle"`
	RepeatMode      string  `json:"repeat_mode"`
}

func (m *PlayerStateModel) ToDto() PlayerStateDto {
	dto := PlayerStateDto{
		ID:              m.ID,
		ClientID:        m.ClientID,
		CurrentPosition: m.CurrentPosition,
		Volume:          m.Volume,
		Shuffle:         m.Shuffle,
		RepeatMode:      m.RepeatMode,
		UpdatedAt:       m.UpdatedAt,
	}
	if m.PlaylistID.Valid {
		v := int(m.PlaylistID.Int64)
		dto.PlaylistID = &v
	}
	if m.CurrentFileID.Valid {
		v := int(m.CurrentFileID.Int64)
		dto.CurrentFileID = &v
	}
	return dto
}

func (m *PlaylistModel) ToDto() PlaylistDto {
	kind := PlaylistKindManual
	if m.IsSystem {
		kind = PlaylistKindSystem
	}

	return PlaylistDto{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		IsSystem:    m.IsSystem,
		IsAuto:      false,
		Kind:        kind,
		SourceKey:   "",
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		TrackCount:  m.TrackCount,
	}
}

func (m *PlaylistTrackModel) ToDto() (PlaylistTrackDto, error) {
	fileModel := files.FileModel{
		ID:              m.FileID,
		Name:            m.FileName,
		Path:            m.FilePath,
		ParentPath:      m.FileParentPath,
		Type:            files.FileType(m.FileType),
		Format:          m.FileFormat,
		Size:            m.FileSize,
		UpdatedAt:       m.FileUpdatedAt,
		CreatedAt:       m.FileCreatedAt,
		DeletedAt:       m.FileDeletedAt,
		LastInteraction: m.LastInteraction,
		LastBackup:      m.LastBackup,
		CheckSum:        m.FileCheckSum,
		Starred:         m.FileStarred,
	}

	fileModel.Metadata = files.AudioMetadataModel{
		ID:                  m.MetadataID,
		FileId:              m.MetadataFileId,
		Path:                m.MetadataPath,
		Mime:                m.MetadataMime,
		Length:              m.MetadataLength,
		Bitrate:             m.MetadataBitrate,
		SampleRate:          m.MetadataSampleRate,
		Channels:            m.MetadataChannels,
		BitrateMode:         m.MetadataBitrateMode,
		EncoderInfo:         m.MetadataEncoderInfo,
		BitDepth:            m.MetadataBitDepth,
		Title:               m.MetadataTitle,
		Artist:              m.MetadataArtist,
		Album:               m.MetadataAlbum,
		AlbumArtist:         m.MetadataAlbumArtist,
		TrackNumber:         m.MetadataTrackNumber,
		Genre:               m.MetadataGenre,
		Composer:            m.MetadataComposer,
		Year:                m.MetadataYear,
		RecordingDate:       m.MetadataRecordingDate,
		Encoder:             m.MetadataEncoder,
		Publisher:           m.MetadataPublisher,
		OriginalReleaseDate: m.MetadataOriginalReleaseDate,
		OriginalArtist:      m.MetadataOriginalArtist,
		Lyricist:            m.MetadataLyricist,
		Lyrics:              m.MetadataLyrics,
		CreatedAt:           m.MetadataCreatedAt,
	}

	fileDto, err := fileModel.ToDto()
	if err != nil {
		return PlaylistTrackDto{}, err
	}
	fileDto.Metadata = fileModel.Metadata

	return PlaylistTrackDto{
		ID:       m.ID,
		Position: m.Position,
		AddedAt:  m.AddedAt,
		File:     fileDto,
	}, nil
}

func ParsePlaylistPaginationToDto(pagination *utils.PaginationResponse[PlaylistModel]) utils.PaginationResponse[PlaylistDto] {
	paginationResponse := utils.PaginationResponse[PlaylistDto]{
		Items: []PlaylistDto{},
		Pagination: utils.Pagination{
			Page:     pagination.Pagination.Page,
			PageSize: pagination.Pagination.PageSize,
			HasNext:  pagination.Pagination.HasNext,
			HasPrev:  pagination.Pagination.HasPrev,
		},
	}

	for _, model := range pagination.Items {
		paginationResponse.Items = append(paginationResponse.Items, model.ToDto())
	}

	return paginationResponse
}

func ParseTrackPaginationToDto(pagination *utils.PaginationResponse[PlaylistTrackModel]) (utils.PaginationResponse[PlaylistTrackDto], error) {
	paginationResponse := utils.PaginationResponse[PlaylistTrackDto]{
		Items: []PlaylistTrackDto{},
		Pagination: utils.Pagination{
			Page:     pagination.Pagination.Page,
			PageSize: pagination.Pagination.PageSize,
			HasNext:  pagination.Pagination.HasNext,
			HasPrev:  pagination.Pagination.HasPrev,
		},
	}

	for _, model := range pagination.Items {
		dto, err := model.ToDto()
		if err != nil {
			return paginationResponse, err
		}
		paginationResponse.Items = append(paginationResponse.Items, dto)
	}

	return paginationResponse, nil
}
