package music

import (
	"reflect"
	"testing"
	"time"
)

func trackFixture(name string, parentPath string, updatedAt time.Time, metadata AudioMetadataModel) catalogTrackFixture {
	return catalogTrackFixture{name: name, parentPath: parentPath, updatedAt: updatedAt, metadata: &metadata}
}

func (environment *catalogTestEnvironment) trackNamesOfAlbum(t *testing.T, albumKey string) []string {
	t.Helper()
	page, err := environment.service.Repository.GetLibraryTrackIDsByAlbum(albumKey, 1, 50)
	if err != nil {
		t.Fatalf("tracks by album: %v", err)
	}
	return environment.namesOf(page.Items)
}

func TestAlbumTracksOrderByDiscThenNumericTrackNumber_Postgres(t *testing.T) {
	environment := newCatalogTestEnvironment(t)
	base := environment.catalogBaseMoment
	environment.seedTracks(t, []catalogTrackFixture{
		trackFixture("d2t1.mp3", "/m", base, AudioMetadataModel{Title: "Disc Two", Artist: "X", Album: "Box", DiscNumberText: "2/2", TrackNumber: "1"}),
		trackFixture("d1t10.mp3", "/m", base, AudioMetadataModel{Title: "Ten", Artist: "X", Album: "Box", DiscNumberText: "1/2", TrackNumber: "10"}),
		trackFixture("d1t2.mp3", "/m", base, AudioMetadataModel{Title: "Two", Artist: "X", Album: "Box", DiscNumberText: "1/2", TrackNumber: "2"}),
		trackFixture("d1t1.mp3", "/m", base, AudioMetadataModel{Title: "One", Artist: "X", Album: "Box", DiscNumberText: "1/2", TrackNumber: "1"}),
		trackFixture("untagged_z.mp3", "/m", base, AudioMetadataModel{Title: "Zed", Artist: "X", Album: "Box"}),
		trackFixture("untagged_a.mp3", "/m", base, AudioMetadataModel{Title: "Aaa", Artist: "X", Album: "Box"}),
		trackFixture("nodisc_t1.mp3", "/m", base, AudioMetadataModel{Title: "Zzz No Disc", Artist: "X", Album: "Box", TrackNumber: "1"}),
	})

	got := environment.trackNamesOfAlbum(t, "x::box")
	want := []string{"d1t1.mp3", "nodisc_t1.mp3", "d1t2.mp3", "d1t10.mp3", "untagged_a.mp3", "untagged_z.mp3", "d2t1.mp3"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("album track order = %v, want %v", got, want)
	}
}

func TestArtistTracksOrderByAlbumYearThenAlbumNameThenDiscAndTrack_Postgres(t *testing.T) {
	environment := newCatalogTestEnvironment(t)
	base := environment.catalogBaseMoment
	environment.seedTracks(t, []catalogTrackFixture{
		trackFixture("late_t1.mp3", "/m", base, AudioMetadataModel{Title: "L1", Artist: "Y", Album: "Late", Year: "2020-05-01", TrackNumber: "1"}),
		trackFixture("early_t2.mp3", "/m", base, AudioMetadataModel{Title: "E2", Artist: "Y", Album: "Early", Year: "2010", TrackNumber: "2"}),
		trackFixture("early_t1.mp3", "/m", base, AudioMetadataModel{Title: "E1", Artist: "Y", Album: "Early", Year: "2010", TrackNumber: "1"}),
		trackFixture("bravo_t1.mp3", "/m", base, AudioMetadataModel{Title: "B1", Artist: "Y", Album: "Bravo", Year: "2015", TrackNumber: "1"}),
		trackFixture("alpha_d2.mp3", "/m", base, AudioMetadataModel{Title: "A2", Artist: "Y", Album: "Alpha", Year: "2015", DiscNumberText: "2", TrackNumber: "1"}),
		trackFixture("alpha_d1.mp3", "/m", base, AudioMetadataModel{Title: "A1", Artist: "Y", Album: "Alpha", Year: "2015", DiscNumberText: "1", TrackNumber: "5"}),
		trackFixture("undated_t1.mp3", "/m", base, AudioMetadataModel{Title: "U1", Artist: "Y", Album: "Aaa Undated", TrackNumber: "1"}),
	})

	page, err := environment.service.Repository.GetLibraryTrackIDsByArtist("y", 1, 50)
	if err != nil {
		t.Fatalf("tracks by artist: %v", err)
	}
	want := []string{"early_t1.mp3", "early_t2.mp3", "alpha_d1.mp3", "alpha_d2.mp3", "bravo_t1.mp3", "late_t1.mp3", "undated_t1.mp3"}
	if got := environment.namesOf(page.Items); !reflect.DeepEqual(got, want) {
		t.Fatalf("artist track order = %v, want %v", got, want)
	}
}

func TestFolderAndGenreTracksOrderByDiscAndTrack_Postgres(t *testing.T) {
	environment := newCatalogTestEnvironment(t)
	base := environment.catalogBaseMoment
	environment.seedTracks(t, []catalogTrackFixture{
		trackFixture("b_t1.mp3", "/m/b", base, AudioMetadataModel{Title: "B1", Artist: "Zed", Album: "Z1", Genre: "Rock", TrackNumber: "1"}),
		trackFixture("a_d2t1.mp3", "/m/a", base, AudioMetadataModel{Title: "A21", Artist: "Ann", Album: "A1", Genre: "Rock", DiscNumberText: "2", TrackNumber: "1"}),
		trackFixture("a_d1t10.mp3", "/m/a", base, AudioMetadataModel{Title: "A110", Artist: "Ann", Album: "A1", Genre: "Rock", DiscNumberText: "1", TrackNumber: "10"}),
		trackFixture("a_d1t2.mp3", "/m/a", base, AudioMetadataModel{Title: "A12", Artist: "Ann", Album: "A1", Genre: "Rock", DiscNumberText: "1", TrackNumber: "2"}),
	})

	want := []string{"a_d1t2.mp3", "a_d1t10.mp3", "a_d2t1.mp3", "b_t1.mp3"}

	folderPage, err := environment.service.Repository.GetLibraryTrackIDsByFolder("/m", 1, 50)
	if err != nil {
		t.Fatalf("tracks by folder: %v", err)
	}
	if got := environment.namesOf(folderPage.Items); !reflect.DeepEqual(got, want) {
		t.Fatalf("folder track order = %v, want %v", got, want)
	}

	genrePage, err := environment.service.Repository.GetLibraryTrackIDsByGenre("rock", 1, 50)
	if err != nil {
		t.Fatalf("tracks by genre: %v", err)
	}
	if got := environment.namesOf(genrePage.Items); !reflect.DeepEqual(got, want) {
		t.Fatalf("genre track order = %v, want %v", got, want)
	}
}

func seedSortableLibrary(t *testing.T, environment *catalogTestEnvironment) {
	t.Helper()
	base := environment.catalogBaseMoment
	environment.seedTracks(t, []catalogTrackFixture{
		trackFixture("a1.mp3", "/m/alpha", base.Add(1*time.Hour), AudioMetadataModel{Title: "A1", Artist: "Artist A", Album: "Old Album", Genre: "Rock", Year: "1999"}),
		trackFixture("a2.mp3", "/m/alpha", base.Add(2*time.Hour), AudioMetadataModel{Title: "A2", Artist: "Artist A", Album: "Old Album", Genre: "Rock", Year: "1999"}),
		trackFixture("a3.mp3", "/m/alpha", base.Add(3*time.Hour), AudioMetadataModel{Title: "A3", Artist: "Artist A", Album: "Old Album", Genre: "Rock", Year: "1999"}),
		trackFixture("c1.mp3", "/m/charlie", base.Add(5*time.Hour), AudioMetadataModel{Title: "C1", Artist: "Artist C", Album: "Mid Album", Genre: "Jazz", Year: "2008"}),
		trackFixture("c2.mp3", "/m/charlie", base.Add(6*time.Hour), AudioMetadataModel{Title: "C2", Artist: "Artist C", Album: "Mid Album", Genre: "Jazz", Year: "2008"}),
		trackFixture("b1.mp3", "/m/bravo", base.Add(9*time.Hour), AudioMetadataModel{Title: "B1", Artist: "Artist B", Album: "New Album", Genre: "Blues"}),
	})
}

func artistNames(page []MusicArtistGroupDto) []string {
	names := []string{}
	for _, group := range page {
		names = append(names, group.Artist)
	}
	return names
}

func albumNames(page []MusicAlbumGroupDto) []string {
	names := []string{}
	for _, group := range page {
		names = append(names, group.Album)
	}
	return names
}

func genreNames(page []MusicGenreGroupDto) []string {
	names := []string{}
	for _, group := range page {
		names = append(names, group.Genre)
	}
	return names
}

func folderNames(page []MusicFolderGroupDto) []string {
	names := []string{}
	for _, group := range page {
		names = append(names, group.Folder)
	}
	return names
}

func TestLibraryListSortingHonoursFieldAndDirection_Postgres(t *testing.T) {
	environment := newCatalogTestEnvironment(t)
	seedSortableLibrary(t, environment)

	cases := []struct {
		name string
		sort CatalogSort
		want []string
	}{
		{"default is tracks desc", CatalogSort{}, []string{"Artist A", "Artist C", "Artist B"}},
		{"tracks asc", CatalogSort{Field: CatalogSortByTracks}, []string{"Artist B", "Artist C", "Artist A"}},
		{"name asc", CatalogSort{Field: CatalogSortByName}, []string{"Artist A", "Artist B", "Artist C"}},
		{"name desc", CatalogSort{Field: CatalogSortByName, IsDescending: true}, []string{"Artist C", "Artist B", "Artist A"}},
		{"recent desc", CatalogSort{Field: CatalogSortByRecent, IsDescending: true}, []string{"Artist B", "Artist C", "Artist A"}},
		{"recent asc", CatalogSort{Field: CatalogSortByRecent}, []string{"Artist A", "Artist C", "Artist B"}},
	}
	for _, testCase := range cases {
		artists, err := environment.service.GetLibraryArtists(1, 10, testCase.sort)
		if err != nil {
			t.Fatalf("%s artists: %v", testCase.name, err)
		}
		if got := artistNames(artists.Items); !reflect.DeepEqual(got, testCase.want) {
			t.Fatalf("%s artists = %v, want %v", testCase.name, got, testCase.want)
		}
	}

	albums, err := environment.service.GetLibraryAlbums(1, 10, CatalogSort{Field: CatalogSortByName})
	if err != nil || !reflect.DeepEqual(albumNames(albums.Items), []string{"Mid Album", "New Album", "Old Album"}) {
		t.Fatalf("albums by name = %v err=%v", albumNames(albums.Items), err)
	}
	albums, err = environment.service.GetLibraryAlbums(1, 10, CatalogSort{Field: CatalogSortByRecent, IsDescending: true})
	if err != nil || !reflect.DeepEqual(albumNames(albums.Items), []string{"New Album", "Mid Album", "Old Album"}) {
		t.Fatalf("albums by recent = %v err=%v", albumNames(albums.Items), err)
	}
	albums, err = environment.service.GetLibraryAlbums(1, 10, CatalogSort{Field: CatalogSortByYear, IsDescending: true})
	if err != nil || !reflect.DeepEqual(albumNames(albums.Items), []string{"Mid Album", "Old Album", "New Album"}) {
		t.Fatalf("albums by year desc = %v err=%v", albumNames(albums.Items), err)
	}
	albums, err = environment.service.GetLibraryAlbums(1, 10, CatalogSort{Field: CatalogSortByYear})
	if err != nil || !reflect.DeepEqual(albumNames(albums.Items), []string{"Old Album", "Mid Album", "New Album"}) {
		t.Fatalf("albums by year asc = %v err=%v", albumNames(albums.Items), err)
	}
	albums, err = environment.service.GetLibraryAlbums(1, 1, CatalogSort{Field: CatalogSortByYear, IsDescending: true})
	if err != nil || !reflect.DeepEqual(albumNames(albums.Items), []string{"Mid Album"}) || !albums.Pagination.HasNext {
		t.Fatalf("paged albums by year = %v err=%v", albumNames(albums.Items), err)
	}

	genres, err := environment.service.GetLibraryGenres(1, 10, CatalogSort{Field: CatalogSortByName, IsDescending: true})
	if err != nil || !reflect.DeepEqual(genreNames(genres.Items), []string{"Rock", "Jazz", "Blues"}) {
		t.Fatalf("genres by name desc = %v err=%v", genreNames(genres.Items), err)
	}
	genres, err = environment.service.GetLibraryGenres(1, 10, CatalogSort{Field: CatalogSortByRecent, IsDescending: true})
	if err != nil || !reflect.DeepEqual(genreNames(genres.Items), []string{"Blues", "Jazz", "Rock"}) {
		t.Fatalf("genres by recent = %v err=%v", genreNames(genres.Items), err)
	}

	folders, err := environment.service.GetLibraryFolders(1, 10, CatalogSort{Field: CatalogSortByName})
	if err != nil || !reflect.DeepEqual(folderNames(folders.Items), []string{"/m/alpha", "/m/bravo", "/m/charlie"}) {
		t.Fatalf("folders by name = %v err=%v", folderNames(folders.Items), err)
	}
	folders, err = environment.service.GetLibraryFolders(1, 10, CatalogSort{Field: CatalogSortByRecent, IsDescending: true})
	if err != nil || !reflect.DeepEqual(folderNames(folders.Items), []string{"/m/bravo", "/m/charlie", "/m/alpha"}) {
		t.Fatalf("folders by recent = %v err=%v", folderNames(folders.Items), err)
	}
}
