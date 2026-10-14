package playlist

import (
	"reflect"
	"testing"

	"nas-go/api/internal/config"
	"nas-go/api/pkg/i18n"
)

func useLocale(t *testing.T, locale string) {
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

func TestMatchNameKeysFollowsLocaleIgnoringCaseAndAccents(t *testing.T) {
	cases := []struct {
		name     string
		locale   string
		term     string
		contains []string
		exact    []string
		prefix   []string
	}{
		{"english exact", "en-US", "movies", []string{"VIDEO_PLAYLIST_NAME_MOVIES"}, []string{"VIDEO_PLAYLIST_NAME_MOVIES"}, []string{"VIDEO_PLAYLIST_NAME_MOVIES"}},
		{"english upper case", "en-US", "MOVIES", []string{"VIDEO_PLAYLIST_NAME_MOVIES"}, []string{"VIDEO_PLAYLIST_NAME_MOVIES"}, []string{"VIDEO_PLAYLIST_NAME_MOVIES"}},
		{"portuguese does not match english name", "pt-BR", "movies", []string{}, []string{}, []string{}},
		{"portuguese exact", "pt-BR", "filmes", []string{"VIDEO_PLAYLIST_NAME_MOVIES"}, []string{"VIDEO_PLAYLIST_NAME_MOVIES"}, []string{"VIDEO_PLAYLIST_NAME_MOVIES"}},
		{"accented term matches folded name", "pt-BR", "Séries", []string{"VIDEO_PLAYLIST_NAME_SERIES"}, []string{"VIDEO_PLAYLIST_NAME_SERIES"}, []string{"VIDEO_PLAYLIST_NAME_SERIES"}},
		{"unaccented term matches accented name", "pt-BR", "serie", []string{"VIDEO_PLAYLIST_NAME_SERIES", "VIDEO_PLAYLIST_NAME_UNKNOWN_SERIES"}, []string{}, []string{"VIDEO_PLAYLIST_NAME_SERIES", "VIDEO_PLAYLIST_NAME_UNKNOWN_SERIES"}},
		{"portuguese contains only", "pt-BR", "ilme", []string{"VIDEO_PLAYLIST_NAME_MOVIES"}, []string{}, []string{}},
		{"blank term matches nothing", "pt-BR", "   ", []string{}, []string{}, []string{}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			useLocale(t, testCase.locale)
			matches := MatchNameKeys(testCase.term)
			if !reflect.DeepEqual(matches.Contains, testCase.contains) {
				t.Fatalf("contains = %v, want %v", matches.Contains, testCase.contains)
			}
			if !reflect.DeepEqual(matches.Exact, testCase.exact) {
				t.Fatalf("exact = %v, want %v", matches.Exact, testCase.exact)
			}
			if !reflect.DeepEqual(matches.Prefix, testCase.prefix) {
				t.Fatalf("prefix = %v, want %v", matches.Prefix, testCase.prefix)
			}
		})
	}
}
