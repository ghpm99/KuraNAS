package takeout

import "testing"

func TestBuildPhotoSidecarsKeepsOnlyFilesWithDescriptionOrLocation(t *testing.T) {
	extractedFiles := []ExtractedFile{
		{DestinationPath: "/lib/described.jpg", Metadata: &TakeoutMetadata{Description: "  praia  "}},
		{DestinationPath: "/lib/located.jpg", Metadata: &TakeoutMetadata{GeoData: takeoutGeoData{Latitude: -23.5, Longitude: -46.6}}},
		{DestinationPath: "/lib/zero-location.jpg", Metadata: &TakeoutMetadata{GeoData: takeoutGeoData{}}},
		{DestinationPath: "/lib/no-json.jpg"},
	}

	sidecars := BuildPhotoSidecars(extractedFiles)

	if len(sidecars) != 2 {
		t.Fatalf("expected 2 sidecars, got %+v", sidecars)
	}
	if sidecars[0].Path != "/lib/described.jpg" || sidecars[0].Description != "praia" || sidecars[0].HasLocation() {
		t.Fatalf("unexpected described sidecar: %+v", sidecars[0])
	}
	if sidecars[1].Path != "/lib/located.jpg" || sidecars[1].Description != "" || !sidecars[1].HasLocation() ||
		*sidecars[1].Latitude != -23.5 || *sidecars[1].Longitude != -46.6 {
		t.Fatalf("unexpected located sidecar: %+v", sidecars[1])
	}
}
