package analytics

import (
	"database/sql"
	"testing"
	"time"

	"nas-go/api/internal/roots"
)

func registerPrimaryRoot(t *testing.T, rootPath string) {
	t.Helper()
	t.Cleanup(roots.Reset)
	roots.Set([]roots.Root{{ID: 1, Path: rootPath, Label: "Principal", Enabled: true}})
}

func TestDuplicateGroupDetailsExposeRelativePaths(t *testing.T) {
	registerPrimaryRoot(t, "/srv/nas")
	groups := []DuplicateGroupDetailModel{{
		DuplicateGroupModel: DuplicateGroupModel{
			Signature: "abc",
			Copies:    2,
			Paths:     []string{"/srv/nas/Fotos/a.jpg", "/srv/nas/Fotos/b.jpg"},
		},
		Files: []DuplicateCopyModel{{ID: 7, Path: "/srv/nas/Fotos/a.jpg", UpdatedAt: time.Unix(0, 0), Tier: "hot"}},
	}}

	details := toDuplicateGroupDetailDtos(groups)

	if details[0].Paths[0] != "/Fotos/a.jpg" || details[0].Paths[1] != "/Fotos/b.jpg" {
		t.Fatalf("paths = %v, want relative to the root", details[0].Paths)
	}
	if details[0].Files[0].Path != "/Fotos/a.jpg" {
		t.Fatalf("copy path = %q, want /Fotos/a.jpg", details[0].Files[0].Path)
	}
}

func TestFolderAndRecentFileDtosExposeRelativePaths(t *testing.T) {
	registerPrimaryRoot(t, "/srv/nas")

	hotFolders := toHotFolderDto([]FolderHotModel{{ParentPath: "/srv/nas/Fotos", LastEvent: sql.NullTime{}}})
	folderUsage := toFolderUsageDto([]FolderUsageModel{{ParentPath: "/srv/nas/Videos"}})
	recentFiles := toRecentFilesDto([]RecentFileModel{{Path: "/srv/nas/Fotos/a.jpg", ParentPath: "/srv/nas/Fotos"}})

	if hotFolders[0].Path != "/Fotos" {
		t.Fatalf("hot folder path = %q, want /Fotos", hotFolders[0].Path)
	}
	if folderUsage[0].Path != "/Videos" {
		t.Fatalf("folder usage path = %q, want /Videos", folderUsage[0].Path)
	}
	if recentFiles[0].Path != "/Fotos/a.jpg" || recentFiles[0].ParentPath != "/Fotos" {
		t.Fatalf("recent file = %+v, want relative paths", recentFiles[0])
	}
}
