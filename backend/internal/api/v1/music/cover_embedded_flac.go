package music

import (
	"encoding/binary"
	"io"
)

const (
	flacMarker              = "fLaC"
	flacBlockHeaderSize     = 4
	flacPictureBlockType    = 6
	flacLastBlockFlag       = 0x80
	flacMaxPictureBlockSize = 32 << 20
)

func extractFLACPicture(source io.Reader) ([]byte, bool) {
	marker := make([]byte, len(flacMarker))
	if _, err := io.ReadFull(source, marker); err != nil || string(marker) != flacMarker {
		return nil, false
	}
	var pictures []embeddedPicture
	for {
		blockHeader := make([]byte, flacBlockHeaderSize)
		if _, err := io.ReadFull(source, blockHeader); err != nil {
			break
		}
		isLastBlock := blockHeader[0]&flacLastBlockFlag != 0
		blockType := blockHeader[0] &^ flacLastBlockFlag
		blockLength := int(blockHeader[1])<<16 | int(blockHeader[2])<<8 | int(blockHeader[3])
		if blockType != flacPictureBlockType {
			if _, err := io.CopyN(io.Discard, source, int64(blockLength)); err != nil {
				break
			}
		} else {
			if blockLength > flacMaxPictureBlockSize {
				break
			}
			blockBody := make([]byte, blockLength)
			if _, err := io.ReadFull(source, blockBody); err != nil {
				break
			}
			if picture, err := parseFLACPictureBlock(blockBody); err == nil {
				pictures = append(pictures, picture)
			}
		}
		if isLastBlock {
			break
		}
	}
	return pickBestPicture(pictures)
}

func parseFLACPictureBlock(blockBody []byte) (embeddedPicture, error) {
	cursor := pictureBlockCursor{buffer: blockBody}
	pictureType, isTypeRead := cursor.readUint32()
	mimeLength, isMimeLengthRead := cursor.readUint32()
	if !isTypeRead || !isMimeLengthRead || !cursor.skip(mimeLength) {
		return embeddedPicture{}, errMalformedEmbeddedPicture
	}
	descriptionLength, isDescriptionLengthRead := cursor.readUint32()
	if !isDescriptionLengthRead || !cursor.skip(descriptionLength) || !cursor.skip(16) {
		return embeddedPicture{}, errMalformedEmbeddedPicture
	}
	dataLength, isDataLengthRead := cursor.readUint32()
	if !isDataLengthRead || dataLength > len(cursor.buffer)-cursor.position {
		return embeddedPicture{}, errMalformedEmbeddedPicture
	}
	return embeddedPicture{
		pictureType: byte(pictureType),
		data:        cursor.buffer[cursor.position : cursor.position+dataLength],
	}, nil
}

type pictureBlockCursor struct {
	buffer   []byte
	position int
}

func (cursor *pictureBlockCursor) readUint32() (int, bool) {
	if cursor.position+4 > len(cursor.buffer) {
		return 0, false
	}
	value := binary.BigEndian.Uint32(cursor.buffer[cursor.position:])
	cursor.position += 4
	return int(value), true
}

func (cursor *pictureBlockCursor) skip(byteCount int) bool {
	if byteCount < 0 || byteCount > len(cursor.buffer)-cursor.position {
		return false
	}
	cursor.position += byteCount
	return true
}
