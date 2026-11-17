package analytics

import (
	"testing"
)

func typeDistributionCountsByCategory(t *testing.T, repo *Repository) map[string]int {
	t.Helper()
	distribution, err := repo.GetTypeDistribution()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	countsByCategory := map[string]int{}
	for _, row := range distribution {
		countsByCategory[row.Category] = int(row.Count)
	}
	return countsByCategory
}

func TestPostgres_TypeDistributionClassifiesCanonicalFormats(t *testing.T) {
	repo := newDuplicateGroupsRepository(t)
	insertDuplicateFixtures(t, repo, []duplicateFileFixture{
		{name: "photo.cr2", size: 10},
		{name: "voice.opus", size: 10},
		{name: "phone.heif", size: 10},
		{name: "upper.NEF", size: 10},
		{name: "backup.zip", size: 10},
		{name: "unknown.xyz", size: 10},
		{name: "deleted.cr2", size: 10, isDeleted: true},
	})

	countsByCategory := typeDistributionCountsByCategory(t, repo)

	expected := map[string]int{"image": 3, "audio": 1, "archive": 1, "other": 1}
	for category, expectedCount := range expected {
		if countsByCategory[category] != expectedCount {
			t.Fatalf("category %s: expected %d, got %d (%v)", category, expectedCount, countsByCategory[category], countsByCategory)
		}
	}
	if len(countsByCategory) != len(expected) {
		t.Fatalf("unexpected categories: %v", countsByCategory)
	}
}

func TestPostgres_TypeDistributionMatchesDuplicateCategorization(t *testing.T) {
	repo := newDuplicateGroupsRepository(t)
	extensions := []string{".cr2", ".opus", ".heif", ".jpg", ".mp3", ".mkv", ".pdf", ".docx", ".md", ".xyz"}
	var fixtures []duplicateFileFixture
	for _, extension := range extensions {
		fixtures = append(fixtures, duplicateGroup("group"+extension, extension, 10, 2)...)
	}
	insertDuplicateFixtures(t, repo, fixtures)

	countsByCategory := typeDistributionCountsByCategory(t, repo)

	categories := []DuplicateCategory{
		DuplicateCategoryImage, DuplicateCategoryAudio, DuplicateCategoryVideo,
		DuplicateCategoryDocument, DuplicateCategoryOther,
	}
	for _, category := range categories {
		page, err := repo.GetDuplicateGroupsPage(category, 1, 50)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		copiesInCategory := 0
		for _, group := range page.Items {
			copiesInCategory += int(group.Copies)
		}
		if copiesInCategory == 0 || countsByCategory[string(category)] != copiesInCategory {
			t.Fatalf("category %s: distribution %d, duplicates %d", category, countsByCategory[string(category)], copiesInCategory)
		}
	}
}
