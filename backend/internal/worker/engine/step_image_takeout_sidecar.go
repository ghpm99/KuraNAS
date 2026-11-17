package engine

import (
	"nas-go/api/internal/api/v1/files"
	imagedom "nas-go/api/internal/api/v1/image"
	"nas-go/api/internal/api/v1/takeout"
	"nas-go/api/pkg/applog"
	"strings"
)

func applyTakeoutSidecarToImageMetadata(context *WorkerContext, fileDto files.FileDto, metadata *imagedom.MetadataModel) {
	if context.TakeoutSidecars == nil {
		return
	}
	sidecar, err := context.TakeoutSidecars.FindSidecarByPath(fileDto.Path)
	if err != nil {
		applog.Warn("metadata step: takeout sidecar lookup failed", "path", fileDto.Path, "error", err.Error())
		return
	}
	if sidecar == nil {
		return
	}
	if fillImageMetadataGapsFromSidecar(metadata, *sidecar) {
		metadata.Classification = imagedom.ClassifyImage(fileDto, *metadata)
	}
}

func fillImageMetadataGapsFromSidecar(metadata *imagedom.MetadataModel, sidecar takeout.PhotoSidecar) bool {
	hasFilledDescription := fillDescriptionFromSidecar(metadata, sidecar)
	hasFilledLocation := fillLocationFromSidecar(metadata, sidecar)
	return hasFilledDescription || hasFilledLocation
}

func fillDescriptionFromSidecar(metadata *imagedom.MetadataModel, sidecar takeout.PhotoSidecar) bool {
	if strings.TrimSpace(metadata.ImageDescription) != "" || sidecar.Description == "" {
		return false
	}
	metadata.ImageDescription = sidecar.Description
	return true
}

func fillLocationFromSidecar(metadata *imagedom.MetadataModel, sidecar takeout.PhotoSidecar) bool {
	hasExifLocation := metadata.GPSLatitude != 0 || metadata.GPSLongitude != 0
	if hasExifLocation || !sidecar.HasLocation() {
		return false
	}
	metadata.GPSLatitude = *sidecar.Latitude
	metadata.GPSLongitude = *sidecar.Longitude
	return true
}
