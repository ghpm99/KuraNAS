package utils

import "strings"

var likeWildcardEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

func EscapeLikeWildcards(text string) string {
	return likeWildcardEscaper.Replace(text)
}

func BuildContainsLikePattern(text string) string {
	return "%" + EscapeLikeWildcards(text) + "%"
}

func BuildPrefixLikePattern(text string) string {
	return EscapeLikeWildcards(text) + "%"
}
