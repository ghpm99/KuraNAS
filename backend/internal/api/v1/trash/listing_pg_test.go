package trash

import (
	"database/sql"
	"slices"
	"testing"
	"time"

	"nas-go/api/internal/testutil"
)

var listingSeedMoment = time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)

func seedListingItems(t *testing.T, repo *Repository) {
	t.Helper()
	seededItems := []TrashItemModel{
		{OriginalPath: `/srv/docs/Zebra.txt`, TrashPath: "/t/1", Size: 30, DeletedAt: listingSeedMoment},
		{OriginalPath: `D:\Fotos\alpha.jpg`, TrashPath: "/t/2", Size: 10, DeletedAt: listingSeedMoment.Add(time.Hour)},
		{OriginalPath: `/srv/fotos/Beta.png`, TrashPath: "/t/3", Size: 20, DeletedAt: listingSeedMoment.Add(2 * time.Hour)},
	}
	err := repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		if _, truncateErr := tx.Exec("TRUNCATE trash_item RESTART IDENTITY"); truncateErr != nil {
			return truncateErr
		}
		for _, seededItem := range seededItems {
			if _, createErr := repo.CreateItem(tx, seededItem); createErr != nil {
				return createErr
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("seed trash items: %v", err)
	}
}

func TestPostgres_TrashListingOrdersAndSearches(t *testing.T) {
	repo := NewRepository(testutil.NewPostgresDB(t, "kuranas_trash_it"))
	seedListingItems(t, repo)

	testCases := []struct {
		name          string
		listing       Listing
		expectedPaths []string
	}{
		{"default lists newest deletions first", DefaultListing, []string{`/srv/fotos/Beta.png`, `D:\Fotos\alpha.jpg`, `/srv/docs/Zebra.txt`}},
		{"oldest deletions first", Listing{SortKey: SortByDeletedAt, Direction: SortAscending}, []string{`/srv/docs/Zebra.txt`, `D:\Fotos\alpha.jpg`, `/srv/fotos/Beta.png`}},
		{"name ignores folder and case", Listing{SortKey: SortByName, Direction: SortAscending}, []string{`D:\Fotos\alpha.jpg`, `/srv/fotos/Beta.png`, `/srv/docs/Zebra.txt`}},
		{"largest first", Listing{SortKey: SortBySize, Direction: SortDescending}, []string{`/srv/docs/Zebra.txt`, `/srv/fotos/Beta.png`, `D:\Fotos\alpha.jpg`}},
		{"search matches the path without case", Listing{SortKey: SortBySize, Direction: SortAscending, SearchText: "fotos"}, []string{`D:\Fotos\alpha.jpg`, `/srv/fotos/Beta.png`}},
		{"search treats backslash literally", Listing{SortKey: SortBySize, Direction: SortAscending, SearchText: `fotos\al`}, []string{`D:\Fotos\alpha.jpg`}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			page, err := repo.GetItems(testCase.listing, 1, 10)
			if err != nil {
				t.Fatalf("GetItems: %v", err)
			}
			listedPaths := make([]string, 0, len(page.Items))
			for _, item := range page.Items {
				listedPaths = append(listedPaths, item.OriginalPath)
			}
			if !slices.Equal(listedPaths, testCase.expectedPaths) {
				t.Fatalf("got %v, want %v", listedPaths, testCase.expectedPaths)
			}
		})
	}
}
