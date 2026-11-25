package vacuum

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"nas-go/api/pkg/tuya"
)

func decodeWrittenMapCommand(t *testing.T, fixture *serviceFixture) map[string]any {
	t.Helper()
	if len(fixture.client.writes) != 1 {
		t.Fatalf("expected exactly one write, got %+v", fixture.client.writes)
	}
	dataPoints := fixture.client.writes[0].dataPoints
	if len(dataPoints) != 1 {
		t.Fatalf("only data point 127 may be written, got %v", dataPoints)
	}
	encoded, isText := dataPoints[dpMapCommand].(string)
	if !isText {
		t.Fatalf("data point 127 must be a string, got %T", dataPoints[dpMapCommand])
	}
	commandJSON, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatal(err)
	}
	var command map[string]any
	if err := json.Unmarshal(commandJSON, &command); err != nil {
		t.Fatal(err)
	}
	return command
}

func decodeJSON(t *testing.T, document string) map[string]any {
	t.Helper()
	var decoded map[string]any
	if err := json.Unmarshal([]byte(document), &decoded); err != nil {
		t.Fatal(err)
	}
	return decoded
}

func TestCleanRoomsWritesTheExactEnvelopeToDataPoint127(t *testing.T) {
	fixture := newServiceFixture(t)

	if err := fixture.service.CleanRooms(context.Background(), RoomCleaningDto{MapID: testMapID, RoomIDs: []int{2, 1, 2}}); err != nil {
		t.Fatal(err)
	}

	want := decodeJSON(t, `{"infoType":30000,"dInfo":{"ts":"`+
		"1794823200000"+`","userId":"0"},"data":{"cmds":[{"infoType":21023,"data":{"cleanId":[-3],"extraAreas":[],"mapId":777,"segmentId":[2,1]}},{"infoType":21005,"data":{"mode":"reAppointClean"}}],"mainCmds":[21005]}}`)
	got := decodeWrittenMapCommand(t, fixture)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v\nwant %v", got, want)
	}
	if fixture.client.writes[0].device.LocalKey != testLocalKey {
		t.Fatal("write must use the stored credentials")
	}
}

func TestCleanSpotWritesTheSquareAroundThePoint(t *testing.T) {
	fixture := newServiceFixture(t)
	x, y := 0.0, 300.0

	if err := fixture.service.CleanSpot(context.Background(), SpotCleaningDto{MapID: testMapID, X: x, Y: y}); err != nil {
		t.Fatal(err)
	}

	want := decodeJSON(t, `{"infoType":30000,"dInfo":{"ts":"1794823200000","userId":"0"},"data":{"cmds":[{"infoType":21023,"data":{"cleanId":[-3],"extraAreas":[{"active":"depth","id":100,"mode":"point","name":"","tag":"room","vertexs":[[-750,-450],[750,-450],[750,1050],[-750,1050]]}],"mapId":777}},{"infoType":21005,"data":{"mode":"reAppointClean"}}],"mainCmds":[21005]}}`)
	got := decodeWrittenMapCommand(t, fixture)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v\nwant %v", got, want)
	}
}

func TestCleanSpotKeepsTheIntegerPointFromAFractionalRequest(t *testing.T) {
	fixture := newServiceFixture(t)

	if err := fixture.service.CleanSpot(context.Background(), SpotCleaningDto{MapID: testMapID, X: 0.4, Y: 299.6}); err != nil {
		t.Fatal(err)
	}
	command := decodeWrittenMapCommand(t, fixture)
	steps := command["data"].(map[string]any)["cmds"].([]any)
	area := steps[0].(map[string]any)["data"].(map[string]any)["extraAreas"].([]any)[0].(map[string]any)
	if !reflect.DeepEqual(area["vertexs"].([]any)[0], []any{-750.0, -450.0}) {
		t.Fatalf("unexpected vertexes: %v", area["vertexs"])
	}
}

func TestCleanRoomsValidation(t *testing.T) {
	tests := []struct {
		name    string
		request RoomCleaningDto
		want    error
	}{
		{"no rooms", RoomCleaningDto{MapID: testMapID}, ErrNoRoomSelected},
		{"map changed", RoomCleaningDto{MapID: 5, RoomIDs: []int{1}}, ErrMapChanged},
		{"unknown room", RoomCleaningDto{MapID: testMapID, RoomIDs: []int{1, 9}}, ErrRoomNotFound},
		{"obstacle value is not a room", RoomCleaningDto{MapID: testMapID, RoomIDs: []int{255}}, ErrRoomNotFound},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newServiceFixture(t)

			err := fixture.service.CleanRooms(context.Background(), test.request)
			if !errors.Is(err, test.want) {
				t.Fatalf("got %v, want %v", err, test.want)
			}
			if len(fixture.client.writes) != 0 {
				t.Fatal("nothing may be written to the robot")
			}
		})
	}
}

func TestCleanSpotValidation(t *testing.T) {
	tests := []struct {
		name    string
		request SpotCleaningDto
		want    error
	}{
		{"map changed", SpotCleaningDto{MapID: 5, X: 0, Y: 300}, ErrMapChanged},
		{"wall", SpotCleaningDto{MapID: testMapID, X: -100, Y: 200}, ErrSpotNotOnFloor},
		{"unknown area", SpotCleaningDto{MapID: testMapID, X: 0, Y: 450}, ErrSpotNotOnFloor},
		{"obstacle", SpotCleaningDto{MapID: testMapID, X: 125, Y: 300}, ErrSpotNotOnFloor},
		{"left of the grid", SpotCleaningDto{MapID: testMapID, X: -101, Y: 300}, ErrSpotNotOnFloor},
		{"below the grid", SpotCleaningDto{MapID: testMapID, X: 0, Y: 199}, ErrSpotNotOnFloor},
		{"right of the grid", SpotCleaningDto{MapID: testMapID, X: 200, Y: 300}, ErrSpotNotOnFloor},
		{"above the grid", SpotCleaningDto{MapID: testMapID, X: 0, Y: 400}, ErrSpotNotOnFloor},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newServiceFixture(t)

			err := fixture.service.CleanSpot(context.Background(), test.request)
			if !errors.Is(err, test.want) {
				t.Fatalf("got %v, want %v", err, test.want)
			}
			if len(fixture.client.writes) != 0 {
				t.Fatal("nothing may be written to the robot")
			}
		})
	}
}

func TestMapCommandsNeedTheCloudAndTheRobot(t *testing.T) {
	fixture := newServiceFixture(t)
	fixture.repository.isCloudFound = false
	if err := fixture.service.CleanRooms(context.Background(), RoomCleaningDto{MapID: testMapID, RoomIDs: []int{1}}); !errors.Is(err, ErrCloudNotConfigured) {
		t.Fatalf("got %v", err)
	}

	fixture = newServiceFixture(t)
	fixture.repository.document = `{"device_id":"x","address":"a"}`
	if err := fixture.service.CleanSpot(context.Background(), SpotCleaningDto{MapID: testMapID, X: 0, Y: 300}); !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("got %v", err)
	}
}

func TestMapCommandReportsADeviceFailure(t *testing.T) {
	fixture := newServiceFixture(t)
	fixture.client.writeFn = func(_ tuya.Device, _ map[string]any) error { return errors.New("write failed") }

	if err := fixture.service.CleanRooms(context.Background(), RoomCleaningDto{MapID: testMapID, RoomIDs: []int{1}}); err == nil {
		t.Fatal("expected the write failure")
	}
}

func TestDataPoint129IsNeverWritten(t *testing.T) {
	fixture := newServiceFixture(t)
	_ = fixture.service.CleanRooms(context.Background(), RoomCleaningDto{MapID: testMapID, RoomIDs: []int{1}})
	_ = fixture.service.CleanSpot(context.Background(), SpotCleaningDto{MapID: testMapID, X: 0, Y: 300})

	for _, write := range fixture.client.writes {
		if _, isPresent := write.dataPoints["129"]; isPresent {
			t.Fatal("data point 129 erases the robot map and must never be written")
		}
	}
}
