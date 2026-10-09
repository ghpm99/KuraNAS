package music

import "database/sql"

func (m *musicRepoMock) ReplacePlayerQueue(tx *sql.Tx, clientID string, fileIDs []int, currentIndex int) error {
	if m.replacePlayerQueueFn != nil {
		return m.replacePlayerQueueFn(clientID, fileIDs, currentIndex)
	}
	return m.playerQueueFailure
}

func (m *musicRepoMock) GetPlayerQueue(clientID string) ([]MusicQueueEntryModel, error) {
	return m.playerQueueEntries, m.playerQueueFailure
}

func (m *musicRepoMock) GetPlayerQueueCurrentIndex(clientID string) (int, error) {
	return m.playerQueueIndex, m.playerQueueFailure
}

func (m *musicHandlerServiceMock) ReplacePlayerQueue(clientID string, request ReplacePlayerQueueRequest) error {
	return nil
}

func (m *musicHandlerServiceMock) GetPlayerQueue(clientID string) (PlayerQueueDto, error) {
	return PlayerQueueDto{Items: []MusicQueueEntryDto{{FileID: 7, Name: "a.mp3"}}, CurrentIndex: 0}, nil
}

func (m *musicHandlerErrServiceMock) ReplacePlayerQueue(clientID string, request ReplacePlayerQueueRequest) error {
	return errQueueFailure
}

func (m *musicHandlerErrServiceMock) GetPlayerQueue(clientID string) (PlayerQueueDto, error) {
	return PlayerQueueDto{}, errQueueFailure
}
