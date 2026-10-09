package music

import (
	"testing"

	"nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/utils"
)

func TestServiceSearchLibraryTracksDelegatesToRepository(t *testing.T) {
	repository := &musicRepoMock{}
	var receivedText string
	repository.searchLibraryTracksFn = func(searchText string, page int, pageSize int) (utils.PaginationResponse[files.FileModel], error) {
		receivedText = searchText
		return utils.PaginationResponse[files.FileModel]{}, nil
	}
	service := &Service{Repository: repository}
	if _, err := service.SearchLibraryTracks("queen", 1, 10); err != nil || receivedText != "queen" {
		t.Fatalf("unexpected result err=%v text=%q", err, receivedText)
	}
}
