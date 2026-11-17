package engine

import "strings"

func isSameOrDescendantPath(directory string, candidate string) bool {
	normalizedDirectory := normalizeSeparators(directory)
	normalizedCandidate := normalizeSeparators(candidate)
	if normalizedCandidate == normalizedDirectory {
		return true
	}
	return strings.HasPrefix(normalizedCandidate, normalizedDirectory+"/")
}

func normalizeSeparators(path string) string {
	forwardSlashed := strings.ReplaceAll(path, `\`, "/")
	return strings.TrimRight(forwardSlashed, "/")
}
