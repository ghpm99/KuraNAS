package image

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"nas-go/api/pkg/utils"
)

type AlbumService struct {
	repository AlbumRepositoryInterface
	library    LibraryServiceInterface
}

func NewAlbumService(repository AlbumRepositoryInterface, library LibraryServiceInterface) *AlbumService {
	return &AlbumService{repository: repository, library: library}
}

func normalizeAlbumName(rawName string) (string, error) {
	name := strings.TrimSpace(rawName)
	if name == "" {
		return "", ErrAlbumNameRequired
	}
	if utf8.RuneCountInString(name) > maxAlbumNameLength {
		return "", ErrAlbumNameTooLong
	}
	return name, nil
}

func normalizeAlbumFileIDs(fileIDs []int) ([]int, error) {
	if len(fileIDs) == 0 || len(fileIDs) > maxAlbumItemsPerCall {
		return nil, ErrAlbumInvalidFileIDs
	}
	uniqueFileIDs := make([]int, 0, len(fileIDs))
	seenFileIDs := make(map[int]struct{}, len(fileIDs))
	for _, fileID := range fileIDs {
		if fileID < 1 {
			return nil, ErrAlbumInvalidFileIDs
		}
		if _, isDuplicate := seenFileIDs[fileID]; isDuplicate {
			continue
		}
		seenFileIDs[fileID] = struct{}{}
		uniqueFileIDs = append(uniqueFileIDs, fileID)
	}
	return uniqueFileIDs, nil
}

func toAlbumDto(album AlbumModel) AlbumDto {
	return AlbumDto{
		ID:          album.ID,
		Name:        album.Name,
		CoverFileID: album.CoverFileID,
		ItemCount:   album.ItemCount,
		CreatedAt:   album.CreatedAt,
		UpdatedAt:   album.UpdatedAt,
	}
}

func (s *AlbumService) ListAlbums(request AlbumListRequest) (utils.PaginationResponse[AlbumDto], error) {
	albums, err := s.repository.ListAlbums(request.PageSize+1, (request.Page-1)*request.PageSize)
	if err != nil {
		return utils.PaginationResponse[AlbumDto]{}, fmt.Errorf("ListAlbums: %w", err)
	}

	hasNext := len(albums) > request.PageSize
	if hasNext {
		albums = albums[:request.PageSize]
	}

	albumDtos := make([]AlbumDto, 0, len(albums))
	for _, album := range albums {
		albumDtos = append(albumDtos, toAlbumDto(album))
	}
	return utils.PaginationResponse[AlbumDto]{
		Items: albumDtos,
		Pagination: utils.Pagination{
			Page:     request.Page,
			PageSize: request.PageSize,
			HasNext:  hasNext,
			HasPrev:  request.Page > 1,
		},
	}, nil
}

func (s *AlbumService) GetAlbum(albumID int) (AlbumDto, error) {
	return s.getAlbumDto(albumID)
}

func (s *AlbumService) CreateAlbum(rawName string) (AlbumDto, error) {
	name, err := normalizeAlbumName(rawName)
	if err != nil {
		return AlbumDto{}, err
	}
	albumID, err := s.repository.CreateAlbum(name)
	if err != nil {
		return AlbumDto{}, fmt.Errorf("CreateAlbum: %w", err)
	}
	return s.getAlbumDto(albumID)
}

func (s *AlbumService) UpdateAlbum(update AlbumUpdate) (AlbumDto, error) {
	if update.Name == nil && update.CoverFileID == nil {
		return AlbumDto{}, ErrAlbumNothingToUpdate
	}
	if update.CoverFileID != nil && *update.CoverFileID < 1 {
		return AlbumDto{}, ErrAlbumInvalidFileIDs
	}
	if update.Name != nil {
		name, err := normalizeAlbumName(*update.Name)
		if err != nil {
			return AlbumDto{}, err
		}
		update.Name = &name
	}
	if _, err := s.repository.GetAlbum(update.AlbumID); err != nil {
		return AlbumDto{}, fmt.Errorf("UpdateAlbum: %w", err)
	}
	if err := s.repository.UpdateAlbum(update); err != nil {
		return AlbumDto{}, fmt.Errorf("UpdateAlbum: %w", err)
	}
	return s.getAlbumDto(update.AlbumID)
}

func (s *AlbumService) DeleteAlbum(albumID int) error {
	if err := s.repository.DeleteAlbum(albumID); err != nil {
		return fmt.Errorf("DeleteAlbum: %w", err)
	}
	return nil
}

func (s *AlbumService) AddAlbumItems(albumID int, fileIDs []int) (AlbumItemsChangeDto, error) {
	return s.changeAlbumItems(albumID, fileIDs, s.repository.AddAlbumItems)
}

func (s *AlbumService) RemoveAlbumItems(albumID int, fileIDs []int) (AlbumItemsChangeDto, error) {
	return s.changeAlbumItems(albumID, fileIDs, s.repository.RemoveAlbumItems)
}

func (s *AlbumService) changeAlbumItems(albumID int, fileIDs []int, change func(int, []int) (int, error)) (AlbumItemsChangeDto, error) {
	uniqueFileIDs, err := normalizeAlbumFileIDs(fileIDs)
	if err != nil {
		return AlbumItemsChangeDto{}, err
	}
	if _, err := s.repository.GetAlbum(albumID); err != nil {
		return AlbumItemsChangeDto{}, fmt.Errorf("changeAlbumItems: %w", err)
	}
	changedCount, err := change(albumID, uniqueFileIDs)
	if err != nil {
		return AlbumItemsChangeDto{}, fmt.Errorf("changeAlbumItems: %w", err)
	}
	return AlbumItemsChangeDto{Requested: len(uniqueFileIDs), Changed: changedCount}, nil
}

func (s *AlbumService) ListAlbumItems(request AlbumItemsRequest) (LibraryPageDto, error) {
	if _, err := s.repository.GetAlbum(request.AlbumID); err != nil {
		return LibraryPageDto{}, fmt.Errorf("ListAlbumItems: %w", err)
	}
	listing := request.Listing
	listing.Filter.AlbumID = request.AlbumID
	page, err := s.library.ListLibraryImages(listing)
	if err != nil {
		return LibraryPageDto{}, fmt.Errorf("ListAlbumItems: %w", err)
	}
	return page, nil
}

func (s *AlbumService) getAlbumDto(albumID int) (AlbumDto, error) {
	album, err := s.repository.GetAlbum(albumID)
	if err != nil {
		return AlbumDto{}, fmt.Errorf("getAlbumDto: %w", err)
	}
	return toAlbumDto(album), nil
}
