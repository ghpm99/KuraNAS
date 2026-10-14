package video

import (
	"bytes"
	"strings"
	"unicode/utf8"
)

var utf8ByteOrderMark = []byte{0xEF, 0xBB, 0xBF}

var windows1252HighControlRunes = [32]rune{
	0x20AC, 0x0081, 0x201A, 0x0192, 0x201E, 0x2026, 0x2020, 0x2021,
	0x02C6, 0x2030, 0x0160, 0x2039, 0x0152, 0x008D, 0x017D, 0x008F,
	0x0090, 0x2018, 0x2019, 0x201C, 0x201D, 0x2022, 0x2013, 0x2014,
	0x02DC, 0x2122, 0x0161, 0x203A, 0x0153, 0x009D, 0x017E, 0x0178,
}

func DecodeSubtitleText(rawBytes []byte) string {
	withoutByteOrderMark := bytes.TrimPrefix(rawBytes, utf8ByteOrderMark)
	if utf8.Valid(withoutByteOrderMark) {
		return string(withoutByteOrderMark)
	}
	return decodeWindows1252(withoutByteOrderMark)
}

func decodeWindows1252(rawBytes []byte) string {
	var decoded strings.Builder
	decoded.Grow(len(rawBytes) * 2)
	for _, currentByte := range rawBytes {
		decoded.WriteRune(windows1252Rune(currentByte))
	}
	return decoded.String()
}

func windows1252Rune(currentByte byte) rune {
	if currentByte >= 0x80 && currentByte <= 0x9F {
		return windows1252HighControlRunes[currentByte-0x80]
	}
	return rune(currentByte)
}
