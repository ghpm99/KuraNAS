package image

import (
	"fmt"
	"slices"

	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/roots"
)

type LibraryService struct {
	repository LibraryRepositoryInterface
}

func NewLibraryService(repository LibraryRepositoryInterface) *LibraryService {
	return &LibraryService{repository: repository}
}

func (s *LibraryService) ListLibraryImages(request LibraryListRequest) (LibraryPageDto, error) {
	isKeyset := IsKeysetOrdering(request.Sort, request.Order)
	query := LibraryListQuery{
		Filter:      resolveFolderOnDisk(request.Filter),
		Sort:        request.Sort,
		Order:       request.Order,
		TakenBefore: request.TakenBefore,
		Limit:       request.PageSize + 1,
	}
	if isKeyset {
		query.Cursor = request.Cursor
	} else {
		query.Offset = (request.Page - 1) * request.PageSize
	}

	itemModels, err := s.repository.ListLibraryImages(query)
	if err != nil {
		return LibraryPageDto{}, fmt.Errorf("ListLibraryImages: %w", err)
	}

	hasNext := len(itemModels) > request.PageSize
	if hasNext {
		itemModels = itemModels[:request.PageSize]
	}

	page := LibraryPageDto{
		Items:    toLibraryItemDtos(itemModels),
		HasNext:  hasNext,
		PageSize: request.PageSize,
	}
	if !isKeyset {
		page.Page = request.Page
	}
	if isKeyset && hasNext {
		page.NextCursor = CursorFromItem(itemModels[len(itemModels)-1]).Encode()
	}
	return page, nil
}

func (s *LibraryService) ListLibraryNeighbors(request LibraryNeighborsRequest) (LibraryNeighborsDto, error) {
	pivot, err := s.repository.GetLibraryItemCursor(request.FileID)
	if err != nil {
		return LibraryNeighborsDto{}, fmt.Errorf("ListLibraryNeighbors: %w", err)
	}

	neighborsQuery := LibraryListQuery{
		Filter: resolveFolderOnDisk(request.Filter),
		Sort:   LibrarySortTakenAt,
		Order:  LibrarySortOrderDesc,
		Limit:  request.Count,
	}

	olderQuery := neighborsQuery
	olderQuery.Cursor = &pivot
	olderModels, err := s.repository.ListLibraryImages(olderQuery)
	if err != nil {
		return LibraryNeighborsDto{}, fmt.Errorf("ListLibraryNeighbors: %w", err)
	}

	newerQuery := neighborsQuery
	newerQuery.NewerThan = &pivot
	newerModels, err := s.repository.ListLibraryImages(newerQuery)
	if err != nil {
		return LibraryNeighborsDto{}, fmt.Errorf("ListLibraryNeighbors: %w", err)
	}
	slices.Reverse(newerModels)

	return LibraryNeighborsDto{
		Before: toLibraryItemDtos(newerModels),
		After:  toLibraryItemDtos(olderModels),
	}, nil
}

func (s *LibraryService) CountLibraryImages(filter LibraryFilter) (LibraryCountDto, error) {
	total, err := s.repository.CountLibraryImages(resolveFolderOnDisk(filter))
	if err != nil {
		return LibraryCountDto{}, fmt.Errorf("CountLibraryImages: %w", err)
	}
	return LibraryCountDto{Total: total}, nil
}

func (s *LibraryService) ListLibraryTimeline(filter LibraryFilter) ([]LibraryTimelineBucketDto, error) {
	bucketModels, err := s.repository.ListLibraryTimeline(resolveFolderOnDisk(filter))
	if err != nil {
		return nil, fmt.Errorf("ListLibraryTimeline: %w", err)
	}

	buckets := make([]LibraryTimelineBucketDto, 0, len(bucketModels))
	for _, bucketModel := range bucketModels {
		buckets = append(buckets, LibraryTimelineBucketDto{
			Year:  bucketModel.Year,
			Month: bucketModel.Month,
			Count: bucketModel.Count,
		})
	}
	return buckets, nil
}

func resolveFolderOnDisk(filter LibraryFilter) LibraryFilter {
	if filter.Folder != "" {
		filter.Folder = roots.ToAbsolutePath(filter.Folder)
	}
	return filter
}

func toLibraryItemDtos(itemModels []LibraryItemModel) []LibraryItemDto {
	itemDtos := make([]LibraryItemDto, 0, len(itemModels))
	for _, itemModel := range itemModels {
		tier := files.TierHot
		if itemModel.IsCold {
			tier = files.TierCold
		}
		itemDtos = append(itemDtos, LibraryItemDto{
			FileID:     itemModel.FileID,
			Name:       itemModel.Name,
			Path:       roots.ToRelativePath(itemModel.Path),
			ParentPath: roots.ToRelativePath(itemModel.ParentPath),
			Format:     itemModel.Format,
			Size:       itemModel.Size,
			Width:      itemModel.Width,
			Height:     itemModel.Height,
			TakenAt:    itemModel.TakenAt,
			Category:   itemModel.Category,
			Starred:    itemModel.Starred,
			Tier:       tier,
			UpdatedAt:  itemModel.UpdatedAt,
		})
	}
	return itemDtos
}
