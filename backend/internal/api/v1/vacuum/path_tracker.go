package vacuum

import (
	"encoding/base64"
	"encoding/json"
	"math"
	"sync"
)

const (
	pathInfoType  = 21011
	maxPathPoints = 100_000
)

type pathPoint struct {
	x         int
	y         int
	isPresent bool
}

type pathReport struct {
	InfoType int `json:"infoType"`
	Data     struct {
		PathID   int64       `json:"pathID"`
		StartPos int         `json:"startPos"`
		PosArray [][]float64 `json:"posArray"`
	} `json:"data"`
}

type pathTracker struct {
	mutex  sync.Mutex
	pathID *int64
	points []pathPoint
}

func (tracker *pathTracker) applyReport(encodedReport string) {
	report, isPathReport := decodePathReport(encodedReport)
	if !isPathReport {
		return
	}

	tracker.mutex.Lock()
	defer tracker.mutex.Unlock()

	if tracker.pathID == nil || *tracker.pathID != report.Data.PathID {
		pathID := report.Data.PathID
		tracker.pathID = &pathID
		tracker.points = nil
	}
	for offset, position := range report.Data.PosArray {
		index := report.Data.StartPos + offset
		if len(position) != 2 || index < 0 || index >= maxPathPoints {
			continue
		}
		tracker.placePoint(index, pathPoint{
			x:         int(math.Round(position[0])),
			y:         int(math.Round(position[1])),
			isPresent: true,
		})
	}
}

func (tracker *pathTracker) placePoint(index int, point pathPoint) {
	for len(tracker.points) <= index {
		tracker.points = append(tracker.points, pathPoint{})
	}
	tracker.points[index] = point
}

func (tracker *pathTracker) reset() {
	tracker.mutex.Lock()
	defer tracker.mutex.Unlock()
	tracker.pathID = nil
	tracker.points = nil
}

func (tracker *pathTracker) snapshot() PathDto {
	tracker.mutex.Lock()
	defer tracker.mutex.Unlock()

	points := [][2]int{}
	for _, point := range tracker.points {
		if point.isPresent {
			points = append(points, [2]int{point.x, point.y})
		}
	}
	snapshot := PathDto{Points: points}
	if tracker.pathID != nil {
		pathID := *tracker.pathID
		snapshot.PathID = &pathID
	}
	if len(points) > 0 {
		lastPoint := points[len(points)-1]
		snapshot.Position = &PointDto{X: lastPoint[0], Y: lastPoint[1]}
	}
	return snapshot
}

func decodePathReport(encodedReport string) (pathReport, bool) {
	reportJSON, err := base64.StdEncoding.DecodeString(encodedReport)
	if err != nil {
		return pathReport{}, false
	}
	var report pathReport
	if err := json.Unmarshal(reportJSON, &report); err != nil {
		return pathReport{}, false
	}
	return report, report.InfoType == pathInfoType
}
