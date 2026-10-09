package music

import (
	"database/sql"
	"testing"

	"nas-go/api/internal/testutil"
)

func TestListAudioWithoutMetadata_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_music_it")
	repository := NewAudioMetadataRepository(dbContext)

	seedErr := dbContext.ExecTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(`TRUNCATE playlist_track, playlist, audio_metadata, home_file RESTART IDENTITY CASCADE`); err != nil {
			return err
		}
		if _, err := tx.Exec(`INSERT INTO home_file (name, path, parent_path, format, size, updated_at, created_at, type, checksum, deleted_at)
			VALUES ('missing.m4a', '/m/missing.m4a', '/m', '.m4a', 1, now(), now(), 2, '', NULL),
			       ('indexed.mp3', '/m/indexed.mp3', '/m', '.mp3', 1, now(), now(), 2, '', NULL),
			       ('deleted.ogg', '/m/deleted.ogg', '/m', '.ogg', 1, now(), now(), 2, '', now()),
			       ('cover.jpg', '/m/cover.jpg', '/m', '.jpg', 1, now(), now(), 2, '', NULL),
			       ('missing2.opus', '/m/missing2.opus', '/m', '.opus', 1, now(), now(), 2, '', NULL)`); err != nil {
			return err
		}
		_, err := tx.Exec(`INSERT INTO audio_metadata (file_id, path) SELECT id, path FROM home_file WHERE name = 'indexed.mp3'`)
		return err
	})
	if seedErr != nil {
		t.Fatalf("seed: %v", seedErr)
	}

	firstPage, err := repository.ListAudioWithoutMetadata(0, 1)
	if err != nil {
		t.Fatalf("ListAudioWithoutMetadata returned error: %v", err)
	}
	if len(firstPage) != 1 || firstPage[0].Path != "/m/missing.m4a" {
		t.Fatalf("unexpected first page: %+v", firstPage)
	}

	secondPage, err := repository.ListAudioWithoutMetadata(firstPage[0].FileID, 10)
	if err != nil {
		t.Fatalf("ListAudioWithoutMetadata returned error: %v", err)
	}
	if len(secondPage) != 1 || secondPage[0].Path != "/m/missing2.opus" {
		t.Fatalf("expected only the remaining active audio without metadata, got %+v", secondPage)
	}
}

func TestUpsertAudioMetadataPersistsNumberingAndListsStaleTags_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_music_it")
	repository := NewAudioMetadataRepository(dbContext)

	seedErr := dbContext.ExecTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(`TRUNCATE playlist_track, playlist, audio_metadata, home_file RESTART IDENTITY CASCADE`); err != nil {
			return err
		}
		_, err := tx.Exec(`INSERT INTO home_file (name, path, parent_path, format, size, updated_at, created_at, type, checksum, deleted_at)
			VALUES ('empty_tags.flac', '/m/empty_tags.flac', '/m', '.flac', 1, now(), now(), 2, '', NULL),
			       ('no_year.ogg', '/m/no_year.ogg', '/m', '.ogg', 1, now(), now(), 2, '', NULL),
			       ('complete.mp3', '/m/complete.mp3', '/m', '.mp3', 1, now(), now(), 2, '', NULL),
			       ('deleted.flac', '/m/deleted.flac', '/m', '.flac', 1, now(), now(), 2, '', now()),
			       ('fresh.m4a', '/m/fresh.m4a', '/m', '.m4a', 1, now(), now(), 2, '', NULL)`)
		if err != nil {
			return err
		}
		_, err = tx.Exec(`INSERT INTO audio_metadata (file_id, path, title, artist, album, year, tags_extracted_version)
			SELECT id, path,
			       CASE name WHEN 'complete.mp3' THEN 'T' WHEN 'no_year.ogg' THEN 'T' ELSE '' END,
			       CASE name WHEN 'complete.mp3' THEN 'A' WHEN 'no_year.ogg' THEN 'A' ELSE '' END,
			       CASE name WHEN 'complete.mp3' THEN 'B' WHEN 'no_year.ogg' THEN 'B' ELSE '' END,
			       CASE name WHEN 'complete.mp3' THEN '2019' ELSE '' END,
			       CASE name WHEN 'fresh.m4a' THEN 2 ELSE 0 END
			FROM home_file`)
		return err
	})
	if seedErr != nil {
		t.Fatalf("seed: %v", seedErr)
	}

	stalePage, err := repository.ListAudioWithStaleTags(0, 10)
	if err != nil {
		t.Fatalf("ListAudioWithStaleTags returned error: %v", err)
	}
	stalePaths := []string{}
	for _, staleAudio := range stalePage {
		stalePaths = append(stalePaths, staleAudio.Path)
	}
	if len(stalePaths) != 2 || stalePaths[0] != "/m/empty_tags.flac" || stalePaths[1] != "/m/no_year.ogg" {
		t.Fatalf("expected empty-tag and no-year audio only, got %v", stalePaths)
	}

	var fileID int
	if err := dbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(`SELECT id FROM home_file WHERE name = 'empty_tags.flac'`).Scan(&fileID)
	}); err != nil {
		t.Fatalf("lookup file: %v", err)
	}

	upsertErr := dbContext.ExecTx(func(tx *sql.Tx) error {
		_, err := repository.UpsertAudioMetadata(tx, AudioMetadataModel{
			FileId: fileID, Path: "/m/empty_tags.flac",
			Title: "Song", Artist: "Artist", Album: "Album", Year: "2019",
			TrackNumber: "3/12", DiscNumberText: "1/2", Lyrics: "la la",
		})
		return err
	})
	if upsertErr != nil {
		t.Fatalf("upsert: %v", upsertErr)
	}

	var trackNo, trackTotal, discNumber, discTotal, extractedVersion int
	var lyrics string
	if err := dbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(`SELECT track_no, track_total, disc_number, disc_total, tags_extracted_version, lyrics
			FROM audio_metadata WHERE file_id = $1`, fileID).
			Scan(&trackNo, &trackTotal, &discNumber, &discTotal, &extractedVersion, &lyrics)
	}); err != nil {
		t.Fatalf("read back: %v", err)
	}
	if trackNo != 3 || trackTotal != 12 || discNumber != 1 || discTotal != 2 || lyrics != "la la" {
		t.Fatalf("unexpected persisted numbering: %d/%d disc %d/%d lyrics %q", trackNo, trackTotal, discNumber, discTotal, lyrics)
	}
	if extractedVersion != CurrentAudioTagsExtractedVersion {
		t.Fatalf("expected tags_extracted_version %d, got %d", CurrentAudioTagsExtractedVersion, extractedVersion)
	}

	remainingStale, err := repository.ListAudioWithStaleTags(0, 10)
	if err != nil {
		t.Fatalf("ListAudioWithStaleTags returned error: %v", err)
	}
	if len(remainingStale) != 1 || remainingStale[0].Path != "/m/no_year.ogg" {
		t.Fatalf("re-extracted audio must leave the stale set, got %+v", remainingStale)
	}
}
