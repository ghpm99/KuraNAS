package img

import (
	"bytes"
	"image"
	"image/jpeg"
	"os"

	"nas-go/api/pkg/utils"
)

type Preview struct {
	Image       image.Image
	Orientation int
}

func OpenPreview(path string, extension string, converter StillImageConverter) (Preview, error) {
	switch {
	case utils.IsRawPhotoExtension(extension):
		return openEmbeddedRawPreview(path)
	case utils.IsHeifFamilyExtension(extension):
		return openConvertedStill(path, converter)
	}
	return openDecodablePreview(path)
}

func PreviewDimensions(path string, extension string, converter StillImageConverter) (int, int, error) {
	switch {
	case utils.IsRawPhotoExtension(extension):
		embedded, err := ExtractEmbeddedPreview(path)
		if err != nil {
			return 0, 0, err
		}
		return jpegDimensions(embedded.JPEG)
	case utils.IsHeifFamilyExtension(extension):
		convertedJPEG, err := converter.ConvertToJPEG(path)
		if err != nil {
			return 0, 0, err
		}
		return jpegDimensions(convertedJPEG)
	}
	return fileDimensions(path)
}

func openEmbeddedRawPreview(path string) (Preview, error) {
	embedded, err := ExtractEmbeddedPreview(path)
	if err != nil {
		return Preview{}, err
	}
	decoded, err := jpeg.Decode(bytes.NewReader(embedded.JPEG))
	if err != nil {
		return Preview{}, err
	}
	return Preview{Image: decoded, Orientation: embedded.Orientation}, nil
}

func openConvertedStill(path string, converter StillImageConverter) (Preview, error) {
	convertedJPEG, err := converter.ConvertToJPEG(path)
	if err != nil {
		return Preview{}, err
	}
	decoded, err := jpeg.Decode(bytes.NewReader(convertedJPEG))
	if err != nil {
		return Preview{}, err
	}
	return Preview{Image: decoded, Orientation: orientationTopLeft}, nil
}

func openDecodablePreview(path string) (Preview, error) {
	decoded, format, err := OpenImageFromFile(path)
	if err != nil {
		return Preview{}, err
	}
	orientation := orientationTopLeft
	if format == "jpeg" {
		orientation = ReadJPEGOrientation(path)
	}
	return Preview{Image: decoded, Orientation: orientation}, nil
}

func jpegDimensions(jpegBytes []byte) (int, int, error) {
	config, err := jpeg.DecodeConfig(bytes.NewReader(jpegBytes))
	if err != nil {
		return 0, 0, err
	}
	return config.Width, config.Height, nil
}

func fileDimensions(path string) (int, int, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	config, _, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, err
	}
	return config.Width, config.Height, nil
}
