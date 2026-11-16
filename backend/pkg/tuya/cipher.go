package tuya

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
)

func encryptECB(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	padded := padPKCS7(plaintext)
	ciphertext := make([]byte, len(padded))
	encryptBlocks(block, ciphertext, padded)
	return ciphertext, nil
}

func decryptECB(key, ciphertext []byte) ([]byte, error) {
	if len(ciphertext) == 0 || len(ciphertext)%aes.BlockSize != 0 {
		return nil, errInvalidCiphertext
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	padded := make([]byte, len(ciphertext))
	for offset := 0; offset < len(ciphertext); offset += aes.BlockSize {
		block.Decrypt(padded[offset:], ciphertext[offset:offset+aes.BlockSize])
	}
	return unpadPKCS7(padded)
}

func encryptBlocks(block cipher.Block, destination, source []byte) {
	for offset := 0; offset < len(source); offset += aes.BlockSize {
		block.Encrypt(destination[offset:], source[offset:offset+aes.BlockSize])
	}
}

func padPKCS7(plaintext []byte) []byte {
	paddingLength := aes.BlockSize - len(plaintext)%aes.BlockSize
	padding := bytes.Repeat([]byte{byte(paddingLength)}, paddingLength)
	return append(append([]byte{}, plaintext...), padding...)
}

func unpadPKCS7(padded []byte) ([]byte, error) {
	paddingLength := int(padded[len(padded)-1])
	if paddingLength == 0 || paddingLength > aes.BlockSize || paddingLength > len(padded) {
		return nil, errInvalidPadding
	}
	padding := padded[len(padded)-paddingLength:]
	if !bytes.Equal(padding, bytes.Repeat([]byte{byte(paddingLength)}, paddingLength)) {
		return nil, errInvalidPadding
	}
	return padded[:len(padded)-paddingLength], nil
}
