package vacuum

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"testing"

	"nas-go/api/pkg/applog"
	"nas-go/api/pkg/tuya"
)

const secretLocalKey = "Zq9!secretKey_16"

func captureApplog(t *testing.T) *bytes.Buffer {
	t.Helper()
	capturedLogs := &bytes.Buffer{}
	applog.Setup(applog.Options{Writer: capturedLogs, Level: slog.LevelDebug})
	t.Cleanup(func() { applog.Setup(applog.Options{Writer: os.Stdout, Level: slog.LevelInfo}) })
	return capturedLogs
}

func TestLocalKeyNeverAppearsInResponsesOrLogs(t *testing.T) {
	capturedLogs := captureApplog(t)
	fixture := newServiceFixture(t)
	fixture.repository.isFound = false
	fixture.client.readFn = func(device tuya.Device) (map[string]any, error) {
		if device.LocalKey != secretLocalKey {
			t.Errorf("device received the wrong key")
		}
		return map[string]any{dpState: "mystery", dpSuction: "auto"}, nil
	}
	router := newHandlerRouter(fixture.service)

	var allResponses strings.Builder
	call := func(method string, path string, body string) int {
		response := perform(router, method, path, body)
		allResponses.WriteString(fmt.Sprintf("%d %v %s\n", response.Code, response.Header(), response.Body.String()))
		return response.Code
	}

	call(http.MethodPut, "/vacuum/settings", `{"address":"robot.local","device_id":"id-1","local_key":"`+secretLocalKey+`"}`)
	call(http.MethodPut, "/vacuum/settings", `{"address":"robot.local","device_id":"id-1","local_key":"short"}`)
	call(http.MethodGet, "/vacuum/settings", "")
	call(http.MethodGet, "/vacuum/status", "")
	call(http.MethodGet, "/vacuum/preferences", "")
	call(http.MethodPatch, "/vacuum/preferences", `{"suction":"max"}`)
	call(http.MethodPatch, "/vacuum/preferences", `{"suction":"bogus"}`)
	call(http.MethodGet, "/vacuum/consumables", "")
	call(http.MethodGet, "/vacuum/totals", "")
	call(http.MethodPost, "/vacuum/commands", `{"command":"start"}`)
	call(http.MethodPost, "/vacuum/commands", `{"command":"bogus"}`)
	call(http.MethodPost, "/vacuum/remote-control", `{"direction":"left"}`)
	call(http.MethodPost, "/vacuum/consumables/filter/reset", "")
	call(http.MethodPost, "/vacuum/connection-test", "")

	fixture.client.readFn = func(tuya.Device) (map[string]any, error) {
		return nil, fmt.Errorf("%w: reply cannot be decrypted", tuya.ErrKeyRejected)
	}
	call(http.MethodPost, "/vacuum/connection-test", "")

	fixture.client.readFn = func(tuya.Device) (map[string]any, error) {
		return nil, fmt.Errorf("%w: dial failed", tuya.ErrUnreachable)
	}
	fixture.client.findFn = func(tuya.Device, []string) (string, error) { return "", tuya.ErrDeviceNotFound }
	call(http.MethodPost, "/vacuum/connection-test", "")
	call(http.MethodPost, "/vacuum/discovery", "")

	fixture.client.findFn = func(tuya.Device, []string) (string, error) { return "10.1.1.2", nil }
	call(http.MethodPost, "/vacuum/discovery", "")

	if strings.Contains(allResponses.String(), secretLocalKey) {
		t.Fatalf("local key leaked in a response:\n%s", allResponses.String())
	}
	if strings.Contains(capturedLogs.String(), secretLocalKey) {
		t.Fatalf("local key leaked in the logs:\n%s", capturedLogs.String())
	}
	if capturedLogs.Len() == 0 {
		t.Fatal("expected the flow to produce log lines, so the assertion is meaningful")
	}
	if !strings.Contains(allResponses.String(), `"has_local_key":true`) {
		t.Fatalf("settings did not report the stored key:\n%s", allResponses.String())
	}
}
