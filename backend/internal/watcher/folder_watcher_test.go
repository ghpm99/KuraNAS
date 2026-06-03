package watcher

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/api/v1/libraries"
	"nas-go/api/internal/api/v1/notifications"
	"nas-go/api/internal/api/v1/watchfolders"
)

// Doubles mínimos: as interfaces pequenas (WatchFolderService, libraries) são
// implementadas de verdade; as gigantes (files com 39 métodos, notifications)
// são embutidas e só o método realmente usado pelo fluxo é sobrescrito. O teste
// exercita o código real do FolderWatcher movendo arquivos de verdade.

type fakeWatchFolderSvc struct {
	folders    []watchfolders.WatchFolderModel
	lastScanID int
	lastScanAt time.Time
}

func (f *fakeWatchFolderSvc) GetEnabledWatchFolders() ([]watchfolders.WatchFolderModel, error) {
	return f.folders, nil
}

func (f *fakeWatchFolderSvc) UpdateWatchFolderLastScan(id int, lastScanAt time.Time) error {
	f.lastScanID = id
	f.lastScanAt = lastScanAt
	return nil
}

type fakeLibrariesSvc struct{ base string }

func (f *fakeLibrariesSvc) GetLibraries() ([]libraries.LibraryDto, error) { return nil, nil }
func (f *fakeLibrariesSvc) GetLibraryByCategory(category libraries.LibraryCategory) (libraries.LibraryDto, error) {
	return libraries.LibraryDto{Category: string(category), Path: filepath.Join(f.base, string(category))}, nil
}
func (f *fakeLibrariesSvc) UpdateLibrary(category libraries.LibraryCategory, dto libraries.UpdateLibraryDto) (libraries.LibraryDto, error) {
	return libraries.LibraryDto{}, nil
}
func (f *fakeLibrariesSvc) ResolveLibraries() error { return nil }

type fakeFilesSvc struct {
	files.ServiceInterface
	jobs [][]string
}

func (f *fakeFilesSvc) CreateUploadProcessJob(paths []string) (int, error) {
	f.jobs = append(f.jobs, paths)
	return len(f.jobs), nil
}

type fakeNotifSvc struct {
	notifications.ServiceInterface
	count int
}

func (f *fakeNotifSvc) GroupOrCreate(dto notifications.CreateNotificationDto) (notifications.NotificationDto, error) {
	f.count++
	return notifications.NotificationDto{}, nil
}

func TestFolderWatcherScanOnceImportsAndEnqueues(t *testing.T) {
	watchDir := t.TempDir()
	libBase := t.TempDir()

	img := filepath.Join(watchDir, "a.jpg")
	vid := filepath.Join(watchDir, "b.mp4")
	for _, p := range []string{img, vid} {
		if err := os.WriteFile(p, []byte("payload"), 0644); err != nil {
			t.Fatalf("write %s: %v", p, err)
		}
	}

	wfSvc := &fakeWatchFolderSvc{folders: []watchfolders.WatchFolderModel{{ID: 7, Path: watchDir, Enabled: true}}}
	fSvc := &fakeFilesSvc{}
	nSvc := &fakeNotifSvc{}

	fw := NewFolderWatcher(wfSvc, &fakeLibrariesSvc{base: libBase}, fSvc, nSvc, time.Minute)
	fw.scanOnce()

	// Os arquivos devem ter saído da pasta monitorada (move, não cópia).
	if _, err := os.Stat(img); !os.IsNotExist(err) {
		t.Fatalf("expected image moved out of watch folder")
	}
	if _, err := os.Stat(vid); !os.IsNotExist(err) {
		t.Fatalf("expected video moved out of watch folder")
	}
	// E aterrissado na biblioteca da categoria correta.
	if _, err := os.Stat(filepath.Join(libBase, string(libraries.LibraryCategoryImages), "a.jpg")); err != nil {
		t.Fatalf("expected image in images library: %v", err)
	}
	if _, err := os.Stat(filepath.Join(libBase, string(libraries.LibraryCategoryVideos), "b.mp4")); err != nil {
		t.Fatalf("expected video in videos library: %v", err)
	}
	// Cada arquivo importado enfileira um job de processamento/indexação.
	if len(fSvc.jobs) != 2 {
		t.Fatalf("expected 2 upload-process jobs, got %d", len(fSvc.jobs))
	}
	// last_scan_at atualizado para a pasta certa.
	if wfSvc.lastScanID != 7 {
		t.Fatalf("expected last scan update for folder 7, got %d", wfSvc.lastScanID)
	}
	// Notificação de importação emitida.
	if nSvc.count != 1 {
		t.Fatalf("expected 1 import notification, got %d", nSvc.count)
	}
}

func TestFolderWatcherStartStopLifecycle(t *testing.T) {
	wfSvc := &fakeWatchFolderSvc{} // nenhuma pasta habilitada
	fw := NewFolderWatcher(wfSvc, &fakeLibrariesSvc{base: t.TempDir()}, &fakeFilesSvc{}, &fakeNotifSvc{}, 10*time.Millisecond)

	fw.Start()
	fw.Start() // idempotente: já rodando, não dispara segunda goroutine
	fw.Stop()
	fw.Stop() // idempotente: já parado
}
