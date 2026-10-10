package playlist

import "testing"

func TestEpisodeDetectionBoundaries(t *testing.T) {
	cases := []struct {
		fileName          string
		isEpisode         bool
		isClassifiedSerie bool
	}{
		{"Movie 1920x1080.mkv", false, false},
		{"Movie 1280x720.mp4", false, false},
		{"Movie 3840x2160.mp4", false, false},
		{"step 2.mp4", false, false},
		{"Show S01E02.mkv", true, true},
		{"Show.s1e12.mkv", true, true},
		{"Show 1x05.mkv", true, true},
		{"Show Ep 3.mkv", true, true},
		{"Show Episódio 12.mkv", true, true},
		{"Show Cap. 7.mkv", true, true},
		{"Show Capítulo 7.mkv", true, true},
		{"Show Episode 4.mkv", true, true},
		{"Show_S01E02_720p.mkv", true, true},
	}

	rules := defaultClassificationRules()
	episodeRule := rules[0]

	for _, testCase := range cases {
		t.Run(testCase.fileName, func(t *testing.T) {
			matchesEpisodePattern := EpisodePattern.MatchString(testCase.fileName) || EpisodeNumeric.MatchString(testCase.fileName)
			if matchesEpisodePattern != testCase.isEpisode {
				t.Fatalf("EpisodePattern/EpisodeNumeric match = %v, want %v", matchesEpisodePattern, testCase.isEpisode)
			}
			spec, isEpisodeSpec := episodeRule.Spec.(*episodePatternSpec)
			if !isEpisodeSpec {
				t.Fatalf("first rule is not episodePatternSpec")
			}
			if spec.pattern.MatchString(testCase.fileName) != testCase.isClassifiedSerie {
				t.Fatalf("classification pattern match = %v, want %v", !testCase.isClassifiedSerie, testCase.isClassifiedSerie)
			}
		})
	}
}

func TestInferTitlePrefixKeepsResolutionAndStripsEpisode(t *testing.T) {
	if got := InferTitlePrefix("Show S01E02.mkv"); got != "show" {
		t.Fatalf("got %q", got)
	}
	if got := InferTitlePrefix("Show 1x05.mkv"); got != "show" {
		t.Fatalf("got %q", got)
	}
}
