package music

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/lib/pq"
)

func genreGroupsByKey(t *testing.T, environment *catalogTestEnvironment) map[string]MusicGenreGroupDto {
	t.Helper()
	genreGroups, err := environment.service.GetLibraryGenres(1, 50, DefaultCatalogSort())
	if err != nil {
		t.Fatalf("genre groups: %v", err)
	}
	groupsByKey := map[string]MusicGenreGroupDto{}
	for _, genreGroup := range genreGroups.Items {
		groupsByKey[genreGroup.Key] = genreGroup
	}
	return groupsByKey
}

func seedGenreTracks(t *testing.T, environment *catalogTestEnvironment, genresByTrackName map[string]string) {
	t.Helper()
	fixtures := []catalogTrackFixture{}
	for trackName, genre := range genresByTrackName {
		fixtures = append(fixtures, catalogTrackFixture{
			name: trackName, parentPath: "/m", updatedAt: environment.catalogBaseMoment,
			metadata: &AudioMetadataModel{Title: trackName, Artist: "Artist", Album: "Album", Genre: genre},
		})
	}
	environment.seedTracks(t, fixtures)
}

func TestLibraryGenreGroupsMergeGenresThatDifferOnlyInAccents_Postgres(t *testing.T) {
	environment := newCatalogTestEnvironment(t)
	seedGenreTracks(t, environment, map[string]string{"accented.mp3": "Música", "plain.mp3": "Musica", "upper.mp3": "MÚSICA "})

	genreGroupsByKey := genreGroupsByKey(t, environment)

	want := map[string]MusicGenreGroupDto{"musica": {Key: "musica", Genre: "Música", TrackCount: 3}}
	if !reflect.DeepEqual(genreGroupsByKey, want) {
		t.Fatalf("genre groups = %+v, want %+v", genreGroupsByKey, want)
	}
}

func TestLibraryGenreGroupsListASlashSeparatedTrackUnderEachGenre_Postgres(t *testing.T) {
	environment := newCatalogTestEnvironment(t)
	seedGenreTracks(t, environment, map[string]string{"both.mp3": "Rock/Pop", "rock.mp3": "Rock", "soul.mp3": "R&B/Soul"})

	genreGroupsByKey := genreGroupsByKey(t, environment)

	if genreGroupsByKey["rock"].TrackCount != 2 || genreGroupsByKey["pop"].TrackCount != 1 {
		t.Fatalf("rock/pop groups = %+v", genreGroupsByKey)
	}
	if genreGroupsByKey["r&b/soul"].Genre != "R&B / Soul" && genreGroupsByKey["r&b / soul"].Genre != "R&B / Soul" {
		t.Fatalf("compound genre must stay one group, got %+v", genreGroupsByKey)
	}
	if len(genreGroupsByKey) != 3 {
		t.Fatalf("expected rock, pop and the compound genre only, got %+v", genreGroupsByKey)
	}
}

func TestBackfillRecomputesTracksWithAnOutdatedCatalogKeysVersion_Postgres(t *testing.T) {
	environment := newCatalogTestEnvironment(t)
	seedGenreTracks(t, environment, map[string]string{"stale.mp3": "Música", "current.mp3": "Rock"})
	staleFileID := environment.fileIDsByName["stale.mp3"]
	currentFileID := environment.fileIDsByName["current.mp3"]
	downgradeErr := environment.dbContext.ExecTx(func(tx *sql.Tx) error {
		_, err := tx.Exec(`UPDATE audio_metadata SET catalog_keys_version = 0, catalog_genre_keys = ARRAY['música']::TEXT[] WHERE file_id = $1`, staleFileID)
		return err
	})
	if downgradeErr != nil {
		t.Fatalf("downgrade stale track: %v", downgradeErr)
	}

	pendingRows, err := environment.audioRepository.ListAudioWithoutCatalogKeys(0, 10)
	if err != nil || len(pendingRows) != 1 || pendingRows[0].FileID != staleFileID {
		t.Fatalf("pending rows = %+v err=%v, want only the stale track", pendingRows, err)
	}
	recomputeErr := environment.dbContext.ExecTx(func(tx *sql.Tx) error {
		return environment.audioRepository.RecomputeCatalogKeys(tx, []int{staleFileID})
	})
	if recomputeErr != nil {
		t.Fatalf("recompute: %v", recomputeErr)
	}

	var genreKeys []string
	var keysVersion int
	var currentKeysVersion int
	readErr := environment.dbContext.QueryTx(func(tx *sql.Tx) error {
		if err := tx.QueryRow(`SELECT catalog_genre_keys, catalog_keys_version FROM audio_metadata WHERE file_id = $1`, staleFileID).Scan(pq.Array(&genreKeys), &keysVersion); err != nil {
			return err
		}
		return tx.QueryRow(`SELECT catalog_keys_version FROM audio_metadata WHERE file_id = $1`, currentFileID).Scan(&currentKeysVersion)
	})
	if readErr != nil {
		t.Fatalf("read back: %v", readErr)
	}
	if !reflect.DeepEqual(genreKeys, []string{"musica"}) || keysVersion != CurrentCatalogKeysVersion || currentKeysVersion != CurrentCatalogKeysVersion {
		t.Fatalf("genre keys = %v version = %d current = %d, want [musica] at version %d", genreKeys, keysVersion, currentKeysVersion, CurrentCatalogKeysVersion)
	}
	remainingRows, err := environment.audioRepository.ListAudioWithoutCatalogKeys(0, 10)
	if err != nil || len(remainingRows) != 0 {
		t.Fatalf("nothing may stay pending, got %+v err=%v", remainingRows, err)
	}
}
