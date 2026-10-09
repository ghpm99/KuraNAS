package music

import (
	"errors"
	"reflect"
	"testing"
)

func queueEntriesOfCount(count int) []MusicQueueEntryModel {
	entries := make([]MusicQueueEntryModel, count)
	for index := range entries {
		entries[index] = MusicQueueEntryModel{FileID: index + 1, Name: "t.mp3", Title: "T", LengthSeconds: 61.5}
	}
	return entries
}

func TestLibraryQueueConvertsEntriesAndAsksForOneExtraRow(t *testing.T) {
	repository := &musicRepoMock{queueEntries: queueEntriesOfCount(2)}
	service := &Service{Repository: repository}

	queue, err := service.GetLibraryQueueByArtist("artist-key")
	if err != nil {
		t.Fatalf("queue: %v", err)
	}
	if queue.Truncated || len(queue.Items) != 2 || queue.Items[1].FileID != 2 || queue.Items[0].LengthSeconds != 61.5 {
		t.Fatalf("queue = %+v", queue)
	}
	if repository.receivedQueueKey != "artist-key" || repository.receivedLimit != maxQueueEntries+1 {
		t.Fatalf("repository received key=%q limit=%d", repository.receivedQueueKey, repository.receivedLimit)
	}
}

func TestLibraryQueueIsCappedAndFlaggedTruncated(t *testing.T) {
	service := &Service{Repository: &musicRepoMock{queueEntries: queueEntriesOfCount(maxQueueEntries + 1)}}

	for name, load := range map[string]func() (MusicQueueDto, error){
		"album":    func() (MusicQueueDto, error) { return service.GetLibraryQueueByAlbum("k") },
		"genre":    func() (MusicQueueDto, error) { return service.GetLibraryQueueByGenre("k") },
		"folder":   func() (MusicQueueDto, error) { return service.GetLibraryQueueByFolder("/m") },
		"playlist": func() (MusicQueueDto, error) { return service.GetPlaylistQueue("ip", 5) },
	} {
		queue, err := load()
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if !queue.Truncated || len(queue.Items) != maxQueueEntries {
			t.Fatalf("%s truncated=%v len=%d", name, queue.Truncated, len(queue.Items))
		}
	}
}

func TestFolderQueueTrimsKeyAndIgnoresBlankFolder(t *testing.T) {
	repository := &musicRepoMock{queueEntries: queueEntriesOfCount(1)}
	service := &Service{Repository: repository}

	if _, err := service.GetLibraryQueueByFolder("  /music/b "); err != nil || repository.receivedQueueKey != "/music/b" {
		t.Fatalf("folder key = %q err=%v", repository.receivedQueueKey, err)
	}

	repository.receivedQueueKey = ""
	blank, err := service.GetLibraryQueueByFolder("   ")
	if err != nil || len(blank.Items) != 0 || repository.receivedQueueKey != "" {
		t.Fatalf("blank folder = %+v err=%v key=%q", blank, err, repository.receivedQueueKey)
	}
}

func TestAutomaticPlaylistQueueResolvesIDsThenLoadsEntries(t *testing.T) {
	repository := &musicRepoMock{
		queueEntries:         queueEntriesOfCount(2),
		getFavoriteFileIDsFn: func(limit int) ([]int, error) { return []int{9, 4}, nil },
	}
	service := &Service{Repository: repository}

	queue, err := service.GetPlaylistQueue("ip", AutoPlaylistFavoritesID)
	if err != nil || len(queue.Items) != 2 {
		t.Fatalf("queue = %+v err=%v", queue, err)
	}
	if !reflect.DeepEqual(repository.receivedFileIDs, []int{9, 4}) {
		t.Fatalf("file ids = %v", repository.receivedFileIDs)
	}

	if _, err := service.GetPlaylistQueue("ip", -99); err == nil {
		t.Fatal("unknown automatic playlist must fail")
	}
}

func TestQueueRepositoryFailuresPropagate(t *testing.T) {
	failure := errors.New("boom")
	service := &Service{Repository: &musicRepoMock{queueFailure: failure}}

	if _, err := service.GetLibraryQueueByArtist("k"); !errors.Is(err, failure) {
		t.Fatalf("artist err = %v", err)
	}
	if _, err := service.GetPlaylistQueue("ip", 2); !errors.Is(err, failure) {
		t.Fatalf("playlist err = %v", err)
	}
	if _, err := service.GetPlaylistQueue("ip", AutoPlaylistRecentlyAddedID); !errors.Is(err, failure) {
		t.Fatalf("automatic playlist err = %v", err)
	}
}
