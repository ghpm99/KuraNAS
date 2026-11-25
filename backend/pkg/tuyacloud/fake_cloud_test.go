package tuyacloud

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/pierrec/lz4/v4"
)

const (
	testAccessID     = "test-access-id"
	testAccessSecret = "test-access-secret"
	testDeviceID     = "device-1"
)

type fakeCloud struct {
	server           *httptest.Server
	mutex            sync.Mutex
	tokenCalls       int
	mapCalls         int
	tokenReply       string
	realtimeMapReply func(layoutURL string) string
	layoutFile       string
	layoutStatus     int
	receivedHeaders  []http.Header
	receivedPaths    []string
}

func startFakeCloud(t *testing.T) *fakeCloud {
	t.Helper()
	cloud := &fakeCloud{
		tokenReply:   `{"success":true,"result":{"access_token":"granted-token","expire_time":7200}}`,
		layoutStatus: http.StatusOK,
		realtimeMapReply: func(layoutURL string) string {
			return fmt.Sprintf(`{"success":true,"result":[{"map_type":1,"map_url":"%[1]s/other"},{"map_type":0,"map_url":"%[1]s/layout"}]}`, layoutURL)
		},
	}
	cloud.server = httptest.NewTLSServer(http.HandlerFunc(cloud.serve))
	t.Cleanup(cloud.server.Close)
	return cloud
}

func (cloud *fakeCloud) serve(writer http.ResponseWriter, request *http.Request) {
	cloud.mutex.Lock()
	defer cloud.mutex.Unlock()
	cloud.receivedHeaders = append(cloud.receivedHeaders, request.Header.Clone())
	cloud.receivedPaths = append(cloud.receivedPaths, request.URL.RequestURI())

	switch request.URL.Path {
	case "/v1.0/token":
		cloud.tokenCalls++
		fmt.Fprint(writer, cloud.tokenReply)
	case "/v1.0/users/sweepers/file/" + testDeviceID + "/realtime-map":
		cloud.mapCalls++
		fmt.Fprint(writer, cloud.realtimeMapReply(cloud.server.URL))
	case "/layout":
		writer.WriteHeader(cloud.layoutStatus)
		fmt.Fprint(writer, cloud.layoutFile)
	default:
		http.NotFound(writer, request)
	}
}

func (cloud *fakeCloud) client() *Client {
	return NewClientWithBaseURL(cloud.server.Client(), cloud.server.URL)
}

func testCredentials() Credentials {
	return Credentials{AccessID: testAccessID, AccessSecret: testAccessSecret, Region: "eu"}
}

func syntheticLayoutFile(t *testing.T, cells []byte, width int, height int) string {
	t.Helper()
	compressed := make([]byte, lz4.CompressBlockBound(len(cells)))
	compressedSize, err := lz4.CompressBlock(cells, compressed, nil)
	if err != nil || compressedSize == 0 {
		t.Fatalf("cannot compress synthetic grid: size=%d err=%v", compressedSize, err)
	}
	layout := map[string]any{
		"infoType": 20002,
		"data": map[string]any{
			"width": width, "height": height, "resolution": 0.05,
			"x_min": -1.5, "y_min": -0.25,
			"map":             base64.StdEncoding.EncodeToString(compressed[:compressedSize]),
			"mapId":           4242,
			"chargeHandlePos": []int{-100, 40},
			"area": []map[string]any{
				{"active": "forbid", "vertexs": [][]int{{0, 0}, {100, 0}, {100, 100}}},
				{"active": "ordinary", "vertexs": [][]int{{5, 5}}},
			},
		},
	}
	encoded, err := json.Marshal(layout)
	if err != nil {
		t.Fatal(err)
	}
	return string(encoded)
}

func syntheticGrid(width int, height int) []byte {
	cells := make([]byte, width*height)
	for index := range cells {
		cells[index] = 127
	}
	for column := 0; column < width; column++ {
		cells[column] = 0
	}
	for column := 1; column < width-1; column++ {
		cells[width+column] = 1
	}
	cells[2*width+2] = 255
	return cells
}

func fixedClock(instant *time.Time) func() time.Time {
	return func() time.Time { return *instant }
}
