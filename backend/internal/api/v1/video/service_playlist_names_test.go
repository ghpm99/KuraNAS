package video

import (
	"database/sql"
	"testing"

	"nas-go/api/internal/api/v1/video/playlist"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/i18n"
)

func useTranslationLocale(t *testing.T, locale string) {
	t.Helper()
	previousLocale := config.AppConfig.Lang
	config.AppConfig.Lang = locale
	if err := i18n.LoadTranslations(); err != nil {
		t.Fatalf("load translations %s: %v", locale, err)
	}
	t.Cleanup(func() {
		config.AppConfig.Lang = previousLocale
		_ = i18n.LoadTranslations()
	})
}

func TestResolvePlaylistNameTranslatesKeyAndPassesDataNameThrough(t *testing.T) {
	useTranslationLocale(t, "pt-BR")
	keyed := sql.NullString{String: "VIDEO_PLAYLIST_NAME_SERIES", Valid: true}

	if got := playlist.ResolveDisplayName("Series", keyed); got != "Séries" {
		t.Fatalf("keyed name = %q, want Séries", got)
	}
	if got := playlist.ResolveDisplayName("Naruto", sql.NullString{}); got != "Naruto" {
		t.Fatalf("data name = %q, want Naruto", got)
	}
	if got := playlist.ResolveDisplayName("Naruto", sql.NullString{String: "", Valid: true}); got != "Naruto" {
		t.Fatalf("empty key must fall back to stored name, got %q", got)
	}
}

func TestResolvePlaylistNameFollowsLanguageWithoutRebuild(t *testing.T) {
	keyed := sql.NullString{String: "VIDEO_PLAYLIST_NAME_MOVIES", Valid: true}

	useTranslationLocale(t, "pt-BR")
	if got := playlist.ResolveDisplayName("Filmes", keyed); got != "Filmes" {
		t.Fatalf("pt-BR = %q", got)
	}
	useTranslationLocale(t, "en-US")
	if got := playlist.ResolveDisplayName("Filmes", keyed); got != "Movies" {
		t.Fatalf("en-US = %q", got)
	}
}

func TestPlaylistModelToDtoExposesResolvedNameAndKey(t *testing.T) {
	useTranslationLocale(t, "en-US")

	keyedDto := (&VideoPlaylistModel{Name: "Filmes", NameKey: sql.NullString{String: "VIDEO_PLAYLIST_NAME_MOVIES", Valid: true}}).ToDto(nil)
	if keyedDto.Name != "Movies" || keyedDto.NameKey == nil || *keyedDto.NameKey != "VIDEO_PLAYLIST_NAME_MOVIES" {
		t.Fatalf("keyed dto = %+v", keyedDto)
	}

	dataDto := (&VideoPlaylistModel{Name: "Naruto"}).ToDto(nil)
	if dataDto.Name != "Naruto" || dataDto.NameKey != nil {
		t.Fatalf("data dto = %+v", dataDto)
	}
}

func TestGetPlaylistsByVideoResolvesKeyedNames(t *testing.T) {
	useTranslationLocale(t, "en-US")
	repo := &videoRepoMock{}
	repo.getVideoPlaylistsByVideoFn = func(videoID int) ([]VideoPlaylistOfVideoModel, error) {
		return []VideoPlaylistOfVideoModel{
			{ID: 1, Name: "Series", NameKey: sql.NullString{String: "VIDEO_PLAYLIST_NAME_SERIES", Valid: true}, Type: "series"},
			{ID: 2, Name: "Naruto", Type: "folder"},
		}, nil
	}
	playlists, err := newVideoServiceForTest(t, repo).GetPlaylistsByVideo(5)
	if err != nil {
		t.Fatalf("GetPlaylistsByVideo: %v", err)
	}
	if playlists[0].Name != "Series" || playlists[0].NameKey == nil || playlists[1].Name != "Naruto" || playlists[1].NameKey != nil {
		t.Fatalf("playlists = %+v", playlists)
	}
}

func TestHomeCatalogSectionTitlesAreTranslated(t *testing.T) {
	repo := &videoRepoMock{
		getCatalogVideosFn:    func(limit int) ([]VideoFileModel, error) { return nil, nil },
		getRecentVideosFn:     func(limit int) ([]VideoFileModel, error) { return nil, nil },
		getContinueWatchingFn: func(clientID string, limit int) ([]ContinueWatchingModel, error) { return nil, nil },
	}
	service := newVideoServiceForTest(t, repo)

	titlesByLocale := map[string][]string{
		"pt-BR": {"Continue assistindo", "Séries", "Filmes", "Vídeos pessoais", "Adicionados recentemente"},
		"en-US": {"Continue watching", "Series", "Movies", "Personal videos", "Recently added"},
	}
	for locale, wantTitles := range titlesByLocale {
		useTranslationLocale(t, locale)
		catalog, err := service.GetHomeCatalog("client", 10)
		if err != nil {
			t.Fatalf("GetHomeCatalog: %v", err)
		}
		if len(catalog.Sections) != len(wantTitles) {
			t.Fatalf("%s: sections = %d", locale, len(catalog.Sections))
		}
		for index, section := range catalog.Sections {
			if section.Title != wantTitles[index] {
				t.Errorf("%s: section %s title = %q, want %q", locale, section.Key, section.Title, wantTitles[index])
			}
		}
	}
}

type upsertedAutoPlaylist struct {
	name    string
	nameKey sql.NullString
}

func rebuildAndCaptureUpserts(t *testing.T, videos []VideoFileModel) map[string]upsertedAutoPlaylist {
	t.Helper()
	videosWithMetadata := make([]VideoWithMetadataModel, 0, len(videos))
	for _, video := range videos {
		videosWithMetadata = append(videosWithMetadata, VideoWithMetadataModel{VideoFileModel: video})
	}
	upsertsBySourcePath := map[string]upsertedAutoPlaylist{}
	repo := &videoRepoMock{
		getAllVideosWithMetadataFn: func() ([]VideoWithMetadataModel, error) { return videosWithMetadata, nil },
		upsertAutoPlaylistFn: func(tx *sql.Tx, contextType, sourcePath, name string, nameKey sql.NullString, groupMode, classification string) (VideoPlaylistModel, error) {
			upsertsBySourcePath[sourcePath] = upsertedAutoPlaylist{name: name, nameKey: nameKey}
			return VideoPlaylistModel{ID: len(upsertsBySourcePath)}, nil
		},
		getPlaylistExclusionsFn:   func(playlistID int) (map[int]bool, error) { return map[int]bool{}, nil },
		deleteAutoPlaylistItemsFn: func(tx *sql.Tx, playlistID int) error { return nil },
		insertPlaylistItemsSrcFn:  func(tx *sql.Tx, playlistID int, videoIDs []int, sourceKind string) error { return nil },
	}
	if err := newVideoServiceForTest(t, repo).RebuildSmartPlaylists(); err != nil {
		t.Fatalf("RebuildSmartPlaylists: %v", err)
	}
	return upsertsBySourcePath
}

func TestRebuildKeepsDetectedSeriesNameAsDataWithoutKey(t *testing.T) {
	useTranslationLocale(t, "pt-BR")
	upserts := rebuildAndCaptureUpserts(t, []VideoFileModel{
		{ID: 1, Name: "Show S01E01.mkv", ParentPath: "/shows/Naruto", Path: "/shows/Naruto/Show S01E01.mkv"},
		{ID: 2, Name: "Show S01E02.mkv", ParentPath: "/shows/Naruto", Path: "/shows/Naruto/Show S01E02.mkv"},
	})

	seriesPlaylist, ok := upserts["series:show"]
	if !ok || seriesPlaylist.name != "Show" || seriesPlaylist.nameKey.Valid {
		t.Fatalf("series playlist must keep the detected prefix as data, got %+v (all=%+v)", seriesPlaylist, upserts)
	}
}

func TestRebuildPersistsKeyAndTranslatedNameForGenericPlaylist(t *testing.T) {
	useTranslationLocale(t, "pt-BR")
	upserts := rebuildAndCaptureUpserts(t, []VideoFileModel{
		{ID: 1, Name: "praia.mp4", ParentPath: "/camera/Viagem", Path: "/camera/Viagem/praia.mp4"},
		{ID: 2, Name: "festa.mp4", ParentPath: "/camera/Festas", Path: "/camera/Festas/festa.mp4"},
		{ID: 3, Name: "natal.mp4", ParentPath: "/camera/Natal", Path: "/camera/Natal/natal.mp4"},
		{ID: 4, Name: "aniversario.mp4", ParentPath: "/camera/Aniversario", Path: "/camera/Aniversario/aniversario.mp4"},
		{ID: 5, Name: "passeio.mp4", ParentPath: "/camera/Passeio", Path: "/camera/Passeio/passeio.mp4"},
	})

	mixedPlaylist, ok := upserts["random:smart_mix"]
	if !ok || mixedPlaylist.nameKey.String != "VIDEO_PLAYLIST_NAME_MIXED" || mixedPlaylist.name != "Mix variado" {
		t.Fatalf("mixed playlist must persist key and translated name, got %+v (all=%+v)", mixedPlaylist, upserts)
	}
}
