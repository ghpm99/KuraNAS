package img

import (
	"bytes"
	"encoding/binary"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
)

func encodedTestJPEG(t *testing.T, width, height int) []byte {
	t.Helper()
	canvas := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			canvas.Set(x, y, color.RGBA{R: uint8(x * 3), G: uint8(y * 3), B: uint8(x + y), A: 255})
		}
	}
	var buffer bytes.Buffer
	if err := jpeg.Encode(&buffer, canvas, nil); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func losslessJPEGStub() []byte {
	return []byte{
		0xFF, 0xD8,
		0xFF, 0xC3, 0x00, 0x0B, 0x08, 0x00, 0x10, 0x00, 0x10, 0x01, 0x01, 0x11, 0x00,
		0xFF, 0xDA, 0x00, 0x08, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00,
		0x12, 0x34, 0x56,
		0xFF, 0xD9,
	}
}

type tiffFixture struct {
	byteOrder      binary.ByteOrder
	orientation    uint16
	subIfdPreviews [][]byte
	losslessStrip  []byte
}

func (fixture tiffFixture) build() []byte {
	order := fixture.byteOrder
	subIfdCount := len(fixture.subIfdPreviews)
	const ifd0Size = 2 + 5*12 + 4
	const subIfdSize = 2 + 2*12 + 4
	subOffsetsPosition := 8 + ifd0Size
	firstSubIfdPosition := subOffsetsPosition + 4*subIfdCount
	dataStart := firstSubIfdPosition + subIfdSize*subIfdCount

	previewOffsets := make([]int, subIfdCount)
	nextDataPosition := dataStart + len(fixture.losslessStrip)
	for index, preview := range fixture.subIfdPreviews {
		previewOffsets[index] = nextDataPosition
		nextDataPosition += len(preview)
	}

	output := make([]byte, 8)
	if order == binary.LittleEndian {
		copy(output, "II")
	} else {
		copy(output, "MM")
	}
	order.PutUint16(output[2:], 42)
	order.PutUint32(output[4:], 8)

	appendEntry := func(tag, fieldType uint16, count uint32, value uint32) {
		entry := make([]byte, 12)
		order.PutUint16(entry[0:], tag)
		order.PutUint16(entry[2:], fieldType)
		order.PutUint32(entry[4:], count)
		if fieldType == 3 {
			order.PutUint16(entry[8:], uint16(value))
		} else {
			order.PutUint32(entry[8:], value)
		}
		output = append(output, entry...)
	}
	appendUint16 := func(value uint16) {
		raw := make([]byte, 2)
		order.PutUint16(raw, value)
		output = append(output, raw...)
	}
	appendUint32 := func(value uint32) {
		raw := make([]byte, 4)
		order.PutUint32(raw, value)
		output = append(output, raw...)
	}

	appendUint16(5)
	appendEntry(0x103, 3, 1, 7)
	appendEntry(0x111, 4, 1, uint32(dataStart))
	appendEntry(0x112, 3, 1, uint32(fixture.orientation))
	appendEntry(0x117, 4, 1, uint32(len(fixture.losslessStrip)))
	subIfdsValue := subOffsetsPosition
	if subIfdCount == 1 {
		subIfdsValue = firstSubIfdPosition
	}
	appendEntry(0x14A, 4, uint32(subIfdCount), uint32(subIfdsValue))
	appendUint32(0)

	for index := 0; index < subIfdCount; index++ {
		appendUint32(uint32(firstSubIfdPosition + index*subIfdSize))
	}
	for index, preview := range fixture.subIfdPreviews {
		appendUint16(2)
		appendEntry(0x201, 4, 1, uint32(previewOffsets[index]))
		appendEntry(0x202, 4, 1, uint32(len(preview)))
		appendUint32(0)
	}

	output = append(output, fixture.losslessStrip...)
	for _, preview := range fixture.subIfdPreviews {
		output = append(output, preview...)
	}
	return output
}

func extractFromBytes(t *testing.T, content []byte) (EmbeddedPreview, error) {
	t.Helper()
	return extractEmbeddedPreview(bytes.NewReader(content), int64(len(content)))
}

func TestExtractEmbeddedPreviewWalksTiffSubIfds(t *testing.T) {
	thumbnailJPEG := encodedTestJPEG(t, 16, 12)
	fullSizeJPEG := encodedTestJPEG(t, 120, 80)

	for name, byteOrder := range map[string]binary.ByteOrder{"little endian": binary.LittleEndian, "big endian": binary.BigEndian} {
		t.Run(name, func(t *testing.T) {
			content := tiffFixture{
				byteOrder:      byteOrder,
				orientation:    6,
				subIfdPreviews: [][]byte{thumbnailJPEG, fullSizeJPEG},
				losslessStrip:  losslessJPEGStub(),
			}.build()

			preview, err := extractFromBytes(t, content)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(preview.JPEG, fullSizeJPEG) {
				t.Fatalf("expected the largest embedded jpeg (%d bytes), got %d bytes", len(fullSizeJPEG), len(preview.JPEG))
			}
			if preview.Orientation != 6 {
				t.Fatalf("expected raw orientation 6, got %d", preview.Orientation)
			}
		})
	}
}

func TestExtractEmbeddedPreviewRejectsLosslessRawStream(t *testing.T) {
	content := tiffFixture{byteOrder: binary.LittleEndian, orientation: 1, losslessStrip: losslessJPEGStub()}.build()

	if _, err := extractFromBytes(t, content); !errors.Is(err, ErrNoEmbeddedPreview) {
		t.Fatalf("expected ErrNoEmbeddedPreview for lossless-only raw, got %v", err)
	}
}

func TestExtractEmbeddedPreviewPrefersLargerScannedJPEGOverSmallStructuredOne(t *testing.T) {
	thumbnailJPEG := encodedTestJPEG(t, 16, 12)
	hiddenFullSizeJPEG := encodedTestJPEG(t, 200, 150)
	content := tiffFixture{byteOrder: binary.LittleEndian, orientation: 1, subIfdPreviews: [][]byte{thumbnailJPEG}}.build()
	content = append(content, hiddenFullSizeJPEG...)

	preview, err := extractFromBytes(t, content)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(preview.JPEG, hiddenFullSizeJPEG) {
		t.Fatalf("expected scan to find the larger jpeg beyond the ifd tree, got %d bytes", len(preview.JPEG))
	}
}

func TestExtractEmbeddedPreviewScansUnstructuredContainer(t *testing.T) {
	previewJPEG := encodedTestJPEG(t, 90, 60)
	noise := make([]byte, 4096)
	rand.New(rand.NewSource(1)).Read(noise)
	content := append(append(append([]byte("ftypcrx "), noise...), previewJPEG...), noise...)

	preview, err := extractFromBytes(t, content)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(preview.JPEG, previewJPEG) {
		t.Fatalf("expected the embedded jpeg, got %d bytes", len(preview.JPEG))
	}
	if preview.Orientation != orientationTopLeft {
		t.Fatalf("expected default orientation, got %d", preview.Orientation)
	}
}

func TestScanIgnoresNestedThumbnailInsideOuterJPEG(t *testing.T) {
	innerThumbnail := encodedTestJPEG(t, 8, 8)
	outerJPEG := encodedTestJPEG(t, 100, 100)
	segmentLength := uint16(len(innerThumbnail) + 2)
	nestedSegment := append([]byte{0xFF, 0xE1, byte(segmentLength >> 8), byte(segmentLength)}, innerThumbnail...)
	outerWithNestedThumbnail := append(append(append([]byte{}, outerJPEG[:2]...), nestedSegment...), outerJPEG[2:]...)

	preview, err := extractFromBytes(t, append([]byte("junk"), outerWithNestedThumbnail...))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(preview.JPEG, outerWithNestedThumbnail) {
		t.Fatalf("expected the whole outer jpeg (%d bytes), got %d bytes", len(outerWithNestedThumbnail), len(preview.JPEG))
	}
	if _, err := jpeg.Decode(bytes.NewReader(preview.JPEG)); err != nil {
		t.Fatalf("extracted jpeg must decode: %v", err)
	}
}

func TestExtractEmbeddedPreviewIgnoresTruncatedJPEG(t *testing.T) {
	fullJPEG := encodedTestJPEG(t, 80, 80)
	truncated := append([]byte("header"), fullJPEG[:len(fullJPEG)/2]...)

	if _, err := extractFromBytes(t, truncated); !errors.Is(err, ErrNoEmbeddedPreview) {
		t.Fatalf("expected ErrNoEmbeddedPreview, got %v", err)
	}
}

func TestExtractEmbeddedPreviewReadsFujiRafHeader(t *testing.T) {
	previewJPEG := encodedTestJPEG(t, 70, 50)
	header := make([]byte, 160)
	copy(header, "FUJIFILMCCD-RAW 0201")
	binary.BigEndian.PutUint32(header[84:], uint32(len(header)))
	binary.BigEndian.PutUint32(header[88:], uint32(len(previewJPEG)))
	content := append(header, previewJPEG...)

	preview, err := extractFromBytes(t, content)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(preview.JPEG, previewJPEG) {
		t.Fatalf("expected the raf jpeg, got %d bytes", len(preview.JPEG))
	}
}

func TestExtractEmbeddedPreviewFromFile(t *testing.T) {
	previewJPEG := encodedTestJPEG(t, 50, 40)
	rawPath := filepath.Join(t.TempDir(), "photo.nef")
	content := tiffFixture{byteOrder: binary.BigEndian, orientation: 3, subIfdPreviews: [][]byte{previewJPEG}}.build()
	if err := os.WriteFile(rawPath, content, 0644); err != nil {
		t.Fatal(err)
	}

	preview, err := ExtractEmbeddedPreview(rawPath)
	if err != nil || preview.Orientation != 3 || !bytes.Equal(preview.JPEG, previewJPEG) {
		t.Fatalf("unexpected preview orientation=%d err=%v", preview.Orientation, err)
	}
	if _, err := ExtractEmbeddedPreview(filepath.Join(t.TempDir(), "missing.nef")); err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestExtractEmbeddedPreviewSurvivesGarbage(t *testing.T) {
	cyclicTiff := []byte{'I', 'I', 42, 0, 8, 0, 0, 0, 0xFF, 0xFF}
	for _, content := range [][]byte{nil, []byte("II"), cyclicTiff, bytes.Repeat([]byte{0xFF}, 64)} {
		if _, err := extractFromBytes(t, content); !errors.Is(err, ErrNoEmbeddedPreview) {
			t.Fatalf("expected ErrNoEmbeddedPreview for %v, got %v", content, err)
		}
	}
}
