package takeout

import "strings"

type PhotoSidecar struct {
	Path        string
	Description string
	Latitude    *float64
	Longitude   *float64
}

func (sidecar PhotoSidecar) HasLocation() bool {
	return sidecar.Latitude != nil && sidecar.Longitude != nil
}

func newPhotoSidecar(destinationPath string, metadata TakeoutMetadata) (PhotoSidecar, bool) {
	sidecar := PhotoSidecar{
		Path:        destinationPath,
		Description: strings.TrimSpace(metadata.Description),
	}
	if metadata.GeoData.hasLocation() {
		latitude, longitude := metadata.GeoData.Latitude, metadata.GeoData.Longitude
		sidecar.Latitude, sidecar.Longitude = &latitude, &longitude
	}
	hasContent := sidecar.Description != "" || sidecar.HasLocation()
	return sidecar, hasContent
}

func (geoData takeoutGeoData) hasLocation() bool {
	return geoData.Latitude != 0 || geoData.Longitude != 0
}

// BuildPhotoSidecars returns one sidecar per extracted file whose Takeout JSON
// carries a description or a location; files without either are omitted.
func BuildPhotoSidecars(extractedFiles []ExtractedFile) []PhotoSidecar {
	sidecars := make([]PhotoSidecar, 0)
	for _, extractedFile := range extractedFiles {
		if extractedFile.Metadata == nil {
			continue
		}
		if sidecar, hasContent := newPhotoSidecar(extractedFile.DestinationPath, *extractedFile.Metadata); hasContent {
			sidecars = append(sidecars, sidecar)
		}
	}
	return sidecars
}
