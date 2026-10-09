package music

import (
	"database/sql"
	"reflect"
	"testing"
)

func queueFileIDs(entries []MusicQueueEntryModel) []int {
	fileIDs := make([]int, 0, len(entries))
	for _, entry := range entries {
		fileIDs = append(fileIDs, entry.FileID)
	}
	return fileIDs
}

func TestQueueFollowsTrackListingOrderPerContext_Postgres(t *testing.T) {
	environment := newCatalogTestEnvironment(t)
	seedSortableLibrary(t, environment)
	repository := environment.service.Repository
	sort := DefaultCatalogSort()

	artists, err := repository.GetLibraryArtistGroups(1, 50, sort)
	if err != nil || len(artists.Items) == 0 {
		t.Fatalf("artists: %v %+v", err, artists)
	}
	albums, _ := repository.GetLibraryAlbumGroups(1, 50, sort)
	genres, _ := repository.GetLibraryGenreGroups(1, 50, sort)

	contexts := map[string]struct {
		listedIDs func() ([]int, error)
		queue     func() ([]MusicQueueEntryModel, error)
	}{
		"artist": {
			func() ([]int, error) {
				page, err := repository.GetLibraryTrackIDsByArtist(artists.Items[0].Key, 1, 50)
				return page.Items, err
			},
			func() ([]MusicQueueEntryModel, error) {
				return repository.GetLibraryQueueByArtist(artists.Items[0].Key, 50)
			},
		},
		"album": {
			func() ([]int, error) {
				page, err := repository.GetLibraryTrackIDsByAlbum(albums.Items[0].Key, 1, 50)
				return page.Items, err
			},
			func() ([]MusicQueueEntryModel, error) {
				return repository.GetLibraryQueueByAlbum(albums.Items[0].Key, 50)
			},
		},
		"genre": {
			func() ([]int, error) {
				page, err := repository.GetLibraryTrackIDsByGenre(genres.Items[0].Key, 1, 50)
				return page.Items, err
			},
			func() ([]MusicQueueEntryModel, error) {
				return repository.GetLibraryQueueByGenre(genres.Items[0].Key, 50)
			},
		},
		"folder": {
			func() ([]int, error) {
				page, err := repository.GetLibraryTrackIDsByFolder("/m", 1, 50)
				return page.Items, err
			},
			func() ([]MusicQueueEntryModel, error) {
				return repository.GetLibraryQueueByFolder("/m", 50)
			},
		},
	}

	for name, context := range contexts {
		listedIDs, err := context.listedIDs()
		if err != nil || len(listedIDs) == 0 {
			t.Fatalf("%s listing: %v %v", name, err, listedIDs)
		}
		entries, err := context.queue()
		if err != nil {
			t.Fatalf("%s queue: %v", name, err)
		}
		if got := queueFileIDs(entries); !reflect.DeepEqual(got, listedIDs) {
			t.Fatalf("%s queue order = %v, listing order = %v", name, got, listedIDs)
		}
	}
}

func TestQueueEntriesCarryPlayerFieldsAndHonourLimit_Postgres(t *testing.T) {
	environment := newCatalogTestEnvironment(t)
	seedSortableLibrary(t, environment)
	repository := environment.service.Repository

	entries, err := repository.GetLibraryQueueByFolder("/m/alpha", 2)
	if err != nil {
		t.Fatalf("queue: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("limit not applied: %d entries", len(entries))
	}
	first := entries[0]
	if first.FileID != environment.fileIDsByName["a1.mp3"] || first.Name != "a1.mp3" || first.Path != "/m/alpha/a1.mp3" ||
		first.Format != ".mp3" || first.Title != "A1" || first.Artist != "Artist A" || first.Album != "Old Album" {
		t.Fatalf("first entry = %+v", first)
	}
}

func TestPlaylistQueueUsesPositionAndSkipsDeletedFiles_Postgres(t *testing.T) {
	environment := newCatalogTestEnvironment(t)
	base := environment.catalogBaseMoment
	environment.seedTracks(t, []catalogTrackFixture{
		trackFixture("one.mp3", "/m", base, AudioMetadataModel{Title: "One", Artist: "X", Album: "Y"}),
		trackFixture("two.mp3", "/m", base, AudioMetadataModel{Title: "Two", Artist: "X", Album: "Y"}),
		trackFixture("three.mp3", "/m", base, AudioMetadataModel{Title: "Three", Artist: "X", Album: "Y"}),
		{name: "gone.mp3", parentPath: "/m", updatedAt: base, isDeleted: true, metadata: &AudioMetadataModel{Title: "Gone"}},
	})
	ids := environment.fileIDsByName

	var playlistID int
	seedErr := environment.dbContext.ExecTx(func(tx *sql.Tx) error {
		if err := tx.QueryRow(`INSERT INTO playlist (name) VALUES ('p') RETURNING id`).Scan(&playlistID); err != nil {
			return err
		}
		_, err := tx.Exec(`INSERT INTO playlist_track (playlist_id, file_id, position) VALUES ($1, $2, 1), ($1, $3, 2), ($1, $4, 3), ($1, $5, 4)`,
			playlistID, ids["three.mp3"], ids["gone.mp3"], ids["one.mp3"], ids["two.mp3"])
		return err
	})
	if seedErr != nil {
		t.Fatalf("seed playlist: %v", seedErr)
	}

	repository := environment.service.Repository
	entries, err := repository.GetPlaylistQueue(playlistID, 10)
	if err != nil {
		t.Fatalf("playlist queue: %v", err)
	}
	want := []int{ids["three.mp3"], ids["one.mp3"], ids["two.mp3"]}
	if got := queueFileIDs(entries); !reflect.DeepEqual(got, want) {
		t.Fatalf("playlist queue = %v, want %v", got, want)
	}

	byIDs, err := repository.GetLibraryQueueByFileIDs([]int{ids["two.mp3"], ids["one.mp3"], ids["gone.mp3"]})
	if err != nil {
		t.Fatalf("queue by ids: %v", err)
	}
	if got, want := queueFileIDs(byIDs), []int{ids["two.mp3"], ids["one.mp3"]}; !reflect.DeepEqual(got, want) {
		t.Fatalf("queue by ids = %v, want %v", got, want)
	}
}
