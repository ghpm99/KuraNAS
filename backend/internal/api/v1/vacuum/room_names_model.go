package vacuum

import (
	"encoding/json"
	"strconv"
)

type roomNamesState struct {
	MapID int64             `json:"map_id"`
	Names map[string]string `json:"names"`
}

func (s roomNamesState) nameOf(roomID int) *string {
	name, isNamed := s.Names[strconv.Itoa(roomID)]
	if !isNamed {
		return nil
	}
	return &name
}

func (s roomNamesState) forMap(mapID int64) roomNamesState {
	if s.MapID == mapID && s.Names != nil {
		return s
	}
	return roomNamesState{MapID: mapID, Names: map[string]string{}}
}

func (s roomNamesState) withName(roomID int, name string) roomNamesState {
	key := strconv.Itoa(roomID)
	if name == "" {
		delete(s.Names, key)
		return s
	}
	s.Names[key] = name
	return s
}

func decodeRoomNames(document string) (roomNamesState, error) {
	var names roomNamesState
	if err := json.Unmarshal([]byte(document), &names); err != nil {
		return roomNamesState{}, err
	}
	return names, nil
}

func encodeRoomNames(names roomNamesState) (string, error) {
	document, err := json.Marshal(names)
	if err != nil {
		return "", err
	}
	return string(document), nil
}
