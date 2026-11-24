package ingest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"testing"
)

func decodeJSONObject(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode object: %v: %s", err, rec.Body.String())
	}
	return body
}

func decodeJSONObjectList(t *testing.T, rec *httptest.ResponseRecorder) []map[string]any {
	t.Helper()
	var body []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode list: %v: %s", err, rec.Body.String())
	}
	return body
}

func sortedKeys(object map[string]any) []string {
	keys := make([]string, 0, len(object))
	for key := range object {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func assertExactFieldNames(t *testing.T, object map[string]any, wantNames []string) {
	t.Helper()
	sort.Strings(wantNames)
	if got := sortedKeys(object); !reflect.DeepEqual(got, wantNames) {
		t.Fatalf("field names = %v, want %v", got, wantNames)
	}
}

func TestYtDlpStatusResponseFieldNames(t *testing.T) {
	router := newYtDlpRouter(&fakeYtDlpService{status: YtDlpStatusDto{
		Installed:       true,
		CurrentVersion:  "2024.08.06",
		LatestVersion:   "2024.09.01",
		UpdateAvailable: true,
		ReleaseURL:      "https://github.com/yt-dlp/yt-dlp/releases/tag/2024.09.01",
		ReleaseDate:     "2024-09-01",
	}})
	rec := doRequest(router, http.MethodGet, "/ingest/ytdlp/status", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	body := decodeJSONObject(t, rec)
	assertExactFieldNames(t, body, []string{"installed", "current_version", "latest_version", "update_available", "release_url", "release_date"})
	if body["installed"] != true || body["update_available"] != true {
		t.Fatalf("boolean fields = %v", body)
	}
	if body["current_version"] != "2024.08.06" || body["latest_version"] != "2024.09.01" || body["release_date"] != "2024-09-01" {
		t.Fatalf("string fields = %v", body)
	}
	if body["release_url"] != "https://github.com/yt-dlp/yt-dlp/releases/tag/2024.09.01" {
		t.Fatalf("release_url = %v", body["release_url"])
	}
}

func TestYtDlpStatusResponseKeepsFieldNamesWhenToolIsMissing(t *testing.T) {
	router := newYtDlpRouter(&fakeYtDlpService{})
	rec := doRequest(router, http.MethodGet, "/ingest/ytdlp/status", "")

	body := decodeJSONObject(t, rec)
	assertExactFieldNames(t, body, []string{"installed", "current_version", "latest_version", "update_available", "release_url", "release_date"})
	if body["installed"] != false {
		t.Fatalf("installed = %v, want false", body["installed"])
	}
}

func TestYtDlpUpdateResponseFieldNames(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		router := newYtDlpRouter(&fakeYtDlpService{})
		rec := doRequest(router, http.MethodPost, "/ingest/ytdlp/update", "")
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		assertExactFieldNames(t, decodeJSONObject(t, rec), []string{"message"})
	})
	t.Run("conflict", func(t *testing.T) {
		router := newYtDlpRouter(&fakeYtDlpService{updateErr: ErrYtDlpUpdateInProgress})
		rec := doRequest(router, http.MethodPost, "/ingest/ytdlp/update", "")
		assertExactFieldNames(t, decodeJSONObject(t, rec), []string{"error"})
	})
	t.Run("failure", func(t *testing.T) {
		router := newYtDlpRouter(&fakeYtDlpService{updateErr: errString("boom")})
		rec := doRequest(router, http.MethodPost, "/ingest/ytdlp/update", "")
		assertExactFieldNames(t, decodeJSONObject(t, rec), []string{"error"})
	})
}

func TestTargetsResponseFieldNames(t *testing.T) {
	router := newTestRouter(&fakeService{targets: []TargetDto{
		{Label: "Midia", Path: "/srv/midia"},
		{Label: "Videos", Path: "/srv/videos"},
	}})
	rec := doRequest(router, http.MethodGet, "/ingest/targets", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	targets := decodeJSONObjectList(t, rec)
	if len(targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(targets))
	}
	for _, target := range targets {
		assertExactFieldNames(t, target, []string{"label", "path"})
	}
	if targets[0]["label"] != "Midia" || targets[0]["path"] != "/srv/midia" {
		t.Fatalf("first target = %v", targets[0])
	}
}

func TestTargetsResponseIsEmptyArrayWhenNoTargetIsEnabled(t *testing.T) {
	router := newTestRouter(&fakeService{targets: []TargetDto{}})
	rec := doRequest(router, http.MethodGet, "/ingest/targets", "")
	if got := rec.Body.String(); got != "[]" {
		t.Fatalf("body = %s, want []", got)
	}
}

func TestPresetsResponseFieldNames(t *testing.T) {
	router := newTestRouter(&fakeService{presets: availablePresets()})
	rec := doRequest(router, http.MethodGet, "/ingest/presets", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	presetItems := decodeJSONObjectList(t, rec)
	if len(presetItems) != len(presets) {
		t.Fatalf("expected %d presets, got %d", len(presets), len(presetItems))
	}
	for index, presetItem := range presetItems {
		assertExactFieldNames(t, presetItem, []string{"key", "label"})
		if presetItem["key"] != presets[index].Key {
			t.Fatalf("preset %d key = %v, want %s", index, presetItem["key"], presets[index].Key)
		}
		if label, _ := presetItem["label"].(string); label == "" {
			t.Fatalf("preset %d has empty label", index)
		}
	}
}

func TestFetchResponseFieldNames(t *testing.T) {
	router := newTestRouter(&fakeService{fetchResponse: FetchResponseDto{JobID: 7, DownloadID: 12}})
	rec := doRequest(router, http.MethodPost, "/ingest/fetch", `{"url":"https://x.test/v","preset":"audio_mp3","target_root":"/srv"}`)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rec.Code)
	}
	body := decodeJSONObject(t, rec)
	assertExactFieldNames(t, body, []string{"download_id", "job_id"})
	if body["job_id"] != float64(7) || body["download_id"] != float64(12) {
		t.Fatalf("unexpected values: %v", body)
	}
}
