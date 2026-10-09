package music

import (
	"database/sql"
	"fmt"
	"reflect"
	"testing"
	"time"

	files "nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/testutil"
	"nas-go/api/pkg/database"
)

type catalogTrackFixture struct {
	name       string
	parentPath string
	format     string
	updatedAt  time.Time
	isStarred  bool
	isDeleted  bool
	metadata   *AudioMetadataModel
}

type catalogTestEnvironment struct {
	dbContext         *database.DbContext
	service           *Service
	audioRepository   *AudioMetadataRepository
	fileIDsByName     map[string]int
	namesByFileID     map[int]string
	catalogBaseMoment time.Time
}

func newCatalogTestEnvironment(t *testing.T) *catalogTestEnvironment {
	t.Helper()
	dbContext := testutil.NewPostgresDB(t, "kuranas_music_it")

	truncateErr := dbContext.ExecTx(func(tx *sql.Tx) error {
		_, err := tx.Exec(`TRUNCATE playlist_track, playlist, player_state, audio_metadata, home_file RESTART IDENTITY CASCADE`)
		return err
	})
	if truncateErr != nil {
		t.Fatalf("truncate: %v", truncateErr)
	}

	return &catalogTestEnvironment{
		dbContext:         dbContext,
		service:           &Service{Repository: NewRepository(dbContext)},
		audioRepository:   NewAudioMetadataRepository(dbContext),
		fileIDsByName:     map[string]int{},
		namesByFileID:     map[int]string{},
		catalogBaseMoment: time.Date(2026, time.March, 10, 8, 0, 0, 0, time.UTC),
	}
}

func (environment *catalogTestEnvironment) seedTracks(t *testing.T, fixtures []catalogTrackFixture) {
	t.Helper()

	seedErr := environment.dbContext.ExecTx(func(tx *sql.Tx) error {
		for _, fixture := range fixtures {
			format := fixture.format
			if format == "" {
				format = ".mp3"
			}
			path := fixture.parentPath + "/" + fixture.name
			var deletedAt any
			if fixture.isDeleted {
				deletedAt = environment.catalogBaseMoment
			}

			var fileID int
			err := tx.QueryRow(`INSERT INTO home_file (name, path, parent_path, format, size, updated_at, created_at, type, checksum, deleted_at, starred)
				VALUES ($1, $2, $3, $4, 1, $5, $5, 2, '', $6, $7) RETURNING id`,
				fixture.name, path, fixture.parentPath, format, fixture.updatedAt, deletedAt, fixture.isStarred).Scan(&fileID)
			if err != nil {
				return err
			}
			environment.fileIDsByName[fixture.name] = fileID
			environment.namesByFileID[fileID] = fixture.name

			if fixture.metadata == nil {
				continue
			}
			metadata := *fixture.metadata
			metadata.FileId = fileID
			metadata.Path = path
			if _, err := environment.audioRepository.UpsertAudioMetadata(tx, metadata); err != nil {
				return err
			}
		}
		return nil
	})
	if seedErr != nil {
		t.Fatalf("seed tracks: %v", seedErr)
	}
}

func (environment *catalogTestEnvironment) namesOf(fileIDs []int) []string {
	names := make([]string, 0, len(fileIDs))
	for _, fileID := range fileIDs {
		names = append(names, environment.namesByFileID[fileID])
	}
	return names
}

func canonicalCatalogFixture(base time.Time) []catalogTrackFixture {
	return []catalogTrackFixture{
		{name: "Beta.mp3", parentPath: "/music/a", updatedAt: base.Add(time.Hour), isStarred: true,
			metadata: &AudioMetadataModel{Title: "Beta", Artist: "Artist A", Album: "Album One", Genre: "Lo-Fi", TrackNumber: "2/12", Year: "2019"}},
		{name: "Alpha.mp3", parentPath: "/music/a", updatedAt: base.Add(2 * time.Hour),
			metadata: &AudioMetadataModel{Title: "Alpha", Artist: "Artist A", Album: "Album One", Genre: "Hip Hop", TrackNumber: "1/12", Year: ""}},
		{name: "Gamma.mp3", parentPath: "/music/b/live", updatedAt: base.Add(3 * time.Hour), isStarred: true,
			metadata: &AudioMetadataModel{Title: "Gamma", Artist: "Artist B", Album: "Album Two", Genre: "Hip Hop; Soundtrack", TrackNumber: "3"}},
		{name: "Delta.mp3", parentPath: "", updatedAt: base.Add(4 * time.Hour),
			metadata: &AudioMetadataModel{Title: "Delta", Artist: "Artist C", Album: "Album Three"}},
		{name: "deleted.mp3", parentPath: "/music/a", updatedAt: base.Add(9 * time.Hour), isDeleted: true,
			metadata: &AudioMetadataModel{Title: "Gone", Artist: "Artist A", Album: "Album One", Genre: "Lo-Fi"}},
		{name: "cover.jpg", parentPath: "/music/a", format: ".jpg", updatedAt: base.Add(9 * time.Hour)},
	}
}

func TestUpsertAudioMetadataPersistsCatalogGroupingKeys_Postgres(t *testing.T) {
	environment := newCatalogTestEnvironment(t)
	environment.seedTracks(t, []catalogTrackFixture{
		{name: "full.mp3", parentPath: "/m", updatedAt: environment.catalogBaseMoment,
			metadata: &AudioMetadataModel{Artist: "Guest", AlbumArtist: " The  Band ", Album: "Live  Album", Genre: "rnb; Hip-Hop"}},
		{name: "empty.mp3", parentPath: "/m", updatedAt: environment.catalogBaseMoment,
			metadata: &AudioMetadataModel{}},
	})

	readKeys := func(fileName string) CatalogGroupingKeys {
		var keys CatalogGroupingKeys
		err := environment.dbContext.QueryTx(func(tx *sql.Tx) error {
			var genreKeys, genreLabels sql.NullString
			scanErr := tx.QueryRow(`SELECT catalog_artist_key, catalog_artist_label, catalog_album_key, catalog_album_label,
				catalog_genre_keys::text, catalog_genre_labels::text FROM audio_metadata WHERE file_id = $1`,
				environment.fileIDsByName[fileName]).
				Scan(&keys.ArtistKey, &keys.ArtistLabel, &keys.AlbumKey, &keys.AlbumLabel, &genreKeys, &genreLabels)
			if scanErr != nil {
				return scanErr
			}
			if !genreKeys.Valid || !genreLabels.Valid {
				t.Fatalf("genre arrays must be non-null once keys are computed")
			}
			keys.GenreKeys = []string{genreKeys.String}
			keys.GenreLabels = []string{genreLabels.String}
			return nil
		})
		if err != nil {
			t.Fatalf("read keys of %s: %v", fileName, err)
		}
		return keys
	}

	fullKeys := readKeys("full.mp3")
	if fullKeys.ArtistKey != "the band" || fullKeys.ArtistLabel != "The Band" ||
		fullKeys.AlbumKey != "the band::live album" || fullKeys.AlbumLabel != "Live Album" ||
		fullKeys.GenreKeys[0] != `{r&b,"hip hop"}` || fullKeys.GenreLabels[0] != `{R&B,Hip-Hop}` {
		t.Fatalf("unexpected persisted keys: %+v", fullKeys)
	}

	emptyKeys := readKeys("empty.mp3")
	if emptyKeys.ArtistKey != "" || emptyKeys.AlbumKey != "" || emptyKeys.GenreKeys[0] != "{}" {
		t.Fatalf("unexpected persisted keys for untagged track: %+v", emptyKeys)
	}
}

func TestCatalogGroupsAndTracksKeepGroupingSemantics_Postgres(t *testing.T) {
	environment := newCatalogTestEnvironment(t)
	environment.seedTracks(t, canonicalCatalogFixture(environment.catalogBaseMoment))
	service := environment.service

	artists, err := service.GetLibraryArtists(1, 10, DefaultCatalogSort())
	if err != nil {
		t.Fatalf("GetLibraryArtists: %v", err)
	}
	expectedArtists := []MusicArtistGroupDto{
		{Key: "artist a", Artist: "Artist A", TrackCount: 2, AlbumCount: 1},
		{Key: "artist b", Artist: "Artist B", TrackCount: 1, AlbumCount: 1},
		{Key: "artist c", Artist: "Artist C", TrackCount: 1, AlbumCount: 1},
	}
	if !reflect.DeepEqual(artists.Items, expectedArtists) || artists.Pagination.HasNext {
		t.Fatalf("GetLibraryArtists returned %+v", artists)
	}

	albums, err := service.GetLibraryAlbums(1, 10, DefaultCatalogSort())
	if err != nil {
		t.Fatalf("GetLibraryAlbums: %v", err)
	}
	expectedAlbums := []MusicAlbumGroupDto{
		{Key: "artist a::album one", Album: "Album One", Artist: "Artist A", Year: "2019", TrackCount: 2},
		{Key: "artist b::album two", Album: "Album Two", Artist: "Artist B", Year: "", TrackCount: 1},
		{Key: "artist c::album three", Album: "Album Three", Artist: "Artist C", Year: "", TrackCount: 1},
	}
	if !reflect.DeepEqual(albums.Items, expectedAlbums) {
		t.Fatalf("GetLibraryAlbums returned %+v", albums.Items)
	}

	genres, err := service.GetLibraryGenres(1, 10, DefaultCatalogSort())
	if err != nil {
		t.Fatalf("GetLibraryGenres: %v", err)
	}
	expectedGenres := []MusicGenreGroupDto{
		{Key: "hip hop", Genre: "Hip-Hop", TrackCount: 2},
		{Key: "lo fi", Genre: "Lo-Fi", TrackCount: 1},
		{Key: "soundtrack", Genre: "Soundtrack", TrackCount: 1},
	}
	if !reflect.DeepEqual(genres.Items, expectedGenres) {
		t.Fatalf("GetLibraryGenres returned %+v", genres.Items)
	}

	folders, err := service.GetLibraryFolders(1, 10, DefaultCatalogSort())
	if err != nil {
		t.Fatalf("GetLibraryFolders: %v", err)
	}
	expectedFolders := []MusicFolderGroupDto{
		{Folder: "/music/a", TrackCount: 2},
		{Folder: "/", TrackCount: 1},
		{Folder: "/music/b/live", TrackCount: 1},
	}
	if !reflect.DeepEqual(folders.Items, expectedFolders) {
		t.Fatalf("GetLibraryFolders returned %+v", folders.Items)
	}

	artistTracks, err := service.GetLibraryTracksByArtist("artist a", 1, 10)
	if err != nil {
		t.Fatalf("GetLibraryTracksByArtist: %v", err)
	}
	if got := trackNames(artistTracks.Items); !reflect.DeepEqual(got, []string{"Alpha.mp3", "Beta.mp3"}) {
		t.Fatalf("tracks by artist returned %v", got)
	}

	albumTracks, err := service.GetLibraryTracksByAlbum("artist a::album one", 1, 10)
	if err != nil {
		t.Fatalf("GetLibraryTracksByAlbum: %v", err)
	}
	if got := trackNames(albumTracks.Items); !reflect.DeepEqual(got, []string{"Alpha.mp3", "Beta.mp3"}) {
		t.Fatalf("tracks by album returned %v", got)
	}

	genreTracks, err := service.GetLibraryTracksByGenre("hip hop", 1, 10)
	if err != nil {
		t.Fatalf("GetLibraryTracksByGenre: %v", err)
	}
	if got := trackNames(genreTracks.Items); !reflect.DeepEqual(got, []string{"Alpha.mp3", "Gamma.mp3"}) {
		t.Fatalf("tracks by genre returned %v", got)
	}

	folderTracks, err := service.GetLibraryTracksByFolder("/music/b", 1, 10)
	if err != nil {
		t.Fatalf("GetLibraryTracksByFolder: %v", err)
	}
	if got := trackNames(folderTracks.Items); !reflect.DeepEqual(got, []string{"Gamma.mp3"}) {
		t.Fatalf("tracks by folder returned %v", got)
	}

	rootFolderTracks, err := service.GetLibraryTracksByFolder("/", 1, 10)
	if err != nil {
		t.Fatalf("GetLibraryTracksByFolder root: %v", err)
	}
	if got := trackNames(rootFolderTracks.Items); !reflect.DeepEqual(got, []string{"Alpha.mp3", "Beta.mp3", "Gamma.mp3"}) {
		t.Fatalf("tracks by root folder returned %v", got)
	}

	emptyFolderTracks, err := service.GetLibraryTracksByFolder("  ", 1, 10)
	if err != nil || len(emptyFolderTracks.Items) != 0 {
		t.Fatalf("blank folder must yield no tracks, got %+v err=%v", emptyFolderTracks, err)
	}
}

func trackNames(tracks []files.FileDto) []string {
	names := make([]string, 0, len(tracks))
	for _, track := range tracks {
		names = append(names, track.Name)
	}
	return names
}

func TestHomeCatalogComposesSummaryAndFirstPages_Postgres(t *testing.T) {
	environment := newCatalogTestEnvironment(t)
	environment.seedTracks(t, canonicalCatalogFixture(environment.catalogBaseMoment))

	home, err := environment.service.GetHomeCatalog("client-1", 2, DefaultCatalogSort())
	if err != nil {
		t.Fatalf("GetHomeCatalog: %v", err)
	}

	expectedSummary := MusicLibrarySummaryDto{TotalTracks: 4, TotalArtists: 3, TotalAlbums: 3, TotalGenres: 3, TotalFolders: 3}
	if home.Summary != expectedSummary {
		t.Fatalf("summary = %+v, want %+v", home.Summary, expectedSummary)
	}
	if len(home.Playlists) != 2 || len(home.Artists) != 2 || len(home.Albums) != 2 {
		t.Fatalf("home catalog must be limited to 2 entries per section: %+v", home)
	}
	if home.Artists[0].Artist != "Artist A" || home.Albums[0].Album != "Album One" {
		t.Fatalf("home catalog leading entries wrong: %+v", home)
	}
}

func TestAutomaticPlaylistsAreBoundedAndOrdered_Postgres(t *testing.T) {
	environment := newCatalogTestEnvironment(t)
	environment.seedTracks(t, canonicalCatalogFixture(environment.catalogBaseMoment))

	playlists, err := environment.service.GetAutomaticPlaylists("client-1")
	if err != nil {
		t.Fatalf("GetAutomaticPlaylists: %v", err)
	}
	if len(playlists) != 3 || playlists[0].TrackCount != 0 || playlists[1].TrackCount != 4 || playlists[2].TrackCount != 2 {
		t.Fatalf("unexpected automatic playlists: %+v", playlists)
	}

	recentIDs, err := environment.service.automaticPlaylistTrackIDs("client-1", AutoPlaylistRecentlyAddedID)
	if err != nil {
		t.Fatalf("recent: %v", err)
	}
	if got := environment.namesOf(recentIDs); !reflect.DeepEqual(got, []string{"Delta.mp3", "Gamma.mp3", "Alpha.mp3", "Beta.mp3"}) {
		t.Fatalf("recent order = %v", got)
	}

	favoriteIDs, err := environment.service.automaticPlaylistTrackIDs("client-1", AutoPlaylistFavoritesID)
	if err != nil {
		t.Fatalf("favorites: %v", err)
	}
	if got := environment.namesOf(favoriteIDs); !reflect.DeepEqual(got, []string{"Gamma.mp3", "Beta.mp3"}) {
		t.Fatalf("favorites order = %v", got)
	}

	limitedRecentIDs, err := NewRepository(environment.dbContext).GetRecentLibraryFileIDs(2)
	if err != nil || len(limitedRecentIDs) != 2 {
		t.Fatalf("recent query must honour its limit, got %v err=%v", limitedRecentIDs, err)
	}
}

func TestContinueListeningStartsFromPlayerStateAndFillsWithRecent_Postgres(t *testing.T) {
	environment := newCatalogTestEnvironment(t)
	environment.seedTracks(t, canonicalCatalogFixture(environment.catalogBaseMoment))

	currentFileID := environment.fileIDsByName["Gamma.mp3"]
	upsertErr := environment.dbContext.ExecTx(func(tx *sql.Tx) error {
		_, err := NewRepository(environment.dbContext).UpsertPlayerState(tx, PlayerStateModel{
			ClientID:      "client-1",
			CurrentFileID: sql.NullInt64{Valid: true, Int64: int64(currentFileID)},
			RepeatMode:    "off",
			Volume:        1,
		})
		return err
	})
	if upsertErr != nil {
		t.Fatalf("player state: %v", upsertErr)
	}

	continueIDs, err := environment.service.automaticPlaylistTrackIDs("client-1", AutoPlaylistContinueListeningID)
	if err != nil {
		t.Fatalf("continue listening: %v", err)
	}
	if got := environment.namesOf(continueIDs); !reflect.DeepEqual(got, []string{"Gamma.mp3", "Delta.mp3", "Alpha.mp3", "Beta.mp3"}) {
		t.Fatalf("continue listening order = %v", got)
	}
}

func TestCatalogPaginationReportsHasNextAndPages_Postgres(t *testing.T) {
	environment := newCatalogTestEnvironment(t)

	fixtures := []catalogTrackFixture{}
	for artistIndex := 1; artistIndex <= 7; artistIndex++ {
		trackCount := 8 - artistIndex
		for trackIndex := 1; trackIndex <= trackCount; trackIndex++ {
			fixtures = append(fixtures, catalogTrackFixture{
				name:       fmt.Sprintf("a%d_t%d.mp3", artistIndex, trackIndex),
				parentPath: "/music/many",
				updatedAt:  environment.catalogBaseMoment.Add(time.Duration(artistIndex*10+trackIndex) * time.Minute),
				metadata: &AudioMetadataModel{
					Title:       fmt.Sprintf("T%02d", trackIndex),
					Artist:      fmt.Sprintf("Artist %d", artistIndex),
					Album:       "Shared Album",
					Genre:       "Rock",
					TrackNumber: fmt.Sprintf("%d", trackIndex),
				},
			})
		}
	}
	environment.seedTracks(t, fixtures)

	firstPage, err := environment.service.GetLibraryArtists(1, 3, DefaultCatalogSort())
	if err != nil {
		t.Fatalf("first page: %v", err)
	}
	if len(firstPage.Items) != 3 || !firstPage.Pagination.HasNext || firstPage.Pagination.HasPrev || firstPage.Items[0].Artist != "Artist 1" || firstPage.Items[0].TrackCount != 7 {
		t.Fatalf("unexpected first page: %+v", firstPage)
	}

	secondPage, err := environment.service.GetLibraryArtists(2, 3, DefaultCatalogSort())
	if err != nil {
		t.Fatalf("second page: %v", err)
	}
	if len(secondPage.Items) != 3 || !secondPage.Pagination.HasNext || !secondPage.Pagination.HasPrev || secondPage.Items[0].Artist != "Artist 4" {
		t.Fatalf("unexpected second page: %+v", secondPage)
	}

	lastPage, err := environment.service.GetLibraryArtists(3, 3, DefaultCatalogSort())
	if err != nil {
		t.Fatalf("last page: %v", err)
	}
	if len(lastPage.Items) != 1 || lastPage.Pagination.HasNext || lastPage.Items[0].Artist != "Artist 7" {
		t.Fatalf("unexpected last page: %+v", lastPage)
	}

	beyondPage, err := environment.service.GetLibraryArtists(9, 3, DefaultCatalogSort())
	if err != nil {
		t.Fatalf("beyond page: %v", err)
	}
	if len(beyondPage.Items) != 0 || beyondPage.Pagination.HasNext || !beyondPage.Pagination.HasPrev {
		t.Fatalf("unexpected page beyond the end: %+v", beyondPage)
	}

	albums, err := environment.service.GetLibraryAlbums(1, 2, DefaultCatalogSort())
	if err != nil || len(albums.Items) != 2 || !albums.Pagination.HasNext {
		t.Fatalf("albums pagination: %+v err=%v", albums, err)
	}

	tracksPageOne, err := environment.service.GetLibraryTracksByGenre("rock", 1, 10)
	if err != nil || len(tracksPageOne.Items) != 10 || !tracksPageOne.Pagination.HasNext {
		t.Fatalf("tracks by genre page 1: len=%d next=%v err=%v", len(tracksPageOne.Items), tracksPageOne.Pagination.HasNext, err)
	}
	tracksPageThree, err := environment.service.GetLibraryTracksByGenre("rock", 3, 10)
	if err != nil || len(tracksPageThree.Items) != 8 || tracksPageThree.Pagination.HasNext {
		t.Fatalf("tracks by genre page 3: len=%d next=%v err=%v", len(tracksPageThree.Items), tracksPageThree.Pagination.HasNext, err)
	}

	artistTracks, err := environment.service.GetLibraryTracksByArtist("artist 1", 2, 3)
	if err != nil {
		t.Fatalf("tracks by artist: %v", err)
	}
	if got := trackNames(artistTracks.Items); !reflect.DeepEqual(got, []string{"a1_t4.mp3", "a1_t5.mp3", "a1_t6.mp3"}) || !artistTracks.Pagination.HasNext {
		t.Fatalf("artist tracks second page = %v next=%v", got, artistTracks.Pagination.HasNext)
	}
}

func TestFolderTracksMatchSubfoldersOnlyAndTreatWildcardsLiterally_Postgres(t *testing.T) {
	environment := newCatalogTestEnvironment(t)
	environment.seedTracks(t, []catalogTrackFixture{
		{name: "direct.mp3", parentPath: "/music/rock", updatedAt: environment.catalogBaseMoment},
		{name: "nested.mp3", parentPath: "/music/rock/live", updatedAt: environment.catalogBaseMoment},
		{name: "sibling.mp3", parentPath: "/music/rockabilly", updatedAt: environment.catalogBaseMoment},
		{name: "percent.mp3", parentPath: "/music/100%", updatedAt: environment.catalogBaseMoment},
		{name: "percent_other.mp3", parentPath: "/music/1000", updatedAt: environment.catalogBaseMoment},
		{name: "underscore.mp3", parentPath: "/music/a_b", updatedAt: environment.catalogBaseMoment},
		{name: "underscore_other.mp3", parentPath: "/music/aXb", updatedAt: environment.catalogBaseMoment},
	})

	cases := []struct {
		folder   string
		expected []string
	}{
		{"/music/rock", []string{"direct.mp3", "nested.mp3"}},
		{"/music/rock/", []string{"nested.mp3"}},
		{"/music/100%", []string{"percent.mp3"}},
		{"/music/a_b", []string{"underscore.mp3"}},
	}
	for _, testCase := range cases {
		tracks, err := environment.service.GetLibraryTracksByFolder(testCase.folder, 1, 10)
		if err != nil {
			t.Fatalf("folder %q: %v", testCase.folder, err)
		}
		if got := trackNames(tracks.Items); !reflect.DeepEqual(got, testCase.expected) {
			t.Fatalf("folder %q returned %v, want %v", testCase.folder, got, testCase.expected)
		}
	}
}

func TestFolderGroupsCountDirectParentAndTracksWithoutMetadata_Postgres(t *testing.T) {
	environment := newCatalogTestEnvironment(t)
	environment.seedTracks(t, []catalogTrackFixture{
		{name: "tagged.mp3", parentPath: "/music/a", updatedAt: environment.catalogBaseMoment,
			metadata: &AudioMetadataModel{Title: "T", Artist: "X"}},
		{name: "untagged.flac", parentPath: "/music/a", format: ".flac", updatedAt: environment.catalogBaseMoment},
		{name: "child.mp3", parentPath: "/music/a/child", updatedAt: environment.catalogBaseMoment},
	})

	folders, err := environment.service.GetLibraryFolders(1, 10, DefaultCatalogSort())
	if err != nil {
		t.Fatalf("folders: %v", err)
	}
	expected := []MusicFolderGroupDto{{Folder: "/music/a", TrackCount: 2}, {Folder: "/music/a/child", TrackCount: 1}}
	if !reflect.DeepEqual(folders.Items, expected) {
		t.Fatalf("folder groups = %+v", folders.Items)
	}

	summary, err := NewRepository(environment.dbContext).GetLibrarySummary()
	if err != nil {
		t.Fatalf("summary: %v", err)
	}
	if summary.TotalTracks != 3 || summary.TotalArtists != 1 || summary.TotalFolders != 2 {
		t.Fatalf("summary must count tracks without metadata: %+v", summary)
	}
}

func TestAlbumGroupsMergeAlbumArtistAndPickFirstNonEmptyYear_Postgres(t *testing.T) {
	environment := newCatalogTestEnvironment(t)
	base := environment.catalogBaseMoment
	environment.seedTracks(t, []catalogTrackFixture{
		{name: "one.mp3", parentPath: "/m", updatedAt: base.Add(3 * time.Hour),
			metadata: &AudioMetadataModel{Title: "One", Artist: "Guest", AlbumArtist: "Main Act", Album: "Record", Year: ""}},
		{name: "two.mp3", parentPath: "/m", updatedAt: base.Add(2 * time.Hour),
			metadata: &AudioMetadataModel{Title: "Two", Artist: "Main Act", Album: "record", Year: "2001"}},
		{name: "three.mp3", parentPath: "/m", updatedAt: base.Add(time.Hour),
			metadata: &AudioMetadataModel{Title: "Three", Artist: "Main Act", Album: "Record", Year: "1999"}},
	})

	albums, err := environment.service.GetLibraryAlbums(1, 10, DefaultCatalogSort())
	if err != nil {
		t.Fatalf("albums: %v", err)
	}
	if len(albums.Items) != 1 || albums.Items[0].TrackCount != 3 || albums.Items[0].Year != "2001" || albums.Items[0].Key != "main act::record" {
		t.Fatalf("unexpected merged album group: %+v", albums.Items)
	}

	artists, err := environment.service.GetLibraryArtists(1, 10, DefaultCatalogSort())
	if err != nil || len(artists.Items) != 1 || artists.Items[0].TrackCount != 3 || artists.Items[0].AlbumCount != 1 {
		t.Fatalf("album artist must group the guest track under the main act: %+v err=%v", artists, err)
	}
}

func TestArtistClusterInputsAndTrackOrdering_Postgres(t *testing.T) {
	environment := newCatalogTestEnvironment(t)
	base := environment.catalogBaseMoment
	environment.seedTracks(t, []catalogTrackFixture{
		{name: "b2.mp3", parentPath: "/m", updatedAt: base, metadata: &AudioMetadataModel{Title: "Something", Artist: "The Beatles", Album: "Abbey Road", Genre: "Rock; Pop", TrackNumber: "2"}},
		{name: "b1.mp3", parentPath: "/m", updatedAt: base, metadata: &AudioMetadataModel{Title: "Come Together", Artist: "The Beatles", Album: "Abbey Road", Genre: "Rock", TrackNumber: "1"}},
		{name: "n1.mp3", parentPath: "/m", updatedAt: base, metadata: &AudioMetadataModel{Title: "Nemo", Artist: "Nightwish", Album: "Once", Genre: "Rock", TrackNumber: "1"}},
		{name: "j1.mp3", parentPath: "/m", updatedAt: base, metadata: &AudioMetadataModel{Title: "Hurt", Artist: "Johnny Cash", Album: "American IV", Genre: "Pop; Country", TrackNumber: "1"}},
		{name: "untagged.mp3", parentPath: "/m", updatedAt: base},
	})
	repository := NewRepository(environment.dbContext)

	inputs, err := repository.GetArtistClusterInputs()
	if err != nil {
		t.Fatalf("GetArtistClusterInputs: %v", err)
	}
	expectedInputs := []artistClusterInput{
		{Key: "the beatles", Artist: "The Beatles", GenreHint: "Rock", TrackCount: 2},
		{Key: "johnny cash", Artist: "Johnny Cash", GenreHint: "Country", TrackCount: 1},
		{Key: "nightwish", Artist: "Nightwish", GenreHint: "Rock", TrackCount: 1},
	}
	if !reflect.DeepEqual(inputs, expectedInputs) {
		t.Fatalf("cluster inputs = %+v", inputs)
	}

	fileIDs, err := repository.GetLibraryFileIDsByArtistKeys([]string{"nightwish", "the beatles"})
	if err != nil {
		t.Fatalf("GetLibraryFileIDsByArtistKeys: %v", err)
	}
	if got := environment.namesOf(fileIDs); !reflect.DeepEqual(got, []string{"n1.mp3", "b1.mp3", "b2.mp3"}) {
		t.Fatalf("cluster track order = %v", got)
	}

	noKeys, err := repository.GetLibraryFileIDsByArtistKeys(nil)
	if err != nil || len(noKeys) != 0 {
		t.Fatalf("no keys must yield no ids, got %v err=%v", noKeys, err)
	}
}

func TestCatalogKeysBackfillMatchesKeysComputedOnUpsert_Postgres(t *testing.T) {
	environment := newCatalogTestEnvironment(t)
	base := environment.catalogBaseMoment
	fixtures := canonicalCatalogFixture(base)
	environment.seedTracks(t, fixtures)

	expectedArtists, err := environment.service.GetLibraryArtists(1, 10, DefaultCatalogSort())
	if err != nil {
		t.Fatalf("artists before reset: %v", err)
	}
	expectedGenres, err := environment.service.GetLibraryGenres(1, 10, DefaultCatalogSort())
	if err != nil {
		t.Fatalf("genres before reset: %v", err)
	}

	resetErr := environment.dbContext.ExecTx(func(tx *sql.Tx) error {
		_, err := tx.Exec(`UPDATE audio_metadata SET catalog_artist_key = NULL, catalog_artist_label = NULL,
			catalog_album_key = NULL, catalog_album_label = NULL, catalog_genre_keys = NULL, catalog_genre_labels = NULL`)
		return err
	})
	if resetErr != nil {
		t.Fatalf("reset keys: %v", resetErr)
	}

	emptyArtists, err := environment.service.GetLibraryArtists(1, 10, DefaultCatalogSort())
	if err != nil || len(emptyArtists.Items) != 0 {
		t.Fatalf("rows without keys must not be grouped yet: %+v err=%v", emptyArtists, err)
	}

	pendingTotal := 0
	afterAudioMetadataID := 0
	for {
		pendingRows, listErr := environment.audioRepository.ListAudioWithoutCatalogKeys(afterAudioMetadataID, 2)
		if listErr != nil {
			t.Fatalf("list pending: %v", listErr)
		}
		if len(pendingRows) == 0 {
			break
		}
		updateErr := environment.dbContext.ExecTx(func(tx *sql.Tx) error {
			for _, pendingRow := range pendingRows {
				groupingKeys := BuildCatalogGroupingKeys(pendingRow.Artist, pendingRow.AlbumArtist, pendingRow.Album, pendingRow.Genre)
				if err := environment.audioRepository.UpdateAudioCatalogKeys(tx, pendingRow.AudioMetadataID, groupingKeys); err != nil {
					return err
				}
			}
			return nil
		})
		if updateErr != nil {
			t.Fatalf("update keys: %v", updateErr)
		}
		pendingTotal += len(pendingRows)
		afterAudioMetadataID = pendingRows[len(pendingRows)-1].AudioMetadataID
	}
	if pendingTotal != 5 {
		t.Fatalf("expected 5 audio_metadata rows to be backfilled, got %d", pendingTotal)
	}

	remaining, err := environment.audioRepository.ListAudioWithoutCatalogKeys(0, 10)
	if err != nil || len(remaining) != 0 {
		t.Fatalf("backfill must leave nothing pending, got %v err=%v", remaining, err)
	}

	artistsAfter, err := environment.service.GetLibraryArtists(1, 10, DefaultCatalogSort())
	if err != nil || !reflect.DeepEqual(artistsAfter.Items, expectedArtists.Items) {
		t.Fatalf("artists after backfill = %+v, want %+v", artistsAfter.Items, expectedArtists.Items)
	}
	genresAfter, err := environment.service.GetLibraryGenres(1, 10, DefaultCatalogSort())
	if err != nil || !reflect.DeepEqual(genresAfter.Items, expectedGenres.Items) {
		t.Fatalf("genres after backfill = %+v, want %+v", genresAfter.Items, expectedGenres.Items)
	}
}
