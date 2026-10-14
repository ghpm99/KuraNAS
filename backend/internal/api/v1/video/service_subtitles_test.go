package video

import (
	"errors"
	"reflect"
	"testing"
)

type subtitleRepositoryFake struct {
	subtitleFiles []SubtitleFileModel
	err           error
}

func (fake *subtitleRepositoryFake) ListSubtitleFilesByNamePrefix(parentPath string, videoName string) ([]SubtitleFileModel, error) {
	return fake.subtitleFiles, fake.err
}

func TestListVideoSubtitlesBuildsDtosAndDropsNonMatchingNames(t *testing.T) {
	service := NewSubtitleService(&subtitleRepositoryFake{subtitleFiles: []SubtitleFileModel{
		{ID: 1, Name: "Filme.srt"},
		{ID: 2, Name: "Filme.en.forced.vtt"},
		{ID: 3, Name: "Filme.2.srt.bak"},
	}})

	subtitles, err := service.ListVideoSubtitles("/data", "Filme.mkv")

	expected := []SubtitleDto{
		{ID: 1, Format: "srt"},
		{ID: 2, Label: "en.forced", Language: "en", Format: "vtt", IsForced: true},
	}
	if err != nil || !reflect.DeepEqual(subtitles, expected) {
		t.Fatalf("unexpected %+v %v", subtitles, err)
	}
}

func TestListVideoSubtitlesPropagatesRepositoryError(t *testing.T) {
	service := NewSubtitleService(&subtitleRepositoryFake{err: errors.New("boom")})

	if _, err := service.ListVideoSubtitles("/data", "Filme.mkv"); err == nil {
		t.Fatal("expected error")
	}
}
