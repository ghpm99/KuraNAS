package tuya

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

func buildReadFrame(device Device, sequence uint32, now time.Time) ([]byte, error) {
	plaintext, err := json.Marshal(map[string]string{
		"gwId":  device.DeviceID,
		"devId": device.DeviceID,
		"uid":   device.DeviceID,
		"t":     strconv.FormatInt(now.Unix(), 10),
	})
	if err != nil {
		return nil, err
	}
	ciphertext, err := encryptECB([]byte(device.LocalKey), plaintext)
	if err != nil {
		return nil, err
	}
	return encodeFrame(sequence, commandRead, ciphertext), nil
}

func buildWriteFrame(device Device, sequence uint32, now time.Time, dataPoints map[string]any) ([]byte, error) {
	plaintext, err := json.Marshal(map[string]any{
		"devId": device.DeviceID,
		"uid":   device.DeviceID,
		"t":     strconv.FormatInt(now.Unix(), 10),
		"dps":   dataPoints,
	})
	if err != nil {
		return nil, err
	}
	ciphertext, err := encryptECB([]byte(device.LocalKey), plaintext)
	if err != nil {
		return nil, err
	}
	versioned := append(versionHeader(), ciphertext...)
	return encodeFrame(sequence, commandWrite, versioned), nil
}

func versionHeader() []byte {
	return append(append([]byte{}, versionTag...), make([]byte, versionHeaderLen-len(versionTag))...)
}

func parseDataPoints(plaintext []byte) (map[string]any, error) {
	var status struct {
		DataPoints map[string]any `json:"dps"`
	}
	if err := json.NewDecoder(bytes.NewReader(plaintext)).Decode(&status); err != nil {
		return nil, fmt.Errorf("%w: reply is not valid JSON", ErrKeyRejected)
	}
	if status.DataPoints == nil {
		return nil, fmt.Errorf("%w: reply has no dps", ErrKeyRejected)
	}
	return status.DataPoints, nil
}
