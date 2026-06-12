package trash

import (
	"database/sql"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"nas-go/api/internal/config"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/utils"
)

// trashRepoMock is an in-memory registry; the service moves real files inside
// a t.TempDir() entry point around it.
type trashRepoMock struct {
	mu        sync.Mutex
	nextID    int
	items     map[int]TrashItemModel
	retention int

	createErr error
	getErr    error
}

func newTrashRepoMock() *trashRepoMock {
	return &trashRepoMock{nextID: 1, items: map[int]TrashItemModel{}}
}

func (m *trashRepoMock) GetDbContext() *database.DbContext { return database.NewDbContext(nil) }

func (m *trashRepoMock) CreateItem(tx *sql.Tx, item TrashItemModel) (TrashItemModel, error) {
	if m.createErr != nil {
		return TrashItemModel{}, m.createErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	item.ID = m.nextID
	m.nextID++
	m.items[item.ID] = item
	return item, nil
}

func (m *trashRepoMock) sortedItems() []TrashItemModel {
	m.mu.Lock()
	defer m.mu.Unlock()
	items := make([]TrashItemModel, 0, len(m.items))
	for id := 1; id < m.nextID; id++ {
		if item, ok := m.items[id]; ok {
			items = append(items, item)
		}
	}
	return items
}

func (m *trashRepoMock) GetItems(page int, pageSize int) (utils.PaginationResponse[TrashItemModel], error) {
	if m.getErr != nil {
		return utils.PaginationResponse[TrashItemModel]{}, m.getErr
	}
	response := utils.PaginationResponse[TrashItemModel]{
		Items:      m.sortedItems(),
		Pagination: utils.Pagination{Page: page, PageSize: pageSize},
	}
	response.UpdatePagination()
	return response, nil
}

func (m *trashRepoMock) GetItemByID(id int) (TrashItemModel, bool, error) {
	if m.getErr != nil {
		return TrashItemModel{}, false, m.getErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	item, ok := m.items[id]
	return item, ok, nil
}

func (m *trashRepoMock) GetExpiredItems(cutoff time.Time) ([]TrashItemModel, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	expired := []TrashItemModel{}
	for _, item := range m.sortedItems() {
		if item.DeletedAt.Before(cutoff) {
			expired = append(expired, item)
		}
	}
	return expired, nil
}

func (m *trashRepoMock) GetAllItems() ([]TrashItemModel, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.sortedItems(), nil
}

func (m *trashRepoMock) DeleteItem(tx *sql.Tx, id int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.items[id]; !ok {
		return sql.ErrNoRows
	}
	delete(m.items, id)
	return nil
}

func (m *trashRepoMock) GetRetentionDays() (int, bool, error) {
	if m.retention <= 0 {
		return 0, false, nil
	}
	return m.retention, true, nil
}

func (m *trashRepoMock) SetRetentionDays(days int) error {
	m.retention = days
	return nil
}

type filesIndexMock struct {
	restoredPaths []string
	scannedDirs   []string
	restoreErr    error
}

func (m *filesIndexMock) RestoreSubtree(path string) error {
	m.restoredPaths = append(m.restoredPaths, path)
	return m.restoreErr
}

func (m *filesIndexMock) ScanDirTask(path string) {
	m.scannedDirs = append(m.scannedDirs, path)
}

func newTrashServiceForTest(t *testing.T) (*Service, *trashRepoMock, *filesIndexMock, string) {
	t.Helper()
	root := t.TempDir()
	originalEntryPoint := config.AppConfig.EntryPoint
	t.Cleanup(func() { config.AppConfig.EntryPoint = originalEntryPoint })
	config.AppConfig.EntryPoint = root

	repo := newTrashRepoMock()
	filesIndex := &filesIndexMock{}
	return NewService(repo, filesIndex), repo, filesIndex, root
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestTrashService_MoveToTrashKeepsBytesAndRegisters(t *testing.T) {
	s, repo, _, root := newTrashServiceForTest(t)
	filePath := filepath.Join(root, "docs", "a.txt")
	writeFile(t, filePath, "conteudo")

	if err := s.MoveToTrash(filePath, 8); err != nil {
		t.Fatalf("MoveToTrash: %v", err)
	}

	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Fatalf("original path must be gone, stat err=%v", err)
	}

	items, err := repo.GetAllItems()
	if err != nil || len(items) != 1 {
		t.Fatalf("expected one registered item, got %v err=%v", items, err)
	}
	item := items[0]
	if item.OriginalPath != filePath || item.Size != 8 {
		t.Fatalf("unexpected registry row: %+v", item)
	}

	data, err := os.ReadFile(item.TrashPath)
	if err != nil || string(data) != "conteudo" {
		t.Fatalf("bytes must survive in the trash, got %q err=%v", data, err)
	}
	if !IsInsideTrash(root, item.TrashPath) {
		t.Fatalf("trash path %q must live inside %s/%s", item.TrashPath, root, DirName)
	}
}

func TestTrashService_MoveToTrashSameNameTwiceDoesNotCollide(t *testing.T) {
	s, repo, _, root := newTrashServiceForTest(t)

	first := filepath.Join(root, "a.txt")
	writeFile(t, first, "primeiro")
	if err := s.MoveToTrash(first, 1); err != nil {
		t.Fatalf("first MoveToTrash: %v", err)
	}

	second := filepath.Join(root, "a.txt")
	writeFile(t, second, "segundo")
	if err := s.MoveToTrash(second, 1); err != nil {
		t.Fatalf("second MoveToTrash: %v", err)
	}

	items, _ := repo.GetAllItems()
	if len(items) != 2 {
		t.Fatalf("expected two items, got %d", len(items))
	}
	if items[0].TrashPath == items[1].TrashPath {
		t.Fatalf("trash paths must be unique, both are %q", items[0].TrashPath)
	}
	for _, item := range items {
		if _, err := os.Stat(item.TrashPath); err != nil {
			t.Fatalf("trashed copy %q missing: %v", item.TrashPath, err)
		}
	}
}

func TestTrashService_MoveToTrashUndoesMoveOnRegistryFailure(t *testing.T) {
	s, repo, _, root := newTrashServiceForTest(t)
	repo.createErr = sql.ErrConnDone

	filePath := filepath.Join(root, "a.txt")
	writeFile(t, filePath, "conteudo")

	if err := s.MoveToTrash(filePath, 8); err == nil {
		t.Fatalf("expected registry failure to surface")
	}
	if _, err := os.Stat(filePath); err != nil {
		t.Fatalf("file must be back at the original path after rollback: %v", err)
	}
}

func TestTrashService_RestoreItemPutsFileBackAndRevivesIndex(t *testing.T) {
	s, repo, filesIndex, root := newTrashServiceForTest(t)
	filePath := filepath.Join(root, "docs", "a.txt")
	writeFile(t, filePath, "conteudo")
	if err := s.MoveToTrash(filePath, 8); err != nil {
		t.Fatalf("MoveToTrash: %v", err)
	}

	restoredPath, err := s.RestoreItem(1)
	if err != nil {
		t.Fatalf("RestoreItem: %v", err)
	}
	if restoredPath != filePath {
		t.Fatalf("expected restore to %q, got %q", filePath, restoredPath)
	}

	data, err := os.ReadFile(filePath)
	if err != nil || string(data) != "conteudo" {
		t.Fatalf("restored bytes wrong: %q err=%v", data, err)
	}

	if items, _ := repo.GetAllItems(); len(items) != 0 {
		t.Fatalf("registry must be empty after restore, got %v", items)
	}
	if len(filesIndex.restoredPaths) != 1 || filesIndex.restoredPaths[0] != filePath {
		t.Fatalf("expected index restore for %q, got %v", filePath, filesIndex.restoredPaths)
	}
	if len(filesIndex.scannedDirs) != 1 || filesIndex.scannedDirs[0] != filepath.Dir(filePath) {
		t.Fatalf("expected rescan of parent dir, got %v", filesIndex.scannedDirs)
	}
}

func TestTrashService_RestoreItemConflictWhenPathOccupied(t *testing.T) {
	s, repo, _, root := newTrashServiceForTest(t)
	filePath := filepath.Join(root, "a.txt")
	writeFile(t, filePath, "original")
	if err := s.MoveToTrash(filePath, 8); err != nil {
		t.Fatalf("MoveToTrash: %v", err)
	}

	// Same path is occupied again by a new file.
	writeFile(t, filePath, "novo conteudo")

	if _, err := s.RestoreItem(1); err != ErrRestoreConflict {
		t.Fatalf("expected ErrRestoreConflict, got %v", err)
	}

	if items, _ := repo.GetAllItems(); len(items) != 1 {
		t.Fatalf("conflicting restore must keep the trash item, got %v", items)
	}
	data, _ := os.ReadFile(filePath)
	if string(data) != "novo conteudo" {
		t.Fatalf("occupying file must be untouched, got %q", data)
	}
}

func TestTrashService_RestoreItemNotFound(t *testing.T) {
	s, _, _, _ := newTrashServiceForTest(t)
	if _, err := s.RestoreItem(99); err != ErrItemNotFound {
		t.Fatalf("expected ErrItemNotFound, got %v", err)
	}
}

func TestTrashService_DeleteItemPermanently(t *testing.T) {
	s, repo, _, root := newTrashServiceForTest(t)
	filePath := filepath.Join(root, "a.txt")
	writeFile(t, filePath, "conteudo")
	if err := s.MoveToTrash(filePath, 8); err != nil {
		t.Fatalf("MoveToTrash: %v", err)
	}
	items, _ := repo.GetAllItems()

	if err := s.DeleteItemPermanently(items[0].ID); err != nil {
		t.Fatalf("DeleteItemPermanently: %v", err)
	}

	if _, err := os.Stat(items[0].TrashPath); !os.IsNotExist(err) {
		t.Fatalf("trashed bytes must be gone, stat err=%v", err)
	}
	if remaining, _ := repo.GetAllItems(); len(remaining) != 0 {
		t.Fatalf("registry must be empty, got %v", remaining)
	}

	if err := s.DeleteItemPermanently(99); err != ErrItemNotFound {
		t.Fatalf("expected ErrItemNotFound for unknown id, got %v", err)
	}
}

func TestTrashService_DeleteRefusesPathOutsideTrashDir(t *testing.T) {
	s, repo, _, root := newTrashServiceForTest(t)

	victim := filepath.Join(root, "vital.txt")
	writeFile(t, victim, "must survive")
	repo.items[1] = TrashItemModel{ID: 1, OriginalPath: victim, TrashPath: victim}
	repo.nextID = 2

	if err := s.DeleteItemPermanently(1); err == nil {
		t.Fatalf("expected refusal for a trash_path outside the trash dir")
	}
	if _, err := os.Stat(victim); err != nil {
		t.Fatalf("file outside the trash dir must never be removed: %v", err)
	}
}

func TestTrashService_EmptyTrash(t *testing.T) {
	s, repo, _, root := newTrashServiceForTest(t)
	for _, name := range []string{"a.txt", "b.txt", "c.txt"} {
		path := filepath.Join(root, name)
		writeFile(t, path, name)
		if err := s.MoveToTrash(path, 1); err != nil {
			t.Fatalf("MoveToTrash %s: %v", name, err)
		}
	}

	purged, err := s.EmptyTrash()
	if err != nil || purged != 3 {
		t.Fatalf("expected 3 purged, got %d err=%v", purged, err)
	}
	if items, _ := repo.GetAllItems(); len(items) != 0 {
		t.Fatalf("registry must be empty, got %v", items)
	}
}

func TestTrashService_PurgeExpiredHonorsRetention(t *testing.T) {
	s, repo, _, root := newTrashServiceForTest(t)

	oldPath := filepath.Join(root, "velho.txt")
	writeFile(t, oldPath, "velho")
	if err := s.MoveToTrash(oldPath, 1); err != nil {
		t.Fatalf("MoveToTrash old: %v", err)
	}
	newPath := filepath.Join(root, "novo.txt")
	writeFile(t, newPath, "novo")
	if err := s.MoveToTrash(newPath, 1); err != nil {
		t.Fatalf("MoveToTrash new: %v", err)
	}

	// Age the first item past the default retention window.
	aged := repo.items[1]
	aged.DeletedAt = time.Now().AddDate(0, 0, -(DefaultRetentionDays + 1))
	repo.items[1] = aged

	purged, err := s.PurgeExpired()
	if err != nil || purged != 1 {
		t.Fatalf("expected 1 purged, got %d err=%v", purged, err)
	}

	items, _ := repo.GetAllItems()
	if len(items) != 1 || items[0].OriginalPath != newPath {
		t.Fatalf("only the aged item should be purged, got %v", items)
	}
}

func TestTrashService_RetentionDaysDefaultAndValidation(t *testing.T) {
	s, _, _, _ := newTrashServiceForTest(t)

	days, err := s.GetRetentionDays()
	if err != nil || days != DefaultRetentionDays {
		t.Fatalf("expected default retention %d, got %d err=%v", DefaultRetentionDays, days, err)
	}

	if err := s.SetRetentionDays(0); err != ErrInvalidRetention {
		t.Fatalf("expected ErrInvalidRetention, got %v", err)
	}
	if err := s.SetRetentionDays(7); err != nil {
		t.Fatalf("SetRetentionDays: %v", err)
	}
	days, err = s.GetRetentionDays()
	if err != nil || days != 7 {
		t.Fatalf("expected retention 7, got %d err=%v", days, err)
	}
}

func TestTrashService_GetItemsConvertsToDto(t *testing.T) {
	s, _, _, root := newTrashServiceForTest(t)
	filePath := filepath.Join(root, "a.txt")
	writeFile(t, filePath, "conteudo")
	if err := s.MoveToTrash(filePath, 8); err != nil {
		t.Fatalf("MoveToTrash: %v", err)
	}

	page, err := s.GetItems(1, 10)
	if err != nil {
		t.Fatalf("GetItems: %v", err)
	}
	if len(page.Items) != 1 || page.Items[0].OriginalPath != filePath || page.Items[0].Size != 8 {
		t.Fatalf("unexpected dto page: %+v", page.Items)
	}
}

func TestTrashService_RepositoryErrorsPropagate(t *testing.T) {
	s, repo, _, _ := newTrashServiceForTest(t)
	repo.getErr = sql.ErrConnDone

	if _, err := s.GetItems(1, 10); err == nil {
		t.Fatalf("expected GetItems to propagate repository error")
	}
	if _, err := s.RestoreItem(1); err == nil {
		t.Fatalf("expected RestoreItem to propagate repository error")
	}
	if err := s.DeleteItemPermanently(1); err == nil {
		t.Fatalf("expected DeleteItemPermanently to propagate repository error")
	}
	if _, err := s.EmptyTrash(); err == nil {
		t.Fatalf("expected EmptyTrash to propagate repository error")
	}
	if _, err := s.PurgeExpired(); err == nil {
		t.Fatalf("expected PurgeExpired to propagate repository error")
	}
}

func TestTrashService_MoveToTrashRequiresEntryPoint(t *testing.T) {
	originalEntryPoint := config.AppConfig.EntryPoint
	t.Cleanup(func() { config.AppConfig.EntryPoint = originalEntryPoint })
	config.AppConfig.EntryPoint = ""

	s := NewService(newTrashRepoMock(), nil)
	if err := s.MoveToTrash("/qualquer/coisa.txt", 1); err == nil {
		t.Fatalf("expected error without a configured entry point")
	}
	if _, err := Dir(); err == nil {
		t.Fatalf("Dir must fail without a configured entry point")
	}
}

func TestTrashService_PurgeItemsSkipsBrokenRegistryRows(t *testing.T) {
	s, repo, _, root := newTrashServiceForTest(t)
	path := filepath.Join(root, "a.txt")
	writeFile(t, path, "a")
	if err := s.MoveToTrash(path, 1); err != nil {
		t.Fatalf("MoveToTrash: %v", err)
	}

	// One healthy item + one whose trash_path points outside the trash dir:
	// the broken one is skipped, the healthy one is purged.
	outside := filepath.Join(root, "fora.txt")
	writeFile(t, outside, "fora")
	repo.items[99] = TrashItemModel{ID: 99, OriginalPath: outside, TrashPath: outside}

	items, _ := repo.GetAllItems()
	purged := s.purgeItems(items)
	if purged != 1 {
		t.Fatalf("expected 1 purged (broken row skipped), got %d", purged)
	}
	if _, err := os.Stat(outside); err != nil {
		t.Fatalf("file outside the trash dir must survive: %v", err)
	}
}

func TestTrashService_IsInsideTrash(t *testing.T) {
	root := filepath.Join(string(filepath.Separator), "data")
	trashDir := filepath.Join(root, DirName)

	if !IsInsideTrash(root, trashDir) {
		t.Fatalf("the trash dir itself must count as inside")
	}
	if !IsInsideTrash(root, filepath.Join(trashDir, "x.txt")) {
		t.Fatalf("children must count as inside")
	}
	if IsInsideTrash(root, filepath.Join(root, "docs", "x.txt")) {
		t.Fatalf("normal paths must not count as inside")
	}
	if IsInsideTrash(root, filepath.Join(root, DirName+"-sufixo", "x.txt")) {
		t.Fatalf("a sibling dir sharing the prefix must not count as inside")
	}
}
