package storageroots

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"nas-go/api/internal/config"
	"nas-go/api/internal/roots"
	"nas-go/api/pkg/database"
)

// rootsRepoMock is an in-memory registry implementing RepositoryInterface.
type rootsRepoMock struct {
	mu     sync.Mutex
	nextID int
	models map[int]StorageRootModel

	getErr    error
	createErr error
}

func newRootsRepoMock() *rootsRepoMock {
	return &rootsRepoMock{nextID: 1, models: map[int]StorageRootModel{}}
}

func (m *rootsRepoMock) GetDbContext() *database.DbContext { return database.NewDbContext(nil) }

func (m *rootsRepoMock) sorted() []StorageRootModel {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]StorageRootModel, 0, len(m.models))
	for id := 1; id < m.nextID; id++ {
		if model, ok := m.models[id]; ok {
			out = append(out, model)
		}
	}
	return out
}

func (m *rootsRepoMock) GetAll() ([]StorageRootModel, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.sorted(), nil
}

func (m *rootsRepoMock) GetByID(id int) (StorageRootModel, bool, error) {
	if m.getErr != nil {
		return StorageRootModel{}, false, m.getErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	model, ok := m.models[id]
	return model, ok, nil
}

func (m *rootsRepoMock) Create(tx *sql.Tx, model StorageRootModel) (StorageRootModel, error) {
	if m.createErr != nil {
		return StorageRootModel{}, m.createErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	model.ID = m.nextID
	model.CreatedAt = time.Now()
	m.nextID++
	m.models[model.ID] = model
	return model, nil
}

func (m *rootsRepoMock) Update(tx *sql.Tx, model StorageRootModel) (StorageRootModel, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	current, ok := m.models[model.ID]
	if !ok {
		return StorageRootModel{}, ErrRootNotFound
	}
	current.Label = model.Label
	current.Enabled = model.Enabled
	m.models[model.ID] = current
	return current, nil
}

func (m *rootsRepoMock) Delete(tx *sql.Tx, id int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.models[id]; !ok {
		return ErrRootNotFound
	}
	delete(m.models, id)
	return nil
}

type indexTriggerMock struct {
	mu    sync.Mutex
	paths []string
}

func (m *indexTriggerMock) ScanDirTask(path string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.paths = append(m.paths, path)
}

func (m *indexTriggerMock) scanned() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]string(nil), m.paths...)
}

func newServiceForTest(t *testing.T) (*Service, *rootsRepoMock, *indexTriggerMock) {
	t.Helper()
	previousEntryPoint := config.AppConfig.EntryPoint
	t.Cleanup(func() {
		config.AppConfig.EntryPoint = previousEntryPoint
		roots.Reset()
	})
	roots.Reset()

	repo := newRootsRepoMock()
	trigger := &indexTriggerMock{}
	return NewService(repo, trigger), repo, trigger
}

func TestReloadRegistrySeedsEntryPointOnEmptyTable(t *testing.T) {
	service, repo, _ := newServiceForTest(t)
	entryPoint := t.TempDir()
	config.AppConfig.EntryPoint = entryPoint

	if err := service.ReloadRegistry(); err != nil {
		t.Fatalf("ReloadRegistry: %v", err)
	}

	registered, err := repo.GetAll()
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if len(registered) != 1 || registered[0].Path != filepath.Clean(entryPoint) {
		t.Fatalf("expected seeded entry point root, got %+v", registered)
	}
	primary, ok := roots.Primary()
	if !ok || primary.Path != filepath.Clean(entryPoint) {
		t.Fatalf("expected primary root %q in registry, got %+v ok=%v", entryPoint, primary, ok)
	}
}

func TestReloadRegistryWithoutEntryPointStaysEmpty(t *testing.T) {
	service, repo, _ := newServiceForTest(t)
	config.AppConfig.EntryPoint = ""

	if err := service.ReloadRegistry(); err != nil {
		t.Fatalf("ReloadRegistry: %v", err)
	}
	if registered, _ := repo.GetAll(); len(registered) != 0 {
		t.Fatalf("expected no seed without ENTRY_POINT, got %+v", registered)
	}
}

func TestReloadRegistryLoadsExistingRows(t *testing.T) {
	service, repo, _ := newServiceForTest(t)
	config.AppConfig.EntryPoint = t.TempDir()

	first := t.TempDir()
	second := t.TempDir()
	repo.Create(nil, StorageRootModel{Path: first, Label: "first", Enabled: true})
	repo.Create(nil, StorageRootModel{Path: second, Label: "second", Enabled: false})

	if err := service.ReloadRegistry(); err != nil {
		t.Fatalf("ReloadRegistry: %v", err)
	}

	enabled := roots.Enabled()
	if len(enabled) != 1 || enabled[0].Path != filepath.Clean(first) {
		t.Fatalf("expected only the enabled root in the registry, got %+v", enabled)
	}
}

func TestCreateRootValidAndIndexTriggered(t *testing.T) {
	service, _, trigger := newServiceForTest(t)
	config.AppConfig.EntryPoint = ""
	newRoot := t.TempDir()

	created, err := service.CreateRoot(CreateStorageRootDto{Path: newRoot, Label: "Midia"})
	if err != nil {
		t.Fatalf("CreateRoot: %v", err)
	}
	if created.Path != filepath.Clean(newRoot) || created.Label != "Midia" || !created.Enabled {
		t.Fatalf("unexpected created root: %+v", created)
	}

	scanned := trigger.scanned()
	if len(scanned) != 1 || scanned[0] != created.Path {
		t.Fatalf("expected index trigger for %q, got %v", created.Path, scanned)
	}
}

func TestCreateRootDisabledSkipsIndex(t *testing.T) {
	service, _, trigger := newServiceForTest(t)
	config.AppConfig.EntryPoint = ""
	disabled := false

	created, err := service.CreateRoot(CreateStorageRootDto{Path: t.TempDir(), Enabled: &disabled})
	if err != nil {
		t.Fatalf("CreateRoot: %v", err)
	}
	if created.Enabled {
		t.Fatalf("expected disabled root, got %+v", created)
	}
	if scanned := trigger.scanned(); len(scanned) != 0 {
		t.Fatalf("disabled root must not trigger indexing, got %v", scanned)
	}
}

func TestCreateRootLabelFallsBackToBaseName(t *testing.T) {
	service, _, _ := newServiceForTest(t)
	config.AppConfig.EntryPoint = ""
	newRoot := t.TempDir()

	created, err := service.CreateRoot(CreateStorageRootDto{Path: newRoot})
	if err != nil {
		t.Fatalf("CreateRoot: %v", err)
	}
	if created.Label != filepath.Base(filepath.Clean(newRoot)) {
		t.Fatalf("expected base-name label, got %q", created.Label)
	}
}

func TestCreateRootRejectsInvalidPaths(t *testing.T) {
	service, _, _ := newServiceForTest(t)
	config.AppConfig.EntryPoint = ""

	filePath := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(filePath, []byte("x"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	cases := []string{"", "relative/path", filepath.Join(t.TempDir(), "missing"), filePath}
	for _, candidate := range cases {
		if _, err := service.CreateRoot(CreateStorageRootDto{Path: candidate}); !errors.Is(err, ErrInvalidRootPath) {
			t.Fatalf("path %q: expected ErrInvalidRootPath, got %v", candidate, err)
		}
	}
}

func TestCreateRootRejectsDuplicateAndOverlap(t *testing.T) {
	service, _, _ := newServiceForTest(t)
	config.AppConfig.EntryPoint = ""

	base := t.TempDir()
	child := filepath.Join(base, "child")
	if err := os.MkdirAll(child, 0755); err != nil {
		t.Fatalf("mkdir child: %v", err)
	}

	if _, err := service.CreateRoot(CreateStorageRootDto{Path: base, Label: "base"}); err != nil {
		t.Fatalf("CreateRoot base: %v", err)
	}

	if _, err := service.CreateRoot(CreateStorageRootDto{Path: base, Label: "again"}); !errors.Is(err, ErrDuplicateRoot) {
		t.Fatalf("same path: expected ErrDuplicateRoot, got %v", err)
	}
	if _, err := service.CreateRoot(CreateStorageRootDto{Path: child, Label: "child"}); !errors.Is(err, ErrOverlappingRoot) {
		t.Fatalf("descendant: expected ErrOverlappingRoot, got %v", err)
	}
	if _, err := service.CreateRoot(CreateStorageRootDto{Path: t.TempDir(), Label: "base"}); !errors.Is(err, ErrDuplicateRoot) {
		t.Fatalf("same label: expected ErrDuplicateRoot, got %v", err)
	}
	if _, err := service.CreateRoot(CreateStorageRootDto{Path: t.TempDir(), Label: "bad/label"}); !errors.Is(err, ErrInvalidRootLabel) {
		t.Fatalf("path-like label: expected ErrInvalidRootLabel, got %v", err)
	}
}

func TestUpdateRootChangesLabelAndEnabled(t *testing.T) {
	service, _, _ := newServiceForTest(t)
	config.AppConfig.EntryPoint = ""

	if _, err := service.CreateRoot(CreateStorageRootDto{Path: t.TempDir(), Label: "primary"}); err != nil {
		t.Fatalf("CreateRoot primary: %v", err)
	}
	created, err := service.CreateRoot(CreateStorageRootDto{Path: t.TempDir(), Label: "secondary"})
	if err != nil {
		t.Fatalf("CreateRoot secondary: %v", err)
	}

	disabled := false
	updated, err := service.UpdateRoot(created.ID, UpdateStorageRootDto{Label: "renamed", Enabled: &disabled})
	if err != nil {
		t.Fatalf("UpdateRoot: %v", err)
	}
	if updated.Label != "renamed" || updated.Enabled {
		t.Fatalf("unexpected update result: %+v", updated)
	}

	enabled := roots.Enabled()
	if len(enabled) != 1 || enabled[0].Label != "primary" {
		t.Fatalf("registry should drop the disabled root, got %+v", enabled)
	}
}

func TestUpdateRootNotFound(t *testing.T) {
	service, _, _ := newServiceForTest(t)
	config.AppConfig.EntryPoint = ""

	if _, err := service.UpdateRoot(42, UpdateStorageRootDto{Label: "x"}); !errors.Is(err, ErrRootNotFound) {
		t.Fatalf("expected ErrRootNotFound, got %v", err)
	}
}

func TestPrimaryRootCannotBeDisabledOrDeleted(t *testing.T) {
	service, _, _ := newServiceForTest(t)
	config.AppConfig.EntryPoint = ""

	primary, err := service.CreateRoot(CreateStorageRootDto{Path: t.TempDir(), Label: "primary"})
	if err != nil {
		t.Fatalf("CreateRoot: %v", err)
	}

	disabled := false
	if _, err := service.UpdateRoot(primary.ID, UpdateStorageRootDto{Enabled: &disabled}); !errors.Is(err, ErrPrimaryRootImmutable) {
		t.Fatalf("disable primary: expected ErrPrimaryRootImmutable, got %v", err)
	}
	if err := service.DeleteRoot(primary.ID); !errors.Is(err, ErrPrimaryRootImmutable) {
		t.Fatalf("delete primary: expected ErrPrimaryRootImmutable, got %v", err)
	}
}

func TestDeleteSecondaryRootRemovesFromRegistry(t *testing.T) {
	service, repo, _ := newServiceForTest(t)
	config.AppConfig.EntryPoint = ""

	if _, err := service.CreateRoot(CreateStorageRootDto{Path: t.TempDir(), Label: "primary"}); err != nil {
		t.Fatalf("CreateRoot primary: %v", err)
	}
	secondary, err := service.CreateRoot(CreateStorageRootDto{Path: t.TempDir(), Label: "secondary"})
	if err != nil {
		t.Fatalf("CreateRoot secondary: %v", err)
	}

	if err := service.DeleteRoot(secondary.ID); err != nil {
		t.Fatalf("DeleteRoot: %v", err)
	}
	if registered, _ := repo.GetAll(); len(registered) != 1 {
		t.Fatalf("expected one remaining root, got %+v", registered)
	}
	if enabled := roots.Enabled(); len(enabled) != 1 || enabled[0].Label != "primary" {
		t.Fatalf("registry should hold only the primary, got %+v", enabled)
	}
}

func TestGetRootsReturnsDtos(t *testing.T) {
	service, _, _ := newServiceForTest(t)
	config.AppConfig.EntryPoint = ""

	created, err := service.CreateRoot(CreateStorageRootDto{Path: t.TempDir(), Label: "primary"})
	if err != nil {
		t.Fatalf("CreateRoot: %v", err)
	}

	dtos, err := service.GetRoots()
	if err != nil {
		t.Fatalf("GetRoots: %v", err)
	}
	if len(dtos) != 1 || dtos[0].ID != created.ID || dtos[0].Label != "primary" {
		t.Fatalf("unexpected dtos: %+v", dtos)
	}
}

func TestServiceSurfacesRepositoryErrors(t *testing.T) {
	service, repo, _ := newServiceForTest(t)
	config.AppConfig.EntryPoint = ""
	repo.getErr = errors.New("boom")

	if _, err := service.GetRoots(); err == nil {
		t.Fatalf("GetRoots: expected error")
	}
	if _, err := service.CreateRoot(CreateStorageRootDto{Path: t.TempDir()}); err == nil {
		t.Fatalf("CreateRoot: expected error")
	}
	if _, err := service.UpdateRoot(1, UpdateStorageRootDto{}); err == nil {
		t.Fatalf("UpdateRoot: expected error")
	}
	if err := service.DeleteRoot(1); err == nil {
		t.Fatalf("DeleteRoot: expected error")
	}
	if err := service.ReloadRegistry(); err == nil {
		t.Fatalf("ReloadRegistry: expected error")
	}
}

func TestReloadRegistrySeedFailureIsReturned(t *testing.T) {
	service, repo, _ := newServiceForTest(t)
	config.AppConfig.EntryPoint = t.TempDir()
	repo.createErr = errors.New("insert failed")

	if err := service.ReloadRegistry(); err == nil {
		t.Fatalf("expected seed failure to surface")
	}
}
