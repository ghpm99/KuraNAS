package vacuum

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type stubService struct {
	settings         SettingsDto
	status           StatusDto
	preferences      PreferencesDto
	consumables      ConsumablesDto
	totals           TotalsDto
	discovery        DiscoveryDto
	failure          error
	receivedSettings UpdateSettingsDto
	receivedCommand  string
	receivedDir      string
	receivedPatch    UpdatePreferencesDto
	receivedPart     string
	calledMethods    []string

	cloudSettings        CloudSettingsDto
	robotMap             MapDto
	rooms                RoomsDto
	room                 RoomDto
	path                 PathDto
	receivedCloud        UpdateCloudSettingsDto
	receivedRoomID       int
	receivedRoomUpdate   UpdateRoomDto
	receivedRoomCleaning RoomCleaningDto
	receivedSpotCleaning SpotCleaningDto
}

func (s *stubService) record(method string) { s.calledMethods = append(s.calledMethods, method) }

func (s *stubService) GetSettings() (SettingsDto, error) {
	s.record("GetSettings")
	return s.settings, s.failure
}

func (s *stubService) UpdateSettings(request UpdateSettingsDto) (SettingsDto, error) {
	s.record("UpdateSettings")
	s.receivedSettings = request
	return s.settings, s.failure
}

func (s *stubService) TestConnection(context.Context) error {
	s.record("TestConnection")
	return s.failure
}

func (s *stubService) Discover(context.Context) (DiscoveryDto, error) {
	s.record("Discover")
	return s.discovery, s.failure
}

func (s *stubService) GetStatus(context.Context) (StatusDto, error) {
	s.record("GetStatus")
	return s.status, s.failure
}

func (s *stubService) SendCommand(_ context.Context, command string) error {
	s.record("SendCommand")
	s.receivedCommand = command
	return s.failure
}

func (s *stubService) SendRemoteControl(_ context.Context, direction string) error {
	s.record("SendRemoteControl")
	s.receivedDir = direction
	return s.failure
}

func (s *stubService) GetPreferences(context.Context) (PreferencesDto, error) {
	s.record("GetPreferences")
	return s.preferences, s.failure
}

func (s *stubService) UpdatePreferences(_ context.Context, patch UpdatePreferencesDto) (PreferencesDto, error) {
	s.record("UpdatePreferences")
	s.receivedPatch = patch
	return s.preferences, s.failure
}

func (s *stubService) GetConsumables(context.Context) (ConsumablesDto, error) {
	s.record("GetConsumables")
	return s.consumables, s.failure
}

func (s *stubService) ResetConsumable(_ context.Context, part string) error {
	s.record("ResetConsumable")
	s.receivedPart = part
	return s.failure
}

func (s *stubService) GetTotals(context.Context) (TotalsDto, error) {
	s.record("GetTotals")
	return s.totals, s.failure
}

func (s *stubService) GetCloudSettings() (CloudSettingsDto, error) {
	s.record("GetCloudSettings")
	return s.cloudSettings, s.failure
}

func (s *stubService) UpdateCloudSettings(request UpdateCloudSettingsDto) (CloudSettingsDto, error) {
	s.record("UpdateCloudSettings")
	s.receivedCloud = request
	return s.cloudSettings, s.failure
}

func (s *stubService) TestCloudConnection(context.Context) error {
	s.record("TestCloudConnection")
	return s.failure
}

func (s *stubService) GetMap(context.Context) (MapDto, error) {
	s.record("GetMap")
	return s.robotMap, s.failure
}

func (s *stubService) GetRooms(context.Context) (RoomsDto, error) {
	s.record("GetRooms")
	return s.rooms, s.failure
}

func (s *stubService) UpdateRoom(_ context.Context, roomID int, request UpdateRoomDto) (RoomDto, error) {
	s.record("UpdateRoom")
	s.receivedRoomID = roomID
	s.receivedRoomUpdate = request
	return s.room, s.failure
}

func (s *stubService) GetPath() PathDto {
	s.record("GetPath")
	return s.path
}

func (s *stubService) CleanRooms(_ context.Context, request RoomCleaningDto) error {
	s.record("CleanRooms")
	s.receivedRoomCleaning = request
	return s.failure
}

func (s *stubService) CleanSpot(_ context.Context, request SpotCleaningDto) error {
	s.record("CleanSpot")
	s.receivedSpotCleaning = request
	return s.failure
}

func newHandlerRouter(service ServiceInterface) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewHandler(service)
	group := router.Group("/vacuum")
	group.GET("/settings", handler.GetSettingsHandler)
	group.PUT("/settings", handler.UpdateSettingsHandler)
	group.POST("/connection-test", handler.TestConnectionHandler)
	group.POST("/discovery", handler.DiscoverHandler)
	group.GET("/status", handler.GetStatusHandler)
	group.POST("/commands", handler.SendCommandHandler)
	group.POST("/remote-control", handler.SendRemoteControlHandler)
	group.GET("/preferences", handler.GetPreferencesHandler)
	group.PATCH("/preferences", handler.UpdatePreferencesHandler)
	group.GET("/consumables", handler.GetConsumablesHandler)
	group.POST("/consumables/:part/reset", handler.ResetConsumableHandler)
	group.GET("/totals", handler.GetTotalsHandler)
	group.GET("/cloud-settings", handler.GetCloudSettingsHandler)
	group.PUT("/cloud-settings", handler.UpdateCloudSettingsHandler)
	group.POST("/cloud-connection-test", handler.TestCloudConnectionHandler)
	group.GET("/map", handler.GetMapHandler)
	group.GET("/rooms", handler.GetRoomsHandler)
	group.PUT("/rooms/:id", handler.UpdateRoomHandler)
	group.GET("/path", handler.GetPathHandler)
	group.POST("/room-cleaning", handler.CleanRoomsHandler)
	group.POST("/spot-cleaning", handler.CleanSpotHandler)
	return router
}

func perform(router *gin.Engine, method string, path string, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func assertJSONEquals(t *testing.T, recorder *httptest.ResponseRecorder, wantJSON string) {
	t.Helper()
	var got, want any
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatalf("response is not JSON: %q", recorder.Body.String())
	}
	if err := json.Unmarshal([]byte(wantJSON), &want); err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("got %s, want %s", recorder.Body.String(), wantJSON)
	}
}

func TestPutSettingsDecodesExactFieldNames(t *testing.T) {
	service := &stubService{settings: SettingsDto{Address: "10.0.0.5", DeviceID: "id-1", HasLocalKey: true}}
	router := newHandlerRouter(service)

	response := perform(router, http.MethodPut, "/vacuum/settings", `{"address":"10.0.0.5","device_id":"id-1","local_key":"abcdefghijklmnop"}`)

	if response.Code != http.StatusOK {
		t.Fatalf("got %d: %s", response.Code, response.Body.String())
	}
	want := UpdateSettingsDto{Address: "10.0.0.5", DeviceID: "id-1", LocalKey: "abcdefghijklmnop"}
	if service.receivedSettings != want {
		t.Fatalf("decoded %+v, want %+v", service.receivedSettings, want)
	}
	assertJSONEquals(t, response, `{"address":"10.0.0.5","device_id":"id-1","has_local_key":true}`)
}

func TestPostCommandDecodesExactFieldName(t *testing.T) {
	service := &stubService{}
	router := newHandlerRouter(service)

	response := perform(router, http.MethodPost, "/vacuum/commands", `{"command":"pause"}`)

	if response.Code != http.StatusNoContent || service.receivedCommand != "pause" {
		t.Fatalf("got %d with command %q", response.Code, service.receivedCommand)
	}
}

func TestPostRemoteControlDecodesExactFieldName(t *testing.T) {
	service := &stubService{}
	router := newHandlerRouter(service)

	response := perform(router, http.MethodPost, "/vacuum/remote-control", `{"direction":"left"}`)

	if response.Code != http.StatusNoContent || service.receivedDir != "left" {
		t.Fatalf("got %d with direction %q", response.Code, service.receivedDir)
	}
}

func TestPatchPreferencesDecodesExactFieldNames(t *testing.T) {
	service := &stubService{}
	router := newHandlerRouter(service)

	body := `{"suction":"strong","water_level":"low","volume":4,"is_do_not_disturb":true,"is_carpet_boost":false,` +
		`"is_resume_after_charge":true,"is_auto_empty":false,"auto_empty_frequency":"2"}`
	response := perform(router, http.MethodPatch, "/vacuum/preferences", body)

	if response.Code != http.StatusOK {
		t.Fatalf("got %d: %s", response.Code, response.Body.String())
	}
	want := UpdatePreferencesDto{
		Suction: textOf("strong"), WaterLevel: textOf("low"), Volume: numberOf(4),
		IsDoNotDisturb: flagOf(true), IsCarpetBoost: flagOf(false), IsResumeAfterCharge: flagOf(true),
		IsAutoEmpty: flagOf(false), AutoEmptyFrequency: textOf("2"),
	}
	if fmt.Sprintf("%+v", dereferenced(service.receivedPatch)) != fmt.Sprintf("%+v", dereferenced(want)) {
		t.Fatalf("decoded %+v, want %+v", dereferenced(service.receivedPatch), dereferenced(want))
	}
}

func TestPatchPreferencesWithPartialBodyLeavesOthersNil(t *testing.T) {
	service := &stubService{}
	router := newHandlerRouter(service)

	perform(router, http.MethodPatch, "/vacuum/preferences", `{"volume":0}`)

	patch := service.receivedPatch
	if patch.Volume == nil || *patch.Volume != 0 || patch.Suction != nil || patch.IsAutoEmpty != nil {
		t.Fatalf("unexpected decoded patch: %+v", dereferenced(patch))
	}
}

func dereferenced(preferences PreferencesDto) map[string]any {
	decoded := map[string]any{}
	encoded, _ := json.Marshal(preferences)
	_ = json.Unmarshal(encoded, &decoded)
	return decoded
}

func TestPostResetDecodesPartFromPath(t *testing.T) {
	service := &stubService{}
	router := newHandlerRouter(service)

	response := perform(router, http.MethodPost, "/vacuum/consumables/main_brush/reset", "")

	if response.Code != http.StatusNoContent || service.receivedPart != "main_brush" {
		t.Fatalf("got %d with part %q", response.Code, service.receivedPart)
	}
}

func TestPostDiscoveryReturnsAddress(t *testing.T) {
	service := &stubService{discovery: DiscoveryDto{Address: "10.0.0.7"}}
	router := newHandlerRouter(service)

	response := perform(router, http.MethodPost, "/vacuum/discovery", "")

	if response.Code != http.StatusOK {
		t.Fatalf("got %d", response.Code)
	}
	assertJSONEquals(t, response, `{"address":"10.0.0.7"}`)
}

func TestPostConnectionTestIsNoContent(t *testing.T) {
	service := &stubService{}
	router := newHandlerRouter(service)

	response := perform(router, http.MethodPost, "/vacuum/connection-test", "")

	if response.Code != http.StatusNoContent || len(service.calledMethods) != 1 || service.calledMethods[0] != "TestConnection" {
		t.Fatalf("got %d calling %v", response.Code, service.calledMethods)
	}
}

func TestGetResponsesUseExactSpecFieldNames(t *testing.T) {
	readAt := time.Date(2026, 11, 16, 10, 0, 0, 0, time.FixedZone("BRT", -3*3600))
	service := &stubService{
		settings:    SettingsDto{Address: "10.0.0.5", DeviceID: "id-1", HasLocalKey: false},
		status:      StatusDto{State: "cleaning", BatteryPercent: 80, CleaningSeconds: 60, CleaningArea: 3, FaultCodes: []int{2}, ReadAt: readAt},
		preferences: PreferencesDto{Suction: textOf("auto"), IsAutoEmpty: flagOf(true)},
		consumables: ConsumablesDto{Items: []ConsumableDto{{Part: "filter", Seconds: 5, MaxSeconds: 540000}}},
		totals:      TotalsDto{CleaningSeconds: 1, CleaningArea: 2, CleaningCount: 3},
	}
	router := newHandlerRouter(service)

	tests := []struct {
		path string
		want string
	}{
		{"/vacuum/settings", `{"address":"10.0.0.5","device_id":"id-1","has_local_key":false}`},
		{"/vacuum/status", `{"state":"cleaning","battery_percent":80,"cleaning_seconds":60,"cleaning_area":3,"fault_codes":[2],"read_at":"2026-11-16T10:00:00-03:00"}`},
		{"/vacuum/preferences", `{"suction":"auto","water_level":null,"volume":null,"is_do_not_disturb":null,"is_carpet_boost":null,"is_resume_after_charge":null,"is_auto_empty":true,"auto_empty_frequency":null}`},
		{"/vacuum/consumables", `{"items":[{"part":"filter","seconds":5,"max_seconds":540000}]}`},
		{"/vacuum/totals", `{"cleaning_seconds":1,"cleaning_area":2,"cleaning_count":3}`},
	}
	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			response := perform(router, http.MethodGet, test.path, "")
			if response.Code != http.StatusOK {
				t.Fatalf("got %d", response.Code)
			}
			assertJSONEquals(t, response, test.want)
		})
	}
}

func TestStatusEmptyFaultCodesSerializeAsArray(t *testing.T) {
	router := newHandlerRouter(&stubService{status: StatusDto{FaultCodes: []int{}}})

	response := perform(router, http.MethodGet, "/vacuum/status", "")

	if !strings.Contains(response.Body.String(), `"fault_codes":[]`) {
		t.Fatalf("got %s", response.Body.String())
	}
}

func TestErrorsMapToTheStatusOfTheSpec(t *testing.T) {
	tests := []struct {
		failure    error
		wantStatus int
	}{
		{ErrNotConfigured, http.StatusConflict},
		{ErrDeviceUnreachable, http.StatusBadGateway},
		{fmt.Errorf("wrapped: %w", ErrLocalKeyRejected), http.StatusBadGateway},
		{ErrDiscoveryNotFound, http.StatusNotFound},
		{ErrInvalidCommand, http.StatusBadRequest},
		{ErrInvalidDirection, http.StatusBadRequest},
		{ErrInvalidPart, http.StatusBadRequest},
		{ErrInvalidPreference, http.StatusBadRequest},
		{ErrInvalidAddress, http.StatusBadRequest},
		{ErrInvalidDeviceID, http.StatusBadRequest},
		{ErrInvalidLocalKey, http.StatusBadRequest},
		{ErrLocalKeyRequired, http.StatusBadRequest},
		{ErrInvalidRequest, http.StatusBadRequest},
		{errors.New("raw database failure with secrets"), http.StatusInternalServerError},
	}
	requests := []struct{ method, path, body string }{
		{http.MethodGet, "/vacuum/settings", ""},
		{http.MethodPut, "/vacuum/settings", `{"device_id":"x"}`},
		{http.MethodPost, "/vacuum/connection-test", ""},
		{http.MethodPost, "/vacuum/discovery", ""},
		{http.MethodGet, "/vacuum/status", ""},
		{http.MethodPost, "/vacuum/commands", `{"command":"start"}`},
		{http.MethodPost, "/vacuum/remote-control", `{"direction":"stop"}`},
		{http.MethodGet, "/vacuum/preferences", ""},
		{http.MethodPatch, "/vacuum/preferences", `{"volume":1}`},
		{http.MethodGet, "/vacuum/consumables", ""},
		{http.MethodPost, "/vacuum/consumables/filter/reset", ""},
		{http.MethodGet, "/vacuum/totals", ""},
	}
	for _, test := range tests {
		for _, request := range requests {
			t.Run(fmt.Sprintf("%v %s %s", test.failure, request.method, request.path), func(t *testing.T) {
				router := newHandlerRouter(&stubService{failure: test.failure})

				response := perform(router, request.method, request.path, request.body)

				if response.Code != test.wantStatus {
					t.Fatalf("got %d, want %d", response.Code, test.wantStatus)
				}
				var body map[string]string
				if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil || body["error"] == "" {
					t.Fatalf("expected an {\"error\": text} body, got %q", response.Body.String())
				}
				if strings.Contains(body["error"], "secrets") {
					t.Fatalf("raw error leaked: %q", body["error"])
				}
			})
		}
	}
}

func TestMalformedBodiesAreBadRequests(t *testing.T) {
	requests := []struct{ method, path string }{
		{http.MethodPut, "/vacuum/settings"},
		{http.MethodPost, "/vacuum/commands"},
		{http.MethodPost, "/vacuum/remote-control"},
		{http.MethodPatch, "/vacuum/preferences"},
	}
	for _, request := range requests {
		t.Run(request.method+" "+request.path, func(t *testing.T) {
			service := &stubService{}
			router := newHandlerRouter(service)

			response := perform(router, request.method, request.path, `{"broken`)

			if response.Code != http.StatusBadRequest || len(service.calledMethods) != 0 {
				t.Fatalf("got %d calling %v", response.Code, service.calledMethods)
			}
		})
	}
}

func TestWrongTypedPreferenceIsBadRequest(t *testing.T) {
	router := newHandlerRouter(&stubService{})

	response := perform(router, http.MethodPatch, "/vacuum/preferences", `{"auto_empty_frequency":2}`)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("got %d", response.Code)
	}
}
