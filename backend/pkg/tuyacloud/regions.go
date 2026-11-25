package tuyacloud

import "fmt"

var baseURLByRegion = map[string]string{
	"us": "https://openapi.tuyaus.com",
	"eu": "https://openapi.tuyaeu.com",
	"cn": "https://openapi.tuyacn.com",
	"in": "https://openapi.tuyain.com",
}

func baseURLForRegion(region string) (string, error) {
	baseURL, isKnown := baseURLByRegion[region]
	if !isKnown {
		return "", fmt.Errorf("%w: %q", ErrUnknownRegion, region)
	}
	return baseURL, nil
}
