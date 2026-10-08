package img

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"testing"

	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
)

func quadrantImage(width, height int) *image.RGBA {
	quadrants := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			quadrants.Set(x, y, quadrantColor(x < width/2, y < height/2))
		}
	}
	return quadrants
}

func quadrantColor(isLeft, isTop bool) color.RGBA {
	switch {
	case isLeft && isTop:
		return color.RGBA{R: 255, A: 255}
	case !isLeft && isTop:
		return color.RGBA{G: 255, A: 255}
	case isLeft && !isTop:
		return color.RGBA{B: 255, A: 255}
	}
	return color.RGBA{R: 255, G: 255, A: 255}
}

func dominantChannel(c color.Color) string {
	red, green, blue, _ := c.RGBA()
	switch {
	case red > 0x8000 && green > 0x8000 && blue < 0x8000:
		return "yellow"
	case red > 0x8000 && green < 0x8000 && blue < 0x8000:
		return "red"
	case green > 0x8000 && red < 0x8000 && blue < 0x8000:
		return "green"
	case blue > 0x8000 && red < 0x8000 && green < 0x8000:
		return "blue"
	}
	return "other"
}

func cornerNames(src image.Image) [4]string {
	bounds := src.Bounds()
	return [4]string{
		dominantChannel(src.At(bounds.Min.X+1, bounds.Min.Y+1)),
		dominantChannel(src.At(bounds.Max.X-2, bounds.Min.Y+1)),
		dominantChannel(src.At(bounds.Min.X+1, bounds.Max.Y-2)),
		dominantChannel(src.At(bounds.Max.X-2, bounds.Max.Y-2)),
	}
}

func TestFitWithinBoxPreservesAspectRatioWithoutCanvas(t *testing.T) {
	wide := image.NewRGBA(image.Rect(0, 0, 400, 200))

	fitted := FitWithinBox(wide, 100, 100, 1)

	if fitted.Bounds().Dx() != 100 || fitted.Bounds().Dy() != 50 {
		t.Fatalf("expected 100x50, got %dx%d", fitted.Bounds().Dx(), fitted.Bounds().Dy())
	}
}

func TestFitWithinBoxDoesNotUpscale(t *testing.T) {
	small := image.NewRGBA(image.Rect(0, 0, 40, 20))

	fitted := FitWithinBox(small, 400, 400, 1)

	if fitted.Bounds().Dx() != 40 || fitted.Bounds().Dy() != 20 {
		t.Fatalf("expected original 40x20, got %dx%d", fitted.Bounds().Dx(), fitted.Bounds().Dy())
	}
}

func TestFitWithinBoxAppliesOrientationTransforms(t *testing.T) {
	source := quadrantImage(40, 20)
	topLeft, topRight, bottomLeft, bottomRight := "red", "green", "blue", "yellow"

	testCases := []struct {
		orientation    int
		expectedWidth  int
		expectedHeight int
		expectedCorner [4]string
	}{
		{1, 40, 20, [4]string{topLeft, topRight, bottomLeft, bottomRight}},
		{2, 40, 20, [4]string{topRight, topLeft, bottomRight, bottomLeft}},
		{3, 40, 20, [4]string{bottomRight, bottomLeft, topRight, topLeft}},
		{4, 40, 20, [4]string{bottomLeft, bottomRight, topLeft, topRight}},
		{5, 20, 40, [4]string{topLeft, bottomLeft, topRight, bottomRight}},
		{6, 20, 40, [4]string{bottomLeft, topLeft, bottomRight, topRight}},
		{7, 20, 40, [4]string{bottomRight, topRight, bottomLeft, topLeft}},
		{8, 20, 40, [4]string{topRight, bottomRight, topLeft, bottomLeft}},
	}

	for _, testCase := range testCases {
		oriented := FitWithinBox(source, 100, 100, testCase.orientation)
		if oriented.Bounds().Dx() != testCase.expectedWidth || oriented.Bounds().Dy() != testCase.expectedHeight {
			t.Fatalf("orientation %d: expected %dx%d, got %dx%d", testCase.orientation, testCase.expectedWidth, testCase.expectedHeight, oriented.Bounds().Dx(), oriented.Bounds().Dy())
		}
		if corners := cornerNames(oriented); corners != testCase.expectedCorner {
			t.Fatalf("orientation %d: expected corners %v, got %v", testCase.orientation, testCase.expectedCorner, corners)
		}
	}
}

func TestFitWithinBoxFitsOrientedDimensionsInsideBox(t *testing.T) {
	landscape := image.NewRGBA(image.Rect(0, 0, 4000, 3000))

	fitted := FitWithinBox(landscape, 400, 400, 6)

	if fitted.Bounds().Dx() != 300 || fitted.Bounds().Dy() != 400 {
		t.Fatalf("expected 300x400 portrait, got %dx%d", fitted.Bounds().Dx(), fitted.Bounds().Dy())
	}
}

func TestEncodeJPEGProducesJPEG(t *testing.T) {
	encoded, err := EncodeJPEG(quadrantImage(16, 16))
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	if _, err := jpeg.Decode(bytes.NewReader(encoded)); err != nil {
		t.Fatalf("expected decodable jpeg: %v", err)
	}
}

func TestIsOpaqueDetectsTransparency(t *testing.T) {
	if !IsOpaque(quadrantImage(4, 4)) {
		t.Fatalf("expected opaque image")
	}
	translucent := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	translucent.Set(0, 0, color.NRGBA{R: 255, A: 10})
	if IsOpaque(translucent) {
		t.Fatalf("expected translucent image to be non-opaque")
	}
}

func TestOpenImageFromFileDecodesBMPAndTIFF(t *testing.T) {
	source := quadrantImage(8, 6)

	var bmpBuffer, tiffBuffer bytes.Buffer
	if err := bmp.Encode(&bmpBuffer, source); err != nil {
		t.Fatal(err)
	}
	if err := tiff.Encode(&tiffBuffer, source, nil); err != nil {
		t.Fatal(err)
	}

	for formatName, encoded := range map[string][]byte{"bmp": bmpBuffer.Bytes(), "tiff": tiffBuffer.Bytes()} {
		decoded, detectedFormat, err := image.Decode(bytes.NewReader(encoded))
		if err != nil {
			t.Fatalf("%s: decode failed: %v", formatName, err)
		}
		if detectedFormat != formatName || decoded.Bounds().Dx() != 8 {
			t.Fatalf("%s: unexpected decode result %s %v", formatName, detectedFormat, decoded.Bounds())
		}
	}
}

func TestWebPDecoderIsRegistered(t *testing.T) {
	tinyWebP := []byte{
		0x52, 0x49, 0x46, 0x46, 0x1a, 0x00, 0x00, 0x00, 0x57, 0x45, 0x42, 0x50, 0x56, 0x50, 0x38, 0x4c,
		0x0d, 0x00, 0x00, 0x00, 0x2f, 0x00, 0x00, 0x00, 0x10, 0x07, 0x10, 0x11, 0x11, 0x88, 0x88, 0xfe,
		0x07, 0x00,
	}

	decoded, detectedFormat, err := image.Decode(bytes.NewReader(tinyWebP))
	if err != nil {
		t.Fatalf("expected webp decoder to be registered: %v", err)
	}
	if detectedFormat != "webp" || decoded.Bounds().Dx() != 1 || decoded.Bounds().Dy() != 1 {
		t.Fatalf("unexpected webp decode: %s %v", detectedFormat, decoded.Bounds())
	}
}

func jpegWithExifOrientation(t *testing.T, orientation uint16, byteOrderTag string) []byte {
	t.Helper()
	var plainJPEG bytes.Buffer
	if err := jpeg.Encode(&plainJPEG, quadrantImage(8, 8), nil); err != nil {
		t.Fatal(err)
	}

	isBigEndian := byteOrderTag == "MM"
	putShort := func(value uint16) []byte {
		if isBigEndian {
			return []byte{byte(value >> 8), byte(value)}
		}
		return []byte{byte(value), byte(value >> 8)}
	}
	putLong := func(value uint32) []byte {
		if isBigEndian {
			return []byte{byte(value >> 24), byte(value >> 16), byte(value >> 8), byte(value)}
		}
		return []byte{byte(value), byte(value >> 8), byte(value >> 16), byte(value >> 24)}
	}

	var tiffBlock bytes.Buffer
	tiffBlock.WriteString(byteOrderTag)
	tiffBlock.Write(putShort(42))
	tiffBlock.Write(putLong(8))
	tiffBlock.Write(putShort(1))
	tiffBlock.Write(putShort(0x0112))
	tiffBlock.Write(putShort(3))
	tiffBlock.Write(putLong(1))
	tiffBlock.Write(putShort(orientation))
	tiffBlock.Write([]byte{0, 0})
	tiffBlock.Write(putLong(0))

	payload := append([]byte("Exif\x00\x00"), tiffBlock.Bytes()...)
	segmentLength := uint16(len(payload) + 2)

	var withExif bytes.Buffer
	withExif.Write([]byte{0xFF, 0xD8, 0xFF, 0xE1, byte(segmentLength >> 8), byte(segmentLength)})
	withExif.Write(payload)
	withExif.Write(plainJPEG.Bytes()[2:])
	return withExif.Bytes()
}

func TestParseJPEGOrientationReadsBothByteOrders(t *testing.T) {
	for _, byteOrderTag := range []string{"II", "MM"} {
		for orientation := uint16(1); orientation <= 8; orientation++ {
			encoded := jpegWithExifOrientation(t, orientation, byteOrderTag)
			if got := parseJPEGOrientation(encoded); got != int(orientation) {
				t.Fatalf("%s: expected orientation %d, got %d", byteOrderTag, orientation, got)
			}
		}
	}
}

func TestParseJPEGOrientationDefaultsWhenAbsentOrInvalid(t *testing.T) {
	var plainJPEG bytes.Buffer
	_ = jpeg.Encode(&plainJPEG, quadrantImage(8, 8), nil)
	var plainPNG bytes.Buffer
	_ = png.Encode(&plainPNG, quadrantImage(4, 4))

	invalidInputs := [][]byte{nil, plainJPEG.Bytes(), plainPNG.Bytes(), jpegWithExifOrientation(t, 9, "II"), jpegWithExifOrientation(t, 6, "II")[:10]}
	for index, invalidInput := range invalidInputs {
		if got := parseJPEGOrientation(invalidInput); got != 1 {
			t.Fatalf("input %d: expected default orientation 1, got %d", index, got)
		}
	}
}

func TestReadJPEGOrientationFromFile(t *testing.T) {
	path := t.TempDir() + "/rotated.jpg"
	if err := writeFile(path, jpegWithExifOrientation(t, 6, "MM")); err != nil {
		t.Fatal(err)
	}
	if got := ReadJPEGOrientation(path); got != 6 {
		t.Fatalf("expected 6, got %d", got)
	}
	if got := ReadJPEGOrientation(path + ".missing"); got != 1 {
		t.Fatalf("expected default 1 for missing file, got %d", got)
	}
}

func writeFile(path string, content []byte) error {
	return os.WriteFile(path, content, 0644)
}
