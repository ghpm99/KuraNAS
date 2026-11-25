package vacuum

import (
	"math"
	"sort"

	"nas-go/api/pkg/tuyacloud"
)

const (
	firstRoomCell       = 1
	lastRoomCell        = 126
	squareMetersPerCell = 0.0025
)

func isRoomCell(cell byte) bool {
	return cell >= firstRoomCell && cell <= lastRoomCell
}

func roomCellCounts(layout tuyacloud.MapLayout) map[int]int {
	cellCounts := map[int]int{}
	for _, cell := range layout.Cells {
		if isRoomCell(cell) {
			cellCounts[int(cell)]++
		}
	}
	return cellCounts
}

func sortedRoomIDs(cellCounts map[int]int) []int {
	roomIDs := make([]int, 0, len(cellCounts))
	for roomID := range cellCounts {
		roomIDs = append(roomIDs, roomID)
	}
	sort.Ints(roomIDs)
	return roomIDs
}

func squareMetersOf(cellCount int) float64 {
	return math.Round(float64(cellCount)*squareMetersPerCell*10) / 10
}

func roomCellAt(layout tuyacloud.MapLayout, xMillimeters int, yMillimeters int) (roomID int, isOnRoomFloor bool) {
	column := floorDivide(xMillimeters-layout.OriginXMM, layout.ResolutionMM)
	row := floorDivide(yMillimeters-layout.OriginYMM, layout.ResolutionMM)
	if column < 0 || column >= layout.Width || row < 0 || row >= layout.Height {
		return 0, false
	}
	cell := layout.Cells[row*layout.Width+column]
	return int(cell), isRoomCell(cell)
}

func floorDivide(dividend int, divisor int) int {
	return int(math.Floor(float64(dividend) / float64(divisor)))
}
