package analytics

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"nas-go/api/internal/testutil"
)

type duplicateFileFixture struct {
	name         string
	checksum     string
	size         int64
	updatedAt    time.Time
	fileType     int
	isDeleted    bool
	physicalPath string
}

func newDuplicateGroupsRepository(t *testing.T) *Repository {
	t.Helper()
	repo := NewRepository(testutil.NewPostgresDB(t, "kuranas_analytics_it"))
	err := repo.DbContext.ExecTx(func(tx *sql.Tx) error {
		_, execErr := tx.Exec("TRUNCATE home_file RESTART IDENTITY CASCADE")
		return execErr
	})
	if err != nil {
		t.Fatalf("truncate home_file: %v", err)
	}
	return repo
}

func insertDuplicateFixtures(t *testing.T, repo *Repository, fixtures []duplicateFileFixture) {
	t.Helper()
	err := repo.DbContext.ExecTx(func(tx *sql.Tx) error {
		for _, fixture := range fixtures {
			fileType := fixture.fileType
			if fileType == 0 {
				fileType = 2
			}
			var deletedAt any
			if fixture.isDeleted {
				deletedAt = time.Now()
			}
			var physicalPath any
			if fixture.physicalPath != "" {
				physicalPath = fixture.physicalPath
			}
			var checksum any
			if fixture.checksum != "" {
				checksum = fixture.checksum
			}
			updatedAt := fixture.updatedAt
			if updatedAt.IsZero() {
				updatedAt = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
			}
			_, execErr := tx.Exec(
				`INSERT INTO home_file (name, path, parent_path, format, size, updated_at, created_at, type, checksum, deleted_at, physical_path)
				 VALUES ($1, $2, '/lib', $3, $4, $5, $5, $6, $7, $8, $9)`,
				fixture.name, "/lib/"+fixture.name, filepath.Ext(fixture.name), fixture.size, updatedAt, fileType, checksum, deletedAt, physicalPath,
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

func duplicateGroup(checksum string, extension string, size int64, copies int) []duplicateFileFixture {
	fixtures := make([]duplicateFileFixture, 0, copies)
	for copyIndex := 0; copyIndex < copies; copyIndex++ {
		fixtures = append(fixtures, duplicateFileFixture{
			name:     fmt.Sprintf("%s-%d%s", checksum, copyIndex, extension),
			checksum: checksum,
			size:     size,
		})
	}
	return fixtures
}

func signaturesOf(groups []DuplicateGroupDetailModel) []string {
	signatures := make([]string, 0, len(groups))
	for _, group := range groups {
		signatures = append(signatures, group.Signature)
	}
	return signatures
}

func TestPostgres_DuplicateGroupsPageOrdersByReclaimableSpaceThenChecksum(t *testing.T) {
	repo := newDuplicateGroupsRepository(t)
	var fixtures []duplicateFileFixture
	fixtures = append(fixtures, duplicateGroup("small-wide", ".jpg", 50, 5)...)
	fixtures = append(fixtures, duplicateGroup("big-pair", ".mp4", 1000, 2)...)
	fixtures = append(fixtures, duplicateGroup("alpha-tie", ".pdf", 100, 3)...)
	fixtures = append(fixtures, duplicateGroup("zulu-tie", ".pdf", 100, 3)...)
	insertDuplicateFixtures(t, repo, fixtures)

	page, err := repo.GetDuplicateGroupsPage(DuplicateCategoryAll, 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"big-pair", "alpha-tie", "small-wide", "zulu-tie"}
	if got := signaturesOf(page.Items); !slices.Equal(got, expected) {
		t.Fatalf("expected order %v, got %v", expected, got)
	}
	bigPair := page.Items[0]
	if bigPair.Copies != 2 || bigPair.ItemSize != 1000 || bigPair.ReclaimableSize != 1000 {
		t.Fatalf("unexpected big-pair numbers: %+v", bigPair.DuplicateGroupModel)
	}
	if page.Pagination.HasNext || page.Pagination.HasPrev || page.Pagination.Page != 1 || page.Pagination.PageSize != 10 {
		t.Fatalf("unexpected pagination: %+v", page.Pagination)
	}
}

func TestPostgres_DuplicateGroupsPageReturnsCopyDetails(t *testing.T) {
	repo := newDuplicateGroupsRepository(t)
	oldest := time.Date(2025, 5, 6, 7, 8, 9, 0, time.UTC)
	newest := time.Date(2026, 6, 7, 8, 9, 10, 0, time.UTC)
	insertDuplicateFixtures(t, repo, []duplicateFileFixture{
		{name: "b-copy.pdf", checksum: "pair", size: 10, updatedAt: newest, physicalPath: "/cold/b-copy.pdf"},
		{name: "a-copy.pdf", checksum: "pair", size: 10, updatedAt: oldest},
	})

	page, err := repo.GetDuplicateGroupsPage(DuplicateCategoryAll, 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("expected 1 group, got %d", len(page.Items))
	}
	group := page.Items[0]
	if !slices.Equal(group.Paths, []string{"/lib/a-copy.pdf", "/lib/b-copy.pdf"}) {
		t.Fatalf("unexpected paths: %v", group.Paths)
	}
	if len(group.Files) != 2 {
		t.Fatalf("expected 2 copies, got %d", len(group.Files))
	}
	hotCopy, coldCopy := group.Files[0], group.Files[1]
	if hotCopy.ID != 2 || hotCopy.Path != "/lib/a-copy.pdf" || hotCopy.Tier != "hot" || !hotCopy.UpdatedAt.Equal(oldest) {
		t.Fatalf("unexpected hot copy: %+v", hotCopy)
	}
	if coldCopy.ID != 1 || coldCopy.Path != "/lib/b-copy.pdf" || coldCopy.Tier != "cold" || !coldCopy.UpdatedAt.Equal(newest) {
		t.Fatalf("unexpected cold copy: %+v", coldCopy)
	}
}

func TestPostgres_DuplicateGroupsPageIgnoresRowsThatAreNotDuplicates(t *testing.T) {
	repo := newDuplicateGroupsRepository(t)
	insertDuplicateFixtures(t, repo, []duplicateFileFixture{
		{name: "lonely.jpg", checksum: "lonely", size: 10},
		{name: "deleted-a.jpg", checksum: "after-delete", size: 10},
		{name: "deleted-b.jpg", checksum: "after-delete", size: 10, isDeleted: true},
		{name: "folder-a", checksum: "folders", size: 10, fileType: 1},
		{name: "folder-b", checksum: "folders", size: 10, fileType: 1},
		{name: "no-checksum-a.jpg", size: 10},
		{name: "no-checksum-b.jpg", size: 10},
	})

	page, err := repo.GetDuplicateGroupsPage(DuplicateCategoryAll, 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(page.Items) != 0 {
		t.Fatalf("expected no groups, got %v", signaturesOf(page.Items))
	}
	if page.Items == nil {
		t.Fatal("items must be an empty slice, never nil")
	}
}

func TestPostgres_DuplicateGroupsPageFiltersByCategory(t *testing.T) {
	repo := newDuplicateGroupsRepository(t)
	var fixtures []duplicateFileFixture
	fixtures = append(fixtures, duplicateGroup("img", ".JPG", 10, 2)...)
	fixtures = append(fixtures, duplicateGroup("aud", ".flac", 10, 2)...)
	fixtures = append(fixtures, duplicateGroup("vid", ".mkv", 10, 2)...)
	fixtures = append(fixtures, duplicateGroup("doc", ".pdf", 10, 2)...)
	fixtures = append(fixtures, duplicateGroup("office", ".xlsx", 10, 2)...)
	fixtures = append(fixtures, duplicateGroup("zip", ".zip", 10, 2)...)
	fixtures = append(fixtures, duplicateGroup("noext", "", 10, 2)...)
	insertDuplicateFixtures(t, repo, fixtures)

	expectedByCategory := map[DuplicateCategory][]string{
		DuplicateCategoryImage:    {"img"},
		DuplicateCategoryAudio:    {"aud"},
		DuplicateCategoryVideo:    {"vid"},
		DuplicateCategoryDocument: {"doc", "office"},
		DuplicateCategoryOther:    {"noext", "zip"},
		DuplicateCategoryAll:      {"aud", "doc", "img", "noext", "office", "vid", "zip"},
	}
	for category, expected := range expectedByCategory {
		page, err := repo.GetDuplicateGroupsPage(category, 1, 50)
		if err != nil {
			t.Fatalf("category %q: %v", category, err)
		}
		got := signaturesOf(page.Items)
		slices.Sort(got)
		if !slices.Equal(got, expected) {
			t.Fatalf("category %q: expected %v, got %v", category, expected, got)
		}
	}
}

func TestPostgres_DuplicateGroupsPageHasNextOnlyWhenMoreGroupsExist(t *testing.T) {
	repo := newDuplicateGroupsRepository(t)
	var fixtures []duplicateFileFixture
	for groupIndex := 0; groupIndex < 4; groupIndex++ {
		fixtures = append(fixtures, duplicateGroup(fmt.Sprintf("g%d", groupIndex), ".jpg", int64(10+groupIndex), 2)...)
	}
	insertDuplicateFixtures(t, repo, fixtures)

	boundaries := []struct {
		page, pageSize int
		itemCount      int
		hasNext        bool
		hasPrev        bool
	}{
		{page: 1, pageSize: 3, itemCount: 3, hasNext: true, hasPrev: false},
		{page: 2, pageSize: 3, itemCount: 1, hasNext: false, hasPrev: true},
		{page: 1, pageSize: 4, itemCount: 4, hasNext: false, hasPrev: false},
		{page: 2, pageSize: 2, itemCount: 2, hasNext: false, hasPrev: true},
		{page: 3, pageSize: 2, itemCount: 0, hasNext: false, hasPrev: true},
	}
	for _, boundary := range boundaries {
		page, err := repo.GetDuplicateGroupsPage(DuplicateCategoryAll, boundary.page, boundary.pageSize)
		if err != nil {
			t.Fatalf("%+v: %v", boundary, err)
		}
		if len(page.Items) != boundary.itemCount || page.Pagination.HasNext != boundary.hasNext || page.Pagination.HasPrev != boundary.hasPrev {
			t.Fatalf("%+v: got items=%d next=%v prev=%v", boundary, len(page.Items), page.Pagination.HasNext, page.Pagination.HasPrev)
		}
	}
}

func TestPostgres_DuplicateGroupsPageWalksMoreThanOneHundredGroupsWithoutRepeatsOrGaps(t *testing.T) {
	repo := newDuplicateGroupsRepository(t)
	const totalGroups = 137
	var fixtures []duplicateFileFixture
	for groupIndex := 0; groupIndex < totalGroups; groupIndex++ {
		sizeWithTies := int64(100 + groupIndex%7)
		fixtures = append(fixtures, duplicateGroup(fmt.Sprintf("group-%03d", groupIndex), ".bin", sizeWithTies, 2)...)
	}
	insertDuplicateFixtures(t, repo, fixtures)

	seenSignatures := map[string]bool{}
	walkedOrder := []string{}
	pageCount := 0
	for page := 1; ; page++ {
		result, err := repo.GetDuplicateGroupsPage(DuplicateCategoryAll, page, 25)
		if err != nil {
			t.Fatalf("page %d: %v", page, err)
		}
		pageCount++
		for _, group := range result.Items {
			if seenSignatures[group.Signature] {
				t.Fatalf("group %q repeated on page %d", group.Signature, page)
			}
			seenSignatures[group.Signature] = true
			walkedOrder = append(walkedOrder, group.Signature)
		}
		if !result.Pagination.HasNext {
			break
		}
		if page > totalGroups {
			t.Fatal("pagination never ended")
		}
	}

	if len(seenSignatures) != totalGroups || pageCount != 6 {
		t.Fatalf("expected %d groups over 6 pages, got %d over %d", totalGroups, len(seenSignatures), pageCount)
	}
	fullPage, err := repo.GetDuplicateGroupsPage(DuplicateCategoryAll, 1, 200)
	if err != nil {
		t.Fatalf("full page: %v", err)
	}
	if !slices.Equal(signaturesOf(fullPage.Items), walkedOrder) {
		t.Fatal("paged order differs from the single full-page order")
	}
}
