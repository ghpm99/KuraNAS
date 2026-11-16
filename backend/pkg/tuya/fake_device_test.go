package tuya

import (
	"bytes"
	"crypto/aes"
	"encoding/binary"
	"encoding/json"
	"hash/crc32"
	"io"
	"net"
	"testing"
	"time"
)

type receivedRequest struct {
	command         uint32
	hasVersionTag   bool
	decryptedFields map[string]any
}

type fakeDevice struct {
	deviceKey            []byte
	dataPoints           map[string]any
	isSilent             bool
	shouldSendEmptyFirst bool
	shouldPrefixVersion  bool
	requests             chan receivedRequest
	listener             net.Listener
}

func startFakeDevice(t *testing.T, device *fakeDevice) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	device.listener = listener
	device.requests = make(chan receivedRequest, 4)
	t.Cleanup(func() { listener.Close() })
	go device.acceptLoop()
	return listener.Addr().String()
}

func (device *fakeDevice) acceptLoop() {
	for {
		connection, err := device.listener.Accept()
		if err != nil {
			return
		}
		go device.serve(connection)
	}
}

func (device *fakeDevice) serve(connection net.Conn) {
	defer connection.Close()
	command, body, err := readRawRequest(connection)
	if err != nil {
		return
	}
	if device.isSilent {
		time.Sleep(2 * time.Second)
		return
	}
	request, isDecryptable := device.interpret(command, body)
	device.requests <- request
	if !isDecryptable {
		connection.Write(rawFrame(command, append([]byte{0, 0, 0, 1}, 'o', 'o', 'p', 's')))
		return
	}
	if device.shouldSendEmptyFirst {
		connection.Write(rawFrame(0x09, nil))
	}
	connection.Write(rawFrame(command, device.replyPayload(command)))
}

func (device *fakeDevice) replyPayload(command uint32) []byte {
	if command == 0x07 {
		return nil
	}
	plain, _ := json.Marshal(map[string]any{"dps": device.dataPoints})
	reply := []byte{0, 0, 0, 0}
	if device.shouldPrefixVersion {
		reply = append(reply, append([]byte("3.3"), make([]byte, 12)...)...)
	}
	return append(reply, aesEncrypt(device.deviceKey, plain)...)
}

func (device *fakeDevice) interpret(command uint32, body []byte) (receivedRequest, bool) {
	request := receivedRequest{command: command}
	if command == 0x07 && bytes.HasPrefix(body, append([]byte("3.3"), make([]byte, 12)...)) {
		request.hasVersionTag = true
		body = body[15:]
	}
	plain, isDecryptable := aesDecrypt(device.deviceKey, body)
	if !isDecryptable {
		return request, false
	}
	if json.Unmarshal(plain, &request.decryptedFields) != nil {
		return request, false
	}
	return request, true
}

func readRawRequest(connection net.Conn) (uint32, []byte, error) {
	header := make([]byte, 16)
	if _, err := io.ReadFull(connection, header); err != nil {
		return 0, nil, err
	}
	rest := make([]byte, binary.BigEndian.Uint32(header[12:]))
	if _, err := io.ReadFull(connection, rest); err != nil {
		return 0, nil, err
	}
	return binary.BigEndian.Uint32(header[8:]), rest[:len(rest)-8], nil
}

func rawFrame(command uint32, payload []byte) []byte {
	head := make([]byte, 16)
	binary.BigEndian.PutUint32(head[0:], 0x55AA)
	binary.BigEndian.PutUint32(head[4:], 99)
	binary.BigEndian.PutUint32(head[8:], command)
	binary.BigEndian.PutUint32(head[12:], uint32(len(payload)+8))
	withPayload := append(head, payload...)
	checksum := make([]byte, 4)
	binary.BigEndian.PutUint32(checksum, crc32.ChecksumIEEE(withPayload))
	return append(append(withPayload, checksum...), 0, 0, 0xAA, 0x55)
}

func aesEncrypt(key, plain []byte) []byte {
	missing := 16 - len(plain)%16
	padded := append(append([]byte{}, plain...), bytes.Repeat([]byte{byte(missing)}, missing)...)
	block, _ := aes.NewCipher(key)
	encrypted := make([]byte, len(padded))
	for offset := 0; offset < len(padded); offset += 16 {
		block.Encrypt(encrypted[offset:offset+16], padded[offset:offset+16])
	}
	return encrypted
}

func aesDecrypt(key, encrypted []byte) ([]byte, bool) {
	if len(encrypted) == 0 || len(encrypted)%16 != 0 {
		return nil, false
	}
	block, _ := aes.NewCipher(key)
	padded := make([]byte, len(encrypted))
	for offset := 0; offset < len(encrypted); offset += 16 {
		block.Decrypt(padded[offset:offset+16], encrypted[offset:offset+16])
	}
	missing := int(padded[len(padded)-1])
	if missing < 1 || missing > 16 {
		return nil, false
	}
	return padded[:len(padded)-missing], true
}
