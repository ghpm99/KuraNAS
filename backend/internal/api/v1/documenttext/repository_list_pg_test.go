package documenttext

import (
	"database/sql"
	"slices"
	"testing"
	"time"

	"nas-go/api/internal/roots"
	"nas-go/api/pkg/database"
)

type libraryEntry struct {
	name      string
	directory string
	size      int64
	updatedAt time.Time
	starred   bool
	isFolder  bool
}

func seedLibraryEntry(t *testing.T, dbContext *database.DbContext, entry libraryEntry) int {
	t.Helper()
	entryType := 2
	if entry.isFolder {
		entryType = 1
	}
	var fileID int
	err := dbContext.ExecTx(func(tx *sql.Tx) error {
		return tx.QueryRow(
			`INSERT INTO home_file (name, path, parent_path, format, size, updated_at, created_at, type, checksum, starred)
			 VALUES ($1::text, $2::text || '/' || $1::text, $2::text, $3::text, $4, $5::timestamptz, $5::timestamptz, $6, '', $7) RETURNING id`,
			entry.name, entry.directory, formatOfName(entry.name), entry.size, entry.updatedAt, entryType, entry.starred,
		).Scan(&fileID)
	})
	if err != nil {
		t.Fatalf("seed %s: %v", entry.name, err)
	}
	return fileID
}

func formatOfName(name string) string {
	for position := len(name) - 1; position >= 0; position-- {
		if name[position] == '.' {
			return name[position:]
		}
	}
	return ""
}

func useTwoRoots(t *testing.T) {
	t.Helper()
	roots.Set([]roots.Root{
		{Path: "/data/main", Label: "main", Enabled: true},
		{Path: "/data/media", Label: "Midia", Enabled: true},
	})
	t.Cleanup(roots.Reset)
}

func listAll(t *testing.T, repository *Repository, filter DocumentFilter, folderPath string) []DocumentListItemModel {
	t.Helper()
	page, err := repository.ListDocuments(DocumentListQuery{Filter: filter, FolderPath: folderPath, Sort: SortName, Order: OrderAscending, Page: 1, PageSize: 100})
	if err != nil {
		t.Fatal(err)
	}
	return page.Items
}

func itemNames(items []DocumentListItemModel) []string {
	names := make([]string, 0, len(items))
	for _, item := range items {
		names = append(names, item.Name)
	}
	return names
}

func assertNames(t *testing.T, items []DocumentListItemModel, want ...string) {
	t.Helper()
	got := itemNames(items)
	if !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestListAndCountDocumentsAcrossFoldersAndRootsExcludeSourceCodeAndLogs_Postgres(t *testing.T) {
	repository, service, dbContext := newDocumentTextPostgres(t)
	useTwoRoots(t)
	now := time.Now().UTC().Truncate(time.Second)
	seedLibraryEntry(t, dbContext, libraryEntry{name: "a.pdf", directory: "/data/main/Casa", size: 10, updatedAt: now})
	seedLibraryEntry(t, dbContext, libraryEntry{name: "b.docx", directory: "/data/main/Trabalho/2024", size: 20, updatedAt: now})
	seedLibraryEntry(t, dbContext, libraryEntry{name: "c.xlsx", directory: "/data/media/docs", size: 30, updatedAt: now})
	seedLibraryEntry(t, dbContext, libraryEntry{name: "main.go", directory: "/data/main/Casa", size: 1, updatedAt: now})
	seedLibraryEntry(t, dbContext, libraryEntry{name: "app.log", directory: "/data/main/Casa", size: 1, updatedAt: now})
	seedLibraryEntry(t, dbContext, libraryEntry{name: "Casa", directory: "/data/main", updatedAt: now, isFolder: true})

	items := listAll(t, repository, DocumentFilter{}, "")
	assertNames(t, items, "a.pdf", "b.docx", "c.xlsx")

	total, err := repository.CountDocuments(DocumentFilter{}, "")
	if err != nil || total != 3 {
		t.Fatalf("count %d err %v", total, err)
	}

	page, err := service.ListDocuments(DocumentListParams{Sort: SortName, Order: OrderAscending, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatal(err)
	}
	parentPaths := []string{}
	for _, item := range page.Items {
		parentPaths = append(parentPaths, item.ParentPath)
	}
	if !slices.Equal(parentPaths, []string{"/Casa", "/Trabalho/2024", "/Midia/docs"}) {
		t.Fatalf("unexpected parent paths %v", parentPaths)
	}
}

func TestListDocumentsFiltersByFormatGroupAndPeriodWithMatchingTotal_Postgres(t *testing.T) {
	repository, _, dbContext := newDocumentTextPostgres(t)
	day := func(dayOfMonth int) time.Time { return time.Date(2024, 3, dayOfMonth, 15, 30, 0, 0, time.UTC) }
	seedLibraryEntry(t, dbContext, libraryEntry{name: "old.pdf", directory: "/docs", updatedAt: day(1)})
	seedLibraryEntry(t, dbContext, libraryEntry{name: "edge.pdf", directory: "/docs", updatedAt: day(10)})
	seedLibraryEntry(t, dbContext, libraryEntry{name: "inside.pdf", directory: "/docs", updatedAt: day(5)})
	seedLibraryEntry(t, dbContext, libraryEntry{name: "after.pdf", directory: "/docs", updatedAt: day(11)})
	seedLibraryEntry(t, dbContext, libraryEntry{name: "inside.docx", directory: "/docs", updatedAt: day(5)})

	from := time.Date(2024, 3, 5, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 3, 10, 0, 0, 0, 0, time.UTC)
	filter := DocumentFilter{Format: GroupPDF, ModifiedFrom: &from, ModifiedTo: &to}

	assertNames(t, listAll(t, repository, filter, ""), "edge.pdf", "inside.pdf")
	total, err := repository.CountDocuments(filter, "")
	if err != nil || total != 2 {
		t.Fatalf("count %d err %v", total, err)
	}
	assertNames(t, listAll(t, repository, DocumentFilter{Format: GroupText}, ""), "inside.docx")
	if items := listAll(t, repository, DocumentFilter{Format: FormatImage}, ""); len(items) != 0 {
		t.Fatalf("image format is reserved and must be empty, got %v", itemNames(items))
	}
}

func TestListDocumentsByFolderIncludesSubfoldersAndExcludesSiblingPrefix_Postgres(t *testing.T) {
	_, service, dbContext := newDocumentTextPostgres(t)
	now := time.Now().UTC()
	folderID := seedLibraryEntry(t, dbContext, libraryEntry{name: "b", directory: "/a", updatedAt: now, isFolder: true})
	seedLibraryEntry(t, dbContext, libraryEntry{name: "direct.pdf", directory: "/a/b", updatedAt: now})
	seedLibraryEntry(t, dbContext, libraryEntry{name: "nested.pdf", directory: "/a/b/c", updatedAt: now})
	seedLibraryEntry(t, dbContext, libraryEntry{name: "sibling.pdf", directory: "/a/bc", updatedAt: now})
	seedLibraryEntry(t, dbContext, libraryEntry{name: "outside.pdf", directory: "/a", updatedAt: now})

	page, err := service.ListDocuments(DocumentListParams{FolderID: folderID, Sort: SortName, Order: OrderAscending, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatal(err)
	}
	names := []string{}
	for _, item := range page.Items {
		names = append(names, item.Name)
	}
	if !slices.Equal(names, []string{"direct.pdf", "nested.pdf"}) {
		t.Fatalf("unexpected documents %v", names)
	}
	count, err := service.CountDocuments(DocumentFilter{}, folderID)
	if err != nil || count.Total != 2 {
		t.Fatalf("count %+v err %v", count, err)
	}

	fileID := seedLibraryEntry(t, dbContext, libraryEntry{name: "plain.pdf", directory: "/a", updatedAt: now})
	deletedFolderID := seedLibraryEntry(t, dbContext, libraryEntry{name: "gone", directory: "/a", updatedAt: now, isFolder: true})
	softDeleteFile(t, dbContext, deletedFolderID)
	for _, invalidFolderID := range []int{fileID, deletedFolderID, 9999} {
		if _, err := service.ListDocuments(DocumentListParams{FolderID: invalidFolderID, Sort: SortName, Page: 1, PageSize: 10}); err != ErrFolderNotFound {
			t.Fatalf("folder %d: want ErrFolderNotFound, got %v", invalidFolderID, err)
		}
		if _, err := service.CountDocuments(DocumentFilter{}, invalidFolderID); err != ErrFolderNotFound {
			t.Fatalf("count folder %d: want ErrFolderNotFound, got %v", invalidFolderID, err)
		}
	}
}

func TestListDocumentsSortsPaginatesAndBreaksTiesById_Postgres(t *testing.T) {
	repository, _, dbContext := newDocumentTextPostgres(t)
	now := time.Now().UTC().Truncate(time.Second)
	seedLibraryEntry(t, dbContext, libraryEntry{name: "small.pdf", directory: "/docs", size: 5, updatedAt: now})
	seedLibraryEntry(t, dbContext, libraryEntry{name: "huge.pdf", directory: "/docs", size: 9000, updatedAt: now})
	seedLibraryEntry(t, dbContext, libraryEntry{name: "Middle.pdf", directory: "/docs", size: 9000, updatedAt: now})

	largest, err := repository.ListDocuments(DocumentListQuery{Sort: SortSize, Order: OrderDescending, Page: 1, PageSize: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(largest.Items) != 1 || largest.Items[0].Name != "huge.pdf" || !largest.Pagination.HasNext || largest.Pagination.HasPrev {
		t.Fatalf("unexpected first page %+v", largest)
	}

	secondPage, err := repository.ListDocuments(DocumentListQuery{Sort: SortSize, Order: OrderDescending, Page: 2, PageSize: 1})
	if err != nil || len(secondPage.Items) != 1 || secondPage.Items[0].Name != "Middle.pdf" || !secondPage.Pagination.HasPrev {
		t.Fatalf("unexpected second page %+v err %v", secondPage, err)
	}

	sameTimestamp, err := repository.ListDocuments(DocumentListQuery{Sort: SortModifiedAt, Order: OrderDescending, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatal(err)
	}
	assertNames(t, sameTimestamp.Items, "small.pdf", "huge.pdf", "Middle.pdf")

	byName, err := repository.ListDocuments(DocumentListQuery{Sort: SortName, Order: OrderAscending, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatal(err)
	}
	assertNames(t, byName.Items, "huge.pdf", "Middle.pdf", "small.pdf")

	byCreatedAt, err := repository.ListDocuments(DocumentListQuery{Sort: SortCreatedAt, Order: OrderAscending, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatal(err)
	}
	assertNames(t, byCreatedAt.Items, "small.pdf", "huge.pdf", "Middle.pdf")
}

func TestListDocumentsExcludesTrashAndFiltersFavorites_Postgres(t *testing.T) {
	repository, _, dbContext := newDocumentTextPostgres(t)
	now := time.Now().UTC()
	seedLibraryEntry(t, dbContext, libraryEntry{name: "kept.pdf", directory: "/docs", updatedAt: now})
	seedLibraryEntry(t, dbContext, libraryEntry{name: "starred.pdf", directory: "/docs", updatedAt: now, starred: true})
	trashed := seedLibraryEntry(t, dbContext, libraryEntry{name: "trashed.pdf", directory: "/docs", updatedAt: now, starred: true})
	softDeleteFile(t, dbContext, trashed)

	assertNames(t, listAll(t, repository, DocumentFilter{}, ""), "kept.pdf", "starred.pdf")
	favorites := listAll(t, repository, DocumentFilter{FavoritesOnly: true}, "")
	assertNames(t, favorites, "starred.pdf")
	if !favorites[0].Starred {
		t.Fatal("starred flag not returned")
	}
	total, err := repository.CountDocuments(DocumentFilter{FavoritesOnly: true}, "")
	if err != nil || total != 1 {
		t.Fatalf("count %d err %v", total, err)
	}
}

func TestListDocumentsFiltersByIndexStateIncludingQueued_Postgres(t *testing.T) {
	repository, _, dbContext := newDocumentTextPostgres(t)
	now := time.Now().UTC().Truncate(time.Second)
	indexed := seedDocumentFile(t, dbContext, "indexed.txt", ".txt", now)
	seedDocumentFile(t, dbContext, "fresh.pdf", ".pdf", now)
	stale := seedDocumentFile(t, dbContext, "stale.md", ".md", now)
	scanned := seedDocumentFile(t, dbContext, "scan.pdf", ".pdf", now)

	upserts := []DocumentTextModel{
		{FileID: indexed, State: StateIndexed, ExtractedText: "ok", IndexedBytes: 2, SourceUpdatedAt: now},
		{FileID: stale, State: StateIndexed, ExtractedText: "old", IndexedBytes: 3, SourceUpdatedAt: now.Add(-time.Hour)},
		{FileID: scanned, State: StateNameOnly, FailureReason: FailureReasonNoText, SourceUpdatedAt: now},
	}
	for _, upsert := range upserts {
		if err := repository.UpsertDocumentText(upsert); err != nil {
			t.Fatal(err)
		}
	}

	testCases := map[string][]string{
		StateIndexed:  {"indexed.txt"},
		StateQueued:   {"fresh.pdf", "stale.md"},
		StateNameOnly: {"scan.pdf"},
		StatePartial:  {},
		StateFailed:   {},
	}
	for indexState, wantNames := range testCases {
		items := listAll(t, repository, DocumentFilter{IndexState: indexState}, "")
		gotNames := itemNames(items)
		if !slices.Equal(gotNames, wantNames) && (len(gotNames) != 0 || len(wantNames) != 0) {
			t.Fatalf("%s: got %v want %v", indexState, gotNames, wantNames)
		}
		for _, item := range items {
			if item.IndexState != indexState {
				t.Fatalf("%s: item %s reports %s", indexState, item.Name, item.IndexState)
			}
		}
		total, err := repository.CountDocuments(DocumentFilter{IndexState: indexState}, "")
		if err != nil || total != len(wantNames) {
			t.Fatalf("%s: count %d err %v", indexState, total, err)
		}
	}
}

func TestListDocumentsRejectsUnknownSort(t *testing.T) {
	if _, _, err := buildListDocumentsQuery(DocumentListQuery{Sort: "bogus", Page: 1, PageSize: 10}); err == nil {
		t.Fatal("expected error for unknown sort")
	}
}
