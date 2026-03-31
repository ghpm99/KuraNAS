package libraries

import (
	"log"
	"nas-go/api/pkg/i18n"
	"os"
	"path/filepath"
	"strings"
)

var categorySlugs = map[LibraryCategory][]string{
	LibraryCategoryImages:    {"Imagens", "Images", "Pictures", "Fotos", "Photos", "Pics", "Fotografias"},
	LibraryCategoryMusic:     {"Musicas", "Music", "Songs", "Audio", "Audios", "Musics"},
	LibraryCategoryVideos:    {"Videos", "Movies", "Filmes", "Gravacoes", "Recordings"},
	LibraryCategoryDocuments: {"Documentos", "Documents", "Docs"},
}

var categoryExtensions = map[LibraryCategory][]string{
	LibraryCategoryImages:    {".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".svg", ".heic", ".heif", ".raw", ".cr2", ".nef"},
	LibraryCategoryMusic:     {".mp3", ".flac", ".wav", ".aac", ".ogg", ".wma", ".m4a", ".opus"},
	LibraryCategoryVideos:    {".mp4", ".mkv", ".avi", ".mov", ".wmv", ".flv", ".webm", ".m4v", ".ts"},
	LibraryCategoryDocuments: {".pdf", ".txt", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".odt", ".ods", ".odp", ".csv", ".md"},
}

var categoryDefaultFolder = map[LibraryCategory]string{
	LibraryCategoryImages:    "Imagens",
	LibraryCategoryMusic:     "Musicas",
	LibraryCategoryVideos:    "Videos",
	LibraryCategoryDocuments: "Documentos",
}

func resolveLibraryPath(entryPoint string, category LibraryCategory) string {
	slugs := categorySlugs[category]
	if match := findSlugMatch(entryPoint, slugs); match != "" {
		return match
	}

	extensions := categoryExtensions[category]
	if match := findBestMatchByFileCount(entryPoint, extensions); match != "" {
		return match
	}

	defaultName := categoryDefaultFolder[category]
	defaultPath := filepath.Join(entryPoint, defaultName)
	if err := os.MkdirAll(defaultPath, 0755); err != nil {
		log.Printf("failed to create default library folder %s: %v", defaultPath, err)
		return defaultPath
	}
	log.Println(i18n.Translate("LIBRARY_CREATED", string(category), defaultPath))

	return defaultPath
}

func findSlugMatch(entryPoint string, slugs []string) string {
	entries, err := os.ReadDir(entryPoint)
	if err != nil {
		return ""
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		for _, slug := range slugs {
			if strings.EqualFold(name, slug) {
				return filepath.Join(entryPoint, name)
			}
		}
	}

	return ""
}

func findBestMatchByFileCount(entryPoint string, extensions []string) string {
	entries, err := os.ReadDir(entryPoint)
	if err != nil {
		return ""
	}

	extSet := make(map[string]struct{}, len(extensions))
	for _, ext := range extensions {
		extSet[strings.ToLower(ext)] = struct{}{}
	}

	var bestPath string
	var bestCount int

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dirPath := filepath.Join(entryPoint, entry.Name())
		count := countFilesByExtension(dirPath, extSet)

		if count > bestCount {
			bestCount = count
			bestPath = dirPath
		}
	}

	return bestPath
}

func countFilesByExtension(dirPath string, extSet map[string]struct{}) int {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return 0
	}

	count := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if _, ok := extSet[ext]; ok {
			count++
		}
	}

	return count
}
