package captures

import (
	"encoding/json"
	"nas-go/api/pkg/utils"
	"time"
)

type CaptureDto struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	FileName   string    `json:"file_name"`
	FilePath   string    `json:"file_path"`
	MediaType  string    `json:"media_type"`
	MimeType   string    `json:"mime_type"`
	Size       int64     `json:"size"`
	EpisodeKey string    `json:"episode_key"`
	CreatedAt  time.Time `json:"created_at"`

	FileID        *int     `json:"file_id,omitempty"`
	Status        string   `json:"status"`
	Title         string   `json:"title,omitempty"`
	EpisodeTitle  string   `json:"episode_title,omitempty"`
	Season        *int     `json:"season,omitempty"`
	Episode       *int     `json:"episode,omitempty"`
	Description   string   `json:"description,omitempty"`
	ReleaseYear   *int     `json:"release_year,omitempty"`
	Genres        []string `json:"genres,omitempty"`
	Cast          []string `json:"cast,omitempty"`
	Directors     []string `json:"directors,omitempty"`
	Studio        string   `json:"studio,omitempty"`
	ContentRating string   `json:"content_rating,omitempty"`
	Platform      string   `json:"platform,omitempty"`
	SourceURL     string   `json:"source_url,omitempty"`
	ThumbnailURL  string   `json:"thumbnail_url,omitempty"`
	ContentType   string   `json:"content_type,omitempty"`
}

type CreateCaptureDto struct {
	Name       string `json:"name" form:"name" binding:"required"`
	MediaType  string `json:"media_type" form:"media_type"`
	MimeType   string `json:"mime_type" form:"mime_type"`
	Size       int64  `json:"size" form:"size"`
	EpisodeKey string `json:"episode_key" form:"episode_key"`
}

type InitCaptureUploadDto struct {
	Name       string `json:"name" binding:"required"`
	MediaType  string `json:"media_type"`
	MimeType   string `json:"mime_type"`
	Size       int64  `json:"size"`
	FileName   string `json:"file_name"`
	EpisodeKey string `json:"episode_key"`
	// Metadata is an opaque, standardized JSON object the client (browser plugin)
	// builds from the source site (title, episode, duration, cast, next episode,
	// origin, …). The server does not interpret it — it persists it verbatim as
	// metadata.json next to the recording. Keeping it opaque means new fields need
	// no backend change; the client owns the schema and the per-site de→para.
	Metadata json.RawMessage `json:"metadata,omitempty"`
}

// InitCaptureUploadResultDto answers an upload init. When EpisodeKey-based
// idempotency kicks in, the optional fields steer the client:
//   - AlreadyComplete=true: a capture with this episode_key already exists; do
//     not re-record. UploadID is empty.
//   - Resumed=true: an open session for this episode_key was reused; UploadID is
//     the existing one and ReceivedSize is how much the server already holds, so
//     the client appends from that offset instead of opening a second file.
type InitCaptureUploadResultDto struct {
	UploadID        string `json:"upload_id"`
	ChunkSize       int64  `json:"chunk_size"`
	ReceivedSize    int64  `json:"received_size"`
	Resumed         bool   `json:"resumed"`
	AlreadyComplete bool   `json:"already_complete"`
}

type UploadCaptureChunkDto struct {
	UploadID string `json:"upload_id" form:"upload_id" binding:"required"`
	Offset   int64  `json:"offset" form:"offset"`
}

type CompleteCaptureUploadDto struct {
	UploadID string `json:"upload_id" binding:"required"`
}

type CaptureFilter struct {
	Name      utils.Optional[string]
	MediaType utils.Optional[string]
}

func (m *CaptureModel) ToDto() CaptureDto {
	return CaptureDto{
		ID:            m.ID,
		Name:          m.Name,
		FileName:      m.FileName,
		FilePath:      m.FilePath,
		MediaType:     m.MediaType,
		MimeType:      m.MimeType,
		Size:          m.Size,
		EpisodeKey:    m.EpisodeKey,
		CreatedAt:     m.CreatedAt,
		FileID:        m.FileID,
		Status:        string(m.Status),
		Title:         m.Title,
		EpisodeTitle:  m.EpisodeTitle,
		Season:        m.Season,
		Episode:       m.Episode,
		Description:   m.Description,
		ReleaseYear:   m.ReleaseYear,
		Genres:        m.Genres,
		Cast:          m.Cast,
		Directors:     m.Directors,
		Studio:        m.Studio,
		ContentRating: m.ContentRating,
		Platform:      m.Platform,
		SourceURL:     m.SourceURL,
		ThumbnailURL:  m.ThumbnailURL,
		ContentType:   m.ContentType,
	}
}

func ParsePaginationToDto(pagination *utils.PaginationResponse[CaptureModel]) utils.PaginationResponse[CaptureDto] {
	result := utils.PaginationResponse[CaptureDto]{
		Items: []CaptureDto{},
		Pagination: utils.Pagination{
			Page:     pagination.Pagination.Page,
			PageSize: pagination.Pagination.PageSize,
			HasNext:  false,
			HasPrev:  false,
		},
	}

	for _, model := range pagination.Items {
		result.Items = append(result.Items, model.ToDto())
	}
	result.Pagination = pagination.Pagination

	return result
}
