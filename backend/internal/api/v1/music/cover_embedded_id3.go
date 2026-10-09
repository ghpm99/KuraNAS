package music

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

const (
	id3HeaderSize          = 10
	id3MaxTagSize          = 32 << 20
	id3UnsynchronisationOn = 0x80
	id3ExtendedHeaderOn    = 0x40
	id3FrameUnsyncOn       = 0x02
	frontCoverPictureType  = 3
)

var errMalformedEmbeddedPicture = errors.New("malformed embedded picture")

type embeddedPicture struct {
	pictureType byte
	data        []byte
}

func extractID3Picture(source io.Reader) ([]byte, bool) {
	header := make([]byte, id3HeaderSize)
	if _, err := io.ReadFull(source, header); err != nil || string(header[:3]) != "ID3" {
		return nil, false
	}
	majorVersion := header[3]
	if majorVersion < 2 || majorVersion > 4 {
		return nil, false
	}
	tagSize := decodeSynchsafe(header[6:10])
	if tagSize <= 0 || tagSize > id3MaxTagSize {
		return nil, false
	}
	tagBody := make([]byte, tagSize)
	if _, err := io.ReadFull(source, tagBody); err != nil {
		return nil, false
	}
	if header[5]&id3UnsynchronisationOn != 0 && majorVersion < 4 {
		tagBody = removeUnsynchronisation(tagBody)
	}
	if header[5]&id3ExtendedHeaderOn != 0 {
		tagBody = skipExtendedHeader(tagBody, majorVersion)
	}
	return pickBestPicture(collectID3Pictures(tagBody, majorVersion))
}

func collectID3Pictures(tagBody []byte, majorVersion byte) []embeddedPicture {
	var pictures []embeddedPicture
	frameIDLength, frameHeaderLength := 4, 10
	if majorVersion == 2 {
		frameIDLength, frameHeaderLength = 3, 6
	}
	position := 0
	for position+frameHeaderLength <= len(tagBody) {
		frameID := string(tagBody[position : position+frameIDLength])
		if tagBody[position] == 0 {
			break
		}
		frameSize := readFrameSize(tagBody[position+frameIDLength:position+frameHeaderLength], majorVersion)
		bodyStart := position + frameHeaderLength
		bodyEnd := bodyStart + frameSize
		if frameSize < 0 || bodyEnd > len(tagBody) {
			break
		}
		if frameID == "APIC" || frameID == "PIC" {
			frameBody := tagBody[bodyStart:bodyEnd]
			if majorVersion == 4 && tagBody[position+9]&id3FrameUnsyncOn != 0 {
				frameBody = removeUnsynchronisation(frameBody)
			}
			if picture, isParsed := parseID3PictureFrame(frameBody, majorVersion == 2); isParsed {
				pictures = append(pictures, picture)
			}
		}
		position = bodyEnd
	}
	return pictures
}

func readFrameSize(sizeBytes []byte, majorVersion byte) int {
	switch majorVersion {
	case 2:
		return int(sizeBytes[0])<<16 | int(sizeBytes[1])<<8 | int(sizeBytes[2])
	case 4:
		return decodeSynchsafe(sizeBytes[:4])
	}
	return int(binary.BigEndian.Uint32(sizeBytes[:4]))
}

func parseID3PictureFrame(frameBody []byte, isLegacyFrame bool) (embeddedPicture, bool) {
	if len(frameBody) < 4 {
		return embeddedPicture{}, false
	}
	textEncoding := frameBody[0]
	position := 1
	if isLegacyFrame {
		position += 3
	} else {
		mimeEnd := bytes.IndexByte(frameBody[position:], 0)
		if mimeEnd < 0 {
			return embeddedPicture{}, false
		}
		position += mimeEnd + 1
	}
	if position >= len(frameBody) {
		return embeddedPicture{}, false
	}
	pictureType := frameBody[position]
	position++
	descriptionEnd := skipEncodedText(frameBody, position, textEncoding)
	if descriptionEnd < 0 || descriptionEnd >= len(frameBody) {
		return embeddedPicture{}, false
	}
	return embeddedPicture{pictureType: pictureType, data: frameBody[descriptionEnd:]}, true
}

func skipEncodedText(buffer []byte, start int, textEncoding byte) int {
	isWideEncoding := textEncoding == 1 || textEncoding == 2
	for position := start; position < len(buffer); {
		if !isWideEncoding {
			if buffer[position] == 0 {
				return position + 1
			}
			position++
			continue
		}
		if position+1 >= len(buffer) {
			return -1
		}
		if buffer[position] == 0 && buffer[position+1] == 0 {
			return position + 2
		}
		position += 2
	}
	return -1
}

func pickBestPicture(pictures []embeddedPicture) ([]byte, bool) {
	var firstUsable []byte
	for _, picture := range pictures {
		if len(picture.data) == 0 {
			continue
		}
		if picture.pictureType == frontCoverPictureType {
			return picture.data, true
		}
		if firstUsable == nil {
			firstUsable = picture.data
		}
	}
	return firstUsable, firstUsable != nil
}

func decodeSynchsafe(sizeBytes []byte) int {
	return int(sizeBytes[0]&0x7f)<<21 | int(sizeBytes[1]&0x7f)<<14 | int(sizeBytes[2]&0x7f)<<7 | int(sizeBytes[3]&0x7f)
}

func removeUnsynchronisation(buffer []byte) []byte {
	return bytes.ReplaceAll(buffer, []byte{0xff, 0x00}, []byte{0xff})
}

func skipExtendedHeader(tagBody []byte, majorVersion byte) []byte {
	if len(tagBody) < 4 {
		return tagBody
	}
	extendedSize := int(binary.BigEndian.Uint32(tagBody[:4]))
	if majorVersion == 4 {
		extendedSize = decodeSynchsafe(tagBody[:4])
	} else {
		extendedSize += 4
	}
	if extendedSize < 4 || extendedSize > len(tagBody) {
		return tagBody
	}
	return tagBody[extendedSize:]
}
