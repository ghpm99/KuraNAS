package tuyacloud

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"

	"github.com/pierrec/lz4/v4"
)

const (
	maxLayoutCells      = 4_000_000
	millimetersPerMeter = 1000
	forbiddenAreaMode   = "forbid"
)

// Point is a position in millimeters.
type Point struct {
	X int
	Y int
}

// ForbiddenZone is a no-go polygon the robot already has, with vertices in millimeters.
type ForbiddenZone struct {
	Vertices []Point
}

// MapLayout is the robot map grid in millimeters. Cells holds Width*Height bytes, row by row,
// starting from the row with the smallest y.
type MapLayout struct {
	MapID          int64
	Width          int
	Height         int
	ResolutionMM   int
	OriginXMM      int
	OriginYMM      int
	Cells          []byte
	Dock           *Point
	ForbiddenZones []ForbiddenZone
}

type layoutFile struct {
	Data struct {
		Width            int          `json:"width"`
		Height           int          `json:"height"`
		ResolutionMeters float64      `json:"resolution"`
		MinXMeters       float64      `json:"x_min"`
		MinYMeters       float64      `json:"y_min"`
		CompressedCells  string       `json:"map"`
		MapID            int64        `json:"mapId"`
		ChargeHandlePos  []float64    `json:"chargeHandlePos"`
		Areas            []layoutArea `json:"area"`
	} `json:"data"`
}

type layoutArea struct {
	Active   string      `json:"active"`
	Vertexes [][]float64 `json:"vertexs"`
}

func decodeMapLayout(fileBody []byte) (MapLayout, error) {
	var file layoutFile
	if err := json.Unmarshal(fileBody, &file); err != nil {
		return MapLayout{}, fmt.Errorf("%w: layout file is not JSON", ErrInvalidResponse)
	}
	data := file.Data

	cells, err := decompressCells(data.CompressedCells, data.Width, data.Height)
	if err != nil {
		return MapLayout{}, err
	}
	return MapLayout{
		MapID:          data.MapID,
		Width:          data.Width,
		Height:         data.Height,
		ResolutionMM:   metersToMillimeters(data.ResolutionMeters),
		OriginXMM:      metersToMillimeters(data.MinXMeters),
		OriginYMM:      metersToMillimeters(data.MinYMeters),
		Cells:          cells,
		Dock:           dockFromPosition(data.ChargeHandlePos),
		ForbiddenZones: forbiddenZonesFromAreas(data.Areas),
	}, nil
}

func decompressCells(encodedCells string, width int, height int) ([]byte, error) {
	if width <= 0 || height <= 0 || width*height > maxLayoutCells {
		return nil, fmt.Errorf("%w: grid size %dx%d", ErrInvalidResponse, width, height)
	}
	compressed, err := base64.StdEncoding.DecodeString(encodedCells)
	if err != nil {
		return nil, fmt.Errorf("%w: cells are not base64", ErrInvalidResponse)
	}

	cells := make([]byte, width*height)
	decompressedSize, err := lz4.UncompressBlock(compressed, cells)
	if err != nil || decompressedSize != len(cells) {
		return nil, fmt.Errorf("%w: cells do not decompress to %d bytes", ErrInvalidResponse, len(cells))
	}
	return cells, nil
}

func metersToMillimeters(meters float64) int {
	return int(math.Round(meters * millimetersPerMeter))
}

func dockFromPosition(position []float64) *Point {
	if len(position) != 2 {
		return nil
	}
	return &Point{X: int(math.Round(position[0])), Y: int(math.Round(position[1]))}
}

func forbiddenZonesFromAreas(areas []layoutArea) []ForbiddenZone {
	zones := []ForbiddenZone{}
	for _, area := range areas {
		if area.Active != forbiddenAreaMode {
			continue
		}
		zones = append(zones, ForbiddenZone{Vertices: pointsFromVertexes(area.Vertexes)})
	}
	return zones
}

func pointsFromVertexes(vertexes [][]float64) []Point {
	points := make([]Point, 0, len(vertexes))
	for _, vertex := range vertexes {
		if len(vertex) != 2 {
			continue
		}
		points = append(points, Point{X: int(math.Round(vertex[0])), Y: int(math.Round(vertex[1]))})
	}
	return points
}
