package documenttext

import (
	"errors"
	"testing"

	"nas-go/api/pkg/utils"
)

func TestListDocumentsResolvesFolderToSlashTerminatedPrefix(t *testing.T) {
	repository := &recordingRepository{folderPath: "/a/b", isFolderFound: true}
	service := &Service{Repository: repository}

	if _, err := service.ListDocuments(DocumentListParams{FolderID: 3, Sort: SortName, Page: 1, PageSize: 5}); err != nil {
		t.Fatal(err)
	}
	if repository.gotListQuery.FolderPath != "/a/b/" {
		t.Fatalf("unexpected prefix %q", repository.gotListQuery.FolderPath)
	}
	if _, err := service.CountDocuments(DocumentFilter{}, 3); err != nil || repository.gotCountPath != "/a/b/" {
		t.Fatalf("count prefix %q err %v", repository.gotCountPath, err)
	}
}

func TestListDocumentsReturnsErrFolderNotFoundAndPropagatesFailures(t *testing.T) {
	missing := &Service{Repository: &recordingRepository{}}
	if _, err := missing.ListDocuments(DocumentListParams{FolderID: 3}); !errors.Is(err, ErrFolderNotFound) {
		t.Fatalf("want ErrFolderNotFound, got %v", err)
	}

	failing := &Service{Repository: &recordingRepository{err: errors.New("db down"), findErr: errors.New("db down")}}
	if _, err := failing.ListDocuments(DocumentListParams{}); err == nil {
		t.Fatal("expected repository error")
	}
	if _, err := failing.CountDocuments(DocumentFilter{}, 0); err == nil {
		t.Fatal("expected repository error")
	}
	if _, err := failing.CountDocuments(DocumentFilter{}, 4); err == nil {
		t.Fatal("expected folder lookup error")
	}
}

func TestListDocumentsMapsParentPathAndKeepsPagination(t *testing.T) {
	repository := &recordingRepository{listedPage: utils.PaginationResponse[DocumentListItemModel]{
		Items:      []DocumentListItemModel{{ID: 1, Name: "a.pdf", Format: ".pdf", ParentPath: "/docs", IndexState: StateQueued}},
		Pagination: utils.Pagination{Page: 2, PageSize: 1, HasNext: true, HasPrev: true},
	}}
	service := &Service{Repository: repository}

	page, err := service.ListDocuments(DocumentListParams{Page: 2, PageSize: 1})
	if err != nil || len(page.Items) != 1 || page.Items[0].ParentPath != "/docs" || !page.Pagination.HasNext || !page.Pagination.HasPrev {
		t.Fatalf("unexpected page %+v err %v", page, err)
	}
}
