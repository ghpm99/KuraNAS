package icons

import (
	"fmt"
	"image"
	"nas-go/api/pkg/img"
	"os"
	"strings"
)

func getIconPath(iconName string) (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	filePath := fmt.Sprintf("%s/icons/%s.png", currentDir, iconName)
	return filePath, nil
}

func getIcon(format string) (image.Image, error) {

	path, err := getIconPath(strings.ToLower(format))
	if err != nil {
		return nil, err
	}
	return img.OpenImageFromFile(path, ".png")
}

func PdfIcon() (image.Image, error) {
	return getIcon("pdf")
}

func FolderIcon() (image.Image, error) {
	return getIcon("folder")
}
func Mp3Icon() (image.Image, error) {
	return getIcon("mp3")
}
func Mp4Icon() (image.Image, error) {
	return getIcon("mp4")
}
func Icon() (image.Image, error) {
	return getIcon("unknown")
}
