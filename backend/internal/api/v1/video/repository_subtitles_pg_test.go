package video

import (
	"reflect"
	"sort"
	"testing"

	"nas-go/api/internal/testutil"
)

func subtitleNames(subtitleFiles []SubtitleFileModel) []string {
	names := make([]string, 0, len(subtitleFiles))
	for _, subtitleFile := range subtitleFiles {
		names = append(names, subtitleFile.Name)
	}
	sort.Strings(names)
	return names
}

func TestListSubtitleFilesByNamePrefixMatchesSameFolderCaseInsensitiveSkippingDeleted_Postgres(t *testing.T) {
	videoRepository := newPostgresVideoRepository(t)
	repository := NewSubtitleRepository(videoRepository.DbContext)
	seedVideoFiles(t, videoRepository, []seededVideoFile{
		{name: "Filme.mkv", folder: "/data/Show"},
		{name: "Filme.srt", folder: "/data/Show"},
		{name: "FILME.PT-BR.SRT", folder: "/data/Show"},
		{name: "filme.en.forced.vtt", folder: "/data/Show"},
		{name: "Filme.old.srt", folder: "/data/Show", deleted: true},
		{name: "Filme 2.srt", folder: "/data/Show"},
		{name: "Filmes.srt", folder: "/data/Show"},
		{name: "Filme.txt", folder: "/data/Show"},
		{name: "Filme.srt", folder: "/data/Other"},
		{name: "Filme.es.srt", folder: "/data/Show/Sub"},
		{name: "Fil_e.srt", folder: "/data/Show"},
		{name: "Fil%.srt", folder: "/data/Show"},
	})

	subtitleFiles, err := repository.ListSubtitleFilesByNamePrefix("/data/Show", "Filme.mkv")
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	expected := []string{"FILME.PT-BR.SRT", "Filme.srt", "filme.en.forced.vtt"}
	sort.Strings(expected)
	if actual := subtitleNames(subtitleFiles); !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected %v, got %v", expected, actual)
	}
}

func TestListSubtitleFilesByNamePrefixEscapesLikeWildcards_Postgres(t *testing.T) {
	videoRepository := newPostgresVideoRepository(t)
	repository := NewSubtitleRepository(videoRepository.DbContext)
	seedVideoFiles(t, videoRepository, []seededVideoFile{
		{name: "Fil_e.srt", folder: "/data/Show"},
		{name: "Filxe.srt", folder: "/data/Show"},
		{name: "Fil%.srt", folder: "/data/Show"},
	})

	subtitleFiles, err := repository.ListSubtitleFilesByNamePrefix("/data/Show", "Fil_e.mkv")
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	if actual := subtitleNames(subtitleFiles); !reflect.DeepEqual(actual, []string{"Fil_e.srt"}) {
		t.Fatalf("unexpected %v", actual)
	}
}

func newPostgresVideoRepository(t *testing.T) *Repository {
	t.Helper()
	return NewRepository(testutil.NewPostgresDB(t, "kuranas_video_it"))
}
