package migrations_test

import (
	"database/sql"
	"testing"

	"nas-go/api/internal/testutil"
)

const truncateOverrideFixtureQuery = `TRUNCATE audio_metadata_override, playlist_track, playlist, audio_metadata, home_file RESTART IDENTITY CASCADE`

type effectiveRow struct {
	title       sql.NullString
	artist      sql.NullString
	album       sql.NullString
	albumArtist sql.NullString
	genre       sql.NullString
	year        sql.NullString
	trackNumber sql.NullString
	trackNo     sql.NullInt64
	discNumber  sql.NullInt64
}

func seedTaggedAudio(t *testing.T, tx *sql.Tx, name string) int {
	t.Helper()
	var fileID int
	err := tx.QueryRow(`INSERT INTO home_file (name, path, parent_path, format, size, updated_at, created_at, type, checksum)
		VALUES ($1, $2, '/m', '.mp3', 1, now(), now(), 2, '') RETURNING id`, name, "/m/"+name).Scan(&fileID)
	if err != nil {
		t.Fatalf("seed home_file: %v", err)
	}
	_, err = tx.Exec(`INSERT INTO audio_metadata (file_id, path, title, artist, album, album_artist, genre, year, track_number, track_no, disc_number)
		VALUES ($1, $2, 'Tag Title', 'Tag Artist', 'Tag Album', 'Tag Album Artist', 'Tag Genre', '1999', '3/12', 3, 1)`, fileID, "/m/"+name)
	if err != nil {
		t.Fatalf("seed audio_metadata: %v", err)
	}
	return fileID
}

func readEffective(t *testing.T, tx *sql.Tx, fileID int) effectiveRow {
	t.Helper()
	var row effectiveRow
	err := tx.QueryRow(`SELECT title, artist, album, album_artist, genre, year, track_number, track_no, disc_number
		FROM audio_metadata_effective WHERE file_id = $1`, fileID).Scan(
		&row.title, &row.artist, &row.album, &row.albumArtist, &row.genre, &row.year, &row.trackNumber, &row.trackNo, &row.discNumber)
	if err != nil {
		t.Fatalf("read effective: %v", err)
	}
	return row
}

func TestAudioMetadataEffectiveViewFallsBackToTagsWithoutOverride_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_migrations_it")

	err := dbContext.ExecTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(truncateOverrideFixtureQuery); err != nil {
			return err
		}
		fileID := seedTaggedAudio(t, tx, "plain.mp3")
		row := readEffective(t, tx, fileID)

		expected := effectiveRow{
			title:       sql.NullString{String: "Tag Title", Valid: true},
			artist:      sql.NullString{String: "Tag Artist", Valid: true},
			album:       sql.NullString{String: "Tag Album", Valid: true},
			albumArtist: sql.NullString{String: "Tag Album Artist", Valid: true},
			genre:       sql.NullString{String: "Tag Genre", Valid: true},
			year:        sql.NullString{String: "1999", Valid: true},
			trackNumber: sql.NullString{String: "3/12", Valid: true},
			trackNo:     sql.NullInt64{Int64: 3, Valid: true},
			discNumber:  sql.NullInt64{Int64: 1, Valid: true},
		}
		if row != expected {
			t.Errorf("without override got %+v, want %+v", row, expected)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("tx: %v", err)
	}
}

func TestAudioMetadataEffectiveViewPrefersOverrideFieldByField_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_migrations_it")

	err := dbContext.ExecTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(truncateOverrideFixtureQuery); err != nil {
			return err
		}
		fileID := seedTaggedAudio(t, tx, "fixed.mp3")
		if _, err := tx.Exec(`INSERT INTO audio_metadata_override (file_id, title, genre, track_number, disc_number)
			VALUES ($1, 'Fixed Title', 'Rock', 7, 2)`, fileID); err != nil {
			return err
		}
		row := readEffective(t, tx, fileID)

		expected := effectiveRow{
			title:       sql.NullString{String: "Fixed Title", Valid: true},
			artist:      sql.NullString{String: "Tag Artist", Valid: true},
			album:       sql.NullString{String: "Tag Album", Valid: true},
			albumArtist: sql.NullString{String: "Tag Album Artist", Valid: true},
			genre:       sql.NullString{String: "Rock", Valid: true},
			year:        sql.NullString{String: "1999", Valid: true},
			trackNumber: sql.NullString{String: "7", Valid: true},
			trackNo:     sql.NullInt64{Int64: 7, Valid: true},
			discNumber:  sql.NullInt64{Int64: 2, Valid: true},
		}
		if row != expected {
			t.Errorf("with override got %+v, want %+v", row, expected)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("tx: %v", err)
	}
}

func TestAudioMetadataEffectiveViewMirrorsTableColumns_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_migrations_it")

	const columnsQuery = `SELECT column_name, data_type FROM information_schema.columns WHERE table_name = $1 ORDER BY ordinal_position`
	readColumns := func(tx *sql.Tx, relation string) []string {
		rows, err := tx.Query(columnsQuery, relation)
		if err != nil {
			t.Fatalf("columns of %s: %v", relation, err)
		}
		defer rows.Close()
		columns := []string{}
		for rows.Next() {
			var name, dataType string
			if err := rows.Scan(&name, &dataType); err != nil {
				t.Fatalf("scan: %v", err)
			}
			columns = append(columns, name+" "+dataType)
		}
		return columns
	}

	err := dbContext.QueryTx(func(tx *sql.Tx) error {
		tableColumns := readColumns(tx, "audio_metadata")
		viewColumns := readColumns(tx, "audio_metadata_effective")
		if len(tableColumns) == 0 || len(tableColumns) != len(viewColumns) {
			t.Fatalf("table has %d columns, view has %d", len(tableColumns), len(viewColumns))
		}
		for position := range tableColumns {
			if tableColumns[position] != viewColumns[position] {
				t.Errorf("column %d: table %q, view %q", position, tableColumns[position], viewColumns[position])
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("tx: %v", err)
	}
}

func TestAudioMetadataOverrideIsDeletedWithHomeFile_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_migrations_it")

	err := dbContext.ExecTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(truncateOverrideFixtureQuery); err != nil {
			return err
		}
		fileID := seedTaggedAudio(t, tx, "doomed.mp3")
		if _, err := tx.Exec(`INSERT INTO audio_metadata_override (file_id, genre) VALUES ($1, 'Rock')`, fileID); err != nil {
			return err
		}
		if _, err := tx.Exec(`DELETE FROM home_file WHERE id = $1`, fileID); err != nil {
			return err
		}
		var remaining int
		if err := tx.QueryRow(`SELECT count(*) FROM audio_metadata_override WHERE file_id = $1`, fileID).Scan(&remaining); err != nil {
			return err
		}
		if remaining != 0 {
			t.Errorf("override rows after deleting home_file: %d, want 0", remaining)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("tx: %v", err)
	}
}
