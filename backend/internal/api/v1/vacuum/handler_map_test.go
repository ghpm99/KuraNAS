package vacuum

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestPutCloudSettingsDecodesExactFieldNamesAndNeverEchoesTheSecret(t *testing.T) {
	service := &stubService{cloudSettings: CloudSettingsDto{AccessID: "id-1", Region: "eu", HasAccessSecret: true}}
	router := newHandlerRouter(service)

	response := perform(router, http.MethodPut, "/vacuum/cloud-settings", `{"access_id":"id-1","access_secret":"`+testCloudAccessSecret+`","region":"eu"}`)

	if response.Code != http.StatusOK {
		t.Fatalf("got %d: %s", response.Code, response.Body.String())
	}
	want := UpdateCloudSettingsDto{AccessID: "id-1", AccessSecret: testCloudAccessSecret, Region: "eu"}
	if service.receivedCloud != want {
		t.Fatalf("decoded %+v, want %+v", service.receivedCloud, want)
	}
	assertJSONEquals(t, response, `{"access_id":"id-1","region":"eu","has_access_secret":true}`)
}

func TestGetCloudSettingsFormat(t *testing.T) {
	router := newHandlerRouter(&stubService{cloudSettings: CloudSettingsDto{}})

	response := perform(router, http.MethodGet, "/vacuum/cloud-settings", "")

	assertJSONEquals(t, response, `{"access_id":"","region":"","has_access_secret":false}`)
}

func TestPostCloudConnectionTestAnswersNoContent(t *testing.T) {
	service := &stubService{}
	router := newHandlerRouter(service)

	response := perform(router, http.MethodPost, "/vacuum/cloud-connection-test", "")

	if response.Code != http.StatusNoContent || !reflect.DeepEqual(service.calledMethods, []string{"TestCloudConnection"}) {
		t.Fatalf("got %d with calls %v", response.Code, service.calledMethods)
	}
}

func TestGetMapFormat(t *testing.T) {
	readAt := time.Date(2026, 11, 25, 12, 0, 0, 0, time.UTC)
	service := &stubService{robotMap: MapDto{
		MapID: 9, Width: 2, Height: 1, Resolution: 50, OriginX: -100, OriginY: 200, Cells: "AX8=",
		Dock: &PointDto{X: 1, Y: 2}, ForbiddenZones: []ForbiddenZoneDto{{Vertices: [][2]int{{1, 2}, {3, 4}}}}, ReadAt: readAt,
	}}
	router := newHandlerRouter(service)

	response := perform(router, http.MethodGet, "/vacuum/map", "")

	assertJSONEquals(t, response, `{"map_id":9,"width":2,"height":1,"resolution":50,"origin_x":-100,"origin_y":200,"cells":"AX8=","dock":{"x":1,"y":2},"forbidden_zones":[{"vertices":[[1,2],[3,4]]}],"read_at":"2026-11-25T12:00:00Z"}`)
}

func TestGetMapWithoutDockSerializesNull(t *testing.T) {
	router := newHandlerRouter(&stubService{robotMap: MapDto{ForbiddenZones: []ForbiddenZoneDto{}}})

	response := perform(router, http.MethodGet, "/vacuum/map", "")

	assertJSONEquals(t, response, `{"map_id":0,"width":0,"height":0,"resolution":0,"origin_x":0,"origin_y":0,"cells":"","dock":null,"forbidden_zones":[],"read_at":"0001-01-01T00:00:00Z"}`)
}

func TestGetRoomsFormat(t *testing.T) {
	name := "Sala"
	service := &stubService{rooms: RoomsDto{MapID: 9, Items: []RoomDto{{ID: 1, Name: &name, Area: 12.5}, {ID: 2, Area: 3}}}}
	router := newHandlerRouter(service)

	response := perform(router, http.MethodGet, "/vacuum/rooms", "")

	assertJSONEquals(t, response, `{"map_id":9,"items":[{"id":1,"name":"Sala","area":12.5},{"id":2,"name":null,"area":3}]}`)
}

func TestPutRoomDecodesExactFieldNamesAndThePathID(t *testing.T) {
	name := "Cozinha"
	service := &stubService{room: RoomDto{ID: 4, Name: &name, Area: 8}}
	router := newHandlerRouter(service)

	response := perform(router, http.MethodPut, "/vacuum/rooms/4", `{"map_id":9,"name":"Cozinha"}`)

	if response.Code != http.StatusOK {
		t.Fatalf("got %d: %s", response.Code, response.Body.String())
	}
	if service.receivedRoomID != 4 || service.receivedRoomUpdate != (UpdateRoomDto{MapID: 9, Name: "Cozinha"}) {
		t.Fatalf("decoded %d %+v", service.receivedRoomID, service.receivedRoomUpdate)
	}
	assertJSONEquals(t, response, `{"id":4,"name":"Cozinha","area":8}`)
}

func TestPutRoomWithNonNumericIDIsABadRequest(t *testing.T) {
	service := &stubService{}
	router := newHandlerRouter(service)

	response := perform(router, http.MethodPut, "/vacuum/rooms/abc", `{"map_id":9,"name":"x"}`)

	if response.Code != http.StatusBadRequest || len(service.calledMethods) != 0 {
		t.Fatalf("got %d with calls %v", response.Code, service.calledMethods)
	}
}

func TestGetPathFormat(t *testing.T) {
	pathID := int64(55)
	service := &stubService{path: PathDto{PathID: &pathID, Points: [][2]int{{1, 2}, {3, 4}}, Position: &PointDto{X: 3, Y: 4}}}
	router := newHandlerRouter(service)

	response := perform(router, http.MethodGet, "/vacuum/path", "")

	assertJSONEquals(t, response, `{"path_id":55,"points":[[1,2],[3,4]],"position":{"x":3,"y":4}}`)
}

func TestGetPathWithoutPathSerializesNullsAndEmptyPoints(t *testing.T) {
	router := newHandlerRouter(&stubService{path: PathDto{Points: [][2]int{}}})

	response := perform(router, http.MethodGet, "/vacuum/path", "")

	assertJSONEquals(t, response, `{"path_id":null,"points":[],"position":null}`)
}

func TestPostRoomCleaningDecodesExactFieldNames(t *testing.T) {
	service := &stubService{}
	router := newHandlerRouter(service)

	response := perform(router, http.MethodPost, "/vacuum/room-cleaning", `{"map_id":9,"room_ids":[1,3]}`)

	if response.Code != http.StatusNoContent {
		t.Fatalf("got %d: %s", response.Code, response.Body.String())
	}
	if !reflect.DeepEqual(service.receivedRoomCleaning, RoomCleaningDto{MapID: 9, RoomIDs: []int{1, 3}}) {
		t.Fatalf("decoded %+v", service.receivedRoomCleaning)
	}
}

func TestPostSpotCleaningDecodesExactFieldNames(t *testing.T) {
	service := &stubService{}
	router := newHandlerRouter(service)

	response := perform(router, http.MethodPost, "/vacuum/spot-cleaning", `{"map_id":9,"x":-250,"y":1300.5}`)

	if response.Code != http.StatusNoContent {
		t.Fatalf("got %d: %s", response.Code, response.Body.String())
	}
	if service.receivedSpotCleaning != (SpotCleaningDto{MapID: 9, X: -250, Y: 1300.5}) {
		t.Fatalf("decoded %+v", service.receivedSpotCleaning)
	}
}

func TestMapEndpointsRejectMalformedBodies(t *testing.T) {
	routes := []struct{ method, path string }{
		{http.MethodPut, "/vacuum/cloud-settings"},
		{http.MethodPut, "/vacuum/rooms/1"},
		{http.MethodPost, "/vacuum/room-cleaning"},
		{http.MethodPost, "/vacuum/spot-cleaning"},
	}
	for _, route := range routes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			service := &stubService{}
			response := perform(newHandlerRouter(service), route.method, route.path, `{"map_id":"not a number"`)

			if response.Code != http.StatusBadRequest || len(service.calledMethods) != 0 {
				t.Fatalf("got %d with calls %v", response.Code, service.calledMethods)
			}
		})
	}
}

func TestMapEndpointsTranslateServiceFailuresToStatuses(t *testing.T) {
	failures := []struct {
		failure error
		status  int
	}{
		{ErrCloudNotConfigured, http.StatusConflict},
		{ErrNotConfigured, http.StatusConflict},
		{ErrMapChanged, http.StatusConflict},
		{ErrCloudUnreachable, http.StatusBadGateway},
		{ErrCloudCredentialsRejected, http.StatusBadGateway},
		{ErrCloudServiceNotAuthorized, http.StatusBadGateway},
		{ErrCloudInvalidResponse, http.StatusBadGateway},
		{ErrRoomNotFound, http.StatusBadRequest},
		{ErrNoRoomSelected, http.StatusBadRequest},
		{ErrInvalidRoomName, http.StatusBadRequest},
		{ErrSpotNotOnFloor, http.StatusBadRequest},
		{ErrInvalidCloudAccessID, http.StatusBadRequest},
		{ErrInvalidCloudAccessSecret, http.StatusBadRequest},
		{ErrCloudAccessSecretRequired, http.StatusBadRequest},
		{ErrInvalidCloudRegion, http.StatusBadRequest},
		{errors.New("unexpected"), http.StatusInternalServerError},
	}
	calls := []struct{ method, path, body string }{
		{http.MethodGet, "/vacuum/cloud-settings", ""},
		{http.MethodPut, "/vacuum/cloud-settings", `{}`},
		{http.MethodPost, "/vacuum/cloud-connection-test", ""},
		{http.MethodGet, "/vacuum/map", ""},
		{http.MethodGet, "/vacuum/rooms", ""},
		{http.MethodPut, "/vacuum/rooms/1", `{}`},
		{http.MethodPost, "/vacuum/room-cleaning", `{}`},
		{http.MethodPost, "/vacuum/spot-cleaning", `{}`},
	}
	for _, failure := range failures {
		for _, call := range calls {
			t.Run(fmt.Sprintf("%v %s", failure.failure, call.path), func(t *testing.T) {
				router := newHandlerRouter(&stubService{failure: failure.failure})

				response := perform(router, call.method, call.path, call.body)

				if response.Code != failure.status {
					t.Fatalf("got %d, want %d: %s", response.Code, failure.status, response.Body.String())
				}
				assertErrorBodyHasMessage(t, response.Body.String())
			})
		}
	}
}

func assertErrorBodyHasMessage(t *testing.T, body string) {
	t.Helper()
	var decoded map[string]string
	if err := json.Unmarshal([]byte(body), &decoded); err != nil || decoded["error"] == "" {
		t.Fatalf("expected an error message, got %q", body)
	}
}
