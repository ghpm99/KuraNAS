package vacuum

import (
	"context"
	"encoding/base64"
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestGetMapConvertsTheLayoutToTheContract(t *testing.T) {
	fixture := newServiceFixture(t)

	robotMap, err := fixture.service.GetMap(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if robotMap.MapID != testMapID || robotMap.Width != 6 || robotMap.Height != 4 || robotMap.Resolution != 50 {
		t.Fatalf("unexpected header: %+v", robotMap)
	}
	if robotMap.OriginX != -100 || robotMap.OriginY != 200 {
		t.Fatalf("unexpected origin: %+v", robotMap)
	}
	cells, err := base64.StdEncoding.DecodeString(robotMap.Cells)
	if err != nil || !reflect.DeepEqual(cells, testMapLayout().Cells) {
		t.Fatalf("unexpected cells: %v %v", cells, err)
	}
	if robotMap.Dock == nil || *robotMap.Dock != (PointDto{X: -50, Y: 250}) {
		t.Fatalf("unexpected dock: %+v", robotMap.Dock)
	}
	if !reflect.DeepEqual(robotMap.ForbiddenZones, []ForbiddenZoneDto{{Vertices: [][2]int{{1, 2}, {3, 4}}}}) {
		t.Fatalf("unexpected zones: %+v", robotMap.ForbiddenZones)
	}
	if !robotMap.ReadAt.Equal(fixture.clock.instant) {
		t.Fatalf("unexpected read_at: %v", robotMap.ReadAt)
	}
}

func TestGetMapWithoutDockAndZonesHasNullDockAndEmptyZones(t *testing.T) {
	fixture := newServiceFixture(t)
	fixture.cloud.layout.Dock = nil
	fixture.cloud.layout.ForbiddenZones = nil

	robotMap, err := fixture.service.GetMap(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if robotMap.Dock != nil || robotMap.ForbiddenZones == nil || len(robotMap.ForbiddenZones) != 0 {
		t.Fatalf("unexpected dock/zones: %+v %+v", robotMap.Dock, robotMap.ForbiddenZones)
	}
}

func TestGetMapUsesTheCloudCredentialsAndTheRobotDeviceID(t *testing.T) {
	fixture := newServiceFixture(t)

	if _, err := fixture.service.GetMap(context.Background()); err != nil {
		t.Fatal(err)
	}
	credentials := fixture.cloud.credentials[0]
	if credentials.AccessID != testCloudAccessID || credentials.AccessSecret != testCloudAccessSecret || credentials.Region != "eu" {
		t.Fatalf("unexpected credentials: %+v", credentials)
	}
	if fixture.cloud.deviceIDs[0] != testDeviceID {
		t.Fatalf("unexpected device id: %v", fixture.cloud.deviceIDs)
	}
}

func TestGetMapIsCachedForThirtySeconds(t *testing.T) {
	fixture := newServiceFixture(t)
	read := func() {
		if _, err := fixture.service.GetMap(context.Background()); err != nil {
			t.Fatal(err)
		}
	}

	read()
	fixture.clock.advance(29 * time.Second)
	read()
	if fixture.cloud.calls != 1 {
		t.Fatalf("map must be served from the cache, got %d cloud reads", fixture.cloud.calls)
	}

	fixture.clock.advance(2 * time.Second)
	read()
	if fixture.cloud.calls != 2 {
		t.Fatalf("map must be read again after 30 seconds, got %d cloud reads", fixture.cloud.calls)
	}
}

func TestFailedMapReadIsNotCached(t *testing.T) {
	fixture := newServiceFixture(t)
	fixture.cloud.failure = errors.New("boom")
	_, _ = fixture.service.GetMap(context.Background())
	fixture.cloud.failure = nil

	if _, err := fixture.service.GetMap(context.Background()); err != nil {
		t.Fatal(err)
	}
	if fixture.cloud.calls != 2 {
		t.Fatalf("got %d cloud reads", fixture.cloud.calls)
	}
}

func TestGetMapWithoutCloudCredentialsIsItsOwnFailure(t *testing.T) {
	fixture := newServiceFixture(t)
	fixture.repository.isCloudFound = false

	_, err := fixture.service.GetMap(context.Background())
	if !errors.Is(err, ErrCloudNotConfigured) || errors.Is(err, ErrNotConfigured) {
		t.Fatalf("got %v", err)
	}
	if fixture.cloud.calls != 0 {
		t.Fatal("the cloud must not be called without credentials")
	}
}

func TestGetMapWithoutRobotDeviceID(t *testing.T) {
	fixture := newServiceFixture(t)
	fixture.repository.isFound = false

	if _, err := fixture.service.GetMap(context.Background()); !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("got %v", err)
	}
}

func TestGetMapPropagatesStorageFailures(t *testing.T) {
	fixture := newServiceFixture(t)
	fixture.repository.getErr = errors.New("db down")

	if _, err := fixture.service.GetMap(context.Background()); err == nil {
		t.Fatal("expected a storage failure")
	}
}

func TestChangingTheRobotDropsTheCachedMap(t *testing.T) {
	fixture := newServiceFixture(t)
	if _, err := fixture.service.GetMap(context.Background()); err != nil {
		t.Fatal(err)
	}

	if _, err := fixture.service.UpdateSettings(UpdateSettingsDto{Address: testAddress, DeviceID: "another-device"}); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.service.GetMap(context.Background()); err != nil {
		t.Fatal(err)
	}
	if fixture.cloud.calls != 2 || fixture.cloud.deviceIDs[1] != "another-device" {
		t.Fatalf("unexpected cloud reads: %d %v", fixture.cloud.calls, fixture.cloud.deviceIDs)
	}
}

func TestGetRoomsListsDistinctRoomsWithAreaAndNames(t *testing.T) {
	fixture := newServiceFixture(t)
	name := "Sala"
	fixture.repository.roomDocument = `{"map_id":777,"names":{"1":"Sala"}}`
	fixture.repository.isRoomFound = true

	rooms, err := fixture.service.GetRooms(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	want := RoomsDto{MapID: testMapID, Items: []RoomDto{
		{ID: 1, Name: &name, Area: 0},
		{ID: 2, Name: nil, Area: 0},
	}}
	if !reflect.DeepEqual(rooms, want) {
		t.Fatalf("got %+v, want %+v", rooms, want)
	}
}

func TestRoomAreaIsCellsTimesZeroPointZeroZeroTwentyFiveRoundedToOneDecimal(t *testing.T) {
	fixture := newServiceFixture(t)
	width, height := 40, 40
	cells := make([]byte, width*height)
	for index := range cells {
		cells[index] = 3
	}
	fixture.cloud.layout.Width, fixture.cloud.layout.Height, fixture.cloud.layout.Cells = width, height, cells

	rooms, err := fixture.service.GetRooms(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(rooms.Items) != 1 || rooms.Items[0].ID != 3 || rooms.Items[0].Area != 4 {
		t.Fatalf("unexpected rooms: %+v", rooms)
	}
}

func TestRoomNamesOfAnotherMapAreNotApplied(t *testing.T) {
	fixture := newServiceFixture(t)
	fixture.repository.roomDocument = `{"map_id":1,"names":{"1":"Sala"}}`
	fixture.repository.isRoomFound = true

	rooms, err := fixture.service.GetRooms(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, room := range rooms.Items {
		if room.Name != nil {
			t.Fatalf("name of an old map leaked: %+v", room)
		}
	}
}

func TestGetRoomsFailures(t *testing.T) {
	fixture := newServiceFixture(t)
	fixture.repository.isCloudFound = false
	if _, err := fixture.service.GetRooms(context.Background()); !errors.Is(err, ErrCloudNotConfigured) {
		t.Fatalf("got %v", err)
	}

	fixture = newServiceFixture(t)
	fixture.repository.roomDocument = "not json"
	fixture.repository.isRoomFound = true
	if _, err := fixture.service.GetRooms(context.Background()); err == nil {
		t.Fatal("expected a decode failure")
	}
}

func TestUpdateRoomStoresTheTrimmedNameAndReturnsTheItem(t *testing.T) {
	fixture := newServiceFixture(t)

	room, err := fixture.service.UpdateRoom(context.Background(), 2, UpdateRoomDto{MapID: testMapID, Name: "  Cozinha "})
	if err != nil {
		t.Fatal(err)
	}
	if room.ID != 2 || room.Name == nil || *room.Name != "Cozinha" {
		t.Fatalf("unexpected room: %+v", room)
	}
	if fixture.repository.roomDocument != `{"map_id":777,"names":{"2":"Cozinha"}}` {
		t.Fatalf("unexpected stored document: %s", fixture.repository.roomDocument)
	}
}

func TestUpdateRoomWithEmptyNameErasesIt(t *testing.T) {
	fixture := newServiceFixture(t)
	fixture.repository.roomDocument = `{"map_id":777,"names":{"1":"Sala","2":"Cozinha"}}`
	fixture.repository.isRoomFound = true

	room, err := fixture.service.UpdateRoom(context.Background(), 2, UpdateRoomDto{MapID: testMapID, Name: "   "})
	if err != nil {
		t.Fatal(err)
	}
	if room.Name != nil {
		t.Fatalf("name must be null after erasing: %+v", room)
	}
	if fixture.repository.roomDocument != `{"map_id":777,"names":{"1":"Sala"}}` {
		t.Fatalf("unexpected stored document: %s", fixture.repository.roomDocument)
	}
}

func TestUpdateRoomDiscardsNamesOfThePreviousMap(t *testing.T) {
	fixture := newServiceFixture(t)
	fixture.repository.roomDocument = `{"map_id":1,"names":{"1":"Sala antiga"}}`
	fixture.repository.isRoomFound = true

	if _, err := fixture.service.UpdateRoom(context.Background(), 2, UpdateRoomDto{MapID: testMapID, Name: "Cozinha"}); err != nil {
		t.Fatal(err)
	}
	if fixture.repository.roomDocument != `{"map_id":777,"names":{"2":"Cozinha"}}` {
		t.Fatalf("old names must be dropped: %s", fixture.repository.roomDocument)
	}
}

func TestUpdateRoomValidation(t *testing.T) {
	longName := ""
	for count := 0; count <= maxRoomNameLength; count++ {
		longName += "ç"
	}
	tests := []struct {
		name    string
		roomID  int
		request UpdateRoomDto
		want    error
	}{
		{"map changed", 2, UpdateRoomDto{MapID: 1, Name: "x"}, ErrMapChanged},
		{"room not on the map", 9, UpdateRoomDto{MapID: testMapID, Name: "x"}, ErrRoomNotFound},
		{"cell value that is not a room", 127, UpdateRoomDto{MapID: testMapID, Name: "x"}, ErrRoomNotFound},
		{"name too long", 2, UpdateRoomDto{MapID: testMapID, Name: longName}, ErrInvalidRoomName},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newServiceFixture(t)

			_, err := fixture.service.UpdateRoom(context.Background(), test.roomID, test.request)
			if !errors.Is(err, test.want) {
				t.Fatalf("got %v, want %v", err, test.want)
			}
			if fixture.repository.isRoomFound {
				t.Fatal("nothing may be stored")
			}
		})
	}
}

func TestUpdateRoomAcceptsAFortyCharacterName(t *testing.T) {
	fixture := newServiceFixture(t)
	name := ""
	for count := 0; count < maxRoomNameLength; count++ {
		name += "ç"
	}

	if _, err := fixture.service.UpdateRoom(context.Background(), 1, UpdateRoomDto{MapID: testMapID, Name: name}); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateRoomStorageFailures(t *testing.T) {
	fixture := newServiceFixture(t)
	fixture.repository.roomUpsertErr = errors.New("db down")
	if _, err := fixture.service.UpdateRoom(context.Background(), 1, UpdateRoomDto{MapID: testMapID, Name: "x"}); err == nil {
		t.Fatal("expected a write failure")
	}

	fixture = newServiceFixture(t)
	fixture.repository.roomDocument = "not json"
	fixture.repository.isRoomFound = true
	if _, err := fixture.service.UpdateRoom(context.Background(), 1, UpdateRoomDto{MapID: testMapID, Name: "x"}); err == nil {
		t.Fatal("expected a decode failure")
	}
}
