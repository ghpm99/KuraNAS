package vacuum

import "fmt"

const cloudAccessSecretLength = 32

var allowedCloudRegions = map[string]bool{"us": true, "eu": true, "cn": true, "in": true}

func validateCloudSettingsFields(accessID string, accessSecret string, region string) error {
	if accessID == "" {
		return fmt.Errorf("%w: empty", ErrInvalidCloudAccessID)
	}
	if !allowedCloudRegions[region] {
		return fmt.Errorf("%w: %q", ErrInvalidCloudRegion, region)
	}
	if accessSecret != "" && len(accessSecret) != cloudAccessSecretLength {
		return fmt.Errorf("%w: must have %d characters", ErrInvalidCloudAccessSecret, cloudAccessSecretLength)
	}
	return nil
}
