package image

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"nas-go/api/internal/testutil"
)

type seededImage struct {
	name     string
	folder   string
	format   string
	takenAt  *time.Time
	category string
	starred  bool
	make     string
	model    string
	deleted  bool
	size     int64
}

func utcTime(year int, month time.Month, day int) *time.Time {
	instant := time.Date(year, month, day, 12, 0, 0, 0, time.UTC)
	return &instant
}

func seedLibraryImages(t *testing.T, repository *LibraryRepository, images []seededImage) map[string]int {
	t.Helper()
	idsByName := map[string]int{}
	seedErr := repository.Db.ExecTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(`TRUNCATE image_metadata, home_file RESTART IDENTITY CASCADE`); err != nil {
			return err
		}
		for _, seed := range images {
			path := seed.folder + "/" + seed.name
			var deletedAt any
			if seed.deleted {
				deletedAt = time.Now()
			}
			var fileID int
			if err := tx.QueryRow(`INSERT INTO home_file (name, path, parent_path, format, size, updated_at, created_at, type, checksum, starred, deleted_at)
				VALUES ($1, $2, $3, $4, $5, now(), now(), 2, '', $6, $7) RETURNING id`,
				seed.name, path, seed.folder, seed.format, seed.size, seed.starred, deletedAt).Scan(&fileID); err != nil {
				return err
			}
			idsByName[seed.name] = fileID
			if _, err := tx.Exec(`INSERT INTO image_metadata (file_id, path, width, height, make, model, classification_category, taken_at, created_at)
				VALUES ($1, $2, 800, 600, $3, $4, $5, $6, now())`,
				fileID, path, seed.make, seed.model, seed.category, seed.takenAt); err != nil {
				return err
			}
		}
		return nil
	})
	if seedErr != nil {
		t.Fatalf("seed: %v", seedErr)
	}
	return idsByName
}

func listNames(t *testing.T, repository *LibraryRepository, query LibraryListQuery) []string {
	t.Helper()
	items, err := repository.ListLibraryImages(query)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	names := make([]string, 0, len(items))
	for _, item := range items {
		names = append(names, item.Name)
	}
	return names
}

func assertNames(t *testing.T, label string, actual []string, expected []string) {
	t.Helper()
	if fmt.Sprint(actual) != fmt.Sprint(expected) {
		t.Fatalf("%s: got %v, want %v", label, actual, expected)
	}
}

func newLibraryPostgresRepository(t *testing.T) *LibraryRepository {
	t.Helper()
	return NewLibraryRepository(testutil.NewPostgresDB(t, "kuranas_image_it"))
}

func TestLibraryKeysetOrdersByTakenAtWithTiesAndNullsLast_Postgres(t *testing.T) {
	repository := newLibraryPostgresRepository(t)
	sameInstant := utcTime(2022, 5, 5)
	seedLibraryImages(t, repository, []seededImage{
		{name: "old.jpg", folder: "/lib", format: ".jpg", takenAt: utcTime(2019, 1, 1)},
		{name: "tie_a.jpg", folder: "/lib", format: ".jpg", takenAt: sameInstant},
		{name: "tie_b.jpg", folder: "/lib", format: ".jpg", takenAt: sameInstant},
		{name: "tie_c.jpg", folder: "/lib", format: ".jpg", takenAt: sameInstant},
		{name: "undated_a.jpg", folder: "/lib", format: ".jpg"},
		{name: "undated_b.jpg", folder: "/lib", format: ".jpg"},
		{name: "new.jpg", folder: "/lib", format: ".jpg", takenAt: utcTime(2024, 1, 1)},
	})

	expectedOrder := []string{"new.jpg", "tie_c.jpg", "tie_b.jpg", "tie_a.jpg", "old.jpg", "undated_b.jpg", "undated_a.jpg"}
	baseQuery := LibraryListQuery{Sort: LibrarySortTakenAt, Order: LibrarySortOrderDesc, Limit: 100}
	assertNames(t, "single page", listNames(t, repository, baseQuery), expectedOrder)

	for _, pageSize := range []int{1, 2, 3} {
		var collected []string
		var cursor *LibraryCursor
		for guard := 0; guard < 20; guard++ {
			query := baseQuery
			query.Limit = pageSize
			query.Cursor = cursor
			items, err := repository.ListLibraryImages(query)
			if err != nil {
				t.Fatalf("page: %v", err)
			}
			if len(items) == 0 {
				break
			}
			for _, item := range items {
				collected = append(collected, item.Name)
			}
			nextCursor := CursorFromItem(items[len(items)-1])
			cursor = &nextCursor
		}
		assertNames(t, fmt.Sprintf("keyset pages of %d", pageSize), collected, expectedOrder)
	}
}

func TestLibraryTakenBeforeSeeksIntoTheTimeline_Postgres(t *testing.T) {
	repository := newLibraryPostgresRepository(t)
	seedLibraryImages(t, repository, []seededImage{
		{name: "y2024.jpg", folder: "/lib", format: ".jpg", takenAt: utcTime(2024, 6, 1)},
		{name: "y2022.jpg", folder: "/lib", format: ".jpg", takenAt: utcTime(2022, 6, 1)},
		{name: "y2020.jpg", folder: "/lib", format: ".jpg", takenAt: utcTime(2020, 6, 1)},
		{name: "undated.jpg", folder: "/lib", format: ".jpg"},
	})
	seekInstant := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	names := listNames(t, repository, LibraryListQuery{Sort: LibrarySortTakenAt, Order: LibrarySortOrderDesc, TakenBefore: &seekInstant, Limit: 10})
	assertNames(t, "taken_before", names, []string{"y2022.jpg", "y2020.jpg", "undated.jpg"})
}

func TestLibraryExcludesDeletedNonImageAndDuplicateMetadata_Postgres(t *testing.T) {
	repository := newLibraryPostgresRepository(t)
	ids := seedLibraryImages(t, repository, []seededImage{
		{name: "keep.jpg", folder: "/lib", format: ".jpg", takenAt: utcTime(2022, 1, 1)},
		{name: "gone.jpg", folder: "/lib", format: ".jpg", takenAt: utcTime(2022, 1, 2), deleted: true},
		{name: "clip.mp4", folder: "/lib", format: ".mp4", takenAt: utcTime(2022, 1, 3)},
	})
	insertErr := repository.Db.ExecTx(func(tx *sql.Tx) error {
		_, err := tx.Exec(`INSERT INTO image_metadata (file_id, path, taken_at, created_at) VALUES ($1, '/old/location/keep.jpg', $2, now())`, ids["keep.jpg"], utcTime(2010, 1, 1))
		return err
	})
	if insertErr != nil {
		t.Fatalf("insert duplicate metadata: %v", insertErr)
	}

	query := LibraryListQuery{Sort: LibrarySortTakenAt, Order: LibrarySortOrderDesc, Limit: 10}
	assertNames(t, "listing", listNames(t, repository, query), []string{"keep.jpg"})

	total, err := repository.CountLibraryImages(LibraryFilter{})
	if err != nil || total != 1 {
		t.Fatalf("count = %d err %v", total, err)
	}
}

func TestLibraryFiltersCombine_Postgres(t *testing.T) {
	repository := newLibraryPostgresRepository(t)
	seedLibraryImages(t, repository, []seededImage{
		{name: "Beach_Day.jpg", folder: "/lib/trips", format: ".jpg", takenAt: utcTime(2021, 7, 1), category: "landscape", starred: true, make: "Canon", model: "EOS R5", size: 300},
		{name: "beach_night.png", folder: "/lib/trips", format: ".png", takenAt: utcTime(2022, 7, 1), category: "photo", make: "Canon", model: "EOS R5", size: 100},
		{name: "100%_sure.jpg", folder: "/lib", format: ".jpg", takenAt: utcTime(2023, 1, 1), category: "photo", make: "Sony", model: "A7", size: 200},
		{name: "shot.png", folder: "/lib/trips/deep", format: ".png", takenAt: utcTime(2020, 1, 1), category: "screenshot_app", size: 400},
	})

	testCases := []struct {
		label    string
		filter   LibraryFilter
		expected []string
	}{
		{"name contains", LibraryFilter{NameQuery: "beach"}, []string{"beach_night.png", "Beach_Day.jpg"}},
		{"literal percent is not a wildcard", LibraryFilter{NameQuery: "100%_"}, []string{"100%_sure.jpg"}},
		{"underscore is not a wildcard", LibraryFilter{NameQuery: "beach.day"}, []string{}},
		{"category any of", LibraryFilter{Categories: []ClassificationCategory{ClassificationCategoryLandscape, ClassificationCategoryScreenshot}}, []string{"Beach_Day.jpg", "shot.png"}},
		{"starred", LibraryFilter{OnlyStarred: true}, []string{"Beach_Day.jpg"}},
		{"format", LibraryFilter{Formats: []string{".png"}}, []string{"beach_night.png", "shot.png"}},
		{"taken range", LibraryFilter{TakenFrom: utcTime(2021, 1, 1), TakenTo: utcTime(2022, 12, 31)}, []string{"beach_night.png", "Beach_Day.jpg"}},
		{"camera exact", LibraryFilter{Camera: "Canon EOS R5"}, []string{"beach_night.png", "Beach_Day.jpg"}},
		{"camera is not a prefix match", LibraryFilter{Camera: "Canon"}, []string{}},
		{"folder direct children only", LibraryFilter{Folder: "/lib/trips"}, []string{"beach_night.png", "Beach_Day.jpg"}},
		{"combined", LibraryFilter{NameQuery: "beach", Formats: []string{".jpg"}, Camera: "Canon EOS R5", Folder: "/lib/trips"}, []string{"Beach_Day.jpg"}},
	}
	for _, testCase := range testCases {
		query := LibraryListQuery{Filter: testCase.filter, Sort: LibrarySortTakenAt, Order: LibrarySortOrderDesc, Limit: 50}
		assertNames(t, testCase.label, listNames(t, repository, query), testCase.expected)

		total, err := repository.CountLibraryImages(testCase.filter)
		if err != nil || total != len(testCase.expected) {
			t.Fatalf("%s: count = %d err %v, want %d", testCase.label, total, err, len(testCase.expected))
		}
	}
}

func TestLibraryOffsetOrderings_Postgres(t *testing.T) {
	repository := newLibraryPostgresRepository(t)
	seedLibraryImages(t, repository, []seededImage{
		{name: "b.jpg", folder: "/lib", format: ".jpg", takenAt: utcTime(2021, 1, 1), size: 20},
		{name: "a.jpg", folder: "/lib", format: ".jpg", takenAt: utcTime(2022, 1, 1), size: 30},
		{name: "c.jpg", folder: "/lib", format: ".jpg", size: 10},
	})

	testCases := []struct {
		label    string
		sort     LibrarySort
		order    LibrarySortOrder
		expected []string
	}{
		{"name asc", LibrarySortName, LibrarySortOrderAsc, []string{"a.jpg", "b.jpg", "c.jpg"}},
		{"name desc", LibrarySortName, LibrarySortOrderDesc, []string{"c.jpg", "b.jpg", "a.jpg"}},
		{"size asc", LibrarySortSize, LibrarySortOrderAsc, []string{"c.jpg", "b.jpg", "a.jpg"}},
		{"size desc", LibrarySortSize, LibrarySortOrderDesc, []string{"a.jpg", "b.jpg", "c.jpg"}},
		{"taken asc, nulls last", LibrarySortTakenAt, LibrarySortOrderAsc, []string{"b.jpg", "a.jpg", "c.jpg"}},
	}
	for _, testCase := range testCases {
		query := LibraryListQuery{Sort: testCase.sort, Order: testCase.order, Limit: 50}
		assertNames(t, testCase.label, listNames(t, repository, query), testCase.expected)
	}

	secondPage := listNames(t, repository, LibraryListQuery{Sort: LibrarySortName, Order: LibrarySortOrderAsc, Limit: 2, Offset: 2})
	assertNames(t, "offset page", secondPage, []string{"c.jpg"})
}

func TestLibraryTimelineGroupsByYearAndMonth_Postgres(t *testing.T) {
	repository := newLibraryPostgresRepository(t)
	endOfYear := time.Date(2022, 12, 31, 23, 30, 0, 0, time.UTC)
	startOfYear := time.Date(2023, 1, 1, 0, 30, 0, 0, time.UTC)
	seedLibraryImages(t, repository, []seededImage{
		{name: "a.jpg", folder: "/lib", format: ".jpg", takenAt: &endOfYear},
		{name: "b.jpg", folder: "/lib", format: ".jpg", takenAt: &startOfYear},
		{name: "c.jpg", folder: "/lib", format: ".jpg", takenAt: utcTime(2023, 1, 20)},
		{name: "d.jpg", folder: "/lib", format: ".jpg", takenAt: utcTime(2021, 3, 3), category: "photo"},
		{name: "undated.jpg", folder: "/lib", format: ".jpg"},
	})

	buckets, err := repository.ListLibraryTimeline(LibraryFilter{})
	if err != nil {
		t.Fatalf("timeline: %v", err)
	}
	expected := []LibraryTimelineBucketModel{{2023, 1, 2}, {2022, 12, 1}, {2021, 3, 1}}
	if fmt.Sprint(buckets) != fmt.Sprint(expected) {
		t.Fatalf("buckets = %v, want %v", buckets, expected)
	}

	filtered, err := repository.ListLibraryTimeline(LibraryFilter{Categories: []ClassificationCategory{ClassificationCategoryPhoto}})
	if err != nil || fmt.Sprint(filtered) != fmt.Sprint([]LibraryTimelineBucketModel{{2021, 3, 1}}) {
		t.Fatalf("filtered buckets = %v err %v", filtered, err)
	}
}

func TestLibraryItemShapeAndTier_Postgres(t *testing.T) {
	repository := newLibraryPostgresRepository(t)
	ids := seedLibraryImages(t, repository, []seededImage{
		{name: "cold.jpg", folder: "/lib", format: ".jpg", takenAt: utcTime(2022, 1, 1), category: "photo", starred: true, size: 99},
	})
	coldErr := repository.Db.ExecTx(func(tx *sql.Tx) error {
		_, err := tx.Exec(`UPDATE home_file SET physical_path = '/cold/cold.jpg' WHERE id = $1`, ids["cold.jpg"])
		return err
	})
	if coldErr != nil {
		t.Fatalf("mark cold: %v", coldErr)
	}

	items, err := repository.ListLibraryImages(LibraryListQuery{Sort: LibrarySortTakenAt, Order: LibrarySortOrderDesc, Limit: 5})
	if err != nil || len(items) != 1 {
		t.Fatalf("items = %v err %v", items, err)
	}
	item := items[0]
	if item.FileID != ids["cold.jpg"] || item.Width != 800 || item.Height != 600 || item.Size != 99 || !item.Starred || !item.IsCold || item.Category != "photo" || item.TakenAt == nil {
		t.Fatalf("unexpected item %+v", item)
	}
}

func TestUpsertImageMetadataSetsTakenAtFromExifThenFileModificationTime_Postgres(t *testing.T) {
	libraryRepository := newLibraryPostgresRepository(t)
	repository := NewRepository(libraryRepository.Db)
	fileModificationTime := time.Date(2018, 2, 3, 4, 5, 6, 0, time.UTC)

	seedErr := repository.Db.ExecTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(`TRUNCATE image_metadata, home_file RESTART IDENTITY CASCADE`); err != nil {
			return err
		}
		_, err := tx.Exec(`INSERT INTO home_file (name, path, parent_path, format, size, updated_at, created_at, type, checksum)
			VALUES ('exif.jpg', '/lib/exif.jpg', '/lib', '.jpg', 1, $1, now(), 2, ''),
			       ('bare.jpg', '/lib/bare.jpg', '/lib', '.jpg', 1, $1, now(), 2, ''),
			       ('broken.jpg', '/lib/broken.jpg', '/lib', '.jpg', 1, $1, now(), 2, '')`, fileModificationTime)
		return err
	})
	if seedErr != nil {
		t.Fatalf("seed: %v", seedErr)
	}

	metadataByFileID := map[int]MetadataModel{
		1: {FileId: 1, Path: "/lib/exif.jpg", DateTimeOriginal: "2020:08:09 10:11:12", DateTime: "2021:01:01 00:00:00"},
		2: {FileId: 2, Path: "/lib/bare.jpg"},
		3: {FileId: 3, Path: "/lib/broken.jpg", DateTimeOriginal: "0000:00:00 00:00:00"},
	}
	for _, metadata := range metadataByFileID {
		upsertErr := repository.Db.ExecTx(func(tx *sql.Tx) error {
			_, err := repository.UpsertImageMetadata(tx, metadata)
			return err
		})
		if upsertErr != nil {
			t.Fatalf("upsert %s: %v", metadata.Path, upsertErr)
		}
	}

	expectedByName := map[string]time.Time{
		"exif.jpg":   time.Date(2020, 8, 9, 10, 11, 12, 0, time.UTC),
		"bare.jpg":   fileModificationTime,
		"broken.jpg": fileModificationTime,
	}
	items, err := libraryRepository.ListLibraryImages(LibraryListQuery{Sort: LibrarySortName, Order: LibrarySortOrderAsc, Limit: 10})
	if err != nil || len(items) != 3 {
		t.Fatalf("items = %v err %v", items, err)
	}
	for _, item := range items {
		if item.TakenAt == nil || !item.TakenAt.Equal(expectedByName[item.Name]) {
			t.Fatalf("%s: taken_at = %v, want %v", item.Name, item.TakenAt, expectedByName[item.Name])
		}
	}

	reupsert := metadataByFileID[2]
	reupsert.DateTimeOriginal = "2019:09:09 09:09:09"
	if err := repository.Db.ExecTx(func(tx *sql.Tx) error {
		_, upsertErr := repository.UpsertImageMetadata(tx, reupsert)
		return upsertErr
	}); err != nil {
		t.Fatalf("re-upsert: %v", err)
	}
	refreshed, err := libraryRepository.ListLibraryImages(LibraryListQuery{Filter: LibraryFilter{NameQuery: "bare"}, Sort: LibrarySortName, Order: LibrarySortOrderAsc, Limit: 1})
	if err != nil || len(refreshed) != 1 || !refreshed[0].TakenAt.Equal(time.Date(2019, 9, 9, 9, 9, 9, 0, time.UTC)) {
		t.Fatalf("re-upsert must refresh taken_at, got %v err %v", refreshed, err)
	}
}
