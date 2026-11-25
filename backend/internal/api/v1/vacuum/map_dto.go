package vacuum

import (
	"encoding/base64"

	"nas-go/api/pkg/tuyacloud"
)

func newMapDto(snapshot mapSnapshot) MapDto {
	layout := snapshot.layout
	return MapDto{
		MapID:          layout.MapID,
		Width:          layout.Width,
		Height:         layout.Height,
		Resolution:     layout.ResolutionMM,
		OriginX:        layout.OriginXMM,
		OriginY:        layout.OriginYMM,
		Cells:          base64.StdEncoding.EncodeToString(layout.Cells),
		Dock:           dockDto(layout.Dock),
		ForbiddenZones: forbiddenZoneDtos(layout.ForbiddenZones),
		ReadAt:         snapshot.readAt,
	}
}

func dockDto(dock *tuyacloud.Point) *PointDto {
	if dock == nil {
		return nil
	}
	return &PointDto{X: dock.X, Y: dock.Y}
}

func forbiddenZoneDtos(zones []tuyacloud.ForbiddenZone) []ForbiddenZoneDto {
	dtos := make([]ForbiddenZoneDto, 0, len(zones))
	for _, zone := range zones {
		vertices := make([][2]int, 0, len(zone.Vertices))
		for _, vertex := range zone.Vertices {
			vertices = append(vertices, [2]int{vertex.X, vertex.Y})
		}
		dtos = append(dtos, ForbiddenZoneDto{Vertices: vertices})
	}
	return dtos
}
