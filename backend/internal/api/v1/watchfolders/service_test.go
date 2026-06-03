package watchfolders

import (
	"database/sql"
	"errors"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/database"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type serviceRepositoryMock struct {
	db     *database.DbContext
	models map[int]WatchFolderModel
	nextID int
}

func newServiceRepositoryMock() *serviceRepositoryMock {
	return &serviceRepositoryMock{
		db:     database.NewDbContext(nil),
		models: make(map[int]WatchFolderModel),
		nextID: 1,
	}
}

func (m *serviceRepositoryMock) GetDbContext() *database.DbContext { return m.db }

func (m *serviceRepositoryMock) GetAll() ([]WatchFolderModel, error) {
	result := make([]WatchFolderModel, 0, len(m.models))
	for _, model := range m.models {
		result = append(result, model)
	}
	return result, nil
}

func (m *serviceRepositoryMock) GetByID(id int) (WatchFolderModel, error) {
	model, ok := m.models[id]
	if !ok {
		return WatchFolderModel{}, sql.ErrNoRows
	}
	return model, nil
}

func (m *serviceRepositoryMock) Create(tx *sql.Tx, model WatchFolderModel) (WatchFolderModel, error) {
	model.ID = m.nextID
	m.nextID++
	now := time.Now()
	model.CreatedAt = now
	model.UpdatedAt = now
	m.models[model.ID] = model
	return model, nil
}

func (m *serviceRepositoryMock) Update(tx *sql.Tx, model WatchFolderModel) (WatchFolderModel, error) {
	if _, ok := m.models[model.ID]; !ok {
		return WatchFolderModel{}, sql.ErrNoRows
	}
	model.UpdatedAt = time.Now()
	m.models[model.ID] = model
	return model, nil
}

func (m *serviceRepositoryMock) Delete(tx *sql.Tx, id int) error {
	if _, ok := m.models[id]; !ok {
		return sql.ErrNoRows
	}
	delete(m.models, id)
	return nil
}

func (m *serviceRepositoryMock) UpdateLastScan(tx *sql.Tx, id int, lastScanAt time.Time) error {
	model, ok := m.models[id]
	if !ok {
		return sql.ErrNoRows
	}
	model.LastScanAt = &lastScanAt
	model.UpdatedAt = time.Now()
	m.models[id] = model
	return nil
}

func TestCreateWatchFolderSuccess(t *testing.T) {
	entry := t.TempDir()
	watch := t.TempDir()

	originalEntry := config.AppConfig.EntryPoint
	t.Cleanup(func() { config.AppConfig.EntryPoint = originalEntry })
	config.AppConfig.EntryPoint = entry

	service := NewService(newServiceRepositoryMock())
	dto, err := service.CreateWatchFolder(CreateWatchFolderDto{Path: watch, Label: "OneDrive"})
	if err != nil {
		t.Fatalf("CreateWatchFolder returned error: %v", err)
	}
	if dto.ID <= 0 {
		t.Fatalf("expected positive id, got %d", dto.ID)
	}
	if dto.Path != watch {
		t.Fatalf("expected path %s, got %s", watch, dto.Path)
	}
}

func TestCreateWatchFolderPathNotExists(t *testing.T) {
	entry := t.TempDir()
	missing := filepath.Join(t.TempDir(), "missing")

	originalEntry := config.AppConfig.EntryPoint
	t.Cleanup(func() { config.AppConfig.EntryPoint = originalEntry })
	config.AppConfig.EntryPoint = entry

	service := NewService(newServiceRepositoryMock())
	_, err := service.CreateWatchFolder(CreateWatchFolderDto{Path: missing})
	if !errors.Is(err, ErrPathNotExists) {
		t.Fatalf("expected ErrPathNotExists, got %v", err)
	}
}

func TestCreateWatchFolderPathIsEntryPointSubfolder(t *testing.T) {
	entry := t.TempDir()
	watch := filepath.Join(entry, "OneDrive")
	if err := os.MkdirAll(watch, 0755); err != nil {
		t.Fatalf("create watch dir: %v", err)
	}

	originalEntry := config.AppConfig.EntryPoint
	t.Cleanup(func() { config.AppConfig.EntryPoint = originalEntry })
	config.AppConfig.EntryPoint = entry

	service := NewService(newServiceRepositoryMock())
	_, err := service.CreateWatchFolder(CreateWatchFolderDto{Path: watch})
	if !errors.Is(err, ErrPathIsSubfolderOfEntryPoint) {
		t.Fatalf("expected ErrPathIsSubfolderOfEntryPoint, got %v", err)
	}
}

func TestCreateWatchFolderDuplicate(t *testing.T) {
	entry := t.TempDir()
	watch := t.TempDir()

	originalEntry := config.AppConfig.EntryPoint
	t.Cleanup(func() { config.AppConfig.EntryPoint = originalEntry })
	config.AppConfig.EntryPoint = entry

	repo := newServiceRepositoryMock()
	repo.models[1] = WatchFolderModel{ID: 1, Path: watch, Enabled: true}
	repo.nextID = 2

	service := NewService(repo)
	_, err := service.CreateWatchFolder(CreateWatchFolderDto{Path: watch})
	if !errors.Is(err, ErrPathAlreadyWatched) {
		t.Fatalf("expected ErrPathAlreadyWatched, got %v", err)
	}
}

func TestUpdateWatchFolderSuccess(t *testing.T) {
	entry := t.TempDir()
	watchA := t.TempDir()
	watchB := t.TempDir()

	originalEntry := config.AppConfig.EntryPoint
	t.Cleanup(func() { config.AppConfig.EntryPoint = originalEntry })
	config.AppConfig.EntryPoint = entry

	repo := newServiceRepositoryMock()
	repo.models[1] = WatchFolderModel{ID: 1, Path: watchA, Label: "A", Enabled: true}
	repo.nextID = 2

	service := NewService(repo)
	newLabel := "Updated"
	enabled := false
	updated, err := service.UpdateWatchFolder(1, UpdateWatchFolderDto{Path: &watchB, Label: &newLabel, Enabled: &enabled})
	if err != nil {
		t.Fatalf("UpdateWatchFolder returned error: %v", err)
	}
	if updated.Path != watchB || updated.Label != newLabel || updated.Enabled != enabled {
		t.Fatalf("unexpected update result: %+v", updated)
	}
}

func TestDeleteWatchFolderSuccess(t *testing.T) {
	repo := newServiceRepositoryMock()
	repo.models[1] = WatchFolderModel{ID: 1, Path: t.TempDir(), Enabled: true}
	service := NewService(repo)

	if err := service.DeleteWatchFolder(1); err != nil {
		t.Fatalf("DeleteWatchFolder returned error: %v", err)
	}
	if len(repo.models) != 0 {
		t.Fatalf("expected empty repo after delete")
	}
}

func TestGetEnabledWatchFolders(t *testing.T) {
	repo := newServiceRepositoryMock()
	repo.models[1] = WatchFolderModel{ID: 1, Path: t.TempDir(), Enabled: true}
	repo.models[2] = WatchFolderModel{ID: 2, Path: t.TempDir(), Enabled: false}

	service := NewService(repo)
	enabled, err := service.GetEnabledWatchFolders()
	if err != nil {
		t.Fatalf("GetEnabledWatchFolders returned error: %v", err)
	}
	if len(enabled) != 1 || enabled[0].ID != 1 {
		t.Fatalf("expected only id=1 enabled, got %+v", enabled)
	}
}
