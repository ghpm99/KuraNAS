package tuya

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
)

const (
	commandWrite uint32 = 0x07
	commandRead  uint32 = 0x0a

	framePrefix      uint32 = 0x000055AA
	frameSuffix      uint32 = 0x0000AA55
	frameHeaderSize         = 16
	frameTrailerSize        = 8
	maxFrameBodySize        = 1 << 16
	versionHeaderLen        = 15
	returnCodeSize          = 4
)

var versionTag = []byte("3.3")

type frame struct {
	sequence uint32
	command  uint32
	payload  []byte
}

func encodeFrame(sequence, command uint32, payload []byte) []byte {
	encoded := make([]byte, 0, frameHeaderSize+len(payload)+frameTrailerSize)
	encoded = binary.BigEndian.AppendUint32(encoded, framePrefix)
	encoded = binary.BigEndian.AppendUint32(encoded, sequence)
	encoded = binary.BigEndian.AppendUint32(encoded, command)
	encoded = binary.BigEndian.AppendUint32(encoded, uint32(len(payload)+frameTrailerSize))
	encoded = append(encoded, payload...)
	encoded = binary.BigEndian.AppendUint32(encoded, crc32.ChecksumIEEE(encoded))
	return binary.BigEndian.AppendUint32(encoded, frameSuffix)
}

func readFrame(reader io.Reader) (frame, error) {
	header := make([]byte, frameHeaderSize)
	if _, err := io.ReadFull(reader, header); err != nil {
		return frame{}, err
	}
	if binary.BigEndian.Uint32(header[0:]) != framePrefix {
		return frame{}, fmt.Errorf("%w: bad prefix", errInvalidFrame)
	}
	bodySize := binary.BigEndian.Uint32(header[12:])
	if bodySize < frameTrailerSize || bodySize > maxFrameBodySize {
		return frame{}, fmt.Errorf("%w: bad size %d", errInvalidFrame, bodySize)
	}
	body := make([]byte, bodySize)
	if _, err := io.ReadFull(reader, body); err != nil {
		return frame{}, err
	}
	return decodeFrame(header, body)
}

func decodeFrame(header, body []byte) (frame, error) {
	payloadSize := len(body) - frameTrailerSize
	trailer := body[payloadSize:]
	expectedChecksum := crc32.Update(crc32.ChecksumIEEE(header), crc32.IEEETable, body[:payloadSize])
	if binary.BigEndian.Uint32(trailer[0:]) != expectedChecksum {
		return frame{}, fmt.Errorf("%w: bad checksum", errInvalidFrame)
	}
	if binary.BigEndian.Uint32(trailer[4:]) != frameSuffix {
		return frame{}, fmt.Errorf("%w: bad suffix", errInvalidFrame)
	}
	return frame{
		sequence: binary.BigEndian.Uint32(header[4:]),
		command:  binary.BigEndian.Uint32(header[8:]),
		payload:  body[:payloadSize],
	}, nil
}

func (received frame) ciphertext() []byte {
	ciphertext := received.payload
	if hasReturnCode(ciphertext) {
		ciphertext = ciphertext[returnCodeSize:]
	}
	if len(ciphertext) >= versionHeaderLen && bytes.HasPrefix(ciphertext, versionTag) {
		ciphertext = ciphertext[versionHeaderLen:]
	}
	return ciphertext
}

func hasReturnCode(payload []byte) bool {
	return len(payload) >= returnCodeSize && binary.BigEndian.Uint32(payload)&0xFFFFFF00 == 0
}

func (received frame) decrypt(localKey string) ([]byte, error) {
	return decryptECB([]byte(localKey), received.ciphertext())
}
