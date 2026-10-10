package video

import (
	"database/sql"
	"testing"

	"nas-go/api/internal/testutil"
)

type seededVideoFile struct {
	name    string
	folder  string
	deleted bool
}

func seedVideoFiles(t *testing.T, repository *Repository, seeds []seededVideoFile) map[string]int {
	t.Helper()
	idsByName := map[string]int{}
	seedErr := repository.DbContext.ExecTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(`TRUNCATE home_file RESTART IDENTITY CASCADE`); err != nil {
			return err
		}
		for _, seed := range seeds {
			var deletedAt any
			if seed.deleted {
				deletedAt = "2024-01-01"
			}
			var fileID int
			err := tx.QueryRow(
				`INSERT INTO home_file (name, path, parent_path, format, size, updated_at, created_at, type, checksum, deleted_at)
				 VALUES ($1, $2, $3, $4, 1, now(), now(), 2, '', $5) RETURNING id`,
				seed.name, seed.folder+"/"+seed.name, seed.folder, fileExtension(seed.name), deletedAt,
			).Scan(&fileID)
			if err != nil {
				return err
			}
			idsByName[seed.name] = fileID
		}
		return nil
	})
	if seedErr != nil {
		t.Fatalf("seed: %v", seedErr)
	}
	return idsByName
}

func fileExtension(name string) string {
	for position := len(name) - 1; position >= 0; position-- {
		if name[position] == '.' {
			return name[position:]
		}
	}
	return ""
}

func TestLibraryFoldersCountRecursivelyAndPickFirstCover_Postgres(t *testing.T) {
	repository := NewRepository(testutil.NewPostgresDB(t, "kuranas_video_it"))
	ids := seedVideoFiles(t, repository, []seededVideoFile{
		{name: "direct.mp4", folder: "/data"},
		{name: "b.mkv", folder: "/data/Series"},
		{name: "a.mkv", folder: "/data/Series/S1"},
		{name: "gone.mkv", folder: "/data/Series", deleted: true},
		{name: "notes.txt", folder: "/data/Series"},
		{name: "solo.mp4", folder: "/data/Album"},
		{name: "extra.mp4", folder: "/mnt/midia/deep"},
		{name: "like.mp4", folder: "/data_other/x"},
	})

	rootScopes := []LibraryFolderScope{{Prefix: "/data/"}, {Prefix: "/mnt/midia/", IsWholeRoot: true, Label: "Midia"}}
	folders, err := repository.ListLibraryFolders(LibraryFolderQuery{Scopes: rootScopes, Separator: "/", Limit: 10})
	if err != nil {
		t.Fatalf("roots: %v", err)
	}
	if len(folders) != 3 {
		t.Fatalf("want 3 folders, got %+v", folders)
	}
	if folders[0].Name != "Album" || folders[0].VideoCount != 1 || folders[0].CoverFileID != ids["solo.mp4"] {
		t.Fatalf("unexpected Album %+v", folders[0])
	}
	if folders[1].Name != "Midia" || folders[1].Path != "/mnt/midia" || folders[1].VideoCount != 1 {
		t.Fatalf("unexpected Midia %+v", folders[1])
	}
	if folders[2].Name != "Series" || folders[2].Path != "/data/Series" || folders[2].VideoCount != 2 || folders[2].CoverFileID != ids["a.mkv"] {
		t.Fatalf("unexpected Series %+v", folders[2])
	}

	children, err := repository.ListLibraryFolders(LibraryFolderQuery{Scopes: []LibraryFolderScope{{Prefix: "/data/Series/"}}, Separator: "/", Limit: 10})
	if err != nil {
		t.Fatalf("children: %v", err)
	}
	if len(children) != 1 || children[0].Name != "S1" || children[0].Path != "/data/Series/S1" || children[0].VideoCount != 1 {
		t.Fatalf("unexpected children %+v", children)
	}

	secondPage, err := repository.ListLibraryFolders(LibraryFolderQuery{Scopes: rootScopes, Separator: "/", Limit: 2, Offset: 2})
	if err != nil || len(secondPage) != 1 || secondPage[0].Name != "Series" {
		t.Fatalf("second page %+v err %v", secondPage, err)
	}
}

func TestLibraryFolderVideosListDirectVideosInNaturalNameOrder_Postgres(t *testing.T) {
	repository := NewRepository(testutil.NewPostgresDB(t, "kuranas_video_it"))
	seedVideoFiles(t, repository, []seededVideoFile{
		{name: "Ep 10.mkv", folder: "/data/Series"},
		{name: "Ep 2.mkv", folder: "/data/Series"},
		{name: "Ep 1.mkv", folder: "/data/Series"},
		{name: "gone.mkv", folder: "/data/Series", deleted: true},
		{name: "notes.txt", folder: "/data/Series"},
		{name: "nested.mkv", folder: "/data/Series/S1"},
	})

	videos, err := repository.ListLibraryFolderVideos("/data/Series", 10, 0)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	names := []string{}
	for _, video := range videos {
		names = append(names, video.Name)
	}
	expected := []string{"Ep 1.mkv", "Ep 2.mkv", "Ep 10.mkv"}
	if len(names) != len(expected) {
		t.Fatalf("want %v, got %v", expected, names)
	}
	for position := range expected {
		if names[position] != expected[position] {
			t.Fatalf("want %v, got %v", expected, names)
		}
	}

	secondPage, err := repository.ListLibraryFolderVideos("/data/Series", 2, 2)
	if err != nil || len(secondPage) != 1 || secondPage[0].Name != "Ep 10.mkv" {
		t.Fatalf("second page %+v err %v", secondPage, err)
	}
}
