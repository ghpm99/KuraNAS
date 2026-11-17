package files

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"unicode"
)

const maxFileNameBytes = 255

const windowsReservedNameCharacters = `<>:"|?*`

func validateItemName(rawName string) (string, error) {
	trimmedName := strings.TrimSpace(rawName)
	if trimmedName == "" || trimmedName == "." || trimmedName == ".." {
		return "", invalidFileNameError(rawName)
	}
	if strings.Contains(trimmedName, "..") || strings.ContainsAny(trimmedName, `/\`) {
		return "", invalidFileNameError(rawName)
	}
	if len(trimmedName) > maxFileNameBytes {
		return "", invalidFileNameError(rawName)
	}
	for _, character := range trimmedName {
		if unicode.IsControl(character) || isReservedNameCharacter(character) {
			return "", invalidFileNameError(rawName)
		}
	}
	return trimmedName, nil
}

func isReservedNameCharacter(character rune) bool {
	return runtime.GOOS == "windows" && strings.ContainsRune(windowsReservedNameCharacters, character)
}

func invalidFileNameError(rawName string) *FileOperationError {
	return newFileOperationError(http.StatusBadRequest, "ERROR_FILE_NAME_INVALID", fmt.Errorf("invalid file name %q", rawName))
}
