package music

import (
	"errors"
	"testing"
)

func TestReplacePlayerQueueValidatesBounds(t *testing.T) {
	tooManyIDs := make([]int, maxPlayerQueueEntries+1)
	cases := []struct {
		name      string
		request   ReplacePlayerQueueRequest
		isAllowed bool
	}{
		{"empty queue at zero", ReplacePlayerQueueRequest{}, true},
		{"empty queue with index", ReplacePlayerQueueRequest{CurrentIndex: 1}, false},
		{"index inside queue", ReplacePlayerQueueRequest{FileIDs: []int{1, 2}, CurrentIndex: 1}, true},
		{"index past the end", ReplacePlayerQueueRequest{FileIDs: []int{1, 2}, CurrentIndex: 2}, false},
		{"negative index", ReplacePlayerQueueRequest{FileIDs: []int{1}, CurrentIndex: -1}, false},
		{"over the limit", ReplacePlayerQueueRequest{FileIDs: tooManyIDs}, false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			var savedFileIDs []int
			repository := &musicRepoMock{replacePlayerQueueFn: func(clientID string, fileIDs []int, currentIndex int) error {
				savedFileIDs = fileIDs
				return nil
			}}
			service := newMusicServiceForTest(t, repository)

			err := service.ReplacePlayerQueue("device-12345", testCase.request)

			if testCase.isAllowed && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !testCase.isAllowed && !errors.Is(err, ErrInvalidPlayerQueue) {
				t.Fatalf("error = %v, want ErrInvalidPlayerQueue", err)
			}
			if !testCase.isAllowed && savedFileIDs != nil {
				t.Fatalf("invalid queue reached the repository")
			}
		})
	}
}

func TestReplacePlayerQueueWrapsRepositoryFailure(t *testing.T) {
	repository := &musicRepoMock{playerQueueFailure: errors.New("boom")}
	service := newMusicServiceForTest(t, repository)

	if err := service.ReplacePlayerQueue("device-12345", ReplacePlayerQueueRequest{FileIDs: []int{1}}); err == nil {
		t.Fatalf("expected repository failure")
	}
}

func TestGetPlayerQueueComposesEntriesAndIndex(t *testing.T) {
	repository := &musicRepoMock{
		playerQueueEntries: []MusicQueueEntryModel{{FileID: 3, Name: "c.mp3"}, {FileID: 8, Name: "h.mp3"}},
		playerQueueIndex:   1,
	}
	service := newMusicServiceForTest(t, repository)

	queue, err := service.GetPlayerQueue("device-12345")

	if err != nil || len(queue.Items) != 2 || queue.Items[1].FileID != 8 || queue.CurrentIndex != 1 {
		t.Fatalf("queue = %+v err=%v", queue, err)
	}
}

func TestGetPlayerQueueReportsRepositoryFailure(t *testing.T) {
	service := newMusicServiceForTest(t, &musicRepoMock{playerQueueFailure: errors.New("boom")})

	if _, err := service.GetPlayerQueue("device-12345"); err == nil {
		t.Fatalf("expected repository failure")
	}
}
