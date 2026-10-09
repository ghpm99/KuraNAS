package img

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
)

const (
	maxRawFileScanBytes       = 64 * 1024 * 1024
	maxEmbeddedPreviewBytes   = 64 * 1024 * 1024
	sufficientPreviewBytes    = 200 * 1024
	maxTiffEntriesPerIfd      = 2048
	maxTiffIfdDepth           = 3
	maxTiffIfdsPerChain       = 16
	tiffCompressionTag        = 0x0103
	tiffStripOffsetsTag       = 0x0111
	tiffStripByteCountsTag    = 0x0117
	tiffSubIfdsTag            = 0x014A
	tiffJpegOffsetTag         = 0x0201
	tiffJpegLengthTag         = 0x0202
	tiffLongFieldType         = 4
	tiffCompressionOldJPEG    = 6
	tiffCompressionJPEG       = 7
	tiffHeaderLength          = 8
	fujiRafSignature          = "FUJIFILMCCD-RAW"
	fujiRafJpegOffsetPosition = 84
	fujiRafHeaderLength       = 92
)

var ErrNoEmbeddedPreview = errors.New("no embedded jpeg preview found")

type EmbeddedPreview struct {
	JPEG        []byte
	Orientation int
}

type embeddedJPEGLocation struct {
	offset int64
	length int64
}

type tiffReader struct {
	source    io.ReaderAt
	size      int64
	byteOrder binary.ByteOrder
}

type tiffEntry struct {
	tag       uint16
	fieldType uint16
	count     uint32
	valueSlot [4]byte
}

func ExtractEmbeddedPreview(path string) (EmbeddedPreview, error) {
	file, err := os.Open(path)
	if err != nil {
		return EmbeddedPreview{}, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return EmbeddedPreview{}, err
	}
	return extractEmbeddedPreview(file, info.Size())
}

func extractEmbeddedPreview(source io.ReaderAt, size int64) (EmbeddedPreview, error) {
	structuredPreview, hasStructuredPreview := extractStructuredPreview(source, size)
	if hasStructuredPreview && len(structuredPreview.JPEG) >= sufficientPreviewBytes {
		return structuredPreview, nil
	}

	scannedJPEG, hasScannedJPEG := scanLargestJPEG(source, size)
	if hasScannedJPEG && (!hasStructuredPreview || len(scannedJPEG) > len(structuredPreview.JPEG)) {
		return EmbeddedPreview{JPEG: scannedJPEG, Orientation: parseJPEGOrientation(scannedJPEG)}, nil
	}
	if hasStructuredPreview {
		return structuredPreview, nil
	}
	return EmbeddedPreview{}, ErrNoEmbeddedPreview
}

func extractStructuredPreview(source io.ReaderAt, size int64) (EmbeddedPreview, bool) {
	header := make([]byte, fujiRafHeaderLength)
	headerLength, _ := source.ReadAt(header, 0)
	header = header[:headerLength]

	if bytes.HasPrefix(header, []byte(fujiRafSignature)) {
		return extractFujiRafPreview(source, size, header)
	}
	return extractTiffPreview(source, size, header)
}

func extractFujiRafPreview(source io.ReaderAt, size int64, header []byte) (EmbeddedPreview, bool) {
	if len(header) < fujiRafHeaderLength {
		return EmbeddedPreview{}, false
	}
	offset := int64(binary.BigEndian.Uint32(header[fujiRafJpegOffsetPosition:]))
	length := int64(binary.BigEndian.Uint32(header[fujiRafJpegOffsetPosition+4:]))
	return loadPreview(source, size, embeddedJPEGLocation{offset: offset, length: length}, orientationTopLeft)
}

func extractTiffPreview(source io.ReaderAt, size int64, header []byte) (EmbeddedPreview, bool) {
	reader, firstIfdOffset, isTiff := newTiffReader(source, size, header)
	if !isTiff {
		return EmbeddedPreview{}, false
	}

	var locations []embeddedJPEGLocation
	rawOrientation := orientationTopLeft
	visitedIfdOffsets := map[int64]bool{}
	reader.collectJPEGLocations(firstIfdOffset, 0, visitedIfdOffsets, &locations, &rawOrientation, true)

	bestPreview := EmbeddedPreview{}
	hasPreview := false
	for _, location := range locations {
		candidate, isLoaded := loadPreview(source, size, location, rawOrientation)
		if isLoaded && (!hasPreview || len(candidate.JPEG) > len(bestPreview.JPEG)) {
			bestPreview, hasPreview = candidate, true
		}
	}
	return bestPreview, hasPreview
}

func newTiffReader(source io.ReaderAt, size int64, header []byte) (*tiffReader, int64, bool) {
	if len(header) < tiffHeaderLength {
		return nil, 0, false
	}
	var byteOrder binary.ByteOrder
	switch string(header[:2]) {
	case "II":
		byteOrder = binary.LittleEndian
	case "MM":
		byteOrder = binary.BigEndian
	default:
		return nil, 0, false
	}
	return &tiffReader{source: source, size: size, byteOrder: byteOrder}, int64(byteOrder.Uint32(header[4:8])), true
}

func (reader *tiffReader) collectJPEGLocations(
	ifdOffset int64,
	depth int,
	visitedIfdOffsets map[int64]bool,
	locations *[]embeddedJPEGLocation,
	rawOrientation *int,
	isFirstChain bool,
) {
	chainLength := 0
	for ifdOffset >= tiffHeaderLength && ifdOffset < reader.size && !visitedIfdOffsets[ifdOffset] && chainLength < maxTiffIfdsPerChain {
		visitedIfdOffsets[ifdOffset] = true
		entries, nextIfdOffset, isReadable := reader.readIfd(ifdOffset)
		if !isReadable {
			return
		}

		if isFirstChain && chainLength == 0 {
			*rawOrientation = reader.orientationOf(entries)
		}
		if location, hasLocation := reader.jpegLocationOf(entries); hasLocation {
			*locations = append(*locations, location)
		}
		if depth < maxTiffIfdDepth {
			for _, subIfdOffset := range reader.subIfdOffsetsOf(entries) {
				reader.collectJPEGLocations(subIfdOffset, depth+1, visitedIfdOffsets, locations, rawOrientation, false)
			}
		}

		ifdOffset = nextIfdOffset
		chainLength++
	}
}

func (reader *tiffReader) readIfd(ifdOffset int64) ([]tiffEntry, int64, bool) {
	countBytes := make([]byte, 2)
	if _, err := reader.source.ReadAt(countBytes, ifdOffset); err != nil {
		return nil, 0, false
	}
	entryCount := int(reader.byteOrder.Uint16(countBytes))
	if entryCount == 0 || entryCount > maxTiffEntriesPerIfd {
		return nil, 0, false
	}

	table := make([]byte, entryCount*tiffIfdEntrySizeBytes+4)
	if _, err := reader.source.ReadAt(table, ifdOffset+2); err != nil {
		return nil, 0, false
	}

	entries := make([]tiffEntry, entryCount)
	for entryIndex := range entries {
		raw := table[entryIndex*tiffIfdEntrySizeBytes:]
		entries[entryIndex] = tiffEntry{
			tag:       reader.byteOrder.Uint16(raw[0:2]),
			fieldType: reader.byteOrder.Uint16(raw[2:4]),
			count:     reader.byteOrder.Uint32(raw[4:8]),
		}
		copy(entries[entryIndex].valueSlot[:], raw[8:12])
	}
	nextIfdOffset := int64(reader.byteOrder.Uint32(table[entryCount*tiffIfdEntrySizeBytes:]))
	return entries, nextIfdOffset, true
}

func (reader *tiffReader) orientationOf(entries []tiffEntry) int {
	for _, entry := range entries {
		if entry.tag != exifOrientationTag || entry.fieldType != tiffShortFieldType {
			continue
		}
		orientation := int(reader.byteOrder.Uint16(entry.valueSlot[:2]))
		if orientation >= 1 && orientation <= 8 {
			return orientation
		}
	}
	return orientationTopLeft
}

func (reader *tiffReader) jpegLocationOf(entries []tiffEntry) (embeddedJPEGLocation, bool) {
	var interchangeOffset, interchangeLength, compression int64
	var stripOffset, stripLength int64
	hasSingleStrip := true

	for _, entry := range entries {
		switch entry.tag {
		case tiffCompressionTag:
			compression = reader.scalarOf(entry)
		case tiffJpegOffsetTag:
			interchangeOffset = reader.scalarOf(entry)
		case tiffJpegLengthTag:
			interchangeLength = reader.scalarOf(entry)
		case tiffStripOffsetsTag:
			hasSingleStrip = hasSingleStrip && entry.count == 1
			stripOffset = reader.scalarOf(entry)
		case tiffStripByteCountsTag:
			hasSingleStrip = hasSingleStrip && entry.count == 1
			stripLength = reader.scalarOf(entry)
		}
	}

	if interchangeOffset > 0 && interchangeLength > 0 {
		return embeddedJPEGLocation{offset: interchangeOffset, length: interchangeLength}, true
	}
	isJPEGCompressed := compression == tiffCompressionJPEG || compression == tiffCompressionOldJPEG
	if isJPEGCompressed && hasSingleStrip && stripOffset > 0 && stripLength > 0 {
		return embeddedJPEGLocation{offset: stripOffset, length: stripLength}, true
	}
	return embeddedJPEGLocation{}, false
}

func (reader *tiffReader) subIfdOffsetsOf(entries []tiffEntry) []int64 {
	for _, entry := range entries {
		if entry.tag != tiffSubIfdsTag || entry.fieldType != tiffLongFieldType {
			continue
		}
		return reader.longValuesOf(entry)
	}
	return nil
}

func (reader *tiffReader) longValuesOf(entry tiffEntry) []int64 {
	const maxSubIfds = 8
	count := int(min(entry.count, maxSubIfds))
	if count == 0 {
		return nil
	}
	if count == 1 {
		return []int64{int64(reader.byteOrder.Uint32(entry.valueSlot[:]))}
	}
	valuesOffset := int64(reader.byteOrder.Uint32(entry.valueSlot[:]))
	raw := make([]byte, count*4)
	if _, err := reader.source.ReadAt(raw, valuesOffset); err != nil {
		return nil
	}
	values := make([]int64, count)
	for valueIndex := range values {
		values[valueIndex] = int64(reader.byteOrder.Uint32(raw[valueIndex*4:]))
	}
	return values
}

func (reader *tiffReader) scalarOf(entry tiffEntry) int64 {
	if entry.fieldType == tiffShortFieldType {
		return int64(reader.byteOrder.Uint16(entry.valueSlot[:2]))
	}
	return int64(reader.byteOrder.Uint32(entry.valueSlot[:]))
}

func loadPreview(source io.ReaderAt, size int64, location embeddedJPEGLocation, rawOrientation int) (EmbeddedPreview, bool) {
	isWithinFile := location.offset > 0 && location.length > 0 && location.offset+location.length <= size
	if !isWithinFile || location.length > maxEmbeddedPreviewBytes {
		return EmbeddedPreview{}, false
	}

	jpegBytes := make([]byte, location.length)
	if _, err := source.ReadAt(jpegBytes, location.offset); err != nil {
		return EmbeddedPreview{}, false
	}
	if _, isDecodable := measureJPEGStream(jpegBytes, 0); !isDecodable {
		return EmbeddedPreview{}, false
	}

	orientation := parseJPEGOrientation(jpegBytes)
	if orientation == orientationTopLeft {
		orientation = rawOrientation
	}
	return EmbeddedPreview{JPEG: jpegBytes, Orientation: orientation}, true
}

func scanLargestJPEG(source io.ReaderAt, size int64) ([]byte, bool) {
	scanLength := min(size, maxRawFileScanBytes)
	content := make([]byte, scanLength)
	readLength, _ := source.ReadAt(content, 0)
	content = content[:readLength]

	var largest []byte
	for searchFrom := 0; searchFrom < len(content); {
		signatureIndex := bytes.Index(content[searchFrom:], jpegSignature)
		if signatureIndex < 0 {
			break
		}
		start := searchFrom + signatureIndex
		stream, isDecodable := measureJPEGStream(content, start)
		if !isDecodable {
			searchFrom = start + 1
			continue
		}
		if stream.length > len(largest) {
			largest = append([]byte(nil), content[start:start+stream.length]...)
		}
		searchFrom = start + stream.length
	}
	return largest, largest != nil
}
