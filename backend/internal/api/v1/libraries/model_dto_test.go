package libraries

import "testing"

func TestLibraryCategoryLabelKey(t *testing.T) {
	cases := []struct {
		category LibraryCategory
		expected string
	}{
		{LibraryCategoryImages, "LIBRARY_IMAGES"},
		{LibraryCategoryMusic, "LIBRARY_MUSIC"},
		{LibraryCategoryVideos, "LIBRARY_VIDEOS"},
		{LibraryCategoryDocuments, "LIBRARY_DOCUMENTS"},
		{LibraryCategory("other"), ""},
	}

	for _, testCase := range cases {
		if got := testCase.category.LabelKey(); got != testCase.expected {
			t.Fatalf("expected %s, got %s", testCase.expected, got)
		}
	}
}

func TestModelAndDtoConversion(t *testing.T) {
	model := LibraryModel{Category: LibraryCategoryImages, Path: "/data/Imagens"}
	dto := model.ToDto()
	if dto.Category != "images" || dto.Path != "/data/Imagens" {
		t.Fatalf("unexpected dto conversion: %+v", dto)
	}

	roundTrip := dto.ToModel()
	if roundTrip.Category != LibraryCategoryImages || roundTrip.Path != "/data/Imagens" {
		t.Fatalf("unexpected model conversion: %+v", roundTrip)
	}
}
