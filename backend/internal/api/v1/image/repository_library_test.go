package image

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"nas-go/api/pkg/database"

	"github.com/DATA-DOG/go-sqlmock"
)

func newLibraryRepoWithMock(t *testing.T) (*LibraryRepository, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	return NewLibraryRepository(database.NewDbContext(db)), mock, db
}

var libraryItemColumns = []string{"id", "name", "path", "parent_path", "format", "size", "width", "height", "taken_at", "category", "starred", "is_cold", "updated_at"}

func TestListLibraryImagesScansDatedAndUndatedRows(t *testing.T) {
	repo, mock, db := newLibraryRepoWithMock(t)
	defer db.Close()
	takenAt := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows(libraryItemColumns).
		AddRow(2, "b.jpg", "/p/b.jpg", "/p", ".jpg", 10, 100, 50, takenAt, "photo", true, false, takenAt).
		AddRow(1, "a.jpg", "/p/a.jpg", "/p", ".jpg", 20, 0, 0, nil, "other", false, true, takenAt))
	mock.ExpectRollback()

	items, err := repo.ListLibraryImages(LibraryListQuery{Sort: LibrarySortTakenAt, Order: LibrarySortOrderDesc, Limit: 10})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(items) != 2 || items[0].TakenAt == nil || items[1].TakenAt != nil || !items[1].IsCold || !items[0].Starred {
		t.Fatalf("unexpected items %+v", items)
	}
}

func TestListLibraryImagesWrapsQueryAndSortErrors(t *testing.T) {
	repo, mock, db := newLibraryRepoWithMock(t)
	defer db.Close()
	queryErr := errors.New("boom")

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT").WillReturnError(queryErr)
	mock.ExpectRollback()

	if _, err := repo.ListLibraryImages(LibraryListQuery{Sort: LibrarySortName, Order: LibrarySortOrderAsc, Limit: 1}); !errors.Is(err, queryErr) {
		t.Fatalf("expected wrapped query error, got %v", err)
	}
	if _, err := repo.ListLibraryImages(LibraryListQuery{Sort: "bogus"}); err == nil {
		t.Fatal("expected error for unsupported sort")
	}
}

func TestListLibraryImagesReportsScanAndRowsErrors(t *testing.T) {
	repo, mock, db := newLibraryRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectRollback()
	if _, err := repo.ListLibraryImages(LibraryListQuery{Sort: LibrarySortSize, Order: LibrarySortOrderAsc, Limit: 1}); err == nil {
		t.Fatal("expected scan error")
	}

	rowsErr := errors.New("rows failed")
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows(libraryItemColumns).RowError(0, rowsErr).AddRow(1, "a", "/a", "/", ".jpg", 1, 1, 1, nil, "other", false, false, time.Now()))
	mock.ExpectRollback()
	if _, err := repo.ListLibraryImages(LibraryListQuery{Sort: LibrarySortSize, Order: LibrarySortOrderAsc, Limit: 1}); !errors.Is(err, rowsErr) {
		t.Fatalf("expected rows error, got %v", err)
	}
}

func TestCountLibraryImages(t *testing.T) {
	repo, mock, db := newLibraryRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery("count").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(321))
	mock.ExpectRollback()
	total, err := repo.CountLibraryImages(LibraryFilter{})
	if err != nil || total != 321 {
		t.Fatalf("total = %d, err = %v", total, err)
	}

	countErr := errors.New("count failed")
	mock.ExpectBegin()
	mock.ExpectQuery("count").WillReturnError(countErr)
	mock.ExpectRollback()
	if _, err := repo.CountLibraryImages(LibraryFilter{}); !errors.Is(err, countErr) {
		t.Fatalf("expected count error, got %v", err)
	}
}

func TestListLibraryTimeline(t *testing.T) {
	repo, mock, db := newLibraryRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery("GROUP BY").WillReturnRows(sqlmock.NewRows([]string{"y", "m", "c"}).AddRow(2022, 5, 3).AddRow(2021, 12, 9))
	mock.ExpectRollback()
	buckets, err := repo.ListLibraryTimeline(LibraryFilter{})
	if err != nil || len(buckets) != 2 || buckets[0].Year != 2022 || buckets[1].Count != 9 {
		t.Fatalf("buckets = %+v, err = %v", buckets, err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery("GROUP BY").WillReturnRows(sqlmock.NewRows([]string{"y"}).AddRow(1))
	mock.ExpectRollback()
	if _, err := repo.ListLibraryTimeline(LibraryFilter{}); err == nil {
		t.Fatal("expected scan error")
	}

	timelineErr := errors.New("timeline failed")
	mock.ExpectBegin()
	mock.ExpectQuery("GROUP BY").WillReturnError(timelineErr)
	mock.ExpectRollback()
	if _, err := repo.ListLibraryTimeline(LibraryFilter{}); !errors.Is(err, timelineErr) {
		t.Fatalf("expected timeline error, got %v", err)
	}
}

func TestListLibraryFoldersScansRowsAndBindsScopes(t *testing.T) {
	repo, mock, db := newLibraryRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery("WITH scope").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "/", 11, 20).
		WillReturnRows(sqlmock.NewRows([]string{"folder_path", "folder_name", "image_count", "cover_file_id"}).
			AddRow("/data/a", "a", 3, 7))
	mock.ExpectRollback()

	folders, err := repo.ListLibraryFolders(LibraryFolderQuery{
		Scopes:    []LibraryFolderScope{{Prefix: "/data/"}, {Prefix: "/mnt/m/", IsWholeRoot: true, Label: "M"}},
		Separator: "/",
		Limit:     11,
		Offset:    20,
	})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(folders) != 1 || folders[0].Name != "a" || folders[0].ImageCount != 3 || folders[0].CoverFileID != 7 {
		t.Fatalf("unexpected folders %+v", folders)
	}
}

func TestListLibraryFoldersWrapsQueryAndScanErrors(t *testing.T) {
	repo, mock, db := newLibraryRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery("WITH scope").WillReturnError(errors.New("boom"))
	mock.ExpectRollback()
	if _, err := repo.ListLibraryFolders(LibraryFolderQuery{Separator: "/", Limit: 1}); err == nil {
		t.Fatal("expected query error")
	}

	mock.ExpectBegin()
	mock.ExpectQuery("WITH scope").WillReturnRows(sqlmock.NewRows([]string{"a", "b", "c", "d"}).AddRow("p", "n", "not-a-number", 1))
	mock.ExpectRollback()
	if _, err := repo.ListLibraryFolders(LibraryFolderQuery{Separator: "/", Limit: 1}); err == nil {
		t.Fatal("expected scan error")
	}
}
