package video

import "testing"

func TestConvertSubtitleToWebVtt(t *testing.T) {
	testCases := []struct {
		name           string
		subtitleText   string
		subtitleFormat string
		expected       string
	}{
		{
			name:           "simple srt drops sequence numbers and converts comma timestamps",
			subtitleText:   "1\n00:00:01,500 --> 00:00:03,000\nOla\n\n2\n00:00:04,000 --> 00:00:05,250\nMundo\nlinha dois\n",
			subtitleFormat: SubtitleFormatSrt,
			expected:       "WEBVTT\n\n00:00:01.500 --> 00:00:03.000\nOla\n\n00:00:04.000 --> 00:00:05.250\nMundo\nlinha dois\n",
		},
		{
			name:           "srt tolerates crlf and extra blank lines",
			subtitleText:   "\r\n1\r\n00:00:01,000 --> 00:00:02,000\r\nUm\r\n\r\n\r\n\r\n2\r\n00:00:03,000 --> 00:00:04,000\r\nDois\r\n\r\n",
			subtitleFormat: SubtitleFormatSrt,
			expected:       "WEBVTT\n\n00:00:01.000 --> 00:00:02.000\nUm\n\n00:00:03.000 --> 00:00:04.000\nDois\n",
		},
		{
			name:           "srt without sequence numbers still converts",
			subtitleText:   "00:00:01,000 --> 00:00:02,000\nUm\n",
			subtitleFormat: SubtitleFormatSrt,
			expected:       "WEBVTT\n\n00:00:01.000 --> 00:00:02.000\nUm\n",
		},
		{
			name:           "srt text containing commas is untouched",
			subtitleText:   "1\n00:00:01,000 --> 00:00:02,000\nSim, claro 00:00:09,123\n",
			subtitleFormat: SubtitleFormatSrt,
			expected:       "WEBVTT\n\n00:00:01.000 --> 00:00:02.000\nSim, claro 00:00:09,123\n",
		},
		{
			name:           "empty srt yields header only",
			subtitleText:   "",
			subtitleFormat: SubtitleFormatSrt,
			expected:       "WEBVTT\n",
		},
		{
			name:           "vtt passes through",
			subtitleText:   "WEBVTT\n\n00:00:01.000 --> 00:00:02.000\nUm\n",
			subtitleFormat: SubtitleFormatVtt,
			expected:       "WEBVTT\n\n00:00:01.000 --> 00:00:02.000\nUm\n",
		},
		{
			name:           "vtt with crlf is normalized",
			subtitleText:   "WEBVTT\r\n\r\n00:00:01.000 --> 00:00:02.000\r\nUm\r\n",
			subtitleFormat: SubtitleFormatVtt,
			expected:       "WEBVTT\n\n00:00:01.000 --> 00:00:02.000\nUm\n",
		},
		{
			name:           "vtt without header receives one",
			subtitleText:   "00:00:01.000 --> 00:00:02.000\nUm\n",
			subtitleFormat: SubtitleFormatVtt,
			expected:       "WEBVTT\n\n00:00:01.000 --> 00:00:02.000\nUm\n",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if converted := ConvertSubtitleToWebVtt(testCase.subtitleText, testCase.subtitleFormat); converted != testCase.expected {
				t.Fatalf("expected %q, got %q", testCase.expected, converted)
			}
		})
	}
}

func TestDecodeSubtitleText(t *testing.T) {
	testCases := []struct {
		name     string
		raw      []byte
		expected string
	}{
		{"utf8 passes", []byte("ação"), "ação"},
		{"utf8 bom is stripped", append([]byte{0xEF, 0xBB, 0xBF}, []byte("ok")...), "ok"},
		{"windows-1252 accents", []byte{0x41, 0xE7, 0xE3, 0x6F, 0x20, 0xE9}, "Ação é"},
		{"windows-1252 punctuation", []byte{0x93, 0x61, 0x94, 0x85, 0x80}, "“a”…€"},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if decoded := DecodeSubtitleText(testCase.raw); decoded != testCase.expected {
				t.Fatalf("expected %q, got %q", testCase.expected, decoded)
			}
		})
	}
}
