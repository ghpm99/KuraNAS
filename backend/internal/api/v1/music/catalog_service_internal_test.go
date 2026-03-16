package music

import (
	"database/sql"
	"errors"
	files "nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/utils"
	"testing"
	"time"
)

func catalogEntry(
	fileID int,
	title string,
	artist string,
	album string,
	genre string,
	parentPath string,
	trackNumber string,
	createdAt time.Time,
	updatedAt time.Time,
	lastInteraction sql.NullTime,
	starred bool,
) MusicLibraryIndexEntryModel {
	return MusicLibraryIndexEntryModel{
		FileID:          fileID,
		FileName:        title + ".mp3",
		FilePath:        parentPath + "/" + title + ".mp3",
		ParentPath:      parentPath,
		Starred:         starred,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
		LastInteraction: lastInteraction,
		Title:           title,
		Artist:          artist,
		Album:           album,
		Genre:           genre,
		TrackNumber:     trackNumber,
	}
}

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
		Metadata: files.AudioMetadataModel{
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

func TestCatalogHelpersNormalizeAggregateAndSort(t *testing.T) {
	base := time.Date(2026, time.March, 10, 8, 0, 0, 0, time.UTC)
	entries := []MusicLibraryIndexEntryModel{
		catalogEntry(1, "Beta", "Artist A", "Album One", "Lo-Fi", "/music/a", "2/12", base, base.Add(time.Hour), sql.NullTime{}, true),
		catalogEntry(2, "Alpha", "Artist A", "Album One", "Hip Hop", "/music/a", "1/12", base.Add(2*time.Hour), base.Add(2*time.Hour), sql.NullTime{Valid: true, Time: base.Add(5 * time.Hour)}, false),
		catalogEntry(3, "Gamma", "Artist B", "Album Two", "Hip Hop; Soundtrack", "/music/b/live", "3", base.Add(3*time.Hour), base.Add(3*time.Hour), sql.NullTime{Valid: true, Time: base.Add(4 * time.Hour)}, true),
		catalogEntry(4, "Delta", "Artist C", "Album Three", "", "", "", base.Add(4*time.Hour), base.Add(4*time.Hour), sql.NullTime{Valid: true, Time: base.Add(6 * time.Hour)}, false),
	}
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
	withAlbumArtist := entries[0]
	withAlbumArtist.AlbumArtist = " Album Artist "
	if got := preferredArtist(withAlbumArtist); got != "Album Artist" {
		t.Fatalf("preferredArtist returned %q", got)
	}
	if got := entryTimestamp(entries[1]); !got.Equal(base.Add(5 * time.Hour)) {
		t.Fatalf("entryTimestamp returned %v", got)
	}
	if got := parseTrackNumber("8/12"); got != 8 {
		t.Fatalf("parseTrackNumber returned %d", got)
	}

	paginated := paginateItems([]int{1, 2, 3}, 2, 2)
	if len(paginated.Items) != 1 || paginated.Items[0] != 3 || paginated.Pagination.HasPrev != true {
		t.Fatalf("paginateItems returned %+v", paginated)
	}

	artists := buildArtistGroups(entries)
	if len(artists) != 3 || artists[0].Artist != "Artist A" || artists[0].TrackCount != 2 || artists[0].AlbumCount != 1 {
		t.Fatalf("buildArtistGroups returned %+v", artists)
	}

	albums := buildAlbumGroups(entries)
	if len(albums) != 3 || albums[0].Album != "Album One" || albums[0].TrackCount != 2 {
		t.Fatalf("buildAlbumGroups returned %+v", albums)
	}

	genres := buildGenreGroups(entries)
	if len(genres) != 3 || genres[0].Genre != "Hip-Hop" || genres[0].TrackCount != 2 {
		t.Fatalf("buildGenreGroups returned %+v", genres)
	}

	folders := buildFolderGroups(entries)
	if len(folders) != 3 || folders[0].Folder != "/music/a" || folders[1].Folder != "/" || folders[2].Folder != "/music/b/live" {
		t.Fatalf("buildFolderGroups returned %+v", folders)
	}

	artistOrdered := append([]MusicLibraryIndexEntryModel(nil), entries[:3]...)
	sortArtistTracks(artistOrdered)
	if artistOrdered[0].FileID != 2 || artistOrdered[1].FileID != 1 {
		t.Fatalf("sortArtistTracks returned %+v", artistOrdered)
	}

	albumOrdered := append([]MusicLibraryIndexEntryModel(nil), entries[:2]...)
	sortAlbumTracks(albumOrdered)
	if albumOrdered[0].FileID != 2 || albumOrdered[1].FileID != 1 {
		t.Fatalf("sortAlbumTracks returned %+v", albumOrdered)
	}

	genreOrdered := append([]MusicLibraryIndexEntryModel(nil), entries[1:3]...)
	sortGenreTracks(genreOrdered)
	if genreOrdered[0].FileID != 2 || genreOrdered[1].FileID != 3 {
		t.Fatalf("sortGenreTracks returned %+v", genreOrdered)
	}

	if ids := limitFileIDs([]MusicLibraryIndexEntryModel{entries[1], entries[1], entries[2]}, 2); len(ids) != 2 || ids[0] != 2 || ids[1] != 3 {
		t.Fatalf("limitFileIDs returned %+v", ids)
	}

	recentIDs := buildRecentPlaylistTrackIDs(entries)
	if len(recentIDs) != 4 || recentIDs[0] != 4 || recentIDs[1] != 2 {
		t.Fatalf("buildRecentPlaylistTrackIDs returned %+v", recentIDs)
	}

	favoriteIDs := buildFavoritePlaylistTrackIDs(entries)
	if len(favoriteIDs) != 2 || favoriteIDs[0] != 3 || favoriteIDs[1] != 1 {
		t.Fatalf("buildFavoritePlaylistTrackIDs returned %+v", favoriteIDs)
	}

	state := &PlayerStateModel{
		PlaylistID:    sql.NullInt64{Valid: true, Int64: 9},
		CurrentFileID: sql.NullInt64{Valid: true, Int64: 3},
	}
	continueIDs := buildContinueListeningTrackIDs(entries, state, []PlaylistTrackModel{{FileID: 3}, {FileID: 2}})
	if len(continueIDs) != 3 || continueIDs[0] != 3 || continueIDs[1] != 2 || continueIDs[2] != 4 {
		t.Fatalf("buildContinueListeningTrackIDs returned %+v", continueIDs)
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

func TestCatalogServiceBuildsPlaylistsAndLibraryViews(t *testing.T) {
	base := time.Date(2026, time.March, 10, 8, 0, 0, 0, time.UTC)
	entries := []MusicLibraryIndexEntryModel{
		catalogEntry(1, "Beta", "Artist A", "Album One", "Lo-Fi", "/music/a", "2/12", base, base.Add(time.Hour), sql.NullTime{}, true),
		catalogEntry(2, "Alpha", "Artist A", "Album One", "Hip Hop", "/music/a", "1/12", base.Add(2*time.Hour), base.Add(2*time.Hour), sql.NullTime{Valid: true, Time: base.Add(5 * time.Hour)}, false),
		catalogEntry(3, "Gamma", "Artist B", "Album Two", "Hip Hop; Soundtrack", "/music/b/live", "3", base.Add(3*time.Hour), base.Add(3*time.Hour), sql.NullTime{Valid: true, Time: base.Add(4 * time.Hour)}, true),
		catalogEntry(4, "Delta", "Artist C", "Album Three", "", "", "", base.Add(4*time.Hour), base.Add(4*time.Hour), sql.NullTime{Valid: true, Time: base.Add(6 * time.Hour)}, false),
	}

	fileModels := map[int]files.FileModel{
		1: musicFileModel(1, "Beta.mp3", "/music/a"),
		2: musicFileModel(2, "Alpha.mp3", "/music/a"),
		3: musicFileModel(3, "Gamma.mp3", "/music/b/live"),
		4: musicFileModel(4, "Delta.mp3", "/music"),
	}

	repo := &musicRepoMock{
		getLibraryIndexFn: func() ([]MusicLibraryIndexEntryModel, error) {
			return entries, nil
		},
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
				Items: []files.FileModel{fileModels[1], fileModels[2]},
				Pagination: utils.Pagination{
					Page:     page,
					PageSize: pageSize,
					HasNext:  false,
					HasPrev:  false,
				},
			}, nil
		},
	}
	service := newMusicServiceForTest(t, repo)

	playlists, err := service.GetAutomaticPlaylists("client-1")
	if err != nil {
		t.Fatalf("GetAutomaticPlaylists returned error: %v", err)
	}
	if len(playlists) != 3 || playlists[0].ID != AutoPlaylistContinueListeningID || playlists[0].TrackCount != 3 || playlists[1].TrackCount != 4 || playlists[2].TrackCount != 2 {
		t.Fatalf("GetAutomaticPlaylists returned %+v", playlists)
	}

	home, err := service.GetHomeCatalog("client-1", 2)
	if err != nil {
		t.Fatalf("GetHomeCatalog returned error: %v", err)
	}
	if home.Summary.TotalTracks != 4 || home.Summary.TotalArtists != 3 || home.Summary.TotalAlbums != 3 || home.Summary.TotalGenres != 3 || home.Summary.TotalFolders != 3 {
		t.Fatalf("GetHomeCatalog summary returned %+v", home.Summary)
	}
	if len(home.Playlists) != 2 || len(home.Artists) != 2 || len(home.Albums) != 2 {
		t.Fatalf("GetHomeCatalog returned %+v", home)
	}

	tracks, err := service.GetLibraryTracks(1, 10)
	if err != nil || len(tracks.Items) != 2 || tracks.Items[0].ID != 1 {
		t.Fatalf("GetLibraryTracks returned %+v err=%v", tracks, err)
	}

	artists, err := service.GetLibraryArtists(1, 10)
	if err != nil || len(artists.Items) != 3 || artists.Items[0].Artist != "Artist A" {
		t.Fatalf("GetLibraryArtists returned %+v err=%v", artists, err)
	}

	albums, err := service.GetLibraryAlbums(1, 10)
	if err != nil || len(albums.Items) != 3 || albums.Items[0].Album != "Album One" {
		t.Fatalf("GetLibraryAlbums returned %+v err=%v", albums, err)
	}

	genres, err := service.GetLibraryGenres(1, 10)
	if err != nil || len(genres.Items) != 3 || genres.Items[0].Genre != "Hip-Hop" {
		t.Fatalf("GetLibraryGenres returned %+v err=%v", genres, err)
	}

	folders, err := service.GetLibraryFolders(1, 10)
	if err != nil || len(folders.Items) != 3 || folders.Items[0].Folder != "/music/a" {
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

	folderTracks, err := service.GetLibraryTracksByFolder("/music/b", 1, 10)
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

func TestCatalogServiceErrorBranchesAndFallbacks(t *testing.T) {
	errBoom := errors.New("boom")
	repo := &musicRepoMock{
		getLibraryIndexFn: func() ([]MusicLibraryIndexEntryModel, error) {
			return nil, errBoom
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
	if state := service.getOptionalPlayerState("client-1"); state != nil {
		t.Fatalf("getOptionalPlayerState returned %+v", state)
	}
	if tracks := service.getContinueListeningSourceTracks(&PlayerStateModel{PlaylistID: sql.NullInt64{Valid: true, Int64: 10}}); tracks != nil {
		t.Fatalf("getContinueListeningSourceTracks returned %+v", tracks)
	}
	if _, err := service.automaticPlaylistTrackIDs("client-1", 99, []MusicLibraryIndexEntryModel{}); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("automaticPlaylistTrackIDs error = %v", err)
	}
	if _, err := service.loadPlaylistTracksByIDs([]int{1, 2}, 1, 10); !errors.Is(err, errBoom) {
		t.Fatalf("loadPlaylistTracksByIDs error = %v", err)
	}
}
