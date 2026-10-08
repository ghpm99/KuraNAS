package img

import "bytes"

const (
	jpegMarkerPrefix        = 0xFF
	jpegStartOfFrameBase    = 0xC0
	jpegStartOfFrameExt     = 0xC1
	jpegStartOfFrameProg    = 0xC2
	jpegRestartFirst        = 0xD0
	jpegRestartLast         = 0xD7
	jpegSegmentHeaderLength = 2
)

var jpegSignature = []byte{jpegMarkerPrefix, jpegStartOfImageMarker, jpegMarkerPrefix}

type jpegStream struct {
	length      int
	isDecodable bool
}

func measureJPEGStream(data []byte, start int) (jpegStream, bool) {
	if start+len(jpegSignature) > len(data) || !bytes.HasPrefix(data[start:], jpegSignature) {
		return jpegStream{}, false
	}

	position := start + 2
	isDecodable := false
	for position+2 <= len(data) {
		if data[position] != jpegMarkerPrefix {
			return jpegStream{}, false
		}
		marker := data[position+1]
		switch {
		case marker == jpegMarkerPrefix:
			position++
		case marker == jpegEndOfImageMarker:
			return jpegStream{length: position + 2 - start, isDecodable: isDecodable}, isDecodable
		case marker == 0x01 || (marker >= jpegRestartFirst && marker <= jpegRestartLast):
			position += 2
		default:
			segmentEnd, isValid := skipJPEGSegment(data, position)
			if !isValid {
				return jpegStream{}, false
			}
			if isStartOfFrame(marker) {
				isDecodable = isDecodableStartOfFrame(marker)
			}
			position = segmentEnd
			if marker == jpegStartOfScanMarker {
				position = skipEntropyCodedData(data, position)
			}
		}
	}
	return jpegStream{}, false
}

func skipJPEGSegment(data []byte, position int) (int, bool) {
	if position+4 > len(data) {
		return 0, false
	}
	segmentLength := int(data[position+2])<<8 | int(data[position+3])
	segmentEnd := position + jpegSegmentHeaderLength + segmentLength
	if segmentLength < jpegSegmentHeaderLength || segmentEnd > len(data) {
		return 0, false
	}
	return segmentEnd, true
}

func skipEntropyCodedData(data []byte, position int) int {
	for position+1 < len(data) {
		if data[position] != jpegMarkerPrefix {
			position++
			continue
		}
		next := data[position+1]
		isStuffedByte := next == 0x00
		isRestart := next >= jpegRestartFirst && next <= jpegRestartLast
		if !isStuffedByte && !isRestart {
			return position
		}
		position += 2
	}
	return len(data)
}

func isStartOfFrame(marker byte) bool {
	isDifferentialOrArithmetic := marker >= 0xC3 && marker <= 0xCF && marker != 0xC4 && marker != 0xC8 && marker != 0xCC
	return isDecodableStartOfFrame(marker) || isDifferentialOrArithmetic
}

func isDecodableStartOfFrame(marker byte) bool {
	return marker == jpegStartOfFrameBase || marker == jpegStartOfFrameExt || marker == jpegStartOfFrameProg
}
