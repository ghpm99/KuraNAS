package music

import (
	"sort"
	"testing"
)

const variousArtistsTestLabel = "Vários artistas"

func reconcileAlbumGroupingsOrFail(t *testing.T, environment *catalogTestEnvironment) {
	t.Helper()
	if _, err := environment.audioRepository.ReconcileAlbumGroupings(variousArtistsTestLabel); err != nil {
		t.Fatalf("ReconcileAlbumGroupings: %v", err)
	}
}

func albumTitlesAndKeys(t *testing.T, environment *catalogTestEnvironment) map[string]MusicAlbumGroupDto {
	t.Helper()
	albums, err := environment.service.GetLibraryAlbums(1, 50, DefaultCatalogSort())
	if err != nil {
		t.Fatalf("GetLibraryAlbums: %v", err)
	}
	albumsByKey := map[string]MusicAlbumGroupDto{}
	for _, album := range albums.Items {
		albumsByKey[album.Key] = album
	}
	return albumsByKey
}

func sortedKeys(albumsByKey map[string]MusicAlbumGroupDto) []string {
	keys := make([]string, 0, len(albumsByKey))
	for key := range albumsByKey {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func albumFixture(name string, parentPath string, artist string, albumArtist string, album string, year string, environment *catalogTestEnvironment) catalogTrackFixture {
	return catalogTrackFixture{name: name, parentPath: parentPath, updatedAt: environment.catalogBaseMoment,
		metadata: &AudioMetadataModel{Title: name, Artist: artist, AlbumArtist: albumArtist, Album: album, Year: year}}
}

func TestReconcileAlbumGroupingsMergesCompilationIntoVariousArtists_Postgres(t *testing.T) {
	environment := newCatalogTestEnvironment(t)
	environment.seedTracks(t, []catalogTrackFixture{
		albumFixture("one.mp3", "/music/hits", "Artist A", "", "Summer Hits", "2020", environment),
		albumFixture("two.mp3", "/music/hits", "Artist B", "", "Summer Hits", "2020", environment),
		albumFixture("three.mp3", "/music/hits", "Artist C", "", "Summer Hits", "2020", environment),
		albumFixture("solo.mp3", "/music/solo", "Artist A", "", "Solo Album", "2018", environment),
	})
	reconcileAlbumGroupingsOrFail(t, environment)

	albumsByKey := albumTitlesAndKeys(t, environment)
	if len(albumsByKey) != 2 {
		t.Fatalf("expected 2 albums, got %v", sortedKeys(albumsByKey))
	}

	var compilation MusicAlbumGroupDto
	for key, album := range albumsByKey {
		if key != "artist a::solo album" {
			compilation = album
		}
	}
	if compilation.Album != "Summer Hits" || compilation.Artist != variousArtistsTestLabel || compilation.TrackCount != 3 {
		t.Fatalf("unexpected compilation: %+v", compilation)
	}
	if solo := albumsByKey["artist a::solo album"]; solo.Artist != "Artist A" || solo.TrackCount != 1 {
		t.Fatalf("normal album must stay unchanged: %+v", solo)
	}

	summary, err := environment.service.GetLibraryAlbumSummary(compilation.Key)
	if err != nil {
		t.Fatalf("GetLibraryAlbumSummary: %v", err)
	}
	if summary.Name != "Summer Hits" || summary.Artist != variousArtistsTestLabel || summary.TrackCount != 3 {
		t.Fatalf("unexpected summary: %+v", summary)
	}

	queue, err := environment.service.GetLibraryQueueByAlbum(compilation.Key)
	if err != nil || len(queue.Items) != 3 {
		t.Fatalf("queue = %+v err=%v", queue, err)
	}

	artists, err := environment.service.GetLibraryArtists(1, 10, DefaultCatalogSort())
	if err != nil || len(artists.Items) != 3 {
		t.Fatalf("artists must stay per track artist: %+v err=%v", artists.Items, err)
	}
	artistAlbums, err := environment.service.GetLibraryAlbumsByArtist("artist b", 1, 10)
	if err != nil || len(artistAlbums.Items) != 1 || artistAlbums.Items[0].Key != compilation.Key {
		t.Fatalf("artist page must list the compilation: %+v err=%v", artistAlbums.Items, err)
	}
}

func TestReconcileAlbumGroupingsSplitsHomonymousAlbumsInDifferentFolders_Postgres(t *testing.T) {
	environment := newCatalogTestEnvironment(t)
	environment.seedTracks(t, []catalogTrackFixture{
		albumFixture("old1.mp3", "/music/band/old", "Band", "", "Greatest Hits", "1999", environment),
		albumFixture("old2.mp3", "/music/band/old", "Band", "", "Greatest Hits", "1999", environment),
		albumFixture("new1.mp3", "/music/band/new", "Band", "", "Greatest Hits", "2015", environment),
	})
	reconcileAlbumGroupingsOrFail(t, environment)

	albumsByKey := albumTitlesAndKeys(t, environment)
	if len(albumsByKey) != 2 {
		t.Fatalf("expected 2 albums, got %v", sortedKeys(albumsByKey))
	}
	trackCounts := []int{}
	for _, album := range albumsByKey {
		trackCounts = append(trackCounts, album.TrackCount)
		if album.Artist != "Band" || album.Album != "Greatest Hits" {
			t.Fatalf("unexpected album: %+v", album)
		}
	}
	sort.Ints(trackCounts)
	if trackCounts[0] != 1 || trackCounts[1] != 2 {
		t.Fatalf("unexpected track counts: %v", trackCounts)
	}
}

func TestReconcileAlbumGroupingsKeepsMultiDiscSubfoldersTogether_Postgres(t *testing.T) {
	environment := newCatalogTestEnvironment(t)
	environment.seedTracks(t, []catalogTrackFixture{
		albumFixture("d1t1.mp3", "/music/band/Big Album/CD1", "Band", "", "Big Album", "2001", environment),
		albumFixture("d2t1.mp3", "/music/band/Big Album/CD2", "Band", "", "Big Album", "2001", environment),
		albumFixture("d2t2.mp3", "/music/band/Big Album/Disc 2", "Band", "", "Big Album", "2002", environment),
	})
	reconcileAlbumGroupingsOrFail(t, environment)

	albumsByKey := albumTitlesAndKeys(t, environment)
	if len(albumsByKey) != 1 || albumsByKey["band::big album"].TrackCount != 3 {
		t.Fatalf("multi-disc album must stay one album: %v", albumsByKey)
	}
}

func TestReconcileAlbumGroupingsIsIdempotentAndRecomputesFromBaseKey_Postgres(t *testing.T) {
	environment := newCatalogTestEnvironment(t)
	environment.seedTracks(t, []catalogTrackFixture{
		albumFixture("one.mp3", "/music/hits", "Artist A", "", "Summer Hits", "2020", environment),
		albumFixture("two.mp3", "/music/hits", "Artist B", "", "Summer Hits", "2020", environment),
	})
	reconcileAlbumGroupingsOrFail(t, environment)

	regroupedAgain, err := environment.audioRepository.ReconcileAlbumGroupings(variousArtistsTestLabel)
	if err != nil || regroupedAgain != 0 {
		t.Fatalf("second pass must be a no-op, regrouped=%d err=%v", regroupedAgain, err)
	}

	environment.seedTracks(t, []catalogTrackFixture{
		albumFixture("three.mp3", "/music/hits", "Artist C", "", "Summer Hits", "2020", environment),
	})
	reconcileAlbumGroupingsOrFail(t, environment)

	albumsByKey := albumTitlesAndKeys(t, environment)
	if len(albumsByKey) != 1 {
		t.Fatalf("late track must join the compilation: %v", sortedKeys(albumsByKey))
	}
	for _, album := range albumsByKey {
		if album.TrackCount != 3 || album.Artist != variousArtistsTestLabel {
			t.Fatalf("unexpected compilation: %+v", album)
		}
	}
}
