package video

import (
	"database/sql"
	"testing"

	"nas-go/api/internal/testutil"
)

func countRows(t *testing.T, repository *Repository, query string, args ...any) int {
	t.Helper()
	var total int
	err := repository.DbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(query, args...).Scan(&total)
	})
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	return total
}

func newCustomPlaylistFixture(t *testing.T) (*Repository, *Service) {
	t.Helper()
	dbContext := testutil.NewPostgresDB(t, "kuranas_video_it")
	repository := NewRepository(dbContext)
	seedPlaylists(t, repository, nil)
	return repository, NewService(repository).(*Service)
}

func TestCreateCustomPlaylistPersistsNonAutoCustomPlaylist_Postgres(t *testing.T) {
	repository, service := newCustomPlaylistFixture(t)

	created, err := service.CreateCustomPlaylist(CreateCustomPlaylistRequest{Name: "Favoritos"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.Type != "custom" || created.IsAuto || created.Classification != "personal" || created.Name != "Favoritos" {
		t.Fatalf("unexpected playlist %+v", created)
	}
	stored, err := repository.GetVideoPlaylistByID(created.ID)
	if err != nil || stored.IsAuto || stored.Type != "custom" {
		t.Fatalf("stored = %+v err=%v", stored, err)
	}
}

func TestCreateCustomPlaylistAllowsDuplicateNames_Postgres(t *testing.T) {
	repository, service := newCustomPlaylistFixture(t)

	for range 2 {
		if _, err := service.CreateCustomPlaylist(CreateCustomPlaylistRequest{Name: "Mesmo nome"}); err != nil {
			t.Fatalf("create: %v", err)
		}
	}
	if total := countRows(t, repository, `SELECT COUNT(*) FROM video_playlist WHERE name = 'Mesmo nome'`); total != 2 {
		t.Fatalf("playlists named alike = %d, want 2", total)
	}
}

func TestCreateCustomPlaylistWithVideoStoresManualItem_Postgres(t *testing.T) {
	repository, service := newCustomPlaylistFixture(t)
	videoID := seedVideoFiles(t, repository, []seededVideoFile{{name: "a.mp4", folder: "/data/A"}})["a.mp4"]

	created, err := service.CreateCustomPlaylist(CreateCustomPlaylistRequest{Name: "Com video", VideoID: &videoID})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	manualItems := countRows(t, repository, `SELECT COUNT(*) FROM video_playlist_item WHERE playlist_id = $1 AND video_id = $2 AND source_kind = 'manual'`, created.ID, videoID)
	if manualItems != 1 || created.ItemCount != 1 {
		t.Fatalf("manualItems=%d itemCount=%d", manualItems, created.ItemCount)
	}
}

func TestCreateCustomPlaylistWithUnknownVideoCreatesNothing_Postgres(t *testing.T) {
	repository, service := newCustomPlaylistFixture(t)

	unknownVideoID := 999999
	if _, err := service.CreateCustomPlaylist(CreateCustomPlaylistRequest{Name: "x", VideoID: &unknownVideoID}); err == nil {
		t.Fatal("expected error")
	}
	if total := countRows(t, repository, `SELECT COUNT(*) FROM video_playlist`); total != 0 {
		t.Fatalf("playlists = %d, want 0", total)
	}
}

func TestDeleteCustomPlaylistCascadesItemsAndKeepsReferences_Postgres(t *testing.T) {
	repository, service := newCustomPlaylistFixture(t)
	videoID := seedVideoFiles(t, repository, []seededVideoFile{{name: "a.mp4", folder: "/data/A"}})["a.mp4"]
	created, err := service.CreateCustomPlaylist(CreateCustomPlaylistRequest{Name: "Apagar", VideoID: &videoID})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	referenceErr := repository.DbContext.ExecTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(`INSERT INTO video_playback_state (client_id, playlist_id, video_id) VALUES ('c1', $1, $2)`, created.ID, videoID); err != nil {
			return err
		}
		_, err := tx.Exec(`INSERT INTO video_behavior_event (client_id, video_id, playlist_id, event_type) VALUES ('c1', $2, $1, 'play')`, created.ID, videoID)
		return err
	})
	if referenceErr != nil {
		t.Fatalf("seed references: %v", referenceErr)
	}

	if err := service.DeleteCustomPlaylist(created.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if total := countRows(t, repository, `SELECT COUNT(*) FROM video_playlist WHERE id = $1`, created.ID); total != 0 {
		t.Fatalf("playlist still present")
	}
	if total := countRows(t, repository, `SELECT COUNT(*) FROM video_playlist_item WHERE playlist_id = $1`, created.ID); total != 0 {
		t.Fatalf("items still present")
	}
	if total := countRows(t, repository, `SELECT COUNT(*) FROM video_playback_state WHERE client_id = 'c1' AND playlist_id IS NULL`); total != 1 {
		t.Fatalf("playback state not detached")
	}
	if total := countRows(t, repository, `SELECT COUNT(*) FROM video_behavior_event WHERE client_id = 'c1' AND playlist_id IS NULL`); total != 1 {
		t.Fatalf("behavior event not detached")
	}
}

func TestDeleteCustomPlaylistRefusesAutomaticPlaylist_Postgres(t *testing.T) {
	repository, service := newCustomPlaylistFixture(t)
	automaticID := seedPlaylists(t, repository, []seededPlaylist{{sourcePath: "/auto", isAuto: true}})["/auto"]

	if err := service.DeleteCustomPlaylist(automaticID); err != ErrPlaylistAutomaticNotDeletable {
		t.Fatalf("err = %v", err)
	}
	if total := countRows(t, repository, `SELECT COUNT(*) FROM video_playlist WHERE id = $1`, automaticID); total != 1 {
		t.Fatal("automatic playlist was deleted")
	}
	if err := service.DeleteCustomPlaylist(999999); err == nil {
		t.Fatal("expected not found error")
	}
}

func TestRebuildKeepsEmptyCustomPlaylist_Postgres(t *testing.T) {
	repository, service := newCustomPlaylistFixture(t)
	created, err := service.CreateCustomPlaylist(CreateCustomPlaylistRequest{Name: "Vazia"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := service.RebuildSmartPlaylists(); err != nil {
		t.Fatalf("rebuild: %v", err)
	}
	stored, err := repository.GetVideoPlaylistByID(created.ID)
	if err != nil {
		t.Fatalf("custom playlist lost after rebuild: %v", err)
	}
	if stored.Name != "Vazia" || stored.IsAuto {
		t.Fatalf("custom playlist altered: %+v", stored)
	}
}
