package music

import "strings"

const maxQueueEntries = 10000

func buildQueueDto(entries []MusicQueueEntryModel) MusicQueueDto {
	isTruncated := len(entries) > maxQueueEntries
	if isTruncated {
		entries = entries[:maxQueueEntries]
	}

	items := make([]MusicQueueEntryDto, 0, len(entries))
	for _, entry := range entries {
		items = append(items, entry.ToDto())
	}
	return MusicQueueDto{Items: items, Truncated: isTruncated}
}

func queueDtoOrError(entries []MusicQueueEntryModel, err error) (MusicQueueDto, error) {
	if err != nil {
		return MusicQueueDto{}, err
	}
	return buildQueueDto(entries), nil
}

func (s *Service) GetLibraryQueueByArtist(artistKey string) (MusicQueueDto, error) {
	return queueDtoOrError(s.Repository.GetLibraryQueueByArtist(artistKey, maxQueueEntries+1))
}

func (s *Service) GetLibraryQueueByAlbum(albumKey string) (MusicQueueDto, error) {
	return queueDtoOrError(s.Repository.GetLibraryQueueByAlbum(albumKey, maxQueueEntries+1))
}

func (s *Service) GetLibraryQueueByGenre(genreKey string) (MusicQueueDto, error) {
	return queueDtoOrError(s.Repository.GetLibraryQueueByGenre(genreKey, maxQueueEntries+1))
}

func (s *Service) GetLibraryQueueByFolder(folderPath string) (MusicQueueDto, error) {
	trimmedFolder := strings.TrimSpace(folderPath)
	if trimmedFolder == "" {
		return buildQueueDto(nil), nil
	}
	return queueDtoOrError(s.Repository.GetLibraryQueueByFolder(trimmedFolder, maxQueueEntries+1))
}

func (s *Service) GetPlaylistQueue(clientID string, playlistID int) (MusicQueueDto, error) {
	if playlistID >= 0 {
		return queueDtoOrError(s.Repository.GetPlaylistQueue(playlistID, maxQueueEntries+1))
	}

	fileIDs, err := s.automaticPlaylistTrackIDs(clientID, playlistID)
	if err != nil {
		return MusicQueueDto{}, err
	}
	return queueDtoOrError(s.Repository.GetLibraryQueueByFileIDs(fileIDs))
}
