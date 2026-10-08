package tiering

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

type promoteFixture struct {
	service      *Service
	repository   *fakeRepo
	hotPath      string
	coldPath     string
	fileContents string
}

func newPromoteFixture(t *testing.T, availableBytes int64) promoteFixture {
	t.Helper()
	hotPath := filepath.Join(t.TempDir(), "docs", "relatorio.pdf")
	coldPath := filepath.Join(t.TempDir(), "Casa", "docs", "relatorio.pdf")
	if err := os.MkdirAll(filepath.Dir(coldPath), 0o755); err != nil {
		t.Fatalf("failed to create cold dir: %v", err)
	}
	fileContents := "conteudo-frio"
	if err := os.WriteFile(coldPath, []byte(fileContents), 0o644); err != nil {
		t.Fatalf("failed to write cold file: %v", err)
	}

	repository := &fakeRepo{files: map[int]TieredFileModel{
		10: {FileID: 10, LogicalPath: hotPath, PhysicalPath: coldPath, Size: int64(len(fileContents))},
		11: {FileID: 11, LogicalPath: hotPath, Size: 5},
	}}
	service := newTestService(repository)
	service.availableBytes = func(string) (int64, error) { return availableBytes, nil }
	return promoteFixture{service: service, repository: repository, hotPath: hotPath, coldPath: coldPath, fileContents: fileContents}
}

func TestPromoteFileMovesBytesToHotPathAndClearsPhysicalPath(t *testing.T) {
	fixture := newPromoteFixture(t, 1<<30)

	location, err := fixture.service.PromoteFile(10)

	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if location.FileID != 10 || location.Tier != "hot" || location.DiskPath != fixture.hotPath {
		t.Fatalf("unexpected location %+v", location)
	}
	hotBytes, readErr := os.ReadFile(fixture.hotPath)
	if readErr != nil || string(hotBytes) != fixture.fileContents {
		t.Fatalf("expected hot copy with original bytes, got %q err=%v", hotBytes, readErr)
	}
	if _, statErr := os.Stat(fixture.coldPath); !os.IsNotExist(statErr) {
		t.Fatalf("expected the cold copy to be removed, stat err: %v", statErr)
	}
	if len(fixture.repository.setCalls) != 1 || fixture.repository.setCalls[0].id != 10 || fixture.repository.setCalls[0].path != "" {
		t.Fatalf("expected physical_path cleared once, got %+v", fixture.repository.setCalls)
	}
}

func TestPromoteFileUnknownFile(t *testing.T) {
	fixture := newPromoteFixture(t, 1<<30)

	_, err := fixture.service.PromoteFile(999)

	if !errors.Is(err, ErrFileNotFound) {
		t.Fatalf("expected ErrFileNotFound, got %v", err)
	}
}

func TestPromoteFileAlreadyHot(t *testing.T) {
	fixture := newPromoteFixture(t, 1<<30)

	_, err := fixture.service.PromoteFile(11)

	if !errors.Is(err, ErrFileAlreadyHot) {
		t.Fatalf("expected ErrFileAlreadyHot, got %v", err)
	}
	if len(fixture.repository.setCalls) != 0 {
		t.Fatalf("an already hot file must not be touched")
	}
}

func TestPromoteFileRefusesWhenHotDiskLacksSpaceAndKeepsColdCopy(t *testing.T) {
	fixture := newPromoteFixture(t, 3)

	_, err := fixture.service.PromoteFile(10)

	if !errors.Is(err, ErrInsufficientHotSpace) {
		t.Fatalf("expected ErrInsufficientHotSpace, got %v", err)
	}
	if _, statErr := os.Stat(fixture.coldPath); statErr != nil {
		t.Fatalf("cold copy must survive a refused promotion: %v", statErr)
	}
	if len(fixture.repository.setCalls) != 0 {
		t.Fatalf("physical_path must not change on a refused promotion")
	}
}

func TestPromoteFileColdCopyUnavailable(t *testing.T) {
	fixture := newPromoteFixture(t, 1<<30)
	if err := os.Remove(fixture.coldPath); err != nil {
		t.Fatalf("failed to simulate unmounted cold volume: %v", err)
	}

	_, err := fixture.service.PromoteFile(10)

	if !errors.Is(err, ErrColdCopyUnavailable) {
		t.Fatalf("expected ErrColdCopyUnavailable, got %v", err)
	}
}

func TestPromoteFilePropagatesRepositoryAndFreeSpaceFailures(t *testing.T) {
	failingRepository := newPromoteFixture(t, 1<<30)
	failingRepository.repository.loadErr = errors.New("db down")
	if _, err := failingRepository.service.PromoteFile(10); err == nil {
		t.Fatalf("expected repository error")
	}

	failingSpaceCheck := newPromoteFixture(t, 1<<30)
	failingSpaceCheck.service.availableBytes = func(string) (int64, error) { return 0, errors.New("statfs failed") }
	if _, err := failingSpaceCheck.service.PromoteFile(10); err == nil || errors.Is(err, ErrInsufficientHotSpace) {
		t.Fatalf("expected a generic free-space error, got %v", err)
	}
}

func TestNearestExistingDirectoryWalksUpToExistingAncestor(t *testing.T) {
	existing := t.TempDir()

	resolved := nearestExistingDirectory(filepath.Join(existing, "missing", "deeper"))

	if resolved != existing {
		t.Fatalf("expected %q, got %q", existing, resolved)
	}
}
