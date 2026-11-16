package tuya

import "errors"

var (
	// ErrUnreachable means the dial failed or no byte arrived from the device within the timeout.
	ErrUnreachable = errors.New("tuya: device unreachable")

	// ErrKeyRejected means the device answered but the answer could not be decrypted into the expected content.
	ErrKeyRejected = errors.New("tuya: device rejected the local key")

	// ErrInvalidDevice means the device description cannot be used to open a conversation.
	ErrInvalidDevice = errors.New("tuya: invalid device")

	errInvalidFrame      = errors.New("tuya: invalid frame")
	errInvalidCiphertext = errors.New("tuya: invalid ciphertext")
	errInvalidPadding    = errors.New("tuya: invalid padding")
)
