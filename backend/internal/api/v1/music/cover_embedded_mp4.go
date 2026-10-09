package music

import (
	"encoding/binary"
	"io"
)

const (
	mp4AtomHeaderSize     = 8
	mp4ExtendedHeaderSize = 16
	mp4MaxMoovSize        = 64 << 20
	mp4MetaVersionSize    = 4
	mp4DataHeaderSize     = 8
)

var mp4CoverAtomPath = []string{"udta", "meta", "ilst", "covr", "data"}

func extractMP4Picture(source io.ReaderAt, sourceSize int64) ([]byte, bool) {
	moovBody, isMoovFound := readTopLevelAtom(source, sourceSize, "moov")
	if !isMoovFound {
		return nil, false
	}
	return findAtomPayload(moovBody, mp4CoverAtomPath)
}

func readTopLevelAtom(source io.ReaderAt, sourceSize int64, wantedType string) ([]byte, bool) {
	for offset := int64(0); offset+mp4AtomHeaderSize <= sourceSize; {
		atomHeader := make([]byte, mp4ExtendedHeaderSize)
		readCount, _ := source.ReadAt(atomHeader, offset)
		if readCount < mp4AtomHeaderSize {
			return nil, false
		}
		atomSize, headerSize := decodeAtomSize(atomHeader[:readCount], sourceSize-offset)
		if atomSize < int64(headerSize) {
			return nil, false
		}
		if string(atomHeader[4:8]) == wantedType {
			bodySize := atomSize - int64(headerSize)
			if bodySize > mp4MaxMoovSize {
				return nil, false
			}
			body := make([]byte, bodySize)
			if _, err := source.ReadAt(body, offset+int64(headerSize)); err != nil && err != io.EOF {
				return nil, false
			}
			return body, true
		}
		offset += atomSize
	}
	return nil, false
}

func decodeAtomSize(atomHeader []byte, remainingBytes int64) (int64, int) {
	declaredSize := int64(binary.BigEndian.Uint32(atomHeader[:4]))
	switch declaredSize {
	case 0:
		return remainingBytes, mp4AtomHeaderSize
	case 1:
		if len(atomHeader) < mp4ExtendedHeaderSize {
			return 0, mp4ExtendedHeaderSize
		}
		return int64(binary.BigEndian.Uint64(atomHeader[8:16])), mp4ExtendedHeaderSize
	}
	return declaredSize, mp4AtomHeaderSize
}

func findAtomPayload(container []byte, remainingPath []string) ([]byte, bool) {
	for position := 0; position+mp4AtomHeaderSize <= len(container); {
		atomSize := int(binary.BigEndian.Uint32(container[position : position+4]))
		if atomSize < mp4AtomHeaderSize || position+atomSize > len(container) {
			return nil, false
		}
		atomType := string(container[position+4 : position+8])
		atomBody := container[position+mp4AtomHeaderSize : position+atomSize]
		position += atomSize
		if atomType != remainingPath[0] {
			continue
		}
		if len(remainingPath) == 1 {
			return dataAtomImage(atomBody)
		}
		if atomType == "meta" && len(atomBody) >= mp4MetaVersionSize {
			atomBody = atomBody[mp4MetaVersionSize:]
		}
		if payload, isFound := findAtomPayload(atomBody, remainingPath[1:]); isFound {
			return payload, true
		}
	}
	return nil, false
}

func dataAtomImage(dataAtomBody []byte) ([]byte, bool) {
	if len(dataAtomBody) <= mp4DataHeaderSize {
		return nil, false
	}
	return dataAtomBody[mp4DataHeaderSize:], true
}
