package takeout

import "time"

type takeoutTimestamp struct {
	Timestamp string `json:"timestamp"`
	Formatted string `json:"formatted"`
}

type takeoutGeoData struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Altitude  float64 `json:"altitude"`
}

type TakeoutMetadata struct {
	Title              string           `json:"title"`
	PhotoTakenTime     takeoutTimestamp `json:"photoTakenTime"`
	GeoData            takeoutGeoData   `json:"geoData"`
	GeoDataExif        takeoutGeoData   `json:"geoDataExif"`
	CreationTime       takeoutTimestamp `json:"creationTime"`
	ModificationTime   takeoutTimestamp `json:"modificationTime"`
	GooglePhotosOrigin struct {
		MobileUpload struct {
			DeviceType string `json:"deviceType"`
		} `json:"mobileUpload"`
	} `json:"googlePhotosOrigin"`
}

type TakeoutUploadSession struct {
	UploadID      string    `json:"upload_id"`
	FileName      string    `json:"file_name"`
	ExpectedSize  int64     `json:"expected_size"`
	ReceivedSize  int64     `json:"received_size"`
	CreatedAt     time.Time `json:"created_at"`
	LastUpdatedAt time.Time `json:"last_updated_at"`
}

type ExtractedFile struct {
	SourcePath      string           `json:"source_path"`
	DestinationPath string           `json:"destination_path"`
	Category        string           `json:"category"`
	Metadata        *TakeoutMetadata `json:"metadata,omitempty"`
}

type ExtractResult struct {
	Files []ExtractedFile `json:"files"`
}
