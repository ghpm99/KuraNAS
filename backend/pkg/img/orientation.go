package img

import (
	"bytes"
	"encoding/binary"
	"image"
	"io"
	"os"
)

const (
	orientationTopLeft              = 1
	exifOrientationTag              = 0x0112
	exifReadLimitBytes              = 128 * 1024
	jpegStartOfImageMarker          = 0xD8
	jpegApp1Marker                  = 0xE1
	jpegStartOfScanMarker           = 0xDA
	jpegEndOfImageMarker            = 0xD9
	tiffIfdEntrySizeBytes           = 12
	tiffShortFieldType              = 3
	firstOrientationNeedingAxisSwap = 5
)

var exifHeaderSignature = []byte("Exif\x00\x00")

func ReadJPEGOrientation(path string) int {
	file, err := os.Open(path)
	if err != nil {
		return orientationTopLeft
	}
	defer file.Close()

	header, err := io.ReadAll(io.LimitReader(file, exifReadLimitBytes))
	if err != nil {
		return orientationTopLeft
	}
	return parseJPEGOrientation(header)
}

func parseJPEGOrientation(jpegHeader []byte) int {
	if len(jpegHeader) < 4 || jpegHeader[0] != 0xFF || jpegHeader[1] != jpegStartOfImageMarker {
		return orientationTopLeft
	}

	position := 2
	for position+4 <= len(jpegHeader) {
		if jpegHeader[position] != 0xFF {
			return orientationTopLeft
		}
		marker := jpegHeader[position+1]
		if marker == jpegStartOfScanMarker || marker == jpegEndOfImageMarker {
			return orientationTopLeft
		}
		segmentLength := int(binary.BigEndian.Uint16(jpegHeader[position+2 : position+4]))
		segmentEnd := position + 2 + segmentLength
		if segmentLength < 2 || segmentEnd > len(jpegHeader) {
			return orientationTopLeft
		}
		payload := jpegHeader[position+4 : segmentEnd]
		if marker == jpegApp1Marker && bytes.HasPrefix(payload, exifHeaderSignature) {
			return parseTIFFOrientation(payload[len(exifHeaderSignature):])
		}
		position = segmentEnd
	}
	return orientationTopLeft
}

func parseTIFFOrientation(tiff []byte) int {
	if len(tiff) < 8 {
		return orientationTopLeft
	}

	var byteOrder binary.ByteOrder
	switch string(tiff[:2]) {
	case "II":
		byteOrder = binary.LittleEndian
	case "MM":
		byteOrder = binary.BigEndian
	default:
		return orientationTopLeft
	}

	firstIfdOffset := int(byteOrder.Uint32(tiff[4:8]))
	if firstIfdOffset < 8 || firstIfdOffset+2 > len(tiff) {
		return orientationTopLeft
	}

	entryCount := int(byteOrder.Uint16(tiff[firstIfdOffset : firstIfdOffset+2]))
	for entryIndex := 0; entryIndex < entryCount; entryIndex++ {
		entryStart := firstIfdOffset + 2 + entryIndex*tiffIfdEntrySizeBytes
		if entryStart+tiffIfdEntrySizeBytes > len(tiff) {
			return orientationTopLeft
		}
		if byteOrder.Uint16(tiff[entryStart:entryStart+2]) != exifOrientationTag {
			continue
		}
		if byteOrder.Uint16(tiff[entryStart+2:entryStart+4]) != tiffShortFieldType {
			return orientationTopLeft
		}
		orientation := int(byteOrder.Uint16(tiff[entryStart+8 : entryStart+10]))
		if orientation < 1 || orientation > 8 {
			return orientationTopLeft
		}
		return orientation
	}
	return orientationTopLeft
}

func orientationSwapsAxes(orientation int) bool {
	return orientation >= firstOrientationNeedingAxisSwap && orientation <= 8
}

func applyOrientation(src *image.RGBA, orientation int) *image.RGBA {
	if orientation <= orientationTopLeft || orientation > 8 {
		return src
	}

	sourceWidth, sourceHeight := src.Bounds().Dx(), src.Bounds().Dy()
	targetWidth, targetHeight := sourceWidth, sourceHeight
	if orientationSwapsAxes(orientation) {
		targetWidth, targetHeight = sourceHeight, sourceWidth
	}

	oriented := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	for targetY := 0; targetY < targetHeight; targetY++ {
		for targetX := 0; targetX < targetWidth; targetX++ {
			sourceX, sourceY := sourceCoordinates(orientation, targetX, targetY, sourceWidth, sourceHeight)
			sourceOffset := src.PixOffset(sourceX, sourceY)
			targetOffset := oriented.PixOffset(targetX, targetY)
			copy(oriented.Pix[targetOffset:targetOffset+4], src.Pix[sourceOffset:sourceOffset+4])
		}
	}
	return oriented
}

func sourceCoordinates(orientation, targetX, targetY, sourceWidth, sourceHeight int) (int, int) {
	switch orientation {
	case 2:
		return sourceWidth - 1 - targetX, targetY
	case 3:
		return sourceWidth - 1 - targetX, sourceHeight - 1 - targetY
	case 4:
		return targetX, sourceHeight - 1 - targetY
	case 5:
		return targetY, targetX
	case 6:
		return targetY, sourceHeight - 1 - targetX
	case 7:
		return sourceWidth - 1 - targetY, sourceHeight - 1 - targetX
	case 8:
		return sourceWidth - 1 - targetY, targetX
	}
	return targetX, targetY
}
