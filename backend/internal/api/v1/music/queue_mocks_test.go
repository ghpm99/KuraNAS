package music

import "errors"

func (m *musicRepoMock) recordQueueRequest(key string, limit int) ([]MusicQueueEntryModel, error) {
	m.receivedQueueKey = key
	m.receivedLimit = limit
	return m.queueEntries, m.queueFailure
}

func (m *musicRepoMock) GetLibraryQueueByArtist(artistKey string, limit int) ([]MusicQueueEntryModel, error) {
	return m.recordQueueRequest(artistKey, limit)
}
func (m *musicRepoMock) GetLibraryQueueByAlbum(albumKey string, limit int) ([]MusicQueueEntryModel, error) {
	return m.recordQueueRequest(albumKey, limit)
}
func (m *musicRepoMock) GetLibraryQueueByGenre(genreKey string, limit int) ([]MusicQueueEntryModel, error) {
	return m.recordQueueRequest(genreKey, limit)
}
func (m *musicRepoMock) GetLibraryQueueByFolder(folderPath string, limit int) ([]MusicQueueEntryModel, error) {
	return m.recordQueueRequest(folderPath, limit)
}
func (m *musicRepoMock) GetPlaylistQueue(playlistID int, limit int) ([]MusicQueueEntryModel, error) {
	return m.recordQueueRequest("playlist", limit)
}
func (m *musicRepoMock) GetLibraryQueueByFileIDs(fileIDs []int) ([]MusicQueueEntryModel, error) {
	m.receivedFileIDs = fileIDs
	return m.queueEntries, m.queueFailure
}

var sampleQueueDto = MusicQueueDto{Items: []MusicQueueEntryDto{{FileID: 7, Name: "a.mp3"}}}

func (m *musicHandlerServiceMock) GetLibraryQueueByArtist(artistKey string) (MusicQueueDto, error) {
	return sampleQueueDto, nil
}
func (m *musicHandlerServiceMock) GetLibraryQueueByAlbum(albumKey string) (MusicQueueDto, error) {
	return sampleQueueDto, nil
}
func (m *musicHandlerServiceMock) GetLibraryQueueByGenre(genreKey string) (MusicQueueDto, error) {
	return sampleQueueDto, nil
}
func (m *musicHandlerServiceMock) GetLibraryQueueByFolder(folderPath string) (MusicQueueDto, error) {
	return sampleQueueDto, nil
}
func (m *musicHandlerServiceMock) GetPlaylistQueue(clientID string, playlistID int) (MusicQueueDto, error) {
	return sampleQueueDto, nil
}

var errQueueFailure = errors.New("queue error")

func (m *musicHandlerErrServiceMock) GetLibraryQueueByArtist(artistKey string) (MusicQueueDto, error) {
	return MusicQueueDto{}, errQueueFailure
}
func (m *musicHandlerErrServiceMock) GetLibraryQueueByAlbum(albumKey string) (MusicQueueDto, error) {
	return MusicQueueDto{}, errQueueFailure
}
func (m *musicHandlerErrServiceMock) GetLibraryQueueByGenre(genreKey string) (MusicQueueDto, error) {
	return MusicQueueDto{}, errQueueFailure
}
func (m *musicHandlerErrServiceMock) GetLibraryQueueByFolder(folderPath string) (MusicQueueDto, error) {
	return MusicQueueDto{}, errQueueFailure
}
func (m *musicHandlerErrServiceMock) GetPlaylistQueue(clientID string, playlistID int) (MusicQueueDto, error) {
	return MusicQueueDto{}, errQueueFailure
}
