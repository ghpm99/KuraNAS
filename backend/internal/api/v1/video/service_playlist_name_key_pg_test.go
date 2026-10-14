package video

import (
	"database/sql"
	"testing"

	"nas-go/api/internal/testutil"
	"nas-go/api/pkg/database/migrations"
)

type legacyPlaylistRow struct {
	playlistType string
	sourcePath   string
	name         string
	isAuto       bool
}

func insertLegacyPlaylists(t *testing.T, repository *Repository, rows []legacyPlaylistRow) {
	t.Helper()
	seedPlaylists(t, repository, nil)
	insertErr := repository.DbContext.ExecTx(func(tx *sql.Tx) error {
		for _, row := range rows {
			if _, err := tx.Exec(
				`INSERT INTO video_playlist (type, source_path, name, is_auto) VALUES ($1, $2, $3, $4)`,
				row.playlistType, row.sourcePath, row.name, row.isAuto,
			); err != nil {
				return err
			}
		}
		return nil
	})
	if insertErr != nil {
		t.Fatalf("insert legacy playlists: %v", insertErr)
	}
}

func playlistIDBySourcePath(t *testing.T, repository *Repository, sourcePath string) int {
	t.Helper()
	var playlistID int
	queryErr := repository.DbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(`SELECT id FROM video_playlist WHERE source_path = $1`, sourcePath).Scan(&playlistID)
	})
	if queryErr != nil {
		t.Fatalf("read id of %s: %v", sourcePath, queryErr)
	}
	return playlistID
}

func storedNameKey(t *testing.T, repository *Repository, playlistType, sourcePath string) sql.NullString {
	t.Helper()
	var nameKey sql.NullString
	queryErr := repository.DbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(
			`SELECT name_key FROM video_playlist WHERE type = $1 AND source_path = $2`, playlistType, sourcePath,
		).Scan(&nameKey)
	})
	if queryErr != nil {
		t.Fatalf("read name_key of %s %s: %v", playlistType, sourcePath, queryErr)
	}
	return nameKey
}

func TestBackfillNameKeyMigrationKeysOnlyGeneratedAutomaticPlaylists_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_video_it")
	repository := NewRepository(dbContext)
	insertLegacyPlaylists(t, repository, []legacyPlaylistRow{
		{"movie", "related:movie", "Filmes", true},
		{"series", "related:series", "Series", true},
		{"course", "related:course", "Cursos", true},
		{"folder", "related:personal", "Videos pessoais", true},
		{"continue", "continue:client-1:7", "Continue assistindo", true},
		{"mixed", "favorites:client-1", "Recomendados para voce", true},
		{"mixed", "random:smart_mix", "Mix variado", true},
		{"folder", "folder:/media/Filmes", "Filmes", true},
		{"series", "series:filmes", "Filmes", true},
		{"movie", "related:anime", "Meus animes favoritos", true},
		{"custom", "related:movie", "Filmes", false},
		{"custom", "/mine", "Mix variado", false},
	})

	if _, err := dbContext.GetDatabase().Exec(migrations.BackfillVideoPlaylistNameKeyQuery); err != nil {
		t.Fatalf("backfill: %v", err)
	}

	keyed := map[[2]string]string{
		{"movie", "related:movie"}:          "VIDEO_PLAYLIST_NAME_MOVIES",
		{"series", "related:series"}:        "VIDEO_PLAYLIST_NAME_SERIES",
		{"course", "related:course"}:        "VIDEO_PLAYLIST_NAME_COURSES",
		{"folder", "related:personal"}:      "VIDEO_PLAYLIST_NAME_PERSONAL",
		{"continue", "continue:client-1:7"}: "VIDEO_PLAYLIST_NAME_CONTINUE_WATCHING",
		{"mixed", "favorites:client-1"}:     "VIDEO_PLAYLIST_NAME_RECOMMENDED",
		{"mixed", "random:smart_mix"}:       "VIDEO_PLAYLIST_NAME_MIXED",
	}
	for identity, wantKey := range keyed {
		if got := storedNameKey(t, repository, identity[0], identity[1]); got.String != wantKey {
			t.Errorf("%v name_key = %+v, want %s", identity, got, wantKey)
		}
	}

	untouched := [][2]string{
		{"folder", "folder:/media/Filmes"},
		{"series", "series:filmes"},
		{"movie", "related:anime"},
		{"custom", "related:movie"},
		{"custom", "/mine"},
	}
	for _, identity := range untouched {
		if got := storedNameKey(t, repository, identity[0], identity[1]); got.Valid {
			t.Errorf("%v must stay without name_key, got %+v", identity, got)
		}
	}

	if _, err := dbContext.GetDatabase().Exec(migrations.BackfillVideoPlaylistNameKeyQuery); err != nil {
		t.Fatalf("backfill must be idempotent: %v", err)
	}
}

func TestRebuildFillsNameKeyOnExistingAutomaticPlaylistsAndSkipsCustom_Postgres(t *testing.T) {
	useTranslationLocale(t, "pt-BR")
	dbContext := testutil.NewPostgresDB(t, "kuranas_video_it")
	repository := NewRepository(dbContext)
	service := NewService(repository).(*Service)
	insertLegacyPlaylists(t, repository, []legacyPlaylistRow{
		{"mixed", "random:smart_mix", "Mix variado", true},
		{"custom", "/mine", "Mix variado", false},
	})
	seedVideoFiles(t, repository, []seededVideoFile{
		{name: "praia.mp4", folder: "/camera/Viagem"},
		{name: "festa.mp4", folder: "/camera/Festas"},
		{name: "natal.mp4", folder: "/camera/Natal"},
		{name: "aniversario.mp4", folder: "/camera/Aniversario"},
		{name: "passeio.mp4", folder: "/camera/Passeio"},
	})

	if err := service.RebuildSmartPlaylists(); err != nil {
		t.Fatalf("rebuild: %v", err)
	}

	if got := storedNameKey(t, repository, "mixed", "random:smart_mix"); got.String != "VIDEO_PLAYLIST_NAME_MIXED" {
		t.Fatalf("automatic playlist name_key = %+v", got)
	}
	if got := storedNameKey(t, repository, "custom", "/mine"); got.Valid {
		t.Fatalf("custom playlist must not get a name_key, got %+v", got)
	}
	customPlaylistID := playlistIDBySourcePath(t, repository, "/mine")
	customPlaylist, err := repository.GetVideoPlaylistByID(customPlaylistID)
	if err != nil || customPlaylist.Name != "Mix variado" || customPlaylist.IsAuto {
		t.Fatalf("custom playlist altered: %+v err=%v", customPlaylist, err)
	}

	playlists, err := service.GetPlaylists(true)
	if err != nil {
		t.Fatalf("GetPlaylists: %v", err)
	}
	namesBySourcePath := map[string]VideoPlaylistDto{}
	for _, playlist := range playlists {
		namesBySourcePath[playlist.SourcePath] = playlist
	}
	if mixed := namesBySourcePath["random:smart_mix"]; mixed.Name != "Mix variado" || mixed.NameKey == nil || *mixed.NameKey != "VIDEO_PLAYLIST_NAME_MIXED" {
		t.Fatalf("mixed dto = %+v", mixed)
	}

	useTranslationLocale(t, "en-US")
	playlists, err = service.GetPlaylists(true)
	if err != nil {
		t.Fatalf("GetPlaylists en-US: %v", err)
	}
	for _, playlist := range playlists {
		if playlist.SourcePath == "random:smart_mix" && playlist.Name != "Mixed picks" {
			t.Fatalf("name must follow LANGUAGE without rebuild, got %q", playlist.Name)
		}
		if playlist.SourcePath == "/mine" && playlist.Name != "Mix variado" {
			t.Fatalf("custom playlist name must never be translated, got %q", playlist.Name)
		}
	}
}

func TestRenamingAutomaticPlaylistClearsNameKey_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_video_it")
	repository := NewRepository(dbContext)
	service := NewService(repository).(*Service)
	insertLegacyPlaylists(t, repository, []legacyPlaylistRow{{"mixed", "random:smart_mix", "Mix variado", true}})
	if _, err := dbContext.GetDatabase().Exec(migrations.BackfillVideoPlaylistNameKeyQuery); err != nil {
		t.Fatalf("backfill: %v", err)
	}
	playlistID := playlistIDBySourcePath(t, repository, "random:smart_mix")

	if err := service.UpdatePlaylistName(playlistID, "Meu mix"); err != nil {
		t.Fatalf("rename: %v", err)
	}

	renamed, err := repository.GetVideoPlaylistByID(playlistID)
	if err != nil || renamed.Name != "Meu mix" || renamed.NameKey.Valid {
		t.Fatalf("renamed playlist = %+v err=%v", renamed, err)
	}
}
