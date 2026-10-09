package music

import (
	"database/sql"
	"errors"

	"nas-go/api/pkg/utils"
)

func (m *musicHandlerServiceMock) GetLibraryAlbumSummary(albumKey string) (MusicAlbumSummaryDto, error) {
	if albumKey == "missing" {
		return MusicAlbumSummaryDto{}, sql.ErrNoRows
	}
	return MusicAlbumSummaryDto{Key: albumKey, Name: "Album", Artist: "Artist", Year: "2020", TrackCount: 2, TotalLengthSeconds: 300, DiscCount: 2}, nil
}

func (m *musicHandlerServiceMock) GetLibraryArtistSummary(artistKey string) (MusicArtistSummaryDto, error) {
	if artistKey == "missing" {
		return MusicArtistSummaryDto{}, sql.ErrNoRows
	}
	return MusicArtistSummaryDto{Key: artistKey, Name: "Artist", TrackCount: 5, AlbumCount: 2, TotalLengthSeconds: 900}, nil
}

func (m *musicHandlerServiceMock) GetLibraryGenreSummary(genreKey string) (MusicGroupSummaryDto, error) {
	if genreKey == "missing" {
		return MusicGroupSummaryDto{}, sql.ErrNoRows
	}
	return MusicGroupSummaryDto{Key: genreKey, Name: "Genre", TrackCount: 3, TotalLengthSeconds: 600}, nil
}

func (m *musicHandlerServiceMock) GetLibraryFolderSummary(folderPath string) (MusicGroupSummaryDto, error) {
	if folderPath == "missing" {
		return MusicGroupSummaryDto{}, sql.ErrNoRows
	}
	return MusicGroupSummaryDto{Key: folderPath, Name: folderPath, TrackCount: 4, TotalLengthSeconds: 700}, nil
}

func (m *musicHandlerServiceMock) GetLibraryAlbumsByArtist(artistKey string, page int, pageSize int) (utils.PaginationResponse[MusicAlbumGroupDto], error) {
	if artistKey == "broken" {
		return utils.PaginationResponse[MusicAlbumGroupDto]{}, errors.New("artist albums error")
	}
	return utils.PaginationResponse[MusicAlbumGroupDto]{Items: []MusicAlbumGroupDto{{Key: "artist::album", Album: "Album", Artist: "Artist", Year: "2020", TrackCount: 2}}}, nil
}

func (m *musicRepoMock) GetLibraryAlbumSummary(albumKey string) (MusicAlbumSummaryDto, error) {
	return MusicAlbumSummaryDto{Key: albumKey}, nil
}

func (m *musicRepoMock) GetLibraryArtistSummary(artistKey string) (MusicArtistSummaryDto, error) {
	return MusicArtistSummaryDto{Key: artistKey}, nil
}

func (m *musicRepoMock) GetLibraryGenreSummary(genreKey string) (MusicGroupSummaryDto, error) {
	return MusicGroupSummaryDto{Key: genreKey}, nil
}

func (m *musicRepoMock) GetLibraryFolderSummary(folderPath string) (MusicGroupSummaryDto, error) {
	return MusicGroupSummaryDto{Key: folderPath}, nil
}

func (m *musicRepoMock) GetLibraryAlbumGroupsByArtist(artistKey string, page int, pageSize int) (utils.PaginationResponse[MusicAlbumGroupDto], error) {
	return utils.PaginationResponse[MusicAlbumGroupDto]{Items: []MusicAlbumGroupDto{{Key: artistKey, TrackCount: page*1000 + pageSize}}}, nil
}
