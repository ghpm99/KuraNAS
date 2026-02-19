package img

import (
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
)

// Thumbnail gera uma thumbnail da imagem com largura especificada
// mantendo a proporção original. Altura é calculada automaticamente.
func Thumbnail(src image.Image, width uint) image.Image {
	if width == 0 {
		width = 320
	}

	// Calcula altura proporcional
	bounds := src.Bounds()
	height := uint(float64(bounds.Dy()) * float64(width) / float64(bounds.Dx()))

	// Cria nova imagem com as dimensões
	dst := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))

	// Implementação de resize bilinear simples
	for y := uint(0); y < height; y++ {
		for x := uint(0); x < width; x++ {
			// Coordenadas na imagem original
			srcX := float64(x) * float64(bounds.Dx()) / float64(width)
			srcY := float64(y) * float64(bounds.Dy()) / float64(height)

			// Bilinear interpolation
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

			// Pega os 4 pixels vizinhos
			p11 := src.At(x1, y1)
			p21 := src.At(x2, y1)
			p12 := src.At(x1, y2)
			p22 := src.At(x2, y2)

			// Interpola bilinear
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

// ThumbnailWithHeight gera uma thumbnail com dimensões exatas (crop ou fit)
func ThumbnailWithHeight(src image.Image, width, height uint, fit bool) image.Image {
	if width == 0 {
		width = 320
	}
	if height == 0 {
		height = 240
	}

	bounds := src.Bounds()
	var dst image.Image

	if fit {
		// Fit: mantém proporção, pode ter letterbox
		aspectRatio := float64(bounds.Dx()) / float64(bounds.Dy())
		targetAspectRatio := float64(width) / float64(height)

		var newWidth uint
		if aspectRatio > targetAspectRatio {
			newWidth = width
		} else {
			newWidth = uint(float64(height) * aspectRatio)
		}

		dst = Thumbnail(src, newWidth)

		// Cria imagem final com letterbox se necessário
		final := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
		// Fundo preto
		for y := 0; y < int(height); y++ {
			for x := 0; x < int(width); x++ {
				final.Set(x, y, image.Black)
			}
		}
		// Centraliza a imagem
		xOffset := (int(width) - dst.Bounds().Dx()) / 2
		yOffset := (int(height) - dst.Bounds().Dy()) / 2
		for y := 0; y < dst.Bounds().Dy(); y++ {
			for x := 0; x < dst.Bounds().Dx(); x++ {
				final.Set(x+xOffset, y+yOffset, dst.At(x, y))
			}
		}
		return final
	} else {
		// Cover: preenche todo o espaço, corta o excesso
		aspectRatio := float64(bounds.Dx()) / float64(bounds.Dy())
		targetAspectRatio := float64(width) / float64(height)

		var cropWidth, cropHeight int
		if aspectRatio > targetAspectRatio {
			// Imagem mais larga que o target - corta lados
			cropHeight = bounds.Dy()
			cropWidth = int(float64(bounds.Dy()) * targetAspectRatio)
		} else {
			// Imagem mais alta que o target - corta topo/baixo
			cropWidth = bounds.Dx()
			cropHeight = int(float64(bounds.Dx()) / targetAspectRatio)
		}

		// Centraliza o crop
		xOffset := (bounds.Dx() - cropWidth) / 2
		yOffset := (bounds.Dy() - cropHeight) / 2

		cropRect := image.Rect(xOffset, yOffset, xOffset+cropWidth, yOffset+cropHeight)
		cropped := image.NewRGBA(cropRect)
		for y := 0; y < cropHeight; y++ {
			for x := 0; x < cropWidth; x++ {
				cropped.Set(x, y, src.At(x+xOffset, y+yOffset))
			}
		}

		return Thumbnail(cropped, width)
	}
}

// OpenImageFromFile abre uma imagem de arquivo suportando múltiplos formatos
func OpenImageFromFile(path string) (image.Image, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	defer file.Close()

	// Detecta formato e decodifica
	img, format, err := image.Decode(file)
	if err != nil {
		return nil, "", err
	}

	return img, format, nil
}

// DecodeJPEG decodifica especificamente JPEG
func DecodeJPEG(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return jpeg.Decode(file)
}

// DecodePNG decodifica especificamente PNG
func DecodePNG(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return png.Decode(file)
}

// DecodeGIF decodifica especificamente GIF
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

// EncodeJPEG codifica imagem para JPEG com qualidade configurável
func EncodeJPEG(img image.Image, quality int) ([]byte, error) {
	if quality == 0 || quality > 100 {
		quality = 85
	}

	var buf []byte
	writer := &sliceWriter{data: &buf}
	err := jpeg.Encode(writer, img, &jpeg.Options{Quality: quality})
	return buf, err
}

type sliceWriter struct {
	data *[]byte
}

func (w *sliceWriter) Write(p []byte) (n int, err error) {
	*w.data = append(*w.data, p...)
	return len(p), nil
}
