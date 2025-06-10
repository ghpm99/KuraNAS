package i18n

import (
	"encoding/json"
	"fmt"
	"log"
	"nas-go/api/internal/config"
	"os"
)

var translations map[string]string

func GetPathFileTranslate() (string, error) {
	var lang = config.AppConfig.Lang

	if lang == "" {
		lang = "en-US"
	}

	translationsPath := config.GetBuildConfig("TranslationsPath")

	filePath := fmt.Sprintf("%s%s.json", translationsPath, lang)
	return filePath, nil
}

func LoadTranslations() error {
	filePath, err := GetPathFileTranslate()
	if err != nil {
		log.Println("Error getting translation file path: " + err.Error())
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
