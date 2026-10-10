package video

import (
	"errors"
	"testing"

	"nas-go/api/internal/roots"
)

func useStorageRoots(t *testing.T, storageRoots []roots.Root) {
	t.Helper()
	roots.Set(storageRoots)
	t.Cleanup(roots.Reset)
}

func TestListLibraryFoldersScopesEveryEnabledRootWhenParentIsAbsent(t *testing.T) {
	useStorageRoots(t, []roots.Root{
		{Path: "/data", Label: "Principal", Enabled: true},
		{Path: "/mnt/midia", Label: "Midia", Enabled: true},
	})
	var receivedQuery LibraryFolderQuery
	repository := &videoRepoMock{listLibraryFoldersFn: func(query LibraryFolderQuery) ([]LibraryFolderModel, error) {
		receivedQuery = query
		return []LibraryFolderModel{{Path: "/data/Series", Name: "Series", VideoCount: 3, CoverFileID: 7}}, nil
	}}
	service := &Service{Repository: repository}

	folderPage, err := service.ListLibraryFolders(LibraryFolderRequest{Page: 2, PageSize: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(receivedQuery.Scopes) != 2 || receivedQuery.Scopes[0].IsWholeRoot || !receivedQuery.Scopes[1].IsWholeRoot {
		t.Fatalf("unexpected scopes %+v", receivedQuery.Scopes)
	}
	if receivedQuery.Scopes[0].Prefix != "/data/" || receivedQuery.Scopes[1].Label != "Midia" {
		t.Fatalf("unexpected scopes %+v", receivedQuery.Scopes)
	}
	if receivedQuery.Limit != 11 || receivedQuery.Offset != 10 {
		t.Fatalf("unexpected window limit=%d offset=%d", receivedQuery.Limit, receivedQuery.Offset)
	}
	if len(folderPage.Items) != 1 || folderPage.Items[0].VideoCount != 3 || folderPage.Items[0].CoverFileID != 7 {
		t.Fatalf("unexpected page %+v", folderPage)
	}
}

func TestListLibraryFoldersScopesOnlyTheParentWhenGiven(t *testing.T) {
	useStorageRoots(t, []roots.Root{{Path: "/data", Label: "Principal", Enabled: true}})
	var receivedQuery LibraryFolderQuery
	repository := &videoRepoMock{listLibraryFoldersFn: func(query LibraryFolderQuery) ([]LibraryFolderModel, error) {
		receivedQuery = query
		return nil, nil
	}}
	service := &Service{Repository: repository}

	if _, err := service.ListLibraryFolders(LibraryFolderRequest{ParentPath: "/Series", Page: 1, PageSize: 5}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(receivedQuery.Scopes) != 1 || receivedQuery.Scopes[0].Prefix != "/data/Series/" {
		t.Fatalf("unexpected scopes %+v", receivedQuery.Scopes)
	}
}

func TestListLibraryFoldersReportsHasNextWhenRepositoryReturnsExtraRow(t *testing.T) {
	useStorageRoots(t, []roots.Root{{Path: "/data", Enabled: true}})
	repository := &videoRepoMock{listLibraryFoldersFn: func(query LibraryFolderQuery) ([]LibraryFolderModel, error) {
		return []LibraryFolderModel{{Name: "a"}, {Name: "b"}}, nil
	}}
	service := &Service{Repository: repository}

	folderPage, err := service.ListLibraryFolders(LibraryFolderRequest{Page: 1, PageSize: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(folderPage.Items) != 1 || !folderPage.Pagination.HasNext {
		t.Fatalf("expected one item and HasNext, got %+v", folderPage)
	}
}

func TestListLibraryFoldersWithoutRootsSkipsRepository(t *testing.T) {
	useStorageRoots(t, nil)
	service := &Service{Repository: &videoRepoMock{listLibraryFoldersFn: func(LibraryFolderQuery) ([]LibraryFolderModel, error) {
		t.Fatal("repository must not be called without scopes")
		return nil, nil
	}}}

	folderPage, err := service.ListLibraryFolders(LibraryFolderRequest{Page: 1, PageSize: 5})
	if err != nil || len(folderPage.Items) != 0 {
		t.Fatalf("unexpected page %+v err %v", folderPage, err)
	}
}

func TestListLibraryFoldersPropagatesRepositoryError(t *testing.T) {
	useStorageRoots(t, []roots.Root{{Path: "/data", Enabled: true}})
	expectedErr := errors.New("boom")
	service := &Service{Repository: &videoRepoMock{listLibraryFoldersFn: func(LibraryFolderQuery) ([]LibraryFolderModel, error) {
		return nil, expectedErr
	}}}

	if _, err := service.ListLibraryFolders(LibraryFolderRequest{Page: 1, PageSize: 5}); !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestListLibraryFolderVideosQueriesAbsoluteFolderAndMapsDtos(t *testing.T) {
	useStorageRoots(t, []roots.Root{{Path: "/data", Enabled: true}})
	var receivedFolder string
	var receivedLimit, receivedOffset int
	repository := &videoRepoMock{listLibraryFolderVideosFn: func(folderPath string, limit int, offset int) ([]VideoFileModel, error) {
		receivedFolder, receivedLimit, receivedOffset = folderPath, limit, offset
		return []VideoFileModel{{ID: 1, Name: "e1.mkv"}, {ID: 2, Name: "e2.mkv"}}, nil
	}}
	service := &Service{Repository: repository}

	videoPage, err := service.ListLibraryFolderVideos(LibraryFolderVideosRequest{FolderPath: "/Series", Page: 1, PageSize: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedFolder != "/data/Series" || receivedLimit != 2 || receivedOffset != 0 {
		t.Fatalf("unexpected query folder=%s limit=%d offset=%d", receivedFolder, receivedLimit, receivedOffset)
	}
	if len(videoPage.Items) != 1 || videoPage.Items[0].ID != 1 || !videoPage.Pagination.HasNext {
		t.Fatalf("unexpected page %+v", videoPage)
	}
}

func TestListLibraryFolderVideosPropagatesRepositoryError(t *testing.T) {
	expectedErr := errors.New("boom")
	service := &Service{Repository: &videoRepoMock{listLibraryFolderVideosFn: func(string, int, int) ([]VideoFileModel, error) {
		return nil, expectedErr
	}}}

	if _, err := service.ListLibraryFolderVideos(LibraryFolderVideosRequest{FolderPath: "/x", Page: 1, PageSize: 5}); !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}
