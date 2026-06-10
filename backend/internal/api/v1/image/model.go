package image

import (
	"fmt"
	"time"
)

// MetadataModel holds the persisted EXIF/image metadata for a single image
// file. The corresponding DB table is image_metadata; the base file record
// lives in home_file (owned by the files package).
type MetadataModel struct {
	ID                int                 `json:"id"`
	FileId            int                 `json:"file_id"`
	Path              string              `json:"path"`
	Format            string              `json:"format"`
	Mode              string              `json:"mode"`
	Width             int                 `json:"width"`
	Height            int                 `json:"height"`
	DPIX              float64             `json:"dpi_x"`
	DPIY              float64             `json:"dpi_y"`
	XResolution       float64             `json:"x_resolution"`
	YResolution       float64             `json:"y_resolution"`
	ResolutionUnit    float64             `json:"resolution_unit"`
	Orientation       float64             `json:"orientation"`
	Compression       float64             `json:"compression"`
	Photometric       float64             `json:"photometric_interpretation"`
	ColorSpace        float64             `json:"color_space"`
	ComponentsConfig  string              `json:"components_configuration"`
	ICCProfile        string              `json:"icc_profile"`
	Make              string              `json:"make"`
	Model             string              `json:"model"`
	Software          string              `json:"software"`
	LensModel         string              `json:"lens_model"`
	SerialNumber      string              `json:"serial_number"`
	DateTime          string              `json:"datetime"`
	DateTimeOriginal  string              `json:"datetime_original"`
	DateTimeDigitized string              `json:"datetime_digitized"`
	SubSecTime        string              `json:"subsec_time"`
	ExposureTime      float64             `json:"exposure_time"`
	FNumber           float64             `json:"f_number"`
	ISO               float64             `json:"iso"`
	ShutterSpeed      float64             `json:"shutter_speed"`
	ApertureValue     float64             `json:"aperture_value"`
	BrightnessValue   float64             `json:"brightness_value"`
	ExposureBias      float64             `json:"exposure_bias"`
	MeteringMode      float64             `json:"metering_mode"`
	Flash             float64             `json:"flash"`
	FocalLength       float64             `json:"focal_length"`
	WhiteBalance      float64             `json:"white_balance"`
	ExposureProgram   float64             `json:"exposure_program"`
	MaxApertureValue  float64             `json:"max_aperture_value"`
	GPSLatitude       float64             `json:"gps_latitude"`
	GPSLongitude      float64             `json:"gps_longitude"`
	GPSAltitude       float64             `json:"gps_altitude"`
	GPSDate           string              `json:"gps_date"`
	GPSTime           string              `json:"gps_time"`
	ImageDescription  string              `json:"image_description"`
	UserComment       string              `json:"user_comment"`
	Copyright         string              `json:"copyright"`
	Artist            string              `json:"artist"`
	Classification    ClassificationModel `json:"classification"`
	CreatedAt         time.Time           `json:"created_at"`
}

// ClassificationModel holds the AI/heuristic classification result for an image.
type ClassificationModel struct {
	Category      ClassificationCategory `json:"category"`
	Confidence    float64                `json:"confidence"`
	SuggestedName string                 `json:"suggested_name"`
}

// ImageGroupBy selects the ordering/grouping of the image listing.
// Moved from the files core: grouping images is an image-domain concept.
type ImageGroupBy string

const (
	ImageGroupByDate ImageGroupBy = "date"
	ImageGroupByType ImageGroupBy = "type"
	ImageGroupByName ImageGroupBy = "name"
)

func ParseImageGroupBy(value string) (ImageGroupBy, error) {
	groupBy := ImageGroupBy(value)
	switch groupBy {
	case "", ImageGroupByDate:
		return ImageGroupByDate, nil
	case ImageGroupByType, ImageGroupByName:
		return groupBy, nil
	default:
		return "", fmt.Errorf("invalid image group_by: %s", value)
	}
}
