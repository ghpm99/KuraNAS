package music

import (
	"bytes"
	"os"
)

const containerSniffLength = 12

func extractEmbeddedCover(audioPath string) ([]byte, bool) {
	audioFile, err := os.Open(audioPath)
	if err != nil {
		return nil, false
	}
	defer audioFile.Close()

	sniffed := make([]byte, containerSniffLength)
	sniffedCount, _ := audioFile.ReadAt(sniffed, 0)
	sniffed = sniffed[:sniffedCount]

	switch {
	case bytes.HasPrefix(sniffed, []byte("ID3")):
		return extractID3Picture(audioFile)
	case bytes.HasPrefix(sniffed, []byte(flacMarker)):
		return extractFLACPicture(audioFile)
	case len(sniffed) >= 8 && string(sniffed[4:8]) == "ftyp":
		audioInfo, err := audioFile.Stat()
		if err != nil {
			return nil, false
		}
		return extractMP4Picture(audioFile, audioInfo.Size())
	}
	return nil, false
}
