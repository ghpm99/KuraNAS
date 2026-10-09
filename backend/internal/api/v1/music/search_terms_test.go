package music

import (
	"testing"

	"nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/utils"
)

func TestBuildTrackSearchPatterns(t *testing.T) {
	if _, hasTerms := buildTrackSearchPatterns("   "); hasTerms {
		t.Fatalf("blank text must not produce patterns")
	}

	patterns, hasTerms := buildTrackSearchPatterns("Queen queen Bohemian")
	if !hasTerms || len(patterns.allPatterns) != 2 {
		t.Fatalf("expected 2 distinct terms, got %+v", patterns)
	}
	if patterns.drivingPattern != "%bohemian%" {
		t.Fatalf("driving pattern should be longest term, got %q", patterns.drivingPattern)
	}

	capped, _ := buildTrackSearchPatterns("a b c d e f g")
	if len(capped.allPatterns) != maxSearchTerms {
		t.Fatalf("expected cap of %d terms, got %d", maxSearchTerms, len(capped.allPatterns))
	}
}

func TestServiceSearchLibraryTracksDelegatesToRepository(t *testing.T) {
	repository := &musicRepoMock{}
	var receivedText string
	repository.searchLibraryTracksFn = func(searchText string, page int, pageSize int) (utils.PaginationResponse[files.FileModel], error) {
		receivedText = searchText
		return utils.PaginationResponse[files.FileModel]{}, nil
	}
	service := &Service{Repository: repository}
	if _, err := service.SearchLibraryTracks("queen", 1, 10); err != nil || receivedText != "queen" {
		t.Fatalf("unexpected result err=%v text=%q", err, receivedText)
	}
}
