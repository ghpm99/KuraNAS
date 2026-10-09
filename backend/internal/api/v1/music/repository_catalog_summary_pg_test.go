package music

import (
	"database/sql"
	"errors"
	"reflect"
	"testing"
	"time"
)

func summaryFixtures(base time.Time) []catalogTrackFixture {
	return []catalogTrackFixture{
		{name: "d1t1.mp3", parentPath: "/music/band", updatedAt: base,
			metadata: &AudioMetadataModel{Title: "D1T1", Artist: "Band", Album: "Double", Genre: "Rock", TrackNumber: "1/2", DiscNumberText: "1/2", Year: "2001", Length: 100.4}},
		{name: "d2t1.mp3", parentPath: "/music/band", updatedAt: base.Add(time.Hour),
			metadata: &AudioMetadataModel{Title: "D2T1", Artist: "Band", Album: "Double", Genre: "Rock", TrackNumber: "1/1", DiscNumberText: "2/2", Year: "2001", Length: 200.3}},
		{name: "single.mp3", parentPath: "/music/band/live", updatedAt: base.Add(2 * time.Hour),
			metadata: &AudioMetadataModel{Title: "S", Artist: "Band", Album: "Aardvark", Genre: "Rock; Jazz", Year: "1990", Length: 50}},
		{name: "noyear.mp3", parentPath: "/music/band", updatedAt: base.Add(3 * time.Hour),
			metadata: &AudioMetadataModel{Title: "N", Artist: "Band", Album: "Zeta", Length: 10}},
		{name: "other.mp3", parentPath: "/music/other", updatedAt: base.Add(4 * time.Hour),
			metadata: &AudioMetadataModel{Title: "O", Artist: "Other", Album: "Elsewhere", Genre: "Jazz", Year: "2010", Length: 30}},
		{name: "gone.mp3", parentPath: "/music/band", updatedAt: base, isDeleted: true,
			metadata: &AudioMetadataModel{Title: "G", Artist: "Band", Album: "Double", Length: 999}},
	}
}

func TestAlbumSummaryReportsTotalsAndDiscCount_Postgres(t *testing.T) {
	environment := newCatalogTestEnvironment(t)
	environment.seedTracks(t, summaryFixtures(environment.catalogBaseMoment))

	summary, err := environment.service.GetLibraryAlbumSummary("band::double")
	if err != nil {
		t.Fatalf("album summary: %v", err)
	}
	expected := MusicAlbumSummaryDto{Key: "band::double", Name: "Double", Artist: "Band", Year: "2001", TrackCount: 2, TotalLengthSeconds: 301, DiscCount: 2}
	if summary != expected {
		t.Fatalf("album summary = %+v, want %+v", summary, expected)
	}

	singleDisc, err := environment.service.GetLibraryAlbumSummary("band::aardvark")
	if err != nil || singleDisc.DiscCount != 1 || singleDisc.TrackCount != 1 {
		t.Fatalf("single disc summary = %+v err=%v", singleDisc, err)
	}

	if _, err := environment.service.GetLibraryAlbumSummary("band::nope"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("unknown album error = %v", err)
	}
}

func TestArtistSummaryCountsAlbumsAndLength_Postgres(t *testing.T) {
	environment := newCatalogTestEnvironment(t)
	environment.seedTracks(t, summaryFixtures(environment.catalogBaseMoment))

	summary, err := environment.service.GetLibraryArtistSummary("band")
	if err != nil {
		t.Fatalf("artist summary: %v", err)
	}
	expected := MusicArtistSummaryDto{Key: "band", Name: "Band", TrackCount: 4, AlbumCount: 3, TotalLengthSeconds: 361}
	if summary != expected {
		t.Fatalf("artist summary = %+v, want %+v", summary, expected)
	}

	if _, err := environment.service.GetLibraryArtistSummary("nobody"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("unknown artist error = %v", err)
	}
}

func TestGenreSummaryCountsMultiGenreTracks_Postgres(t *testing.T) {
	environment := newCatalogTestEnvironment(t)
	environment.seedTracks(t, summaryFixtures(environment.catalogBaseMoment))

	summary, err := environment.service.GetLibraryGenreSummary("jazz")
	if err != nil {
		t.Fatalf("genre summary: %v", err)
	}
	expected := MusicGroupSummaryDto{Key: "jazz", Name: "Jazz", TrackCount: 2, TotalLengthSeconds: 80}
	if summary != expected {
		t.Fatalf("genre summary = %+v, want %+v", summary, expected)
	}

	if _, err := environment.service.GetLibraryGenreSummary("polka"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("unknown genre error = %v", err)
	}
}

func TestFolderSummaryMatchesFolderTrackListing_Postgres(t *testing.T) {
	environment := newCatalogTestEnvironment(t)
	environment.seedTracks(t, summaryFixtures(environment.catalogBaseMoment))

	summary, err := environment.service.GetLibraryFolderSummary("/music/band")
	if err != nil {
		t.Fatalf("folder summary: %v", err)
	}
	listing, err := environment.service.GetLibraryTracksByFolder("/music/band", 1, 50)
	if err != nil {
		t.Fatalf("folder tracks: %v", err)
	}
	if summary.TrackCount != len(listing.Items) || summary.TrackCount != 4 || summary.TotalLengthSeconds != 361 || summary.Name != "/music/band" {
		t.Fatalf("folder summary = %+v listing=%d", summary, len(listing.Items))
	}

	if _, err := environment.service.GetLibraryFolderSummary("/music/absent"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("unknown folder error = %v", err)
	}
}

func TestArtistAlbumsAreOrderedByYearThenNameAndPaginated_Postgres(t *testing.T) {
	environment := newCatalogTestEnvironment(t)
	environment.seedTracks(t, summaryFixtures(environment.catalogBaseMoment))

	firstPage, err := environment.service.GetLibraryAlbumsByArtist("band", 1, 2)
	if err != nil {
		t.Fatalf("artist albums: %v", err)
	}
	expectedFirstPage := []MusicAlbumGroupDto{
		{Key: "band::aardvark", Album: "Aardvark", Artist: "Band", Year: "1990", TrackCount: 1},
		{Key: "band::double", Album: "Double", Artist: "Band", Year: "2001", TrackCount: 2},
	}
	if !reflect.DeepEqual(firstPage.Items, expectedFirstPage) || !firstPage.Pagination.HasNext {
		t.Fatalf("first page = %+v", firstPage)
	}

	secondPage, err := environment.service.GetLibraryAlbumsByArtist("band", 2, 2)
	if err != nil {
		t.Fatalf("artist albums page 2: %v", err)
	}
	if len(secondPage.Items) != 1 || secondPage.Items[0].Key != "band::zeta" || secondPage.Pagination.HasNext {
		t.Fatalf("second page = %+v", secondPage)
	}

	unknown, err := environment.service.GetLibraryAlbumsByArtist("nobody", 1, 10)
	if err != nil || len(unknown.Items) != 0 {
		t.Fatalf("unknown artist albums = %+v err=%v", unknown, err)
	}
}
