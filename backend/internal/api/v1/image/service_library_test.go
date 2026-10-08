package image

import (
	"errors"
	"testing"
	"time"

	"nas-go/api/internal/roots"
)

type fakeLibraryRepository struct {
	items       []LibraryItemModel
	total       int
	buckets     []LibraryTimelineBucketModel
	err         error
	listQuery   LibraryListQuery
	countFilter LibraryFilter
}

func (f *fakeLibraryRepository) ListLibraryImages(query LibraryListQuery) ([]LibraryItemModel, error) {
	f.listQuery = query
	return f.items, f.err
}

func (f *fakeLibraryRepository) CountLibraryImages(filter LibraryFilter) (int, error) {
	f.countFilter = filter
	return f.total, f.err
}

func (f *fakeLibraryRepository) ListLibraryTimeline(filter LibraryFilter) ([]LibraryTimelineBucketModel, error) {
	f.countFilter = filter
	return f.buckets, f.err
}

func useLibraryTestRoot(t *testing.T) {
	t.Helper()
	roots.Set([]roots.Root{{Path: "/data", Label: "data", Enabled: true}})
	t.Cleanup(roots.Reset)
}

func libraryTestItems(count int) []LibraryItemModel {
	takenAt := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	items := make([]LibraryItemModel, 0, count)
	for position := 0; position < count; position++ {
		items = append(items, LibraryItemModel{
			FileID:     100 - position,
			Name:       "a.jpg",
			Path:       "/data/photos/a.jpg",
			ParentPath: "/data/photos",
			TakenAt:    &takenAt,
			Category:   "photo",
			IsCold:     position == 0,
		})
	}
	return items
}

func TestListLibraryImagesKeysetBuildsNextCursorAndRelativePaths(t *testing.T) {
	useLibraryTestRoot(t)
	repository := &fakeLibraryRepository{items: libraryTestItems(3)}
	service := NewLibraryService(repository)

	page, err := service.ListLibraryImages(LibraryListRequest{
		Filter:   LibraryFilter{Folder: "/photos"},
		Sort:     LibrarySortTakenAt,
		Order:    LibrarySortOrderDesc,
		Page:     4,
		PageSize: 2,
	})
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	if repository.listQuery.Limit != 3 || repository.listQuery.Offset != 0 {
		t.Fatalf("keyset must fetch pageSize+1 without offset, got %+v", repository.listQuery)
	}
	if repository.listQuery.Filter.Folder != "/data/photos" {
		t.Fatalf("folder was not resolved to disk: %q", repository.listQuery.Filter.Folder)
	}
	if len(page.Items) != 2 || !page.HasNext || page.Page != 0 || page.PageSize != 2 {
		t.Fatalf("unexpected page %+v", page)
	}
	if page.Items[0].Path != "/photos/a.jpg" || page.Items[0].ParentPath != "/photos" || page.Items[0].Tier != "cold" || page.Items[1].Tier != "hot" {
		t.Fatalf("unexpected item dto %+v", page.Items)
	}

	cursor, err := DecodeLibraryCursor(page.NextCursor)
	if err != nil || cursor.FileID != 99 {
		t.Fatalf("next cursor should point at the last returned item, got %+v err %v", cursor, err)
	}
}

func TestListLibraryImagesKeysetLastPageHasNoCursor(t *testing.T) {
	useLibraryTestRoot(t)
	service := NewLibraryService(&fakeLibraryRepository{items: libraryTestItems(2)})

	page, err := service.ListLibraryImages(LibraryListRequest{Sort: LibrarySortTakenAt, Order: LibrarySortOrderDesc, Page: 1, PageSize: 2})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if page.HasNext || page.NextCursor != "" || len(page.Items) != 2 {
		t.Fatalf("unexpected page %+v", page)
	}
}

func TestListLibraryImagesOffsetOrderingUsesPageAndNoCursor(t *testing.T) {
	useLibraryTestRoot(t)
	repository := &fakeLibraryRepository{items: libraryTestItems(3)}
	service := NewLibraryService(repository)

	page, err := service.ListLibraryImages(LibraryListRequest{Sort: LibrarySortName, Order: LibrarySortOrderAsc, Page: 3, PageSize: 2})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if repository.listQuery.Offset != 4 || repository.listQuery.Limit != 3 || repository.listQuery.Cursor != nil {
		t.Fatalf("unexpected query %+v", repository.listQuery)
	}
	if page.Page != 3 || !page.HasNext || page.NextCursor != "" {
		t.Fatalf("unexpected page %+v", page)
	}
}

func TestLibraryServiceCountAndTimeline(t *testing.T) {
	useLibraryTestRoot(t)
	repository := &fakeLibraryRepository{total: 12, buckets: []LibraryTimelineBucketModel{{Year: 2022, Month: 3, Count: 4}}}
	service := NewLibraryService(repository)

	count, err := service.CountLibraryImages(LibraryFilter{Folder: "/photos"})
	if err != nil || count.Total != 12 || repository.countFilter.Folder != "/data/photos" {
		t.Fatalf("count = %+v err %v filter %+v", count, err, repository.countFilter)
	}

	buckets, err := service.ListLibraryTimeline(LibraryFilter{})
	if err != nil || len(buckets) != 1 || buckets[0] != (LibraryTimelineBucketDto{Year: 2022, Month: 3, Count: 4}) {
		t.Fatalf("buckets = %+v err %v", buckets, err)
	}
}

func TestLibraryServicePropagatesRepositoryErrors(t *testing.T) {
	repositoryErr := errors.New("db down")
	service := NewLibraryService(&fakeLibraryRepository{err: repositoryErr})

	if _, err := service.ListLibraryImages(LibraryListRequest{Sort: LibrarySortTakenAt, Order: LibrarySortOrderDesc, Page: 1, PageSize: 1}); !errors.Is(err, repositoryErr) {
		t.Fatalf("list: %v", err)
	}
	if _, err := service.CountLibraryImages(LibraryFilter{}); !errors.Is(err, repositoryErr) {
		t.Fatalf("count: %v", err)
	}
	if _, err := service.ListLibraryTimeline(LibraryFilter{}); !errors.Is(err, repositoryErr) {
		t.Fatalf("timeline: %v", err)
	}
}
