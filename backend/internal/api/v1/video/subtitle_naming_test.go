package video

import "testing"

func TestParseSubtitleNameExtractsLabelLanguageAndForcedFlag(t *testing.T) {
	testCases := []struct {
		videoName    string
		subtitleName string
		isMatch      bool
		label        string
		language     string
		format       string
		isForced     bool
	}{
		{"Filme.mkv", "Filme.srt", true, "", "", "srt", false},
		{"Filme.mkv", "Filme.pt-BR.srt", true, "pt-BR", "pt-BR", "srt", false},
		{"Filme.mkv", "Filme.en.forced.srt", true, "en.forced", "en", "srt", true},
		{"Filme.mkv", "Filme.Portuguese.srt", true, "Portuguese", "", "srt", false},
		{"Filme.2019.1080p.mkv", "Filme.2019.1080p.pt.srt", true, "pt", "pt", "srt", false},
		{"Filme.2019.1080p.mkv", "Filme.2019.1080p.srt", true, "", "", "srt", false},
		{"Filme.mkv", "FILME.EN.VTT", true, "EN", "en", "vtt", false},
		{"Filme.mkv", "Filme.eng.srt", true, "eng", "en", "srt", false},
		{"Filme.mkv", "Filme.pt_br.srt", true, "pt_br", "pt-BR", "srt", false},
		{"Filme.mkv", "Filme.zh-Hans.srt", true, "zh-Hans", "zh-Hans", "srt", false},
		{"Filme.mkv", "Filme.sub.srt", true, "sub", "", "srt", false},
		{"Filme.mkv", "Filme 2.srt", false, "", "", "", false},
		{"Filme.mkv", "Filmes.srt", false, "", "", "", false},
		{"Filme.mkv", "Filme.txt", false, "", "", "", false},
		{"Filme.mkv", "Filme..srt", false, "", "", "", false},
		{"Filme", "Filme.en.srt", true, "en", "en", "srt", false},
	}
	for _, testCase := range testCases {
		t.Run(testCase.videoName+"|"+testCase.subtitleName, func(t *testing.T) {
			nameParts, isMatch := ParseSubtitleName(testCase.videoName, testCase.subtitleName)
			if isMatch != testCase.isMatch {
				t.Fatalf("expected match %v, got %v", testCase.isMatch, isMatch)
			}
			if !isMatch {
				return
			}
			subtitle := BuildSubtitleDto(7, nameParts)
			if subtitle.ID != 7 || subtitle.Label != testCase.label || subtitle.Language != testCase.language ||
				subtitle.Format != testCase.format || subtitle.IsForced != testCase.isForced {
				t.Fatalf("unexpected subtitle %+v", subtitle)
			}
		})
	}
}
