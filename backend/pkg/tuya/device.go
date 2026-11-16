package tuya

import (
	"fmt"
	"net"
)

const (
	defaultPort  = "6668"
	localKeySize = 16
)

// Device identifies a Tuya device reachable on the local network.
// Address may omit the port, in which case 6668 is used.
type Device struct {
	Address  string
	DeviceID string
	LocalKey string
}

func validateDevice(device Device) error {
	if device.Address == "" {
		return fmt.Errorf("%w: empty address", ErrInvalidDevice)
	}
	if device.DeviceID == "" {
		return fmt.Errorf("%w: empty device id", ErrInvalidDevice)
	}
	if len(device.LocalKey) != localKeySize {
		return fmt.Errorf("%w: local key must have %d bytes", ErrInvalidDevice, localKeySize)
	}
	return nil
}

func normalizeAddress(address string) string {
	if _, _, err := net.SplitHostPort(address); err == nil {
		return address
	}
	return net.JoinHostPort(address, defaultPort)
}
