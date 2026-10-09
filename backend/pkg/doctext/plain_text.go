package doctext

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"
	"unicode/utf16"
	"unicode/utf8"
)

var (
	utf8ByteOrderMark    = []byte{0xEF, 0xBB, 0xBF}
	utf16LittleEndianBOM = []byte{0xFF, 0xFE}
	utf16BigEndianBOM    = []byte{0xFE, 0xFF}
)

func readLeadingBytes(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return io.ReadAll(io.LimitReader(file, MaxReadBytes))
}

func extractPlainText(path string, sizeBytes int64) (Result, error) {
	rawBytes, err := readLeadingBytes(path)
	if err != nil {
		return Result{}, err
	}
	decodedText, err := decodeTextBytes(rawBytes, sizeBytes > MaxReadBytes)
	if err != nil {
		return Result{}, err
	}
	return finalizeText(decodedText, sizeBytes > MaxReadBytes), nil
}

func decodeTextBytes(rawBytes []byte, isReadPartial bool) (string, error) {
	switch {
	case bytes.HasPrefix(rawBytes, utf16LittleEndianBOM):
		return decodeUTF16(rawBytes[len(utf16LittleEndianBOM):], binary.LittleEndian), nil
	case bytes.HasPrefix(rawBytes, utf16BigEndianBOM):
		return decodeUTF16(rawBytes[len(utf16BigEndianBOM):], binary.BigEndian), nil
	}

	rawBytes = bytes.TrimPrefix(rawBytes, utf8ByteOrderMark)
	if bytes.IndexByte(rawBytes, 0) >= 0 {
		return "", ErrBinary
	}
	if isReadPartial {
		rawBytes = dropIncompleteTrailingRune(rawBytes)
	}
	if utf8.Valid(rawBytes) {
		return string(rawBytes), nil
	}
	return decodeLatin1(rawBytes), nil
}

func dropIncompleteTrailingRune(rawBytes []byte) []byte {
	for offset := 1; offset <= utf8.UTFMax-1 && offset <= len(rawBytes); offset++ {
		trailingStart := len(rawBytes) - offset
		if !utf8.RuneStart(rawBytes[trailingStart]) {
			continue
		}
		if utf8.FullRune(rawBytes[trailingStart:]) {
			return rawBytes
		}
		return rawBytes[:trailingStart]
	}
	return rawBytes
}

func decodeUTF16(rawBytes []byte, byteOrder binary.ByteOrder) string {
	codeUnits := make([]uint16, 0, len(rawBytes)/2)
	for index := 0; index+1 < len(rawBytes); index += 2 {
		codeUnits = append(codeUnits, byteOrder.Uint16(rawBytes[index:]))
	}
	return string(utf16.Decode(codeUnits))
}

func decodeLatin1(rawBytes []byte) string {
	characters := make([]rune, len(rawBytes))
	for index, rawByte := range rawBytes {
		characters[index] = rune(rawByte)
	}
	return string(characters)
}
