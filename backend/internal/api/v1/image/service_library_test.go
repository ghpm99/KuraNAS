package image

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"nas-go/api/internal/roots"
)

type fakeLibraryRepository struct {
	items       []LibraryItemModel
	total       int
	buckets     []LibraryTimelineBucketModel
	cameras     []LibraryCameraFacetModel
	formats     []LibraryFormatFacetModel
	err         error
	listQuery   LibraryListQuery
	countFilter LibraryFilter
	folders     []LibraryFolderModel
	folderQuery LibraryFolderQuery
	folderCalls int
	newerItems  []LibraryItemModel
	pivot       LibraryCursor
	pivotErr    error
	listQueries []LibraryListQuery
}

func (f *fakeLibraryRepository) ListLibraryImages(query LibraryListQuery) ([]LibraryItemModel, error) {
	f.listQuery = query
	f.listQueries = append(f.listQueries, query)
	if query.NewerThan != nil {
		return f.newerItems, f.err
	}
	return f.items, f.err
}

func (f *fakeLibraryRepository) GetLibraryItemCursor(fileID int) (LibraryCursor, error) {
	return f.pivot, f.pivotErr
}

func (f *fakeLibraryRepository) CountLibraryImages(filter LibraryFilter) (int, error) {
	f.countFilter = filter
	return f.total, f.err
}

func (f *fakeLibraryRepository) ListLibraryTimeline(filter LibraryFilter) ([]LibraryTimelineBucketModel, error) {
	f.countFilter = filter
	return f.buckets, f.err
}

func (f *fakeLibraryRepository) ListLibraryCameraFacets(filter LibraryFilter) ([]LibraryCameraFacetModel, error) {
	f.countFilter = filter
	return f.cameras, f.err
}

func (f *fakeLibraryRepository) ListLibraryFormatFacets(filter LibraryFilter) ([]LibraryFormatFacetModel, error) {
	f.countFilter = filter
	return f.formats, f.err
}

func (f *fakeLibraryRepository) ListLibraryFolders(query LibraryFolderQuery) ([]LibraryFolderModel, error) {
	f.folderCalls++
	f.folderQuery = query
	return f.folders, f.err
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

func TestListLibraryFoldersScopesParentAndMapsRelativePaths(t *testing.T) {
	useLibraryTestRoot(t)
	repository := &fakeLibraryRepository{folders: []LibraryFolderModel{
		{Path: "/data/photos/a", Name: "a", ImageCount: 3, CoverFileID: 7},
		{Path: "/data/photos/b", Name: "b", ImageCount: 1, CoverFileID: 8},
		{Path: "/data/photos/c", Name: "c", ImageCount: 1, CoverFileID: 9},
	}}
	service := NewLibraryService(repository)

	page, err := service.ListLibraryFolders(LibraryFolderRequest{ParentPath: "/photos", Page: 2, PageSize: 2})
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	if len(repository.folderQuery.Scopes) != 1 || repository.folderQuery.Scopes[0].Prefix != "/data/photos/" || repository.folderQuery.Scopes[0].IsWholeRoot {
		t.Fatalf("unexpected scopes %+v", repository.folderQuery.Scopes)
	}
	if repository.folderQuery.Limit != 3 || repository.folderQuery.Offset != 2 {
		t.Fatalf("unexpected window %+v", repository.folderQuery)
	}
	if len(page.Items) != 2 || !page.Pagination.HasNext || !page.Pagination.HasPrev || page.Items[0].Path != "/photos/a" || page.Items[0].ImageCount != 3 {
		t.Fatalf("unexpected page %+v", page)
	}
}

func TestListLibraryFoldersWithoutParentCoversEveryRoot(t *testing.T) {
	roots.Set([]roots.Root{
		{Path: "/data", Label: "data", Enabled: true},
		{Path: "/mnt/midia", Label: "Midia", Enabled: true},
	})
	t.Cleanup(roots.Reset)
	repository := &fakeLibraryRepository{folders: []LibraryFolderModel{
		{Path: "/mnt/midia", Name: "Midia", ImageCount: 2, CoverFileID: 5},
	}}

	page, err := NewLibraryService(repository).ListLibraryFolders(LibraryFolderRequest{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	scopes := repository.folderQuery.Scopes
	if len(scopes) != 2 || scopes[0].Prefix != "/data/" || scopes[0].IsWholeRoot || scopes[1].Prefix != "/mnt/midia/" || !scopes[1].IsWholeRoot || scopes[1].Label != "Midia" {
		t.Fatalf("unexpected scopes %+v", scopes)
	}
	if page.Items[0].Path != "/Midia" || page.Pagination.HasNext {
		t.Fatalf("unexpected page %+v", page)
	}
}

func TestListLibraryFoldersWithoutRootsSkipsRepository(t *testing.T) {
	roots.Reset()
	repository := &fakeLibraryRepository{}

	page, err := NewLibraryService(repository).ListLibraryFolders(LibraryFolderRequest{Page: 1, PageSize: 10})
	if err != nil || repository.folderCalls != 0 || len(page.Items) != 0 || page.Items == nil {
		t.Fatalf("err %v calls %d page %+v", err, repository.folderCalls, page)
	}
}

func TestListLibraryFoldersWrapsRepositoryError(t *testing.T) {
	useLibraryTestRoot(t)
	repository := &fakeLibraryRepository{err: errors.New("boom")}

	if _, err := NewLibraryService(repository).ListLibraryFolders(LibraryFolderRequest{ParentPath: "/photos", Page: 1, PageSize: 10}); err == nil {
		t.Fatal("expected error")
	}
}

func TestListLibraryNeighborsReturnsNewerBeforeAndOlderAfterInListOrder(t *testing.T) {
	useLibraryTestRoot(t)
	pivotTakenAt := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	repository := &fakeLibraryRepository{
		pivot:      LibraryCursor{TakenAt: &pivotTakenAt, FileID: 50},
		newerItems: []LibraryItemModel{{FileID: 51, Path: "/data/p/a.jpg"}, {FileID: 52, Path: "/data/p/b.jpg"}},
		items:      []LibraryItemModel{{FileID: 49, Path: "/data/p/c.jpg"}},
	}
	service := NewLibraryService(repository)

	neighbors, err := service.ListLibraryNeighbors(LibraryNeighborsRequest{FileID: 50, Filter: LibraryFilter{Folder: "/p"}, Count: 2})
	if err != nil {
		t.Fatalf("neighbors: %v", err)
	}

	if len(neighbors.Before) != 2 || neighbors.Before[0].FileID != 52 || neighbors.Before[1].FileID != 51 {
		t.Fatalf("before must list the farthest newer item first and end next to the pivot, got %+v", neighbors.Before)
	}
	if len(neighbors.After) != 1 || neighbors.After[0].FileID != 49 || neighbors.After[0].Path != "/p/c.jpg" {
		t.Fatalf("unexpected after %+v", neighbors.After)
	}
	if len(repository.listQueries) != 2 {
		t.Fatalf("expected one query per side, got %d", len(repository.listQueries))
	}
	for _, query := range repository.listQueries {
		if query.Limit != 2 || query.Sort != LibrarySortTakenAt || query.Order != LibrarySortOrderDesc || query.Filter.Folder != "/data/p" {
			t.Fatalf("unexpected neighbors query %+v", query)
		}
	}
}

func TestListLibraryNeighborsEmptySidesAreEmptyLists(t *testing.T) {
	service := NewLibraryService(&fakeLibraryRepository{})

	neighbors, err := service.ListLibraryNeighbors(LibraryNeighborsRequest{FileID: 1, Count: 5})
	if err != nil || neighbors.Before == nil || neighbors.After == nil {
		t.Fatalf("sides must be empty lists, got %+v err %v", neighbors, err)
	}
}

func TestListLibraryNeighborsPropagatesPivotAndListErrors(t *testing.T) {
	missing := NewLibraryService(&fakeLibraryRepository{pivotErr: sql.ErrNoRows})
	if _, err := missing.ListLibraryNeighbors(LibraryNeighborsRequest{FileID: 1, Count: 5}); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}

	failing := NewLibraryService(&fakeLibraryRepository{err: errors.New("db down")})
	if _, err := failing.ListLibraryNeighbors(LibraryNeighborsRequest{FileID: 1, Count: 5}); err == nil {
		t.Fatal("expected the list error")
	}
}

type olderFailingLibraryRepository struct {
	fakeLibraryRepository
}

func (f *olderFailingLibraryRepository) ListLibraryImages(query LibraryListQuery) ([]LibraryItemModel, error) {
	if query.NewerThan != nil {
		return nil, errors.New("newer side down")
	}
	return nil, nil
}

func TestListLibraryNeighborsPropagatesNewerSideError(t *testing.T) {
	service := NewLibraryService(&olderFailingLibraryRepository{})
	if _, err := service.ListLibraryNeighbors(LibraryNeighborsRequest{FileID: 1, Count: 5}); err == nil {
		t.Fatal("expected the newer side error")
	}
}

func TestLibraryServiceFacetsMapModelsToDtos(t *testing.T) {
	useLibraryTestRoot(t)
	repository := &fakeLibraryRepository{
		cameras: []LibraryCameraFacetModel{{Camera: "Canon EOS", Count: 5}},
		formats: []LibraryFormatFacetModel{{Format: ".jpg", Count: 9}},
	}
	service := NewLibraryService(repository)

	cameras, err := service.ListLibraryCameraFacets(LibraryFilter{Folder: "data/trip"})
	if err != nil || len(cameras) != 1 || cameras[0] != (LibraryCameraFacetDto{Camera: "Canon EOS", Count: 5}) {
		t.Fatalf("cameras = %+v err %v", cameras, err)
	}
	if repository.countFilter.Folder == "data/trip" {
		t.Fatalf("folder was not resolved on disk: %q", repository.countFilter.Folder)
	}

	formats, err := service.ListLibraryFormatFacets(LibraryFilter{})
	if err != nil || len(formats) != 1 || formats[0] != (LibraryFormatFacetDto{Format: "jpg", Count: 9}) {
		t.Fatalf("formats = %+v err %v", formats, err)
	}
}

func TestLibraryServiceFacetsPropagateRepositoryError(t *testing.T) {
	repositoryErr := errors.New("db down")
	service := NewLibraryService(&fakeLibraryRepository{err: repositoryErr})

	if _, err := service.ListLibraryCameraFacets(LibraryFilter{}); !errors.Is(err, repositoryErr) {
		t.Fatalf("cameras err = %v", err)
	}
	if _, err := service.ListLibraryFormatFacets(LibraryFilter{}); !errors.Is(err, repositoryErr) {
		t.Fatalf("formats err = %v", err)
	}
}
