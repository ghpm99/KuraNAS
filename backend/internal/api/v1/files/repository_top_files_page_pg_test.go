package files

import (
	"database/sql"
	"fmt"
	"slices"
	"testing"
	"time"

	"nas-go/api/internal/testutil"
)

type topFileFixture struct {
	name         string
	size         int64
	fileType     int
	isDeleted    bool
	physicalPath string
}

func newTopFilesRepository(t *testing.T) *Repository {
	t.Helper()
	repo := NewRepository(testutil.NewPostgresDB(t, "kuranas_files_it"))
	truncateHomeFile(t, repo)
	return repo
}

func insertTopFileFixtures(t *testing.T, repo *Repository, fixtures []topFileFixture) {
	t.Helper()
	updatedAt := time.Date(2026, 2, 3, 4, 5, 6, 0, time.UTC)
	err := repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		for _, fixture := range fixtures {
			fileType := fixture.fileType
			if fileType == 0 {
				fileType = int(File)
			}
			var deletedAt any
			if fixture.isDeleted {
				deletedAt = time.Now()
			}
			var physicalPath any
			if fixture.physicalPath != "" {
				physicalPath = fixture.physicalPath
			}
			_, execErr := tx.Exec(
				`INSERT INTO home_file (name, path, parent_path, format, size, updated_at, created_at, type, deleted_at, physical_path)
				 VALUES ($1, $2, '/lib/albums', '.bin', $3, $4, $4, $5, $6, $7)`,
				fixture.name, "/lib/albums/"+fixture.name, fixture.size, updatedAt, fileType, deletedAt, physicalPath,
			)
			if execErr != nil {
				return execErr
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("insert fixtures: %v", err)
	}
}

func topFileNamesOf(files []FileModel) []string {
	names := make([]string, 0, len(files))
	for _, file := range files {
		names = append(names, file.Name)
	}
	return names
}

func TestPostgres_TopFilesBySizePageOrdersBySizeDescendingThenId(t *testing.T) {
	repo := newTopFilesRepository(t)
	insertTopFileFixtures(t, repo, []topFileFixture{
		{name: "medium.bin", size: 500},
		{name: "huge.bin", size: 9000},
		{name: "tie-first.bin", size: 100},
		{name: "tie-second.bin", size: 100},
		{name: "tiny.bin", size: 1},
	})

	page, err := repo.GetTopFilesBySizePage(1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"huge.bin", "medium.bin", "tie-first.bin", "tie-second.bin", "tiny.bin"}
	if got := topFileNamesOf(page.Items); !slices.Equal(got, expected) {
		t.Fatalf("expected %v, got %v", expected, got)
	}
	if page.Pagination.HasNext || page.Pagination.HasPrev || page.Pagination.Page != 1 || page.Pagination.PageSize != 10 {
		t.Fatalf("unexpected pagination: %+v", page.Pagination)
	}
}

func TestPostgres_TopFilesBySizePageReturnsFieldsNeededToOpenTheFile(t *testing.T) {
	repo := newTopFilesRepository(t)
	insertTopFileFixtures(t, repo, []topFileFixture{
		{name: "hot.bin", size: 20},
		{name: "cold.bin", size: 30, physicalPath: "/cold/cold.bin"},
	})

	page, err := repo.GetTopFilesBySizePage(1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	coldFile, hotFile := page.Items[0], page.Items[1]
	if coldFile.ID != 2 || coldFile.Name != "cold.bin" || coldFile.Path != "/lib/albums/cold.bin" ||
		coldFile.ParentPath != "/lib/albums" || coldFile.Format != ".bin" || coldFile.Type != File || coldFile.Size != 30 {
		t.Fatalf("unexpected cold file: %+v", coldFile)
	}
	if !coldFile.PhysicalPath.Valid || hotFile.PhysicalPath.Valid {
		t.Fatalf("tier source lost: cold=%+v hot=%+v", coldFile.PhysicalPath, hotFile.PhysicalPath)
	}
	if coldFile.UpdatedAt.IsZero() || coldFile.CreatedAt.IsZero() {
		t.Fatalf("timestamps not loaded: %+v", coldFile)
	}
}

func TestPostgres_TopFilesBySizePageExcludesDirectoriesAndDeletedFiles(t *testing.T) {
	repo := newTopFilesRepository(t)
	insertTopFileFixtures(t, repo, []topFileFixture{
		{name: "folder", size: 99999, fileType: int(Directory)},
		{name: "removed.bin", size: 88888, isDeleted: true},
		{name: "kept.bin", size: 10},
	})

	page, err := repo.GetTopFilesBySizePage(1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := topFileNamesOf(page.Items); !slices.Equal(got, []string{"kept.bin"}) {
		t.Fatalf("expected only kept.bin, got %v", got)
	}
}

func TestPostgres_TopFilesBySizePageHasNextOnlyWhenMoreFilesExist(t *testing.T) {
	repo := newTopFilesRepository(t)
	var fixtures []topFileFixture
	for fileIndex := 0; fileIndex < 5; fileIndex++ {
		fixtures = append(fixtures, topFileFixture{name: fmt.Sprintf("f%d.bin", fileIndex), size: int64(100 - fileIndex)})
	}
	insertTopFileFixtures(t, repo, fixtures)

	boundaries := []struct {
		page, pageSize int
		expectedNames  []string
		hasNext        bool
		hasPrev        bool
	}{
		{page: 1, pageSize: 2, expectedNames: []string{"f0.bin", "f1.bin"}, hasNext: true, hasPrev: false},
		{page: 3, pageSize: 2, expectedNames: []string{"f4.bin"}, hasNext: false, hasPrev: true},
		{page: 1, pageSize: 5, expectedNames: []string{"f0.bin", "f1.bin", "f2.bin", "f3.bin", "f4.bin"}, hasNext: false, hasPrev: false},
		{page: 2, pageSize: 3, expectedNames: []string{"f3.bin", "f4.bin"}, hasNext: false, hasPrev: true},
		{page: 4, pageSize: 2, expectedNames: []string{}, hasNext: false, hasPrev: true},
	}
	for _, boundary := range boundaries {
		page, err := repo.GetTopFilesBySizePage(boundary.page, boundary.pageSize)
		if err != nil {
			t.Fatalf("%+v: %v", boundary, err)
		}
		if got := topFileNamesOf(page.Items); !slices.Equal(got, boundary.expectedNames) ||
			page.Pagination.HasNext != boundary.hasNext || page.Pagination.HasPrev != boundary.hasPrev {
			t.Fatalf("%+v: got names=%v next=%v prev=%v", boundary, got, page.Pagination.HasNext, page.Pagination.HasPrev)
		}
	}
}

func TestPostgres_TopFilesBySizePageWalksMoreThanOneHundredFilesWithoutRepeatsOrGaps(t *testing.T) {
	repo := newTopFilesRepository(t)
	const totalFiles = 143
	var fixtures []topFileFixture
	for fileIndex := 0; fileIndex < totalFiles; fileIndex++ {
		fixtures = append(fixtures, topFileFixture{name: fmt.Sprintf("file-%03d.bin", fileIndex), size: int64(1000 + fileIndex%9)})
	}
	insertTopFileFixtures(t, repo, fixtures)

	walkedNames := []string{}
	seenNames := map[string]bool{}
	for page := 1; ; page++ {
		result, err := repo.GetTopFilesBySizePage(page, 30)
		if err != nil {
			t.Fatalf("page %d: %v", page, err)
		}
		for _, file := range result.Items {
			if seenNames[file.Name] {
				t.Fatalf("%q repeated on page %d", file.Name, page)
			}
			seenNames[file.Name] = true
			walkedNames = append(walkedNames, file.Name)
		}
		if !result.Pagination.HasNext {
			break
		}
		if page > totalFiles {
			t.Fatal("pagination never ended")
		}
	}

	if len(walkedNames) != totalFiles {
		t.Fatalf("expected %d files, walked %d", totalFiles, len(walkedNames))
	}
	fullPage, err := repo.GetTopFilesBySizePage(1, 200)
	if err != nil {
		t.Fatalf("full page: %v", err)
	}
	if !slices.Equal(topFileNamesOf(fullPage.Items), walkedNames) {
		t.Fatal("paged order differs from the single full-page order")
	}
	sizes := make([]int64, 0, len(fullPage.Items))
	for _, file := range fullPage.Items {
		sizes = append(sizes, file.Size)
	}
	if !slices.IsSortedFunc(sizes, func(left, right int64) int { return int(right - left) }) {
		t.Fatal("sizes are not in descending order")
	}
}
