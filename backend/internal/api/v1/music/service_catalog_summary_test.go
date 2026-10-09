package music

import (
	"database/sql"
	"errors"
	"testing"
)

func TestCatalogSummaryServiceDelegatesToRepository(t *testing.T) {
	service := newMusicServiceForTest(t, &musicRepoMock{})

	albumSummary, err := service.GetLibraryAlbumSummary("album-key")
	if err != nil || albumSummary.Key != "album-key" {
		t.Fatalf("album summary = %+v err=%v", albumSummary, err)
	}
	artistSummary, err := service.GetLibraryArtistSummary("artist-key")
	if err != nil || artistSummary.Key != "artist-key" {
		t.Fatalf("artist summary = %+v err=%v", artistSummary, err)
	}
	genreSummary, err := service.GetLibraryGenreSummary("genre-key")
	if err != nil || genreSummary.Key != "genre-key" {
		t.Fatalf("genre summary = %+v err=%v", genreSummary, err)
	}
	folderSummary, err := service.GetLibraryFolderSummary("  /music/a ")
	if err != nil || folderSummary.Key != "/music/a" {
		t.Fatalf("folder summary = %+v err=%v", folderSummary, err)
	}
}

func TestFolderSummaryServiceRejectsBlankFolderAsNotFound(t *testing.T) {
	service := newMusicServiceForTest(t, &musicRepoMock{})
	if _, err := service.GetLibraryFolderSummary("   "); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("blank folder error = %v", err)
	}
}

func TestArtistAlbumsServiceNormalizesPagination(t *testing.T) {
	service := newMusicServiceForTest(t, &musicRepoMock{})
	albums, err := service.GetLibraryAlbumsByArtist("artist", 0, 0)
	if err != nil || len(albums.Items) != 1 || albums.Items[0].TrackCount != 1001 {
		t.Fatalf("albums = %+v err=%v", albums, err)
	}
}
