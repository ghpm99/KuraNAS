package image

import (
	"errors"
	"time"
)

var (
	ErrAlbumNotFound        = errors.New("image album not found")
	ErrAlbumNameTaken       = errors.New("image album name already in use")
	ErrAlbumNameRequired    = errors.New("image album name is required")
	ErrAlbumNameTooLong     = errors.New("image album name is too long")
	ErrAlbumInvalidID       = errors.New("invalid image album id")
	ErrAlbumInvalidFileIDs  = errors.New("invalid image album file ids")
	ErrAlbumCoverNotInAlbum = errors.New("cover image is not part of the album")
	ErrAlbumNothingToUpdate = errors.New("nothing to update in the image album")
)

const (
	maxAlbumNameLength   = 100
	maxAlbumItemsPerCall = 1000
)

type AlbumModel struct {
	ID          int
	Name        string
	CoverFileID *int
	ItemCount   int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type AlbumUpdate struct {
	AlbumID     int
	Name        *string
	CoverFileID *int
}
