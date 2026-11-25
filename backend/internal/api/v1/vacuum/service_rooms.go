package vacuum

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"nas-go/api/pkg/tuyacloud"
)

const maxRoomNameLength = 40

func (s *Service) GetRooms(ctx context.Context) (RoomsDto, error) {
	snapshot, err := s.currentMap(ctx)
	if err != nil {
		return RoomsDto{}, err
	}
	names, err := s.loadRoomNames()
	if err != nil {
		return RoomsDto{}, err
	}

	namesOfThisMap := names.forMap(snapshot.layout.MapID)
	cellCounts := roomCellCounts(snapshot.layout)
	rooms := []RoomDto{}
	for _, roomID := range sortedRoomIDs(cellCounts) {
		rooms = append(rooms, newRoomDto(roomID, cellCounts[roomID], namesOfThisMap))
	}
	return RoomsDto{MapID: snapshot.layout.MapID, Items: rooms}, nil
}

func (s *Service) UpdateRoom(ctx context.Context, roomID int, request UpdateRoomDto) (RoomDto, error) {
	name := strings.TrimSpace(request.Name)
	if utf8.RuneCountInString(name) > maxRoomNameLength {
		return RoomDto{}, fmt.Errorf("%w: more than %d characters", ErrInvalidRoomName, maxRoomNameLength)
	}
	snapshot, err := s.mapWithID(ctx, request.MapID)
	if err != nil {
		return RoomDto{}, err
	}
	cellCount, isOnMap := roomCellCounts(snapshot.layout)[roomID]
	if !isOnMap {
		return RoomDto{}, fmt.Errorf("%w: %d", ErrRoomNotFound, roomID)
	}

	names, err := s.storeRoomName(snapshot.layout, roomID, name)
	if err != nil {
		return RoomDto{}, err
	}
	return newRoomDto(roomID, cellCount, names), nil
}

func newRoomDto(roomID int, cellCount int, names roomNamesState) RoomDto {
	return RoomDto{ID: roomID, Name: names.nameOf(roomID), Area: squareMetersOf(cellCount)}
}

func (s *Service) storeRoomName(layout tuyacloud.MapLayout, roomID int, name string) (roomNamesState, error) {
	s.roomNamesMutex.Lock()
	defer s.roomNamesMutex.Unlock()

	stored, err := s.loadRoomNames()
	if err != nil {
		return roomNamesState{}, err
	}
	updated := stored.forMap(layout.MapID).withName(roomID, name)
	if err := s.saveRoomNames(updated); err != nil {
		return roomNamesState{}, err
	}
	return updated, nil
}

func (s *Service) loadRoomNames() (roomNamesState, error) {
	document, isFound, err := s.repository.GetRoomNamesDocument()
	if err != nil {
		return roomNamesState{}, err
	}
	if !isFound {
		return roomNamesState{}, nil
	}
	return decodeRoomNames(document)
}

func (s *Service) saveRoomNames(names roomNamesState) error {
	document, err := encodeRoomNames(names)
	if err != nil {
		return err
	}
	return s.repository.UpsertRoomNamesDocument(document)
}
