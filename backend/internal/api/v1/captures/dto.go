package captures

import (
	"nas-go/api/pkg/utils"
	"time"
)

type CaptureDto struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	FileName  string    `json:"file_name"`
	FilePath  string    `json:"file_path"`
	MediaType string    `json:"media_type"`
	MimeType  string    `json:"mime_type"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateCaptureDto struct {
	Name      string `json:"name" form:"name" binding:"required"`
	MediaType string `json:"media_type" form:"media_type"`
	MimeType  string `json:"mime_type" form:"mime_type"`
	Size      int64  `json:"size" form:"size"`
}

type CaptureFilter struct {
	Name      utils.Optional[string]
	MediaType utils.Optional[string]
}

func (m *CaptureModel) ToDto() CaptureDto {
	return CaptureDto{
		ID:        m.ID,
		Name:      m.Name,
		FileName:  m.FileName,
		FilePath:  m.FilePath,
		MediaType: m.MediaType,
		MimeType:  m.MimeType,
		Size:      m.Size,
		CreatedAt: m.CreatedAt,
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
