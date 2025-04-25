package i18n

import (
	"encoding/json"
	"fmt"
	"nas-go/api/internal/config"
	"os"
)

var translations map[string]string

func LoadTranslations() error {
	var lang = config.AppConfig.Lang
	if lang == "" {
		lang = "en-US"
	}
	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}

	filePath := fmt.Sprintf("%s/translations/%s.json", currentDir, lang)

	file, err := os.Open(filePath)
	fmt.Println("Loading translations from: " + filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(&translations)
}

func Translate(key string, args ...any) string {
	if msg, ok := translations[key]; ok {
		return fmt.Sprintf(msg, args...)
	}

	return key

}

func PrintTranslate(key string, args ...any) {
	fmt.Println(Translate(key, args...))
}
