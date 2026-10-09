package img

import (
	"bytes"
	"image"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"math"
	"os"

	_ "golang.org/x/image/bmp"
	xdraw "golang.org/x/image/draw"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

const (
	defaultThumbnailBoxSize   = 320
	photoThumbnailJPEGQuality = 82
)

func Thumbnail(src image.Image, maxWidth, maxHeight uint) image.Image {
	if maxWidth == 0 {
		maxWidth = defaultThumbnailBoxSize
	}
	if maxHeight == 0 {
		maxHeight = defaultThumbnailBoxSize
	}

	fittedWidth, fittedHeight := fitDimensions(src.Bounds().Dx(), src.Bounds().Dy(), int(maxWidth), int(maxHeight), true)
	resized := scaleTo(src, fittedWidth, fittedHeight)

	canvas := image.NewRGBA(image.Rect(0, 0, int(maxWidth), int(maxHeight)))
	offsetX := (int(maxWidth) - fittedWidth) / 2
	offsetY := (int(maxHeight) - fittedHeight) / 2
	draw.Draw(canvas, image.Rect(offsetX, offsetY, offsetX+fittedWidth, offsetY+fittedHeight), resized, image.Point{}, draw.Over)

	return canvas
}

func FitWithinBox(src image.Image, maxWidth, maxHeight int, orientation int) image.Image {
	sourceWidth, sourceHeight := src.Bounds().Dx(), src.Bounds().Dy()
	isTransposing := orientationSwapsAxes(orientation)

	orientedWidth, orientedHeight := sourceWidth, sourceHeight
	if isTransposing {
		orientedWidth, orientedHeight = sourceHeight, sourceWidth
	}

	fittedWidth, fittedHeight := fitDimensions(orientedWidth, orientedHeight, maxWidth, maxHeight, false)
	if isTransposing {
		fittedWidth, fittedHeight = fittedHeight, fittedWidth
	}

	return applyOrientation(scaleTo(src, fittedWidth, fittedHeight), orientation)
}

func fitDimensions(sourceWidth, sourceHeight, maxWidth, maxHeight int, canUpscale bool) (int, int) {
	if sourceWidth <= 0 || sourceHeight <= 0 {
		return 1, 1
	}
	if !canUpscale && sourceWidth <= maxWidth && sourceHeight <= maxHeight {
		return sourceWidth, sourceHeight
	}

	widthScale := float64(maxWidth) / float64(sourceWidth)
	heightScale := float64(maxHeight) / float64(sourceHeight)
	scale := math.Min(widthScale, heightScale)

	return max(1, int(math.Round(float64(sourceWidth)*scale))), max(1, int(math.Round(float64(sourceHeight)*scale)))
}

func scaleTo(src image.Image, width, height int) *image.RGBA {
	scaled := image.NewRGBA(image.Rect(0, 0, width, height))
	xdraw.CatmullRom.Scale(scaled, scaled.Bounds(), src, src.Bounds(), xdraw.Src, nil)
	return scaled
}

func IsOpaque(src image.Image) bool {
	opacityReporter, isReporter := src.(interface{ Opaque() bool })
	if !isReporter {
		return true
	}
	return opacityReporter.Opaque()
}

func OpenImageFromFile(path string) (image.Image, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		return nil, "", err
	}

	return img, format, nil
}

func DecodeJPEG(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return jpeg.Decode(file)
}

func DecodePNG(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return png.Decode(file)
}

func DecodeGIF(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, err := gif.Decode(file)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func EncodePNG(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	return buf.Bytes(), err
}

func EncodeJPEG(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: photoThumbnailJPEGQuality})
	return buf.Bytes(), err
}
