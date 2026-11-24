package music

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

type playlistRulesRecordingServiceMock struct {
	musicHandlerServiceMock
	receivedPreviewRules PlaylistRules
	receivedCreate       CreatePlaylistRequest
	receivedUpdate       UpdatePlaylistRequest
	failure              error
}

func (m *playlistRulesRecordingServiceMock) PreviewPlaylistRules(rules PlaylistRules) (PlaylistRulePreviewDto, error) {
	m.receivedPreviewRules = rules
	if m.failure != nil {
		return PlaylistRulePreviewDto{}, m.failure
	}
	return PlaylistRulePreviewDto{TrackCount: 412}, nil
}

func (m *playlistRulesRecordingServiceMock) CreatePlaylist(req CreatePlaylistRequest) (PlaylistDto, error) {
	m.receivedCreate = req
	if m.failure != nil {
		return PlaylistDto{}, m.failure
	}
	return PlaylistDto{ID: 3, Name: req.Name, Kind: PlaylistKindRule, Rules: req.Rules, TotalLengthSeconds: 90.5}, nil
}

func (m *playlistRulesRecordingServiceMock) UpdatePlaylist(id int, req UpdatePlaylistRequest) (PlaylistDto, error) {
	m.receivedUpdate = req
	return PlaylistDto{ID: id, Name: req.Name, Kind: PlaylistKindRule, Rules: req.Rules}, m.failure
}

func newPlaylistRulesRouter(service ServiceInterface) *gin.Engine {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(service, nil, &musicRecentServiceMock{}, &musicLoggerMock{})
	router := gin.New()
	router.POST("/music/playlists", handler.CreatePlaylistHandler)
	router.POST("/music/playlists/rules/preview", handler.PreviewPlaylistRulesHandler)
	router.PUT("/music/playlists/:id", handler.UpdatePlaylistHandler)
	return router
}

const rulesPayload = `{"match":"all","genres":["Rock","Metal"],"artists":[],"albums":["a::b"],"folders":["/Music"],"year_from":1990,"year_to":null,"favorites_only":true}`

func assertDecodedRules(t *testing.T, rules PlaylistRules) {
	t.Helper()
	if rules.Match != "all" || len(rules.Genres) != 2 || rules.Genres[1] != "Metal" || len(rules.Albums) != 1 ||
		len(rules.Folders) != 1 || rules.YearFrom == nil || *rules.YearFrom != 1990 || rules.YearTo != nil || !rules.FavoritesOnly {
		t.Fatalf("rules were not decoded from the spec payload: %+v", rules)
	}
}

func TestPreviewHandlerDecodesRulesAndReturnsTrackCount(t *testing.T) {
	service := &playlistRulesRecordingServiceMock{}
	router := newPlaylistRulesRouter(service)

	recorder := performPlaylistRequest(router, http.MethodPost, "/music/playlists/rules/preview", rulesPayload)

	if recorder.Code != http.StatusOK || recorder.Body.String() != `{"track_count":412}` {
		t.Fatalf("unexpected response %d %s", recorder.Code, recorder.Body.String())
	}
	assertDecodedRules(t, service.receivedPreviewRules)
}

func TestCreateHandlerDecodesRulesAndEncodesKindRulesAndLength(t *testing.T) {
	service := &playlistRulesRecordingServiceMock{}
	router := newPlaylistRulesRouter(service)

	recorder := performPlaylistRequest(router, http.MethodPost, "/music/playlists", `{"name":"Heavy","rules":`+rulesPayload+`}`)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", recorder.Code, recorder.Body.String())
	}
	assertDecodedRules(t, *service.receivedCreate.Rules)

	var body map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["kind"] != "rule" || body["total_length_seconds"] != 90.5 || body["rules"] == nil {
		t.Fatalf("unexpected playlist body: %v", body)
	}
}

func TestCreateHandlerWithoutRulesLeavesRulesNil(t *testing.T) {
	service := &playlistRulesRecordingServiceMock{}
	router := newPlaylistRulesRouter(service)

	performPlaylistRequest(router, http.MethodPost, "/music/playlists", `{"name":"Plain"}`)

	if service.receivedCreate.Rules != nil {
		t.Fatalf("rules must stay nil, got %+v", service.receivedCreate.Rules)
	}
}

func TestUpdateHandlerDecodesRules(t *testing.T) {
	service := &playlistRulesRecordingServiceMock{}
	router := newPlaylistRulesRouter(service)

	recorder := performPlaylistRequest(router, http.MethodPut, "/music/playlists/3", `{"name":"Heavy","rules":`+rulesPayload+`}`)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	assertDecodedRules(t, *service.receivedUpdate.Rules)
}

func TestRuleValidationRefusalsAnswer400WithTranslatedMessageAndField(t *testing.T) {
	refusals := []struct {
		name    string
		failure error
		field   string
	}{
		{"empty rule", &fieldValidationError{Field: "rules", MessageKey: "ERROR_MUSIC_PLAYLIST_RULE_EMPTY"}, "rules"},
		{"rules on manual playlist", ErrPlaylistRulesNotAllowed, "rules"},
	}

	for _, refusal := range refusals {
		service := &playlistRulesRecordingServiceMock{failure: refusal.failure}
		router := newPlaylistRulesRouter(service)
		requests := []struct{ method, path, body string }{
			{http.MethodPost, "/music/playlists/rules/preview", `{"match":"any"}`},
			{http.MethodPost, "/music/playlists", `{"name":"x","rules":{"match":"any"}}`},
			{http.MethodPut, "/music/playlists/3", `{"name":"x","rules":{"match":"any"}}`},
		}
		for _, request := range requests {
			recorder := performPlaylistRequest(router, request.method, request.path, request.body)
			var body map[string]string
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
				t.Fatalf("%s %s: decode: %v", refusal.name, request.path, err)
			}
			if recorder.Code != http.StatusBadRequest || body["field"] != refusal.field || body["error"] == "" {
				t.Fatalf("%s %s: expected 400 with field, got %d %s", refusal.name, request.path, recorder.Code, recorder.Body.String())
			}
		}
	}
}

func TestPreviewHandlerRejectsMalformedBody(t *testing.T) {
	router := newPlaylistRulesRouter(&playlistRulesRecordingServiceMock{})

	recorder := performPlaylistRequest(router, http.MethodPost, "/music/playlists/rules/preview", `{"match":`)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}
