package vacuum

import (
	"context"
	"fmt"
	"math"
)

func (s *Service) CleanRooms(ctx context.Context, request RoomCleaningDto) error {
	roomIDs := distinctRoomIDs(request.RoomIDs)
	if len(roomIDs) == 0 {
		return ErrNoRoomSelected
	}
	snapshot, err := s.mapWithID(ctx, request.MapID)
	if err != nil {
		return err
	}

	cellCounts := roomCellCounts(snapshot.layout)
	for _, roomID := range roomIDs {
		if _, isOnMap := cellCounts[roomID]; !isOnMap {
			return fmt.Errorf("%w: %d", ErrRoomNotFound, roomID)
		}
	}
	return s.writeMapCommand(ctx, roomCleaningCommand(s.now(), request.MapID, roomIDs))
}

func (s *Service) CleanSpot(ctx context.Context, request SpotCleaningDto) error {
	snapshot, err := s.mapWithID(ctx, request.MapID)
	if err != nil {
		return err
	}

	centerX, centerY := int(math.Round(request.X)), int(math.Round(request.Y))
	if _, isOnRoomFloor := roomCellAt(snapshot.layout, centerX, centerY); !isOnRoomFloor {
		return fmt.Errorf("%w: (%d, %d)", ErrSpotNotOnFloor, centerX, centerY)
	}
	return s.writeMapCommand(ctx, spotCleaningCommand(s.now(), request.MapID, centerX, centerY))
}

func (s *Service) writeMapCommand(ctx context.Context, command mapCommandEnvelope) error {
	commandValue, err := command.asDataPointValue()
	if err != nil {
		return err
	}
	return s.writeDataPoints(ctx, dataPointWrite{dpMapCommand, commandValue}.asPayload())
}

func distinctRoomIDs(roomIDs []int) []int {
	seen := map[int]bool{}
	distinct := []int{}
	for _, roomID := range roomIDs {
		if seen[roomID] {
			continue
		}
		seen[roomID] = true
		distinct = append(distinct, roomID)
	}
	return distinct
}
