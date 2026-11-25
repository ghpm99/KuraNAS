package tuyacloud

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"
)

func TestReadMapLayoutDecodesTheGridInMillimeters(t *testing.T) {
	cloud := startFakeCloud(t)
	grid := syntheticGrid(6, 4)
	cloud.layoutFile = syntheticLayoutFile(t, grid, 6, 4)

	layout, err := cloud.client().ReadMapLayout(context.Background(), testCredentials(), testDeviceID)
	if err != nil {
		t.Fatal(err)
	}

	if layout.MapID != 4242 || layout.Width != 6 || layout.Height != 4 {
		t.Fatalf("unexpected header: %+v", layout)
	}
	if layout.ResolutionMM != 50 || layout.OriginXMM != -1500 || layout.OriginYMM != -250 {
		t.Fatalf("unexpected geometry: %+v", layout)
	}
	if !bytes.Equal(layout.Cells, grid) {
		t.Fatalf("cells differ: %v", layout.Cells)
	}
	if layout.Dock == nil || layout.Dock.X != -100 || layout.Dock.Y != 40 {
		t.Fatalf("unexpected dock: %+v", layout.Dock)
	}
	if len(layout.ForbiddenZones) != 1 || len(layout.ForbiddenZones[0].Vertices) != 3 {
		t.Fatalf("only the forbid area must be kept: %+v", layout.ForbiddenZones)
	}
}

func TestReadMapLayoutSendsTheSignedHeaders(t *testing.T) {
	cloud := startFakeCloud(t)
	cloud.layoutFile = syntheticLayoutFile(t, syntheticGrid(6, 4), 6, 4)

	if _, err := cloud.client().ReadMapLayout(context.Background(), testCredentials(), testDeviceID); err != nil {
		t.Fatal(err)
	}

	tokenHeaders, mapHeaders := cloud.receivedHeaders[0], cloud.receivedHeaders[1]
	if tokenHeaders.Get("client_id") != testAccessID || tokenHeaders.Get("sign_method") != "HMAC-SHA256" {
		t.Fatalf("token headers: %v", tokenHeaders)
	}
	if tokenHeaders.Get("access_token") != "" || tokenHeaders.Get("sign") == "" || tokenHeaders.Get("t") == "" {
		t.Fatalf("token call must be signed without access_token: %v", tokenHeaders)
	}
	if mapHeaders.Get("access_token") != "granted-token" {
		t.Fatalf("business call must carry the token: %v", mapHeaders)
	}
	if cloud.receivedPaths[0] != "/v1.0/token?grant_type=1" {
		t.Fatalf("token path: %s", cloud.receivedPaths[0])
	}
	layoutHeaders := cloud.receivedHeaders[2]
	if layoutHeaders.Get("sign") != "" || layoutHeaders.Get("client_id") != "" {
		t.Fatalf("layout download must be a plain GET: %v", layoutHeaders)
	}
}

func TestTokenIsCachedUntilItExpires(t *testing.T) {
	cloud := startFakeCloud(t)
	cloud.layoutFile = syntheticLayoutFile(t, syntheticGrid(6, 4), 6, 4)
	instant := time.Date(2026, 11, 25, 10, 0, 0, 0, time.UTC)
	client := cloud.client()
	client.now = fixedClock(&instant)

	read := func() {
		if _, err := client.ReadMapLayout(context.Background(), testCredentials(), testDeviceID); err != nil {
			t.Fatal(err)
		}
	}
	read()
	instant = instant.Add(time.Hour)
	read()
	if cloud.tokenCalls != 1 {
		t.Fatalf("token must be reused, got %d token calls", cloud.tokenCalls)
	}

	instant = instant.Add(2 * time.Hour)
	read()
	if cloud.tokenCalls != 2 {
		t.Fatalf("expired token must be renewed, got %d token calls", cloud.tokenCalls)
	}
}

func TestServiceNotAuthorizedIsItsOwnFailure(t *testing.T) {
	cloud := startFakeCloud(t)
	cloud.realtimeMapReply = func(string) string {
		return `{"success":false,"code":28841101,"msg":"No permissions. This API is not subscribed."}`
	}

	_, err := cloud.client().ReadMapLayout(context.Background(), testCredentials(), testDeviceID)
	if !errors.Is(err, ErrServiceNotAuthorized) {
		t.Fatalf("got %v", err)
	}
}

func TestRejectedTokenIsCredentialsRejected(t *testing.T) {
	cloud := startFakeCloud(t)
	cloud.tokenReply = `{"success":false,"code":1004,"msg":"sign invalid"}`

	_, err := cloud.client().ReadMapLayout(context.Background(), testCredentials(), testDeviceID)
	if !errors.Is(err, ErrCredentialsRejected) {
		t.Fatalf("got %v", err)
	}
}

func TestRejectedBusinessCallDropsTheCachedToken(t *testing.T) {
	cloud := startFakeCloud(t)
	cloud.realtimeMapReply = func(string) string { return `{"success":false,"code":1010,"msg":"token invalid"}` }
	client := cloud.client()

	for attempt := 0; attempt < 2; attempt++ {
		if _, err := client.ReadMapLayout(context.Background(), testCredentials(), testDeviceID); !errors.Is(err, ErrCredentialsRejected) {
			t.Fatalf("got %v", err)
		}
	}
	if cloud.tokenCalls != 2 {
		t.Fatalf("a rejected token must not be reused, got %d token calls", cloud.tokenCalls)
	}
}

func TestInvalidRealtimeMapReplies(t *testing.T) {
	tests := map[string]string{
		"not json":             `nope`,
		"unknown cloud error":  `{"success":false,"code":999,"msg":"boom"}`,
		"result is not a list": `{"success":true,"result":{}}`,
		"no layout map":        `{"success":true,"result":[{"map_type":1,"map_url":"https://x"}]}`,
		"plain http map url":   `{"success":true,"result":[{"map_type":0,"map_url":"http://x/layout"}]}`,
	}
	for name, reply := range tests {
		t.Run(name, func(t *testing.T) {
			cloud := startFakeCloud(t)
			cloud.realtimeMapReply = func(string) string { return reply }

			_, err := cloud.client().ReadMapLayout(context.Background(), testCredentials(), testDeviceID)
			if !errors.Is(err, ErrInvalidResponse) {
				t.Fatalf("got %v", err)
			}
		})
	}
}

func TestInvalidTokenReplies(t *testing.T) {
	for name, reply := range map[string]string{
		"no token":     `{"success":true,"result":{"expire_time":7200}}`,
		"not json":     `nope`,
		"result wrong": `{"success":true,"result":[]}`,
	} {
		t.Run(name, func(t *testing.T) {
			cloud := startFakeCloud(t)
			cloud.tokenReply = reply

			_, err := cloud.client().ReadMapLayout(context.Background(), testCredentials(), testDeviceID)
			if !errors.Is(err, ErrInvalidResponse) {
				t.Fatalf("got %v", err)
			}
		})
	}
}

func TestInvalidLayoutFiles(t *testing.T) {
	grid := syntheticGrid(6, 4)
	tests := map[string]string{
		"not json":         `nope`,
		"wrong dimensions": syntheticLayoutFile(t, grid, 5, 4),
		"too many cells":   syntheticLayoutFile(t, grid, 4001, 1000),
		"zero size":        syntheticLayoutFile(t, grid, 0, 0),
		"not base64":       `{"data":{"width":6,"height":4,"map":"***"}}`,
	}
	for name, layoutFile := range tests {
		t.Run(name, func(t *testing.T) {
			cloud := startFakeCloud(t)
			cloud.layoutFile = layoutFile

			_, err := cloud.client().ReadMapLayout(context.Background(), testCredentials(), testDeviceID)
			if !errors.Is(err, ErrInvalidResponse) {
				t.Fatalf("got %v", err)
			}
		})
	}
}

func TestLayoutDownloadFailureStatusIsInvalidResponse(t *testing.T) {
	cloud := startFakeCloud(t)
	cloud.layoutStatus = 403

	_, err := cloud.client().ReadMapLayout(context.Background(), testCredentials(), testDeviceID)
	if !errors.Is(err, ErrInvalidResponse) {
		t.Fatalf("got %v", err)
	}
}

func TestUnreachableCloud(t *testing.T) {
	cloud := startFakeCloud(t)
	client := cloud.client()
	cloud.server.Close()

	_, err := client.ReadMapLayout(context.Background(), testCredentials(), testDeviceID)
	if !errors.Is(err, ErrUnreachable) {
		t.Fatalf("got %v", err)
	}
}

func TestUnknownRegionWithoutBaseURLOverride(t *testing.T) {
	client := NewClient(nil)

	credentials := testCredentials()
	credentials.Region = "mars"
	_, err := client.ReadMapLayout(context.Background(), credentials, testDeviceID)
	if !errors.Is(err, ErrUnknownRegion) {
		t.Fatalf("got %v", err)
	}
}
