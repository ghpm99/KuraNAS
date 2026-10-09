package music

import (
	"reflect"
	"testing"
)

func seedPlayerQueueLibrary(t *testing.T) (*catalogTestEnvironment, map[string]int) {
	t.Helper()
	environment := newCatalogTestEnvironment(t)
	base := environment.catalogBaseMoment
	environment.seedTracks(t, []catalogTrackFixture{
		trackFixture("one.mp3", "/m", base, AudioMetadataModel{Title: "One", Artist: "X", Album: "Y"}),
		trackFixture("two.mp3", "/m", base, AudioMetadataModel{Title: "Two", Artist: "X", Album: "Y"}),
		trackFixture("three.mp3", "/m", base, AudioMetadataModel{Title: "Three", Artist: "X", Album: "Y"}),
		{name: "gone.mp3", parentPath: "/m", updatedAt: base, isDeleted: true, metadata: &AudioMetadataModel{Title: "Gone"}},
	})
	return environment, environment.fileIDsByName
}

func TestPlayerQueueIsIsolatedPerClient_Postgres(t *testing.T) {
	environment, ids := seedPlayerQueueLibrary(t)
	service := environment.service

	if err := service.ReplacePlayerQueue("device-aaaaaaaa", ReplacePlayerQueueRequest{FileIDs: []int{ids["one.mp3"], ids["two.mp3"]}, CurrentIndex: 1}); err != nil {
		t.Fatalf("save a: %v", err)
	}
	if err := service.ReplacePlayerQueue("device-bbbbbbbb", ReplacePlayerQueueRequest{FileIDs: []int{ids["three.mp3"]}, CurrentIndex: 0}); err != nil {
		t.Fatalf("save b: %v", err)
	}

	queueA, err := service.GetPlayerQueue("device-aaaaaaaa")
	if err != nil {
		t.Fatalf("get a: %v", err)
	}
	queueB, _ := service.GetPlayerQueue("device-bbbbbbbb")
	queueUnknown, _ := service.GetPlayerQueue("device-cccccccc")

	if got := queueEntryIDs(queueA); !reflect.DeepEqual(got, []int{ids["one.mp3"], ids["two.mp3"]}) || queueA.CurrentIndex != 1 {
		t.Fatalf("queue a = %+v", queueA)
	}
	if got := queueEntryIDs(queueB); !reflect.DeepEqual(got, []int{ids["three.mp3"]}) || queueB.CurrentIndex != 0 {
		t.Fatalf("queue b = %+v", queueB)
	}
	if len(queueUnknown.Items) != 0 || queueUnknown.CurrentIndex != 0 {
		t.Fatalf("unknown client queue = %+v", queueUnknown)
	}
}

func queueEntryIDs(queue PlayerQueueDto) []int {
	fileIDs := make([]int, 0, len(queue.Items))
	for _, entry := range queue.Items {
		fileIDs = append(fileIDs, entry.FileID)
	}
	return fileIDs
}

func TestPlayerQueueReplaceOverwritesKeepsDuplicatesAndCarriesPlayerFields_Postgres(t *testing.T) {
	environment, ids := seedPlayerQueueLibrary(t)
	service := environment.service
	clientID := "device-aaaaaaaa"

	_ = service.ReplacePlayerQueue(clientID, ReplacePlayerQueueRequest{FileIDs: []int{ids["one.mp3"]}})
	err := service.ReplacePlayerQueue(clientID, ReplacePlayerQueueRequest{FileIDs: []int{ids["two.mp3"], ids["two.mp3"], ids["three.mp3"]}, CurrentIndex: 2})
	if err != nil {
		t.Fatalf("replace: %v", err)
	}

	queue, _ := service.GetPlayerQueue(clientID)

	if got := queueEntryIDs(queue); !reflect.DeepEqual(got, []int{ids["two.mp3"], ids["two.mp3"], ids["three.mp3"]}) || queue.CurrentIndex != 2 {
		t.Fatalf("queue = %+v", queue)
	}
	first := queue.Items[0]
	if first.Name != "two.mp3" || first.Path != "/m/two.mp3" || first.Format != ".mp3" || first.Title != "Two" || first.Artist != "X" || first.Album != "Y" {
		t.Fatalf("entry = %+v", first)
	}

	if err := service.ReplacePlayerQueue(clientID, ReplacePlayerQueueRequest{}); err != nil {
		t.Fatalf("clear: %v", err)
	}
	cleared, _ := service.GetPlayerQueue(clientID)
	if len(cleared.Items) != 0 || cleared.CurrentIndex != 0 {
		t.Fatalf("cleared queue = %+v", cleared)
	}
}

func TestPlayerQueueSkipsDeletedAndUnknownFilesAndAdjustsIndex_Postgres(t *testing.T) {
	environment, ids := seedPlayerQueueLibrary(t)
	service := environment.service
	clientID := "device-aaaaaaaa"
	unknownFileID := 987654

	err := service.ReplacePlayerQueue(clientID, ReplacePlayerQueueRequest{
		FileIDs:      []int{ids["gone.mp3"], ids["one.mp3"], unknownFileID, ids["two.mp3"]},
		CurrentIndex: 3,
	})
	if err != nil {
		t.Fatalf("replace: %v", err)
	}

	queue, _ := service.GetPlayerQueue(clientID)

	if got := queueEntryIDs(queue); !reflect.DeepEqual(got, []int{ids["one.mp3"], ids["two.mp3"]}) {
		t.Fatalf("queue = %v", got)
	}
	if queue.CurrentIndex != 1 {
		t.Fatalf("current index = %d, want 1 (points at two.mp3)", queue.CurrentIndex)
	}
}

func TestPlayerQueueSaveDoesNotClobberPlayerState_Postgres(t *testing.T) {
	environment, ids := seedPlayerQueueLibrary(t)
	service := environment.service
	clientID := "device-aaaaaaaa"
	fileID := ids["one.mp3"]

	_, err := service.UpdatePlayerState(clientID, UpdatePlayerStateRequest{CurrentFileID: &fileID, CurrentPosition: 42, Volume: 0.4, Shuffle: true, RepeatMode: "all"})
	if err != nil {
		t.Fatalf("state: %v", err)
	}
	if err := service.ReplacePlayerQueue(clientID, ReplacePlayerQueueRequest{FileIDs: []int{fileID}}); err != nil {
		t.Fatalf("queue: %v", err)
	}

	state, err := service.GetPlayerState(clientID)

	if err != nil || state.CurrentPosition != 42 || state.Volume != 0.4 || !state.Shuffle || state.RepeatMode != "all" {
		t.Fatalf("state = %+v err=%v", state, err)
	}
}
