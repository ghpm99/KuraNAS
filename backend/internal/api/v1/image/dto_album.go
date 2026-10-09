package image

import "time"

type AlbumDto struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	CoverFileID *int      `json:"cover_file_id"`
	ItemCount   int       `json:"item_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type AlbumNameBody struct {
	Name string `json:"name"`
}

type AlbumUpdateBody struct {
	Name        *string `json:"name"`
	CoverFileID *int    `json:"cover_file_id"`
}

type AlbumFileIDsBody struct {
	FileIDs []int `json:"file_ids"`
}

type AlbumItemsChangeDto struct {
	Requested int `json:"requested"`
	Changed   int `json:"changed"`
}
