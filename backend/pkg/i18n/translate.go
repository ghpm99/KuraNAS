package i18n

import (
	"encoding/json"
	"fmt"
	"log"
	"nas-go/api/internal/config"
	"os"
	"path/filepath"
)

var translations map[string]string

func GetPathFileTranslate() string {
	var lang = config.AppConfig.Lang

	if lang == "" {
		lang = "en-US"
	}

	return GetPathFileTranslateByLang(lang)
}

func GetPathFileTranslateByLang(lang string) string {
	if lang == "" {
		lang = "en-US"
	}

	translationsPath := ResolveTranslationsPath()
	return fmt.Sprintf("%s%s.json", translationsPath, lang)
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
	filePath := GetPathFileTranslate()

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
