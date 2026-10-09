package image

import (
	"database/sql"
	"errors"
	"testing"

	"nas-go/api/internal/testutil"
)

type albumPostgresFixture struct {
	albums        *AlbumRepository
	library       *LibraryRepository
	fileIDsByName map[string]int
}

func newAlbumPostgresFixture(t *testing.T) albumPostgresFixture {
	t.Helper()
	dbContext := testutil.NewPostgresDB(t, "kuranas_image_it")
	library := NewLibraryRepository(dbContext)
	fileIDsByName := seedLibraryImages(t, library, []seededImage{
		{name: "old.jpg", folder: "/lib", format: ".jpg", takenAt: utcTime(2019, 1, 1)},
		{name: "mid.jpg", folder: "/lib", format: ".jpg", takenAt: utcTime(2021, 1, 1)},
		{name: "new.jpg", folder: "/lib", format: ".jpg", takenAt: utcTime(2024, 1, 1)},
		{name: "trashed.jpg", folder: "/lib", format: ".jpg", takenAt: utcTime(2023, 1, 1), deleted: true},
		{name: "clip.mp4", folder: "/lib", format: ".mp4"},
	})
	truncateErr := dbContext.ExecTx(func(tx *sql.Tx) error {
		_, err := tx.Exec(`TRUNCATE image_album RESTART IDENTITY CASCADE`)
		return err
	})
	if truncateErr != nil {
		t.Fatalf("truncate albums: %v", truncateErr)
	}
	return albumPostgresFixture{albums: NewAlbumRepository(dbContext), library: library, fileIDsByName: fileIDsByName}
}

func (fixture albumPostgresFixture) createAlbum(t *testing.T, name string) int {
	t.Helper()
	albumID, err := fixture.albums.CreateAlbum(name)
	if err != nil {
		t.Fatalf("create album %q: %v", name, err)
	}
	return albumID
}

func (fixture albumPostgresFixture) albumItemNames(t *testing.T, albumID int) []string {
	t.Helper()
	return listNames(t, fixture.library, LibraryListQuery{
		Filter: LibraryFilter{AlbumID: albumID},
		Sort:   LibrarySortTakenAt,
		Order:  LibrarySortOrderDesc,
		Limit:  100,
	})
}

func (fixture albumPostgresFixture) fileIDs(names ...string) []int {
	fileIDs := make([]int, 0, len(names))
	for _, name := range names {
		fileIDs = append(fileIDs, fixture.fileIDsByName[name])
	}
	return fileIDs
}

func TestAlbumNamesAreUnique_Postgres(t *testing.T) {
	fixture := newAlbumPostgresFixture(t)
	fixture.createAlbum(t, "Viagem")

	if _, err := fixture.albums.CreateAlbum("Viagem"); !errors.Is(err, ErrAlbumNameTaken) {
		t.Fatalf("duplicate create error = %v", err)
	}

	otherAlbumID := fixture.createAlbum(t, "Família")
	renameTaken := "Viagem"
	if err := fixture.albums.UpdateAlbum(AlbumUpdate{AlbumID: otherAlbumID, Name: &renameTaken}); !errors.Is(err, ErrAlbumNameTaken) {
		t.Fatalf("duplicate rename error = %v", err)
	}
	renameFree := "Família 2024"
	if err := fixture.albums.UpdateAlbum(AlbumUpdate{AlbumID: otherAlbumID, Name: &renameFree}); err != nil {
		t.Fatalf("rename: %v", err)
	}
	renamed, err := fixture.albums.GetAlbum(otherAlbumID)
	if err != nil || renamed.Name != renameFree {
		t.Fatalf("renamed album = %+v err %v", renamed, err)
	}
}

func TestAddAlbumItemsOnlyAcceptsActiveImagesAndIgnoresDuplicates_Postgres(t *testing.T) {
	fixture := newAlbumPostgresFixture(t)
	albumID := fixture.createAlbum(t, "Viagem")

	added, err := fixture.albums.AddAlbumItems(albumID, fixture.fileIDs("old.jpg", "new.jpg", "trashed.jpg", "clip.mp4"))
	if err != nil || added != 2 {
		t.Fatalf("first add = %d err %v", added, err)
	}
	addedAgain, err := fixture.albums.AddAlbumItems(albumID, fixture.fileIDs("old.jpg", "mid.jpg"))
	if err != nil || addedAgain != 1 {
		t.Fatalf("second add = %d err %v", addedAgain, err)
	}

	assertNames(t, "album items", fixture.albumItemNames(t, albumID), []string{"new.jpg", "mid.jpg", "old.jpg"})
}

func TestAlbumListingReportsCountAndEffectiveCover_Postgres(t *testing.T) {
	fixture := newAlbumPostgresFixture(t)
	albumID := fixture.createAlbum(t, "Viagem")
	emptyAlbumID := fixture.createAlbum(t, "Vazio")

	if _, err := fixture.albums.AddAlbumItems(albumID, fixture.fileIDs("old.jpg")); err != nil {
		t.Fatalf("add old: %v", err)
	}
	if _, err := fixture.albums.AddAlbumItems(albumID, fixture.fileIDs("mid.jpg")); err != nil {
		t.Fatalf("add mid: %v", err)
	}

	album, err := fixture.albums.GetAlbum(albumID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if album.ItemCount != 2 || album.CoverFileID == nil || *album.CoverFileID != fixture.fileIDsByName["mid.jpg"] {
		t.Fatalf("most recently added must be the default cover, got %+v", album)
	}

	oldFileID := fixture.fileIDsByName["old.jpg"]
	if err := fixture.albums.UpdateAlbum(AlbumUpdate{AlbumID: albumID, CoverFileID: &oldFileID}); err != nil {
		t.Fatalf("set cover: %v", err)
	}
	album, _ = fixture.albums.GetAlbum(albumID)
	if album.CoverFileID == nil || *album.CoverFileID != oldFileID {
		t.Fatalf("explicit cover not honored: %+v", album)
	}

	albums, err := fixture.albums.ListAlbums(10, 0)
	if err != nil || len(albums) != 2 {
		t.Fatalf("list = %+v err %v", albums, err)
	}
	if albums[1].ID != albumID || albums[0].ID != emptyAlbumID || albums[0].CoverFileID != nil || albums[0].ItemCount != 0 {
		t.Fatalf("unexpected listing order or empty album shape: %+v", albums)
	}

	pagedAlbums, err := fixture.albums.ListAlbums(1, 1)
	if err != nil || len(pagedAlbums) != 1 || pagedAlbums[0].ID != albumID {
		t.Fatalf("offset page = %+v err %v", pagedAlbums, err)
	}
}

func TestSetAlbumCoverRequiresAlbumMember_Postgres(t *testing.T) {
	fixture := newAlbumPostgresFixture(t)
	albumID := fixture.createAlbum(t, "Viagem")
	outsiderFileID := fixture.fileIDsByName["new.jpg"]

	err := fixture.albums.UpdateAlbum(AlbumUpdate{AlbumID: albumID, CoverFileID: &outsiderFileID})
	if !errors.Is(err, ErrAlbumCoverNotInAlbum) {
		t.Fatalf("cover outside album error = %v", err)
	}
}

func TestTrashedPhotosAreHiddenFromAlbumListingAndCount_Postgres(t *testing.T) {
	fixture := newAlbumPostgresFixture(t)
	albumID := fixture.createAlbum(t, "Viagem")
	if _, err := fixture.albums.AddAlbumItems(albumID, fixture.fileIDs("old.jpg", "new.jpg")); err != nil {
		t.Fatalf("add: %v", err)
	}

	trashErr := fixture.albums.Db.ExecTx(func(tx *sql.Tx) error {
		_, err := tx.Exec(`UPDATE home_file SET deleted_at = now() WHERE id = $1`, fixture.fileIDsByName["new.jpg"])
		return err
	})
	if trashErr != nil {
		t.Fatalf("trash: %v", trashErr)
	}

	assertNames(t, "listing without trashed", fixture.albumItemNames(t, albumID), []string{"old.jpg"})
	album, _ := fixture.albums.GetAlbum(albumID)
	if album.ItemCount != 1 || album.CoverFileID == nil || *album.CoverFileID != fixture.fileIDsByName["old.jpg"] {
		t.Fatalf("count and cover must ignore trashed photos: %+v", album)
	}
}

func TestRemoveAlbumItemsKeepsPhotosInGallery_Postgres(t *testing.T) {
	fixture := newAlbumPostgresFixture(t)
	albumID := fixture.createAlbum(t, "Viagem")
	if _, err := fixture.albums.AddAlbumItems(albumID, fixture.fileIDs("old.jpg", "mid.jpg")); err != nil {
		t.Fatalf("add: %v", err)
	}

	removed, err := fixture.albums.RemoveAlbumItems(albumID, fixture.fileIDs("old.jpg", "new.jpg"))
	if err != nil || removed != 1 {
		t.Fatalf("remove = %d err %v", removed, err)
	}
	assertNames(t, "album after remove", fixture.albumItemNames(t, albumID), []string{"mid.jpg"})
	assertNames(t, "gallery untouched", listNames(t, fixture.library, LibraryListQuery{Sort: LibrarySortTakenAt, Order: LibrarySortOrderDesc, Limit: 100}),
		[]string{"new.jpg", "mid.jpg", "old.jpg"})
}

func TestDeleteAlbumCascadesItemsButKeepsFiles_Postgres(t *testing.T) {
	fixture := newAlbumPostgresFixture(t)
	albumID := fixture.createAlbum(t, "Viagem")
	if _, err := fixture.albums.AddAlbumItems(albumID, fixture.fileIDs("old.jpg", "mid.jpg")); err != nil {
		t.Fatalf("add: %v", err)
	}

	if err := fixture.albums.DeleteAlbum(albumID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if err := fixture.albums.DeleteAlbum(albumID); !errors.Is(err, ErrAlbumNotFound) {
		t.Fatalf("second delete error = %v", err)
	}
	if _, err := fixture.albums.GetAlbum(albumID); !errors.Is(err, ErrAlbumNotFound) {
		t.Fatalf("get after delete error = %v", err)
	}

	var orphanItems int
	countErr := fixture.albums.Db.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(`SELECT count(*) FROM image_album_item`).Scan(&orphanItems)
	})
	if countErr != nil || orphanItems != 0 {
		t.Fatalf("cascade left %d items (err %v)", orphanItems, countErr)
	}
	assertNames(t, "gallery untouched", listNames(t, fixture.library, LibraryListQuery{Sort: LibrarySortTakenAt, Order: LibrarySortOrderDesc, Limit: 100}),
		[]string{"new.jpg", "mid.jpg", "old.jpg"})
}
