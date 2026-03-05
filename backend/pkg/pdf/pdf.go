package pdf

import (
	"image"
	"os"
	"path/filepath"
	"runtime"
)

func Thumbnail() (image.Image, error) {
	_, filename, _, _ := runtime.Caller(0)

	filePath := filepath.Join(filepath.Dir(filename), "pdf.jpg")

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	img, _, err := image.Decode(file)

	if err != nil {
		return nil, err
	}

	return img, nil

}
