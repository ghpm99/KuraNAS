package files

import (
	"testing"

	"nas-go/api/internal/roots"
)

func TestUploadResponsesExposeLogicalPaths(t *testing.T) {
	t.Cleanup(roots.Reset)
	roots.Set([]roots.Root{{ID: 1, Path: "/srv/nas", Label: "Principal", Enabled: true}})

	fileResult := toClientUploadFileResult(UploadFileResult{Name: "a.jpg", Path: "/srv/nas/Fotos/a.jpg", Status: UploadStatusUploaded})
	fileResults := toClientUploadFileResults([]UploadFileResult{{Name: "b.jpg", Path: "/srv/nas/Fotos/b.jpg", Status: UploadStatusRenamed}})
	uploadedPaths := toClientPaths([]string{"/srv/nas/Fotos/a.jpg"})

	if fileResult.Path != "/Fotos/a.jpg" {
		t.Fatalf("complete path = %q, want /Fotos/a.jpg", fileResult.Path)
	}
	if fileResults[0].Path != "/Fotos/b.jpg" || fileResults[0].Status != UploadStatusRenamed {
		t.Fatalf("file results = %+v, want logical path with status kept", fileResults)
	}
	if uploadedPaths[0] != "/Fotos/a.jpg" {
		t.Fatalf("uploaded = %v, want logical paths", uploadedPaths)
	}
}
