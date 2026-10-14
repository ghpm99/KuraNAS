package video

import (
	"regexp"
	"strings"
)

const (
	SubtitleFormatSrt = "srt"
	SubtitleFormatVtt = "vtt"
)

var subtitleExtensions = []string{"." + SubtitleFormatSrt, "." + SubtitleFormatVtt}

var languageTagPattern = regexp.MustCompile(`^([a-z]{2,3})(-([a-z]{4}))?(-([a-z]{2}))?$`)

var twoLetterLanguageCodes = map[string]bool{
	"ar": true, "bg": true, "ca": true, "cs": true, "da": true, "de": true, "el": true, "en": true,
	"es": true, "et": true, "fa": true, "fi": true, "fr": true, "he": true, "hi": true, "hr": true,
	"hu": true, "id": true, "it": true, "ja": true, "ko": true, "lt": true, "lv": true, "nl": true,
	"no": true, "pl": true, "pt": true, "ro": true, "ru": true, "sk": true, "sl": true, "sr": true,
	"sv": true, "th": true, "tr": true, "uk": true, "vi": true, "zh": true,
}

var threeLetterLanguageCodes = map[string]string{
	"ara": "ar", "bul": "bg", "cat": "ca", "ces": "cs", "cze": "cs", "dan": "da", "deu": "de",
	"ger": "de", "ell": "el", "gre": "el", "eng": "en", "spa": "es", "est": "et", "fas": "fa",
	"per": "fa", "fin": "fi", "fra": "fr", "fre": "fr", "heb": "he", "hin": "hi", "hrv": "hr",
	"hun": "hu", "ind": "id", "ita": "it", "jpn": "ja", "kor": "ko", "lit": "lt", "lav": "lv",
	"nld": "nl", "dut": "nl", "nor": "no", "pol": "pl", "por": "pt", "ron": "ro", "rum": "ro",
	"rus": "ru", "slk": "sk", "slo": "sk", "slv": "sl", "srp": "sr", "swe": "sv", "tha": "th",
	"tur": "tr", "ukr": "uk", "vie": "vi", "zho": "zh", "chi": "zh",
}

type SubtitleNameParts struct {
	Label  string
	Format string
}

func SubtitleNamePrefixFor(videoName string) string {
	return strings.ToLower(videoBaseName(videoName)) + "."
}

func videoBaseName(videoName string) string {
	extensionStart := strings.LastIndex(videoName, ".")
	if extensionStart <= 0 {
		return videoName
	}
	return videoName[:extensionStart]
}

func SubtitleFormatOfName(fileName string) (string, bool) {
	lowerName := strings.ToLower(fileName)
	for _, extension := range subtitleExtensions {
		if strings.HasSuffix(lowerName, extension) {
			return strings.TrimPrefix(extension, "."), true
		}
	}
	return "", false
}

func ParseSubtitleName(videoName string, subtitleName string) (SubtitleNameParts, bool) {
	subtitleFormat, isSubtitle := SubtitleFormatOfName(subtitleName)
	if !isSubtitle {
		return SubtitleNameParts{}, false
	}
	nameWithoutExtension := subtitleName[:len(subtitleName)-len(subtitleFormat)-1]
	videoBase := videoBaseName(videoName)
	if strings.EqualFold(nameWithoutExtension, videoBase) {
		return SubtitleNameParts{Format: subtitleFormat}, true
	}
	labelSeparator := len(videoBase) + 1
	if len(nameWithoutExtension) <= labelSeparator ||
		!strings.EqualFold(nameWithoutExtension[:labelSeparator], videoBase+".") {
		return SubtitleNameParts{}, false
	}
	return SubtitleNameParts{Label: nameWithoutExtension[labelSeparator:], Format: subtitleFormat}, true
}

func BuildSubtitleDto(subtitleID int, nameParts SubtitleNameParts) SubtitleDto {
	return SubtitleDto{
		ID:       subtitleID,
		Label:    nameParts.Label,
		Language: extractSubtitleLanguage(nameParts.Label),
		Format:   nameParts.Format,
		IsForced: strings.Contains(strings.ToLower(nameParts.Label), "forced"),
	}
}

func extractSubtitleLanguage(label string) string {
	for _, labelToken := range strings.Split(label, ".") {
		if language, isLanguage := normalizeLanguageTag(labelToken); isLanguage {
			return language
		}
	}
	return ""
}

func normalizeLanguageTag(token string) (string, bool) {
	tagParts := languageTagPattern.FindStringSubmatch(strings.ToLower(strings.ReplaceAll(token, "_", "-")))
	if tagParts == nil {
		return "", false
	}
	primaryLanguage, isKnown := resolvePrimaryLanguage(tagParts[1])
	if !isKnown {
		return "", false
	}
	normalizedTag := primaryLanguage
	if tagParts[3] != "" {
		normalizedTag += "-" + strings.ToUpper(tagParts[3][:1]) + tagParts[3][1:]
	}
	if tagParts[5] != "" {
		normalizedTag += "-" + strings.ToUpper(tagParts[5])
	}
	return normalizedTag, true
}

func resolvePrimaryLanguage(code string) (string, bool) {
	if len(code) == 2 {
		return code, twoLetterLanguageCodes[code]
	}
	twoLetterCode, isKnown := threeLetterLanguageCodes[code]
	return twoLetterCode, isKnown
}
