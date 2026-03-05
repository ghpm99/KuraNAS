package img

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
)

func Thumbnail(src image.Image, maxWidth, maxHeight uint) image.Image {
	if maxWidth == 0 {
		maxWidth = 320
	}
	if maxHeight == 0 {
		maxHeight = 320
	}

	bounds := src.Bounds()
	srcWidth := uint(bounds.Dx())
	srcHeight := uint(bounds.Dy())

	aspectRatio := float64(srcWidth) / float64(srcHeight)
	targetAspectRatio := float64(maxWidth) / float64(maxHeight)

	var newWidth, newHeight uint
	if aspectRatio > targetAspectRatio {
		newWidth = maxWidth
		newHeight = uint(float64(maxWidth) / aspectRatio)
	} else {
		newHeight = maxHeight
		newWidth = uint(float64(maxHeight) * aspectRatio)
	}

	resized := resizeBilinear(src, newWidth, newHeight)

	canvas := image.NewRGBA(image.Rect(0, 0, int(maxWidth), int(maxHeight)))

	x := (int(maxWidth) - resized.Bounds().Dx()) / 2
	y := (int(maxHeight) - resized.Bounds().Dy()) / 2

	draw.Draw(canvas, image.Rect(x, y, x+resized.Bounds().Dx(), y+resized.Bounds().Dy()), resized, image.Point{0, 0}, draw.Over)

	return canvas
}

func resizeBilinear(src image.Image, width, height uint) image.Image {
	bounds := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))

	for y := uint(0); y < height; y++ {
		for x := uint(0); x < width; x++ {
			srcX := float64(x) * float64(bounds.Dx()) / float64(width)
			srcY := float64(y) * float64(bounds.Dy()) / float64(height)

			x1 := int(srcX)
			y1 := int(srcY)
			x2 := x1 + 1
			y2 := y1 + 1

			if x2 >= bounds.Dx() {
				x2 = bounds.Dx() - 1
			}
			if y2 >= bounds.Dy() {
				y2 = bounds.Dy() - 1
			}

			dx := srcX - float64(x1)
			dy := srcY - float64(y1)

			p11 := src.At(x1, y1)
			p21 := src.At(x2, y1)
			p12 := src.At(x1, y2)
			p22 := src.At(x2, y2)

			r1, g1, b1, a1 := p11.RGBA()
			r2, g2, b2, a2 := p21.RGBA()
			r3, g3, b3, a3 := p12.RGBA()
			r4, g4, b4, a4 := p22.RGBA()

			r := uint32((float64(r1)*(1-dx)*(1-dy) + float64(r2)*dx*(1-dy) + float64(r3)*(1-dx)*dy + float64(r4)*dx*dy) / 65535.0 * 65535.0)
			g := uint32((float64(g1)*(1-dx)*(1-dy) + float64(g2)*dx*(1-dy) + float64(g3)*(1-dx)*dy + float64(g4)*dx*dy) / 65535.0 * 65535.0)
			b := uint32((float64(b1)*(1-dx)*(1-dy) + float64(b2)*dx*(1-dy) + float64(b3)*(1-dx)*dy + float64(b4)*dx*dy) / 65535.0 * 65535.0)
			a := uint32((float64(a1)*(1-dx)*(1-dy) + float64(a2)*dx*(1-dy) + float64(a3)*(1-dx)*dy + float64(a4)*dx*dy) / 65535.0 * 65535.0)

			dst.Set(int(x), int(y), color.NRGBA64{R: uint16(r), G: uint16(g), B: uint16(b), A: uint16(a)})
		}
	}

	return dst
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
