package libraries

import (
	"database/sql"
	"errors"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/database"
	"os"
	"path/filepath"
	"testing"
)

type serviceRepositoryMock struct {
	db      *database.DbContext
	models  map[LibraryCategory]LibraryModel
	upserts int
}

func newServiceRepositoryMock() *serviceRepositoryMock {
	return &serviceRepositoryMock{
		db:     database.NewDbContext(nil),
		models: make(map[LibraryCategory]LibraryModel),
	}
}

func (m *serviceRepositoryMock) GetDbContext() *database.DbContext {
	return m.db
}

func (m *serviceRepositoryMock) GetAll() ([]LibraryModel, error) {
	result := make([]LibraryModel, 0, len(m.models))
	for _, category := range AllCategories {
		if model, ok := m.models[category]; ok {
			result = append(result, model)
		}
	}
	return result, nil
}

func (m *serviceRepositoryMock) GetByCategory(category LibraryCategory) (LibraryModel, error) {
	model, ok := m.models[category]
	if !ok {
		return LibraryModel{}, sql.ErrNoRows
	}
	return model, nil
}

func (m *serviceRepositoryMock) Upsert(tx *sql.Tx, model LibraryModel) (LibraryModel, error) {
	m.upserts++
	m.models[model.Category] = model
	return model, nil
}

func TestServiceGetLibraries(t *testing.T) {
	root := t.TempDir()
	for _, folder := range []string{"Imagens", "Musicas", "Videos", "Documentos"} {
		if err := os.MkdirAll(filepath.Join(root, folder), 0755); err != nil {
			t.Fatalf("failed to create %s: %v", folder, err)
		}
	}

	originalEntryPoint := config.AppConfig.EntryPoint
	t.Cleanup(func() {
		config.AppConfig.EntryPoint = originalEntryPoint
	})
	config.AppConfig.EntryPoint = root

	repository := newServiceRepositoryMock()
	service := NewService(repository)

	libraries, err := service.GetLibraries()
	if err != nil {
		t.Fatalf("GetLibraries returned error: %v", err)
	}
	if len(libraries) != 4 {
		t.Fatalf("expected 4 libraries, got %d", len(libraries))
	}
}

func TestServiceGetLibraryByCategoryInvalid(t *testing.T) {
	service := NewService(newServiceRepositoryMock())
	_, err := service.GetLibraryByCategory("invalid")
	if !errors.Is(err, ErrInvalidCategory) {
		t.Fatalf("expected invalid category error, got %v", err)
	}
}

func TestServiceGetLibraryByCategoryValid(t *testing.T) {
	repository := newServiceRepositoryMock()
	repository.models[LibraryCategoryImages] = LibraryModel{
		Category: LibraryCategoryImages,
		Path:     "/data/Imagens",
	}

	service := NewService(repository)
	library, err := service.GetLibraryByCategory(LibraryCategoryImages)
	if err != nil {
		t.Fatalf("GetLibraryByCategory returned error: %v", err)
	}
	if library.Path != "/data/Imagens" {
		t.Fatalf("expected /data/Imagens, got %s", library.Path)
	}
}

func TestServiceGetLibraryByCategoryResolvesMissing(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "Imagens"), 0755); err != nil {
		t.Fatalf("failed to create folder: %v", err)
	}

	originalEntryPoint := config.AppConfig.EntryPoint
	t.Cleanup(func() {
		config.AppConfig.EntryPoint = originalEntryPoint
	})
	config.AppConfig.EntryPoint = root

	repository := newServiceRepositoryMock()
	service := NewService(repository)

	library, err := service.GetLibraryByCategory(LibraryCategoryImages)
	if err != nil {
		t.Fatalf("GetLibraryByCategory returned error: %v", err)
	}
	if library.Path != filepath.Join(root, "Imagens") {
		t.Fatalf("expected resolved path, got %s", library.Path)
	}
}

func TestServiceUpdateLibraryValid(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "Imagens")
	if err := os.MkdirAll(target, 0755); err != nil {
		t.Fatalf("failed to create target folder: %v", err)
	}

	originalEntryPoint := config.AppConfig.EntryPoint
	t.Cleanup(func() {
		config.AppConfig.EntryPoint = originalEntryPoint
	})
	config.AppConfig.EntryPoint = root

	repository := newServiceRepositoryMock()
	service := NewService(repository)

	library, err := service.UpdateLibrary(LibraryCategoryImages, UpdateLibraryDto{Path: target})
	if err != nil {
		t.Fatalf("UpdateLibrary returned error: %v", err)
	}
	if library.Path != target {
		t.Fatalf("expected updated path %s, got %s", target, library.Path)
	}
}

func TestServiceUpdateLibraryNotSubfolder(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()

	originalEntryPoint := config.AppConfig.EntryPoint
	t.Cleanup(func() {
		config.AppConfig.EntryPoint = originalEntryPoint
	})
	config.AppConfig.EntryPoint = root

	service := NewService(newServiceRepositoryMock())
	_, err := service.UpdateLibrary(LibraryCategoryImages, UpdateLibraryDto{Path: outside})
	if !errors.Is(err, ErrPathNotSubfolder) {
		t.Fatalf("expected ErrPathNotSubfolder, got %v", err)
	}
}

func TestServiceUpdateLibraryPathNotExists(t *testing.T) {
	root := t.TempDir()
	missing := filepath.Join(root, "missing")

	originalEntryPoint := config.AppConfig.EntryPoint
	t.Cleanup(func() {
		config.AppConfig.EntryPoint = originalEntryPoint
	})
	config.AppConfig.EntryPoint = root

	service := NewService(newServiceRepositoryMock())
	_, err := service.UpdateLibrary(LibraryCategoryImages, UpdateLibraryDto{Path: missing})
	if !errors.Is(err, ErrPathNotExists) {
		t.Fatalf("expected ErrPathNotExists, got %v", err)
	}
}

func TestServiceResolveLibraries(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "Imagens"), 0755); err != nil {
		t.Fatalf("failed to create images folder: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "Videos"), 0755); err != nil {
		t.Fatalf("failed to create videos folder: %v", err)
	}

	originalEntryPoint := config.AppConfig.EntryPoint
	t.Cleanup(func() {
		config.AppConfig.EntryPoint = originalEntryPoint
	})
	config.AppConfig.EntryPoint = root

	repository := newServiceRepositoryMock()
	service := NewService(repository)

	if err := service.ResolveLibraries(); err != nil {
		t.Fatalf("ResolveLibraries returned error: %v", err)
	}
	if len(repository.models) != 4 {
		t.Fatalf("expected 4 resolved libraries, got %d", len(repository.models))
	}
}

func TestServiceResolveLibrariesWithoutEntryPoint(t *testing.T) {
	originalEntryPoint := config.AppConfig.EntryPoint
	t.Cleanup(func() {
		config.AppConfig.EntryPoint = originalEntryPoint
	})
	config.AppConfig.EntryPoint = ""

	service := NewService(newServiceRepositoryMock())
	if err := service.ResolveLibraries(); err == nil {
		t.Fatalf("expected error when entry point is missing")
	}
}

func TestValidatePathIsSubfolder(t *testing.T) {
	root := t.TempDir()
	valid := filepath.Join(root, "Imagens")
	if err := os.MkdirAll(valid, 0755); err != nil {
		t.Fatalf("failed to create folder: %v", err)
	}

	originalEntryPoint := config.AppConfig.EntryPoint
	t.Cleanup(func() {
		config.AppConfig.EntryPoint = originalEntryPoint
	})
	config.AppConfig.EntryPoint = root

	if err := validatePathIsSubfolder(valid); err != nil {
		t.Fatalf("expected valid subfolder, got %v", err)
	}

	if err := validatePathIsSubfolder(root); !errors.Is(err, ErrPathNotSubfolder) {
		t.Fatalf("expected ErrPathNotSubfolder for root path, got %v", err)
	}
}
