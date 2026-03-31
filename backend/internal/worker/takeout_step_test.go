package worker

import (
	"archive/zip"
	"encoding/json"
	"nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/api/v1/libraries"
	"os"
	"path/filepath"
	"testing"
)

type librariesServiceMock struct {
	paths map[libraries.LibraryCategory]string
}

func (m *librariesServiceMock) GetLibraries() ([]libraries.LibraryDto, error) {
	return nil, nil
}

func (m *librariesServiceMock) GetLibraryByCategory(category libraries.LibraryCategory) (libraries.LibraryDto, error) {
	return libraries.LibraryDto{Category: string(category), Path: m.paths[category]}, nil
}

func (m *librariesServiceMock) UpdateLibrary(category libraries.LibraryCategory, dto libraries.UpdateLibraryDto) (libraries.LibraryDto, error) {
	return libraries.LibraryDto{}, nil
}

func (m *librariesServiceMock) ResolveLibraries() error {
	return nil
}

func createWorkerZip(t *testing.T, zipPath string) {
	t.Helper()
	file, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("failed to create zip: %v", err)
	}
	defer file.Close()

	writer := zip.NewWriter(file)
	entry, err := writer.Create("Google Fotos/IMG_01.jpg")
	if err != nil {
		t.Fatalf("failed to create zip entry: %v", err)
	}
	_, _ = entry.Write([]byte("img"))
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close zip writer: %v", err)
	}
}

func TestExecuteTakeoutExtractStepSuccess(t *testing.T) {
	tempDir := t.TempDir()
	zipPath := filepath.Join(tempDir, "upload", "takeout.zip")
	if err := os.MkdirAll(filepath.Dir(zipPath), 0755); err != nil {
		t.Fatalf("failed to create upload dir: %v", err)
	}
	createWorkerZip(t, zipPath)

	imageLib := filepath.Join(tempDir, "Imagens")
	if err := os.MkdirAll(imageLib, 0755); err != nil {
		t.Fatalf("failed to create image lib: %v", err)
	}

	payload, err := json.Marshal(TakeoutStepPayload{
		ZipPath:  zipPath,
		UploadID: "u1",
	})
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	err = executeTakeoutExtractStep(&WorkerContext{
		LibrariesService: &librariesServiceMock{
			paths: map[libraries.LibraryCategory]string{
				libraries.LibraryCategoryImages: imageLib,
				libraries.LibraryCategoryVideos: filepath.Join(tempDir, "Videos"),
			},
		},
	}, jobs.StepModel{Payload: payload})
	if err != nil {
		t.Fatalf("executeTakeoutExtractStep returned error: %v", err)
	}
}

func TestExecuteTakeoutExtractStepInvalidPayload(t *testing.T) {
	err := executeTakeoutExtractStep(&WorkerContext{
		LibrariesService: &librariesServiceMock{
			paths: map[libraries.LibraryCategory]string{},
		},
	}, jobs.StepModel{Payload: []byte(`{`)})
	if err == nil {
		t.Fatalf("expected payload decode error")
	}
}
