package music

import (
	"database/sql"
	"errors"
	files "nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/utils"
	"testing"
	"time"
)

func musicFileModel(id int, name string, parentPath string) files.FileModel {
	now := time.Date(2026, time.March, 10, 10, 0, 0, 0, time.UTC)
	return files.FileModel{
		ID:         id,
		Name:       name,
		Path:       parentPath + "/" + name,
		ParentPath: parentPath,
		Type:       files.File,
		Format:     ".mp3",
		Size:       1024,
		CreatedAt:  now,
		UpdatedAt:  now,
		Metadata: AudioMetadataModel{
			ID:        id,
			FileId:    id,
			Path:      parentPath + "/" + name,
			Title:     name,
			Artist:    "Artist",
			Album:     "Album",
			Year:      "2026",
			Genre:     "Pop",
			CreatedAt: now,
		},
	}
}

func TestCatalogHelpersNormalizeAndPaginate(t *testing.T) {
	if got := normalizeText("  hello   world  "); got != "hello world" {
		t.Fatalf("normalizeText returned %q", got)
	}
	if got := normalizeLookupKey(" Hello_World.MP3 "); got != "hello world mp3" {
		t.Fatalf("normalizeLookupKey returned %q", got)
	}
	if got := normalizeGenreLabel("rnb/soul"); got != "R&B / Soul" {
		t.Fatalf("normalizeGenreLabel returned %q", got)
	}
	if got := normalizeGenreLabels("hip hop; hip-hop | soundtrack"); len(got) != 2 || got[0] != "Hip-Hop" || got[1] != "Soundtrack" {
		t.Fatalf("normalizeGenreLabels returned %+v", got)
	}
	if got := preferredArtist(" Album Artist ", "Artist"); got != "Album Artist" {
		t.Fatalf("preferredArtist returned %q", got)
	}
	if got := preferredArtist("  ", " Artist  Name "); got != "Artist Name" {
		t.Fatalf("preferredArtist fallback returned %q", got)
	}

	paginated := paginateItems([]int{1, 2, 3}, 2, 2)
	if len(paginated.Items) != 1 || paginated.Items[0] != 3 || paginated.Pagination.HasPrev != true {
		t.Fatalf("paginateItems returned %+v", paginated)
	}

	beyondEnd := paginateItems([]int{1, 2, 3}, 5, 2)
	if len(beyondEnd.Items) != 0 || beyondEnd.Pagination.HasNext || !beyondEnd.Pagination.HasPrev {
		t.Fatalf("paginateItems beyond end returned %+v", beyondEnd)
	}

	if page, pageSize := normalizePagination(0, -3); page != 1 || pageSize != 1 {
		t.Fatalf("normalizePagination returned %d/%d", page, pageSize)
	}

	state := &PlayerStateModel{
		PlaylistID:    sql.NullInt64{Valid: true, Int64: 9},
		CurrentFileID: sql.NullInt64{Valid: true, Int64: 3},
	}
	continueIDs := buildContinueListeningTrackIDs([]int{2, 1, 4}, state, []PlaylistTrackModel{{FileID: 3}, {FileID: 2}})
	if len(continueIDs) != 4 || continueIDs[0] != 3 || continueIDs[1] != 2 || continueIDs[2] != 1 || continueIDs[3] != 4 {
		t.Fatalf("buildContinueListeningTrackIDs returned %+v", continueIDs)
	}
	if ids := buildContinueListeningTrackIDs([]int{1}, nil, nil); len(ids) != 0 {
		t.Fatalf("expected no continue-listening tracks without player state, got %+v", ids)
	}

	playlist := buildAutomaticPlaylistDto(AutoPlaylistFavoritesID, "PLAYLIST_NAME", "PLAYLIST_DESC", autoPlaylistFavoritesKey, 2)
	if !playlist.IsAuto || playlist.Kind != PlaylistKindAutomatic || playlist.TrackCount != 2 {
		t.Fatalf("buildAutomaticPlaylistDto returned %+v", playlist)
	}

	trackDto, err := fileModelToPlaylistTrackDto(musicFileModel(10, "Track", "/music"), 4)
	if err != nil {
		t.Fatalf("fileModelToPlaylistTrackDto returned error: %v", err)
	}
	if trackDto.Position != 4 || trackDto.File.ID != 10 {
		t.Fatalf("fileModelToPlaylistTrackDto returned %+v", trackDto)
	}
}

func TestCatalogServiceComposesPlaylistsHomeAndLibraryViews(t *testing.T) {
	fileModels := map[int]files.FileModel{
		1: musicFileModel(1, "Beta.mp3", "/music/a"),
		2: musicFileModel(2, "Alpha.mp3", "/music/a"),
		3: musicFileModel(3, "Gamma.mp3", "/music/b/live"),
	}
	requestedPages := []int{}
	idPage := func(ids ...int) utils.PaginationResponse[int] {
		return utils.PaginationResponse[int]{Items: ids, Pagination: utils.Pagination{Page: 1, PageSize: 10}}
	}

	repo := &musicRepoMock{
		getLibrarySummaryFn: func() (MusicLibrarySummaryDto, error) {
			return MusicLibrarySummaryDto{TotalTracks: 4, TotalArtists: 3, TotalAlbums: 3, TotalGenres: 3, TotalFolders: 3}, nil
		},
		getArtistGroupsFn: func(page int, pageSize int) (utils.PaginationResponse[MusicArtistGroupDto], error) {
			requestedPages = append(requestedPages, page, pageSize)
			return utils.PaginationResponse[MusicArtistGroupDto]{Items: []MusicArtistGroupDto{{Key: "artist a", Artist: "Artist A", TrackCount: 2, AlbumCount: 1}}}, nil
		},
		getAlbumGroupsFn: func(page int, pageSize int) (utils.PaginationResponse[MusicAlbumGroupDto], error) {
			return utils.PaginationResponse[MusicAlbumGroupDto]{Items: []MusicAlbumGroupDto{{Key: "artist a::album one", Album: "Album One"}}}, nil
		},
		getGenreGroupsFn: func(page int, pageSize int) (utils.PaginationResponse[MusicGenreGroupDto], error) {
			requestedPages = append(requestedPages, page, pageSize)
			return utils.PaginationResponse[MusicGenreGroupDto]{Items: []MusicGenreGroupDto{{Key: "hip hop", Genre: "Hip-Hop", TrackCount: 2}}}, nil
		},
		getFolderGroupsFn: func(page int, pageSize int) (utils.PaginationResponse[MusicFolderGroupDto], error) {
			return utils.PaginationResponse[MusicFolderGroupDto]{Items: []MusicFolderGroupDto{{Folder: "/music/a", TrackCount: 2}}}, nil
		},
		getTrackIDsByArtistFn: func(artistKey string, page int, pageSize int) (utils.PaginationResponse[int], error) {
			return idPage(2, 1), nil
		},
		getTrackIDsByAlbumFn: func(albumKey string, page int, pageSize int) (utils.PaginationResponse[int], error) {
			return idPage(2, 1), nil
		},
		getTrackIDsByGenreFn: func(genreKey string, page int, pageSize int) (utils.PaginationResponse[int], error) {
			return idPage(2, 3), nil
		},
		getTrackIDsByFolderFn: func(folderPath string, page int, pageSize int) (utils.PaginationResponse[int], error) {
			if folderPath != "/music/b" {
				t.Fatalf("folder path must be trimmed, got %q", folderPath)
			}
			return idPage(3), nil
		},
		getRecentFileIDsFn:   func(limit int) ([]int, error) { return []int{4, 3, 2, 1}, nil },
		getFavoriteFileIDsFn: func(limit int) ([]int, error) { return []int{3, 1}, nil },
		getPlayerStateFn: func(clientID string) (PlayerStateModel, error) {
			return PlayerStateModel{
				ClientID:      clientID,
				PlaylistID:    sql.NullInt64{Valid: true, Int64: 9},
				CurrentFileID: sql.NullInt64{Valid: true, Int64: 3},
			}, nil
		},
		getPlaylistTracksFn: func(playlistID int, page int, pageSize int) (utils.PaginationResponse[PlaylistTrackModel], error) {
			return utils.PaginationResponse[PlaylistTrackModel]{
				Items: []PlaylistTrackModel{{FileID: 3}, {FileID: 2}},
			}, nil
		},
		getLibraryFilesByIDsFn: func(fileIDs []int) ([]files.FileModel, error) {
			results := make([]files.FileModel, 0, len(fileIDs))
			for index := len(fileIDs) - 1; index >= 0; index-- {
				results = append(results, fileModels[fileIDs[index]])
			}
			return results, nil
		},
		getLibraryTracksFn: func(page int, pageSize int) (utils.PaginationResponse[files.FileModel], error) {
			return utils.PaginationResponse[files.FileModel]{
				Items:      []files.FileModel{fileModels[1], fileModels[2]},
				Pagination: utils.Pagination{Page: page, PageSize: pageSize},
			}, nil
		},
	}
	service := newMusicServiceForTest(t, repo)

	playlists, err := service.GetAutomaticPlaylists("client-1")
	if err != nil {
		t.Fatalf("GetAutomaticPlaylists returned error: %v", err)
	}
	if len(playlists) != 3 || playlists[0].ID != AutoPlaylistContinueListeningID || playlists[0].TrackCount != 4 || playlists[1].TrackCount != 4 || playlists[2].TrackCount != 2 {
		t.Fatalf("GetAutomaticPlaylists returned %+v", playlists)
	}

	home, err := service.GetHomeCatalog("client-1", 2)
	if err != nil {
		t.Fatalf("GetHomeCatalog returned error: %v", err)
	}
	if home.Summary.TotalTracks != 4 || home.Summary.TotalArtists != 3 || home.Summary.TotalAlbums != 3 || home.Summary.TotalGenres != 3 || home.Summary.TotalFolders != 3 {
		t.Fatalf("GetHomeCatalog summary returned %+v", home.Summary)
	}
	if len(home.Playlists) != 2 || len(home.Artists) != 1 || len(home.Albums) != 1 {
		t.Fatalf("GetHomeCatalog returned %+v", home)
	}
	if len(requestedPages) != 2 || requestedPages[0] != 1 || requestedPages[1] != 2 {
		t.Fatalf("home artists must be a first page limited by the home limit, got %v", requestedPages)
	}

	tracks, err := service.GetLibraryTracks(1, 10)
	if err != nil || len(tracks.Items) != 2 || tracks.Items[0].ID != 1 {
		t.Fatalf("GetLibraryTracks returned %+v err=%v", tracks, err)
	}

	artists, err := service.GetLibraryArtists(0, 0)
	if err != nil || len(artists.Items) != 1 || artists.Items[0].Artist != "Artist A" {
		t.Fatalf("GetLibraryArtists returned %+v err=%v", artists, err)
	}

	albums, err := service.GetLibraryAlbums(1, 10)
	if err != nil || len(albums.Items) != 1 || albums.Items[0].Album != "Album One" {
		t.Fatalf("GetLibraryAlbums returned %+v err=%v", albums, err)
	}

	genres, err := service.GetLibraryGenres(1, 10)
	if err != nil || len(genres.Items) != 1 || genres.Items[0].Genre != "Hip-Hop" {
		t.Fatalf("GetLibraryGenres returned %+v err=%v", genres, err)
	}

	folders, err := service.GetLibraryFolders(1, 10)
	if err != nil || len(folders.Items) != 1 || folders.Items[0].Folder != "/music/a" {
		t.Fatalf("GetLibraryFolders returned %+v err=%v", folders, err)
	}

	artistTracks, err := service.GetLibraryTracksByArtist("artist a", 1, 10)
	if err != nil || len(artistTracks.Items) != 2 || artistTracks.Items[0].ID != 2 || artistTracks.Items[1].ID != 1 {
		t.Fatalf("GetLibraryTracksByArtist returned %+v err=%v", artistTracks, err)
	}

	albumTracks, err := service.GetLibraryTracksByAlbum("artist a::album one", 1, 10)
	if err != nil || len(albumTracks.Items) != 2 || albumTracks.Items[0].ID != 2 {
		t.Fatalf("GetLibraryTracksByAlbum returned %+v err=%v", albumTracks, err)
	}

	genreTracks, err := service.GetLibraryTracksByGenre("hip hop", 1, 10)
	if err != nil || len(genreTracks.Items) != 2 || genreTracks.Items[0].ID != 2 || genreTracks.Items[1].ID != 3 {
		t.Fatalf("GetLibraryTracksByGenre returned %+v err=%v", genreTracks, err)
	}

	folderTracks, err := service.GetLibraryTracksByFolder(" /music/b ", 1, 10)
	if err != nil || len(folderTracks.Items) != 1 || folderTracks.Items[0].ID != 3 {
		t.Fatalf("GetLibraryTracksByFolder returned %+v err=%v", folderTracks, err)
	}

	emptyFolder, err := service.GetLibraryTracksByFolder(" ", 1, 10)
	if err != nil || len(emptyFolder.Items) != 0 {
		t.Fatalf("GetLibraryTracksByFolder blank returned %+v err=%v", emptyFolder, err)
	}

	loadedTracks, err := service.loadPlaylistTracksByIDs([]int{3, 2, 4}, 1, 2)
	if err != nil || len(loadedTracks.Items) != 2 || loadedTracks.Items[0].File.ID != 3 || loadedTracks.Items[0].Position != 1 {
		t.Fatalf("loadPlaylistTracksByIDs returned %+v err=%v", loadedTracks, err)
	}
}

func TestNormalizeGenreLabelAllBranches(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"r&b", "R&B"},
		{"r & b", "R&B"},
		{"rnb", "R&B"},
		{"rhythm and blues", "R&B"},
		{"r&b/soul", "R&B / Soul"},
		{"rnb/soul", "R&B / Soul"},
		{"rnb / soul", "R&B / Soul"},
		{"soul/r&b", "R&B / Soul"},
		{"hip hop", "Hip-Hop"},
		{"hiphop", "Hip-Hop"},
		{"hip-hop", "Hip-Hop"},
		{"lo fi", "Lo-Fi"},
		{"lofi", "Lo-Fi"},
		{"lo-fi", "Lo-Fi"},
		{"soundtrack", "Soundtrack"},
		{"ost", "Soundtrack"},
		{"", ""},
		{"  ", ""},
		{"rock", "Rock"},
		{"indie pop", "Indie Pop"},
		{"rock & roll", "Rock & Roll"},
	}

	for _, tc := range tests {
		t.Run("input="+tc.input, func(t *testing.T) {
			got := normalizeGenreLabel(tc.input)
			if got != tc.want {
				t.Fatalf("normalizeGenreLabel(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestCatalogServiceErrorBranchesAndFallbacks(t *testing.T) {
	errBoom := errors.New("boom")
	repo := &musicRepoMock{
		getLibrarySummaryFn:  func() (MusicLibrarySummaryDto, error) { return MusicLibrarySummaryDto{}, errBoom },
		getRecentFileIDsFn:   func(limit int) ([]int, error) { return nil, errBoom },
		getFavoriteFileIDsFn: func(limit int) ([]int, error) { return nil, errBoom },
		getArtistGroupsFn: func(page int, pageSize int) (utils.PaginationResponse[MusicArtistGroupDto], error) {
			return utils.PaginationResponse[MusicArtistGroupDto]{}, errBoom
		},
		getTrackIDsByArtistFn: func(artistKey string, page int, pageSize int) (utils.PaginationResponse[int], error) {
			return utils.PaginationResponse[int]{}, errBoom
		},
		getTrackIDsByFolderFn: func(folderPath string, page int, pageSize int) (utils.PaginationResponse[int], error) {
			return utils.PaginationResponse[int]{}, errBoom
		},
		getPlayerStateFn: func(clientID string) (PlayerStateModel, error) {
			return PlayerStateModel{}, errBoom
		},
		getPlaylistTracksFn: func(playlistID int, page int, pageSize int) (utils.PaginationResponse[PlaylistTrackModel], error) {
			return utils.PaginationResponse[PlaylistTrackModel]{}, errBoom
		},
		getLibraryFilesByIDsFn: func(fileIDs []int) ([]files.FileModel, error) {
			return nil, errBoom
		},
	}
	service := newMusicServiceForTest(t, repo)

	if _, err := service.GetAutomaticPlaylists("client-1"); !errors.Is(err, errBoom) {
		t.Fatalf("GetAutomaticPlaylists error = %v", err)
	}
	if _, err := service.GetHomeCatalog("client-1", 0); !errors.Is(err, errBoom) {
		t.Fatalf("GetHomeCatalog error = %v", err)
	}
	if _, err := service.GetLibraryArtists(1, 10); !errors.Is(err, errBoom) {
		t.Fatalf("GetLibraryArtists error = %v", err)
	}
	if _, err := service.GetLibraryTracksByArtist("a", 1, 10); !errors.Is(err, errBoom) {
		t.Fatalf("GetLibraryTracksByArtist error = %v", err)
	}
	if _, err := service.GetLibraryTracksByFolder("/m", 1, 10); !errors.Is(err, errBoom) {
		t.Fatalf("GetLibraryTracksByFolder error = %v", err)
	}
	if state := service.getOptionalPlayerState("client-1"); state != nil {
		t.Fatalf("getOptionalPlayerState returned %+v", state)
	}
	if tracks := service.getContinueListeningSourceTracks(&PlayerStateModel{PlaylistID: sql.NullInt64{Valid: true, Int64: 10}}); tracks != nil {
		t.Fatalf("getContinueListeningSourceTracks returned %+v", tracks)
	}
	if _, err := service.automaticPlaylistTrackIDs("client-1", 99); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("automaticPlaylistTrackIDs error = %v", err)
	}
	if _, err := service.automaticPlaylistTrackIDs("client-1", AutoPlaylistContinueListeningID); !errors.Is(err, errBoom) {
		t.Fatalf("continue listening error = %v", err)
	}
	if _, err := service.loadPlaylistTracksByIDs([]int{1, 2}, 1, 10); !errors.Is(err, errBoom) {
		t.Fatalf("loadPlaylistTracksByIDs error = %v", err)
	}
	if _, err := service.loadLibraryTracksOfIDPage(utils.PaginationResponse[int]{Items: []int{1}}); !errors.Is(err, errBoom) {
		t.Fatalf("loadLibraryTracksOfIDPage error = %v", err)
	}
}
