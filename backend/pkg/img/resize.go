package img

import (
	"fmt"
	"image"
	"os"
	"slices"
)

var suportedFormats = []string{
	".jpg",
	".png",
}

func ResizeFromFile(path string, format string) (image.Image, error) {
	if !slices.Contains(suportedFormats, format) {
		return nil, fmt.Errorf("formato nao suportado")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	image, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	thumbnail := Resize(image, 320, 0)

	switch format {
	case ".jpg":
		return thumbnail, nil

	}
}

func Resize(img image.Image, width int, height int) image.Image {

	return img
}
