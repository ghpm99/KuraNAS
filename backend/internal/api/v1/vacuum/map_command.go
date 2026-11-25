package vacuum

import (
	"encoding/base64"
	"encoding/json"
	"strconv"
	"time"
)

const (
	mapCommandInfoType      = 30000
	cleanAreaInfoType       = 21023
	startCleaningInfoType   = 21005
	startCleaningMode       = "reAppointClean"
	mapCommandUserID        = "0"
	wholeMapCleanID         = -3
	spotAreaID              = 100
	spotAreaMode            = "point"
	spotAreaActive          = "depth"
	spotAreaTag             = "room"
	spotHalfSideMillimeters = 750
)

type mapCommandEnvelope struct {
	InfoType int               `json:"infoType"`
	Info     mapCommandInfo    `json:"dInfo"`
	Data     mapCommandPayload `json:"data"`
}

type mapCommandInfo struct {
	Timestamp string `json:"ts"`
	UserID    string `json:"userId"`
}

type mapCommandPayload struct {
	Commands     []mapCommandStep `json:"cmds"`
	MainCommands []int            `json:"mainCmds"`
}

type mapCommandStep struct {
	InfoType int `json:"infoType"`
	Data     any `json:"data"`
}

type cleanAreaData struct {
	CleanIDs   []int       `json:"cleanId"`
	ExtraAreas []extraArea `json:"extraAreas"`
	MapID      int64       `json:"mapId"`
	SegmentIDs []int       `json:"segmentId,omitempty"`
}

type extraArea struct {
	Active   string   `json:"active"`
	ID       int      `json:"id"`
	Mode     string   `json:"mode"`
	Name     string   `json:"name"`
	Tag      string   `json:"tag"`
	Vertexes [][2]int `json:"vertexs"`
}

type startCleaningData struct {
	Mode string `json:"mode"`
}

func newMapCommand(issuedAt time.Time, cleanArea cleanAreaData) mapCommandEnvelope {
	return mapCommandEnvelope{
		InfoType: mapCommandInfoType,
		Info:     mapCommandInfo{Timestamp: strconv.FormatInt(issuedAt.UnixMilli(), 10), UserID: mapCommandUserID},
		Data: mapCommandPayload{
			Commands: []mapCommandStep{
				{InfoType: cleanAreaInfoType, Data: cleanArea},
				{InfoType: startCleaningInfoType, Data: startCleaningData{Mode: startCleaningMode}},
			},
			MainCommands: []int{startCleaningInfoType},
		},
	}
}

func roomCleaningCommand(issuedAt time.Time, mapID int64, roomIDs []int) mapCommandEnvelope {
	return newMapCommand(issuedAt, cleanAreaData{
		CleanIDs:   []int{wholeMapCleanID},
		ExtraAreas: []extraArea{},
		MapID:      mapID,
		SegmentIDs: roomIDs,
	})
}

func spotCleaningCommand(issuedAt time.Time, mapID int64, centerX int, centerY int) mapCommandEnvelope {
	square := extraArea{
		Active: spotAreaActive,
		ID:     spotAreaID,
		Mode:   spotAreaMode,
		Name:   "",
		Tag:    spotAreaTag,
		Vertexes: [][2]int{
			{centerX - spotHalfSideMillimeters, centerY - spotHalfSideMillimeters},
			{centerX + spotHalfSideMillimeters, centerY - spotHalfSideMillimeters},
			{centerX + spotHalfSideMillimeters, centerY + spotHalfSideMillimeters},
			{centerX - spotHalfSideMillimeters, centerY + spotHalfSideMillimeters},
		},
	}
	return newMapCommand(issuedAt, cleanAreaData{
		CleanIDs:   []int{wholeMapCleanID},
		ExtraAreas: []extraArea{square},
		MapID:      mapID,
	})
}

func (command mapCommandEnvelope) asDataPointValue() (string, error) {
	commandJSON, err := json.Marshal(command)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(commandJSON), nil
}
