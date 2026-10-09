package image

import (
	"database/sql"
	"testing"

	searchqueries "nas-go/api/pkg/database/queries/search"

	"github.com/lib/pq"
)

func seedContentLibrary(t *testing.T) (*LibraryRepository, *Repository, map[string]int) {
	t.Helper()
	libraryRepository := newLibraryPostgresRepository(t)
	idsByName := seedLibraryImages(t, libraryRepository, []seededImage{
		{name: "IMG_0001.jpg", folder: "/lib", format: ".jpg", takenAt: utcTime(2021, 7, 1), category: "photo"},
		{name: "IMG_0002.jpg", folder: "/lib", format: ".jpg", takenAt: utcTime(2022, 7, 1), category: "photo"},
		{name: "red_car_name.jpg", folder: "/lib", format: ".jpg", takenAt: utcTime(2023, 7, 1), category: "photo"},
		{name: "IMG_0004.jpg", folder: "/lib", format: ".jpg", takenAt: utcTime(2024, 7, 1), category: "photo"},
	})
	imageRepository := NewRepository(libraryRepository.Db)
	fillSummaryTextColumns(t, imageRepository)

	contentByName := map[string]ContentDescription{
		"IMG_0001.jpg": {Caption: "A red car parked by the beach", Tags: []string{"car", "beach"}, OCRText: "PLATE 42"},
		"IMG_0002.jpg": {Caption: "A dog sleeping", Tags: []string{"dog"}, OCRText: ""},
	}
	for name, content := range contentByName {
		classification := ClassificationModel{Category: ClassificationCategoryPhoto, Confidence: 0.9, Content: content}
		if err := imageRepository.UpdateAIClassification(idsByName[name], classification); err != nil {
			t.Fatalf("update classification of %s: %v", name, err)
		}
	}
	return libraryRepository, imageRepository, idsByName
}

func fillSummaryTextColumns(t *testing.T, imageRepository *Repository) {
	t.Helper()
	err := imageRepository.Db.ExecTx(func(tx *sql.Tx) error {
		_, execErr := tx.Exec(`UPDATE image_metadata SET lens_model = COALESCE(lens_model, ''), datetime_original = COALESCE(datetime_original, ''), exposure_time = COALESCE(exposure_time, 0), f_number = COALESCE(f_number, 0), iso = COALESCE(iso, 0), focal_length = COALESCE(focal_length, 0)`)
		return execErr
	})
	if err != nil {
		t.Fatalf("fill text columns: %v", err)
	}
}

func TestUpdateAIClassificationPersistsContent_Postgres(t *testing.T) {
	_, imageRepository, idsByName := seedContentLibrary(t)

	summaryRepository := NewImageSummaryRepository(imageRepository.Db)
	summary, err := summaryRepository.GetImageSummaryByFileID(idsByName["IMG_0001.jpg"])
	if err != nil {
		t.Fatalf("summary: %v", err)
	}
	if summary.Caption != "A red car parked by the beach" || summary.OCRText != "PLATE 42" || len(summary.Tags) != 2 || summary.Tags[0] != "car" {
		t.Fatalf("unexpected content in summary %+v", summary)
	}

	var searchText string
	if err := imageRepository.Db.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(`SELECT ai_search_text FROM image_metadata WHERE file_id = $1`, idsByName["IMG_0001.jpg"]).Scan(&searchText)
	}); err != nil {
		t.Fatalf("select search text: %v", err)
	}
	if searchText != "a red car parked by the beach car beach plate 42" {
		t.Fatalf("unexpected search text %q", searchText)
	}

	untouchedSummary, err := summaryRepository.GetImageSummaryByFileID(idsByName["IMG_0004.jpg"])
	if err != nil || untouchedSummary.Caption != "" || len(untouchedSummary.Tags) != 0 || untouchedSummary.Tags == nil {
		t.Fatalf("image without AI content must expose empty caption and empty tags, got %+v err=%v", untouchedSummary, err)
	}
}

func TestListPendingAIClassificationIncludesClassifiedWithoutCaption_Postgres(t *testing.T) {
	_, imageRepository, idsByName := seedContentLibrary(t)

	err := imageRepository.Db.ExecTx(func(tx *sql.Tx) error {
		_, execErr := tx.Exec(`UPDATE image_metadata SET ai_classified_at = now(), classification_confidence = 0.95 WHERE file_id = $1`, idsByName["IMG_0004.jpg"])
		return execErr
	})
	if err != nil {
		t.Fatalf("mark previously classified: %v", err)
	}

	pendingImages, err := imageRepository.ListPendingAIClassification(AIClassificationConfidenceThreshold, 0, 10)
	if err != nil {
		t.Fatalf("list pending: %v", err)
	}
	pendingFileIDs := map[int]bool{}
	for _, pendingImage := range pendingImages {
		pendingFileIDs[pendingImage.FileID] = true
	}
	if !pendingFileIDs[idsByName["IMG_0004.jpg"]] {
		t.Fatalf("classified image without caption must be pending, got %v", pendingImages)
	}
	if pendingFileIDs[idsByName["IMG_0001.jpg"]] || pendingFileIDs[idsByName["IMG_0002.jpg"]] {
		t.Fatalf("images with caption must not be pending, got %v", pendingImages)
	}

	pendingCount, err := imageRepository.CountPendingAIClassification(AIClassificationConfidenceThreshold)
	if err != nil || pendingCount != len(pendingImages) {
		t.Fatalf("count = %d err=%v, want %d", pendingCount, err, len(pendingImages))
	}

	firstPage, err := imageRepository.ListPendingAIClassification(AIClassificationConfidenceThreshold, 0, 1)
	if err != nil || len(firstPage) != 1 {
		t.Fatalf("pending must stay paged, got %v err=%v", firstPage, err)
	}
}

func TestLibraryContentSearch_Postgres(t *testing.T) {
	libraryRepository, _, _ := seedContentLibrary(t)

	testCases := []struct {
		label    string
		filter   LibraryFilter
		expected []string
	}{
		{"content only", LibraryFilter{ContentQuery: "red car"}, []string{"IMG_0001.jpg"}},
		{"content matches a tag", LibraryFilter{ContentQuery: "dog"}, []string{"IMG_0002.jpg"}},
		{"content matches ocr text", LibraryFilter{ContentQuery: "plate 42"}, []string{"IMG_0001.jpg"}},
		{"content ignores images without ai content", LibraryFilter{ContentQuery: "img"}, []string{}},
		{"name or content", LibraryFilter{NameQuery: "car", ContentQuery: "car"}, []string{"red_car_name.jpg", "IMG_0001.jpg"}},
		{"name and content", LibraryFilter{NameQuery: "car", ContentQuery: "car", MustMatchNameAndContent: true}, []string{}},
		{"name and content intersect", LibraryFilter{NameQuery: "img_0001", ContentQuery: "beach", MustMatchNameAndContent: true}, []string{"IMG_0001.jpg"}},
		{"content with other filters", LibraryFilter{ContentQuery: "a ", Formats: []string{".png"}}, []string{}},
		{"literal percent is not a wildcard", LibraryFilter{ContentQuery: "%"}, []string{}},
	}
	for _, testCase := range testCases {
		query := LibraryListQuery{Filter: testCase.filter, Sort: LibrarySortTakenAt, Order: LibrarySortOrderDesc, Limit: 50}
		assertNames(t, testCase.label, listNames(t, libraryRepository, query), testCase.expected)

		total, err := libraryRepository.CountLibraryImages(testCase.filter)
		if err != nil || total != len(testCase.expected) {
			t.Fatalf("%s: count = %d err %v, want %d", testCase.label, total, err, len(testCase.expected))
		}
	}
}

func TestLibraryContentSearchWorksWithoutTrigramIndex_Postgres(t *testing.T) {
	libraryRepository, _, _ := seedContentLibrary(t)

	dropErr := libraryRepository.Db.ExecTx(func(tx *sql.Tx) error {
		_, execErr := tx.Exec(`DROP INDEX IF EXISTS idx_image_metadata_ai_search_text_trigram`)
		return execErr
	})
	if dropErr != nil {
		t.Fatalf("drop index: %v", dropErr)
	}

	t.Cleanup(func() {
		_ = libraryRepository.Db.ExecTx(func(tx *sql.Tx) error {
			_, execErr := tx.Exec(`DO $$ BEGIN
				IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'pg_trgm') THEN
					CREATE INDEX IF NOT EXISTS idx_image_metadata_ai_search_text_trigram
						ON image_metadata USING gin (ai_search_text gin_trgm_ops) WHERE ai_search_text IS NOT NULL;
				END IF;
			END $$`)
			return execErr
		})
	})

	query := LibraryListQuery{Filter: LibraryFilter{NameQuery: "dog", ContentQuery: "dog"}, Sort: LibrarySortTakenAt, Order: LibrarySortOrderDesc, Limit: 50}
	assertNames(t, "content search without trigram index", listNames(t, libraryRepository, query), []string{"IMG_0002.jpg"})
}

func TestMigrationCreatesContentTrigramIndexWhenExtensionIsAvailable_Postgres(t *testing.T) {
	libraryRepository := newLibraryPostgresRepository(t)

	var isTrigramAvailable, hasIndex bool
	err := libraryRepository.Db.QueryTx(func(tx *sql.Tx) error {
		if scanErr := tx.QueryRow(`SELECT EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'pg_trgm')`).Scan(&isTrigramAvailable); scanErr != nil {
			return scanErr
		}
		return tx.QueryRow(`SELECT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_image_metadata_ai_search_text_trigram')`).Scan(&hasIndex)
	})
	if err != nil {
		t.Fatalf("inspect index: %v", err)
	}
	if isTrigramAvailable && !hasIndex {
		t.Fatal("pg_trgm is available but the ai_search_text trigram index is missing")
	}
}

func TestGlobalImageSearchMatchesAIContent_Postgres(t *testing.T) {
	libraryRepository, _, idsByName := seedContentLibrary(t)

	var matchedFileIDs []int
	err := libraryRepository.Db.QueryTx(func(tx *sql.Tx) error {
		rows, queryErr := tx.Query(searchqueries.SearchImagesQuery, "Beach", pq.Array([]string{".jpg"}), 10)
		if queryErr != nil {
			return queryErr
		}
		defer rows.Close()
		for rows.Next() {
			var fileID int
			var name, path, parentPath, format, category, camera string
			if scanErr := rows.Scan(&fileID, &name, &path, &parentPath, &format, &category, &camera); scanErr != nil {
				return scanErr
			}
			matchedFileIDs = append(matchedFileIDs, fileID)
		}
		return rows.Err()
	})
	if err != nil {
		t.Fatalf("global search: %v", err)
	}
	if len(matchedFileIDs) != 1 || matchedFileIDs[0] != idsByName["IMG_0001.jpg"] {
		t.Fatalf("expected only the image whose AI content mentions the beach, got %v", matchedFileIDs)
	}
}
