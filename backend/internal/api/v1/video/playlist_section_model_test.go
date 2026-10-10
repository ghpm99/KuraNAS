package video

import (
	"errors"
	"reflect"
	"testing"
)

func TestParsePlaylistSectionAcceptsOnlyKnownSections(t *testing.T) {
	for _, rawSection := range []string{"series", "movies", "personal", "clips", "folders"} {
		section, isValid := ParsePlaylistSection(rawSection)
		if !isValid || string(section) != rawSection {
			t.Fatalf("section %q: got %q valid=%v", rawSection, section, isValid)
		}
	}
	for _, rawSection := range []string{"", "anime", "home", "SERIES"} {
		if _, isValid := ParsePlaylistSection(rawSection); isValid {
			t.Fatalf("section %q must be invalid", rawSection)
		}
	}
}

func TestPlaylistSectionFilterMapsClassificationsAndType(t *testing.T) {
	expectedFilters := map[PlaylistSection]PlaylistSectionFilter{
		PlaylistSectionSeries:   {Classifications: []string{"series", "anime"}},
		PlaylistSectionMovies:   {Classifications: []string{"movie"}},
		PlaylistSectionPersonal: {Classifications: []string{"personal"}},
		PlaylistSectionClips:    {Classifications: []string{"clip", "program"}},
		PlaylistSectionFolders:  {Classifications: []string{}, PlaylistType: "folder"},
		PlaylistSection("zzz"):  {Classifications: []string{}},
	}
	for section, expectedFilter := range expectedFilters {
		if got := section.Filter(); !reflect.DeepEqual(got, expectedFilter) {
			t.Fatalf("section %s filter = %+v, want %+v", section, got, expectedFilter)
		}
	}
}

func TestGetPlaylistsBySectionPagesWithOneExtraRow(t *testing.T) {
	var receivedFilter PlaylistSectionFilter
	var receivedLimit, receivedOffset int
	repository := &videoRepoMock{getVideoPlaylistsBySectionFn: func(filter PlaylistSectionFilter, limit int, offset int) ([]VideoPlaylistModel, error) {
		receivedFilter, receivedLimit, receivedOffset = filter, limit, offset
		return []VideoPlaylistModel{{ID: 1}, {ID: 2}, {ID: 3}}, nil
	}}

	sectionPage, err := newVideoServiceForTest(t, repository).GetPlaylistsBySection(PlaylistSectionRequest{Section: PlaylistSectionClips, Page: 2, PageSize: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedLimit != 3 || receivedOffset != 2 || !reflect.DeepEqual(receivedFilter.Classifications, []string{"clip", "program"}) {
		t.Fatalf("unexpected window filter=%+v limit=%d offset=%d", receivedFilter, receivedLimit, receivedOffset)
	}
	if len(sectionPage.Items) != 2 || !sectionPage.Pagination.HasNext || !sectionPage.Pagination.HasPrev {
		t.Fatalf("unexpected page %+v", sectionPage)
	}
}

func TestGetPlaylistsBySectionPropagatesRepositoryError(t *testing.T) {
	repository := &videoRepoMock{getVideoPlaylistsBySectionFn: func(PlaylistSectionFilter, int, int) ([]VideoPlaylistModel, error) {
		return nil, errors.New("boom")
	}}
	if _, err := newVideoServiceForTest(t, repository).GetPlaylistsBySection(PlaylistSectionRequest{Section: PlaylistSectionSeries, Page: 1, PageSize: 5}); err == nil {
		t.Fatal("expected error")
	}
}
