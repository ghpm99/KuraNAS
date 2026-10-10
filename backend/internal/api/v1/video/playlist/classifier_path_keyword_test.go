package playlist

import (
	"regexp"
	"testing"
)

func TestPathKeywordSpecsMatchTokensNotSubstrings(t *testing.T) {
	programSpec := &pathContainsSpec{name: "program_path", keywords: []string{"steam", "program", "programa", "programas", "sample", "samples", "benchmark"}}
	capturePathSpec := &pathContainsSpec{name: "capture_path", keywords: []string{"camera", "screen recording", "whatsapp video"}}
	musicSpec := &musicVideoSpec{}

	cases := []struct {
		description string
		spec        ClassificationSpec
		video       VideoEntry
		isMatch     bool
	}{
		{"programacao filename", programSpec, VideoEntry{Name: "programação.mp4", Path: "/Videos/programação.mp4", ParentPath: "/Videos"}, false},
		{"steamboat filename", programSpec, VideoEntry{Name: "steamboat.mp4", Path: "/Videos/steamboat.mp4", ParentPath: "/Videos"}, false},
		{"mvp demo filename", musicSpec, VideoEntry{Name: "mvp_demo.mkv", Path: "/Videos/mvp_demo.mkv", ParentPath: "/Videos"}, false},
		{"mv inside word", musicSpec, VideoEntry{Name: "summvary.mp4", Path: "/Videos/summvary.mp4", ParentPath: "/Videos"}, false},
		{"musicians folder", musicSpec, VideoEntry{Name: "a.mp4", Path: "/Musicians/a.mp4", ParentPath: "/Musicians"}, false},
		{"cameraman filename", capturePathSpec, VideoEntry{Name: "cameraman.mp4", Path: "/Videos/cameraman.mp4", ParentPath: "/Videos"}, false},
		{"programas folder", programSpec, VideoEntry{Name: "a.mp4", Path: "/programas/a.mp4", ParentPath: "/programas"}, true},
		{"Samples folder", programSpec, VideoEntry{Name: "a.mp4", Path: "/Media/Samples/a.mp4", ParentPath: "/Media/Samples"}, true},
		{"steam token", programSpec, VideoEntry{Name: "steam_trailer.mp4", Path: "/Videos/steam_trailer.mp4", ParentPath: "/Videos"}, true},
		{"programa accent-insensitive", programSpec, VideoEntry{Name: "a.mp4", Path: "/Vídeos/Programa/a.mp4", ParentPath: "/Vídeos/Programa"}, true},
		{"Music Videos folder", musicSpec, VideoEntry{Name: "a.mp4", Path: "/Music Videos/a.mp4", ParentPath: "/Music Videos"}, true},
		{"mv token in name", musicSpec, VideoEntry{Name: "Artist - Song (MV).mp4", Path: "/Videos/Artist - Song (MV).mp4", ParentPath: "/Videos"}, true},
		{"multi word keyword", capturePathSpec, VideoEntry{Name: "a.mp4", Path: "/Screen Recording/a.mp4", ParentPath: "/Screen Recording"}, true},
		{"windows separator", capturePathSpec, VideoEntry{Name: "a.mp4", Path: `D:\Camera\a.mp4`, ParentPath: `D:\Camera`}, true},
	}

	for _, testCase := range cases {
		t.Run(testCase.description, func(t *testing.T) {
			if got := testCase.spec.IsSatisfiedBy(testCase.video); got != testCase.isMatch {
				t.Fatalf("expected %v, got %v", testCase.isMatch, got)
			}
		})
	}
}

func TestCourseAndClipSpecsMatchTokensNotSubstrings(t *testing.T) {
	courseRegexpSpec := &coursePatternSpec{pattern: regexp.MustCompile(`(?i)(?:aula|lesson|lecture|module|modulo)\s*\d+`)}
	clipSpecification := &clipSpec{}

	cases := []struct {
		description string
		spec        ClassificationSpec
		video       VideoEntry
		isMatch     bool
	}{
		{"discurso filename", courseRegexpSpec, VideoEntry{Name: "discurso.mp4", Path: "/Videos/discurso.mp4", ParentPath: "/Videos"}, false},
		{"eclipse filename", clipSpecification, VideoEntry{Name: "eclipse.mp4", Path: "/Videos/eclipse.mp4", ParentPath: "/Videos"}, false},
		{"curso folder", courseRegexpSpec, VideoEntry{Name: "aula 1.mp4", Path: "/Curso de Go/aula 1.mp4", ParentPath: "/Curso de Go"}, true},
		{"cursos folder", courseRegexpSpec, VideoEntry{Name: "a.mp4", Path: "/Cursos/a.mp4", ParentPath: "/Cursos"}, true},
		{"clips folder", clipSpecification, VideoEntry{Name: "x.mp4", Path: "/clips/x.mp4", ParentPath: "/clips"}, true},
		{"reels folder", clipSpecification, VideoEntry{Name: "x.mp4", Path: "/reels/x.mp4", ParentPath: "/reels"}, true},
		{"shorts folder", clipSpecification, VideoEntry{Name: "x.mp4", Path: "/shorts/x.mp4", ParentPath: "/shorts"}, true},
	}

	for _, testCase := range cases {
		t.Run(testCase.description, func(t *testing.T) {
			if got := testCase.spec.IsSatisfiedBy(testCase.video); got != testCase.isMatch {
				t.Fatalf("expected %v, got %v", testCase.isMatch, got)
			}
		})
	}
}
