package indexing

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/api/v1/storageroots"
	"nas-go/api/pkg/utils"
)

type launchCall struct {
	reindexKey string
	paths      []string
}

type jobLauncherFake struct {
	calls            []launchCall
	jobID            int
	isAlreadyRunning bool
	err              error
}

func (fake *jobLauncherFake) LaunchReindex(reindexKey string, paths []string) (int, bool, error) {
	fake.calls = append(fake.calls, launchCall{reindexKey: reindexKey, paths: paths})
	return fake.jobID, fake.isAlreadyRunning, fake.err
}

type rootsReaderFake struct {
	registeredRoots []storageroots.StorageRootDto
	err             error
}

func (fake *rootsReaderFake) GetRoots() ([]storageroots.StorageRootDto, error) {
	return fake.registeredRoots, fake.err
}

type folderReaderFake struct {
	foldersByID map[int]files.FileDto
	err         error
}

func (fake *folderReaderFake) GetFileById(id int) (files.FileDto, error) {
	if fake.err != nil {
		return files.FileDto{}, fake.err
	}
	folder, exists := fake.foldersByID[id]
	if !exists {
		return files.FileDto{}, sql.ErrNoRows
	}
	return folder, nil
}

func newServiceFixture() (*Service, *jobLauncherFake) {
	launcher := &jobLauncherFake{jobID: 11}
	roots := &rootsReaderFake{registeredRoots: []storageroots.StorageRootDto{
		{ID: 1, Path: "/data/main", Enabled: true},
		{ID: 2, Path: "/data/media", Enabled: true},
		{ID: 3, Path: "/data/off", Enabled: false},
	}}
	folders := &folderReaderFake{foldersByID: map[int]files.FileDto{
		345: {ID: 345, Path: "/data/main/photos", Type: files.Directory},
		346: {ID: 346, Path: "/data/main/a.txt", Type: files.File},
		347: {ID: 347, Path: "/data/main/old", Type: files.Directory, DeletedAt: utils.Optional[time.Time]{HasValue: true, Value: time.Now()}},
	}}
	return NewService(launcher, roots, folders), launcher
}

func TestReindexAllLaunchesEveryEnabledRoot(t *testing.T) {
	service, launcher := newServiceFixture()

	jobID, err := service.Reindex(ReindexRequestDto{Scope: ScopeAll})
	if err != nil || jobID != 11 {
		t.Fatalf("unexpected result: id=%d err=%v", jobID, err)
	}
	call := launcher.calls[0]
	if call.reindexKey != "all" || len(call.paths) != 2 || call.paths[0] != "/data/main" || call.paths[1] != "/data/media" {
		t.Fatalf("unexpected launch: %+v", call)
	}
}

func TestReindexRootLaunchesOnlyThatRoot(t *testing.T) {
	service, launcher := newServiceFixture()

	if _, err := service.Reindex(ReindexRequestDto{Scope: ScopeRoot, RootID: 2}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	call := launcher.calls[0]
	if call.reindexKey != "root:2" || len(call.paths) != 1 || call.paths[0] != "/data/media" {
		t.Fatalf("unexpected launch: %+v", call)
	}
}

func TestReindexFolderLaunchesTheFolderPath(t *testing.T) {
	service, launcher := newServiceFixture()

	if _, err := service.Reindex(ReindexRequestDto{Scope: ScopeFolder, FolderID: 345}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	call := launcher.calls[0]
	if call.reindexKey != "folder:345" || len(call.paths) != 1 || call.paths[0] != "/data/main/photos" {
		t.Fatalf("unexpected launch: %+v", call)
	}
}

func TestReindexRejectsInvalidRequests(t *testing.T) {
	cases := []struct {
		name     string
		request  ReindexRequestDto
		expected error
	}{
		{"unknown scope", ReindexRequestDto{Scope: "galaxy"}, ErrInvalidScope},
		{"empty scope", ReindexRequestDto{}, ErrInvalidScope},
		{"root without id", ReindexRequestDto{Scope: ScopeRoot}, ErrIdentifierRequired},
		{"folder without id", ReindexRequestDto{Scope: ScopeFolder}, ErrIdentifierRequired},
		{"unknown root", ReindexRequestDto{Scope: ScopeRoot, RootID: 99}, ErrRootNotFound},
		{"disabled root", ReindexRequestDto{Scope: ScopeRoot, RootID: 3}, ErrRootNotFound},
		{"unknown folder", ReindexRequestDto{Scope: ScopeFolder, FolderID: 99}, ErrFolderNotFound},
		{"file instead of folder", ReindexRequestDto{Scope: ScopeFolder, FolderID: 346}, ErrFolderNotFound},
		{"deleted folder", ReindexRequestDto{Scope: ScopeFolder, FolderID: 347}, ErrFolderNotFound},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			service, launcher := newServiceFixture()
			_, err := service.Reindex(testCase.request)
			if !errors.Is(err, testCase.expected) {
				t.Fatalf("expected %v, got %v", testCase.expected, err)
			}
			if len(launcher.calls) != 0 {
				t.Fatalf("no job may be launched for an invalid request")
			}
		})
	}
}

func TestReindexAllWithoutEnabledRootsIsNotFound(t *testing.T) {
	service := NewService(&jobLauncherFake{}, &rootsReaderFake{}, &folderReaderFake{})
	if _, err := service.Reindex(ReindexRequestDto{Scope: ScopeAll}); !errors.Is(err, ErrRootNotFound) {
		t.Fatalf("expected ErrRootNotFound, got %v", err)
	}
}

func TestReindexReportsAlreadyRunningJob(t *testing.T) {
	service, launcher := newServiceFixture()
	launcher.jobID = 77
	launcher.isAlreadyRunning = true

	_, err := service.Reindex(ReindexRequestDto{Scope: ScopeFolder, FolderID: 345})
	var alreadyRunning *AlreadyRunningError
	if !errors.As(err, &alreadyRunning) || alreadyRunning.JobID != 77 {
		t.Fatalf("expected AlreadyRunningError for job 77, got %v", err)
	}
}

func TestReindexPropagatesDependencyFailures(t *testing.T) {
	failure := errors.New("boom")

	launcherFailure := NewService(&jobLauncherFake{err: failure}, &rootsReaderFake{registeredRoots: []storageroots.StorageRootDto{{ID: 1, Path: "/a", Enabled: true}}}, &folderReaderFake{})
	if _, err := launcherFailure.Reindex(ReindexRequestDto{Scope: ScopeAll}); !errors.Is(err, failure) {
		t.Fatalf("expected launcher error, got %v", err)
	}
	rootsFailure := NewService(&jobLauncherFake{}, &rootsReaderFake{err: failure}, &folderReaderFake{})
	if _, err := rootsFailure.Reindex(ReindexRequestDto{Scope: ScopeRoot, RootID: 1}); !errors.Is(err, failure) {
		t.Fatalf("expected roots error, got %v", err)
	}
	foldersFailure := NewService(&jobLauncherFake{}, &rootsReaderFake{}, &folderReaderFake{err: failure})
	if _, err := foldersFailure.Reindex(ReindexRequestDto{Scope: ScopeFolder, FolderID: 1}); !errors.Is(err, failure) {
		t.Fatalf("expected folders error, got %v", err)
	}
}
