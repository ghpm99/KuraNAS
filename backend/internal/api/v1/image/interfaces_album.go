package image

import "nas-go/api/pkg/utils"

type AlbumRepositoryInterface interface {
	ListAlbums(limit int, offset int) ([]AlbumModel, error)
	GetAlbum(albumID int) (AlbumModel, error)
	CreateAlbum(name string) (int, error)
	UpdateAlbum(update AlbumUpdate) error
	DeleteAlbum(albumID int) error
	AddAlbumItems(albumID int, fileIDs []int) (int, error)
	RemoveAlbumItems(albumID int, fileIDs []int) (int, error)
}

type AlbumListRequest struct {
	Page     int
	PageSize int
}

type AlbumItemsRequest struct {
	AlbumID int
	Listing LibraryListRequest
}

type AlbumServiceInterface interface {
	ListAlbums(request AlbumListRequest) (utils.PaginationResponse[AlbumDto], error)
	GetAlbum(albumID int) (AlbumDto, error)
	CreateAlbum(name string) (AlbumDto, error)
	UpdateAlbum(update AlbumUpdate) (AlbumDto, error)
	DeleteAlbum(albumID int) error
	AddAlbumItems(albumID int, fileIDs []int) (AlbumItemsChangeDto, error)
	RemoveAlbumItems(albumID int, fileIDs []int) (AlbumItemsChangeDto, error)
	ListAlbumItems(request AlbumItemsRequest) (LibraryPageDto, error)
}
