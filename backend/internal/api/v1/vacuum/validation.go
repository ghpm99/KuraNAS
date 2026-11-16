package vacuum

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
)

const localKeyLength = 16

var hostnamePattern = regexp.MustCompile(`^[A-Za-z0-9]([A-Za-z0-9-]{0,61}[A-Za-z0-9])?(\.[A-Za-z0-9]([A-Za-z0-9-]{0,61}[A-Za-z0-9])?)*$`)

func validateSettingsFields(address string, deviceID string, localKey string) error {
	if err := validateAddress(address); err != nil {
		return err
	}
	if deviceID == "" {
		return fmt.Errorf("%w: empty", ErrInvalidDeviceID)
	}
	if localKey != "" && len(localKey) != localKeyLength {
		return fmt.Errorf("%w: must have %d characters", ErrInvalidLocalKey, localKeyLength)
	}
	return nil
}

func validateAddress(address string) error {
	if address == "" {
		return nil
	}

	host := address
	if splitHost, port, err := net.SplitHostPort(address); err == nil {
		if !isValidPort(port) {
			return fmt.Errorf("%w: port", ErrInvalidAddress)
		}
		host = splitHost
	}

	if net.ParseIP(host) == nil && !hostnamePattern.MatchString(host) {
		return fmt.Errorf("%w: host", ErrInvalidAddress)
	}
	return nil
}

func isValidPort(port string) bool {
	portNumber, err := strconv.Atoi(port)
	return err == nil && portNumber >= 1 && portNumber <= 65535
}

func validatePreferencesPatch(patch UpdatePreferencesDto) error {
	if patch.Suction != nil && !allowedSuctionLevels[*patch.Suction] {
		return fmt.Errorf("%w: suction", ErrInvalidPreference)
	}
	if patch.WaterLevel != nil && !allowedWaterLevels[*patch.WaterLevel] {
		return fmt.Errorf("%w: water_level", ErrInvalidPreference)
	}
	if patch.Volume != nil && (*patch.Volume < 0 || *patch.Volume > maxVolume) {
		return fmt.Errorf("%w: volume", ErrInvalidPreference)
	}
	if patch.AutoEmptyFrequency != nil && !allowedAutoEmptyFrequencies[*patch.AutoEmptyFrequency] {
		return fmt.Errorf("%w: auto_empty_frequency", ErrInvalidPreference)
	}
	return nil
}
