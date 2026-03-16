package i18n

import (
	"encoding/json"
	"fmt"
	"log"
	"nas-go/api/internal/config"
	"os"
	"path/filepath"
	"strings"
)

var translations map[string]string

const defaultLocale = "en-US"

func GetPathFileTranslate() string {
	var lang = config.AppConfig.Lang

	filePath, err := resolveTranslationFilePath(lang)
	if err == nil {
		return filePath
	}

	return filepath.Join(ResolveTranslationsPath(), defaultLocale+".json")
}

func GetPathFileTranslateByLang(lang string) string {
	filePath, err := resolveTranslationFilePath(lang)
	if err == nil {
		return filePath
	}

	return filepath.Join(ResolveTranslationsPath(), defaultLocale+".json")
}

func ResolveTranslationsPath() string {
	translationsPath := config.GetBuildConfig("TranslationsPath")
	if info, err := os.Stat(translationsPath); err == nil && info.IsDir() {
		return translationsPath
	}

	fallbackPath, ok := findFallbackTranslationsPath()
	if ok {
		return fallbackPath
	}

	return translationsPath
}

func LoadTranslations() error {
	filePath, err := resolveTranslationFilePath(config.AppConfig.Lang)
	if err != nil {
		return err
	}

	file, err := os.Open(filePath)
	log.Println("Loading translations from: " + filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(&translations)
}

func GetMessage(key string) string {
	if msg, ok := translations[key]; ok {
		return msg
	}

	return key
}

func Translate(key string, args ...any) string {
	msg := GetMessage(key)
	if msg != key && len(args) > 0 {
		return fmt.Sprintf(msg, args...)
	}

	return msg

}

func PrintTranslate(key string, args ...any) {
	fmt.Println(Translate(key, args...))
}

func LogTranslate(key string, args ...any) {
	log.Println(Translate(key, args...))
}

func resolveTranslationFilePath(lang string) (string, error) {
	translationsPath, err := filepath.Abs(ResolveTranslationsPath())
	if err != nil {
		return "", fmt.Errorf("resolve translations directory: %w", err)
	}

	sanitizedLang := sanitizeLocale(lang)
	filePath, err := filepath.Abs(filepath.Join(translationsPath, sanitizedLang+".json"))
	if err != nil {
		return "", fmt.Errorf("resolve translation file path: %w", err)
	}

	if !isPathWithinDirectory(filePath, translationsPath) {
		return "", fmt.Errorf("translation file path escapes translations directory")
	}

	return filePath, nil
}

func sanitizeLocale(lang string) string {
	trimmedLang := strings.TrimSpace(lang)
	if trimmedLang == "" {
		return defaultLocale
	}

	if strings.Contains(trimmedLang, "..") || strings.ContainsAny(trimmedLang, `/\`) {
		return defaultLocale
	}

	for _, char := range trimmedLang {
		if isLocaleCharacter(char) {
			continue
		}
		return defaultLocale
	}

	return trimmedLang
}

func isLocaleCharacter(char rune) bool {
	switch {
	case char >= 'a' && char <= 'z':
		return true
	case char >= 'A' && char <= 'Z':
		return true
	case char >= '0' && char <= '9':
		return true
	case char == '-' || char == '_':
		return true
	default:
		return false
	}
}

func isPathWithinDirectory(path string, directory string) bool {
	relativePath, err := filepath.Rel(directory, path)
	if err != nil {
		return false
	}

	return relativePath != ".." && !strings.HasPrefix(relativePath, ".."+string(os.PathSeparator))
}

func findFallbackTranslationsPath() (string, bool) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", false
	}

	for {
		candidates := []string{
			filepath.Join(currentDir, "translations"),
			filepath.Join(currentDir, "backend", "translations"),
		}
		for _, candidate := range candidates {
			if info, err := os.Stat(candidate); err == nil && info.IsDir() {
				return candidate + string(os.PathSeparator), true
			}
		}

		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			break
		}
		currentDir = parentDir
	}

	return "", false
}
