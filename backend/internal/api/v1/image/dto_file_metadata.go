package image

type ImageSummaryDto struct {
	Width            int     `json:"width"`
	Height           int     `json:"height"`
	Make             string  `json:"make"`
	Model            string  `json:"model"`
	LensModel        string  `json:"lens_model"`
	DateTimeOriginal string  `json:"datetime_original"`
	ExposureTime     float64 `json:"exposure_time"`
	FNumber          float64 `json:"f_number"`
	ISO              float64 `json:"iso"`
	FocalLength      float64 `json:"focal_length"`
}
