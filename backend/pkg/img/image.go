package img

import (
	"image"
	"image/jpeg"
	"image/png"
	"math"
	"os"
)

func Thumbnail(src image.Image) (image.Image, error) {

	var width = 652
	var height = 489

	bounds := src.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()

	scaleX := float64(width) / float64(srcWidth)
	scaleY := float64(height) / float64(srcHeight)
	scale := math.Min(scaleX, scaleY)

	newWidth := int(float64(srcWidth) * scale)
	newHeight := int(float64(srcHeight) * scale)

	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	for y := range newHeight {
		for x := range newWidth {
			srcX := int(float64(x) / scale)
			srcY := int(float64(y) / scale)

			dst.Set(x, y, src.At(srcX, srcY))
		}
	}

	return dst, nil
}

func OpenImageFromFile(path string, format string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()
	var img image.Image
	switch format {
	case ".jpg":
		img, err = jpeg.Decode(file)
	case ".png":
		img, err = png.Decode(file)
	default:
		img, _, err = image.Decode(file)
	}

	if err != nil {
		return nil, err
	}

	return img, nil
}
