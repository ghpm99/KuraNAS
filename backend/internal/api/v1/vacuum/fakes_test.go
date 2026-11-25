package vacuum

import (
	"context"
	"sync"
	"testing"
	"time"

	"nas-go/api/pkg/tuya"
	"nas-go/api/pkg/tuyacloud"
)

const (
	testDeviceID = "device-123"
	testLocalKey = "abcdefghijklmnop"
	testAddress  = "robot.local"
)

type fakeRepository struct {
	mutex       sync.Mutex
	document    string
	isFound     bool
	getErr      error
	upsertErr   error
	upsertCalls int

	cloudDocument string
	isCloudFound  bool
	roomDocument  string
	isRoomFound   bool
	roomUpsertErr error
}

func (r *fakeRepository) GetCloudSettingsDocument() (string, bool, error) {
	return r.cloudDocument, r.isCloudFound, r.getErr
}

func (r *fakeRepository) UpsertCloudSettingsDocument(document string) error {
	if r.upsertErr != nil {
		return r.upsertErr
	}
	r.cloudDocument = document
	r.isCloudFound = true
	return nil
}

func (r *fakeRepository) GetRoomNamesDocument() (string, bool, error) {
	return r.roomDocument, r.isRoomFound, r.getErr
}

func (r *fakeRepository) UpsertRoomNamesDocument(document string) error {
	if r.roomUpsertErr != nil {
		return r.roomUpsertErr
	}
	r.roomDocument = document
	r.isRoomFound = true
	return nil
}

type fakeCloudClient struct {
	mutex       sync.Mutex
	layout      tuyacloud.MapLayout
	failure     error
	calls       int
	credentials []tuyacloud.Credentials
	deviceIDs   []string
}

func (c *fakeCloudClient) ReadMapLayout(_ context.Context, credentials tuyacloud.Credentials, deviceID string) (tuyacloud.MapLayout, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.calls++
	c.credentials = append(c.credentials, credentials)
	c.deviceIDs = append(c.deviceIDs, deviceID)
	return c.layout, c.failure
}

const (
	testCloudAccessID     = "cloud-access-id"
	testCloudAccessSecret = "0123456789abcdef0123456789abcdef"
	testMapID             = int64(777)
)

func testMapLayout() tuyacloud.MapLayout {
	cells := []byte{
		0, 0, 0, 0, 0, 0,
		0, 1, 1, 2, 2, 0,
		0, 1, 1, 2, 255, 0,
		0, 127, 127, 127, 127, 0,
	}
	return tuyacloud.MapLayout{
		MapID: testMapID, Width: 6, Height: 4, ResolutionMM: 50, OriginXMM: -100, OriginYMM: 200, Cells: cells,
		Dock:           &tuyacloud.Point{X: -50, Y: 250},
		ForbiddenZones: []tuyacloud.ForbiddenZone{{Vertices: []tuyacloud.Point{{X: 1, Y: 2}, {X: 3, Y: 4}}}},
	}
}

func (r *fakeRepository) GetSettingsDocument() (string, bool, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.document, r.isFound, r.getErr
}

func (r *fakeRepository) storeSettingsConcurrently(document string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.document = document
	r.isFound = true
}

func (r *fakeRepository) UpsertSettingsDocument(document string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.upsertCalls++
	if r.upsertErr != nil {
		return r.upsertErr
	}
	r.document = document
	r.isFound = true
	return nil
}

type writeCall struct {
	device     tuya.Device
	dataPoints map[string]any
}

type fakeDeviceClient struct {
	mutex          sync.Mutex
	readFn         func(device tuya.Device) (map[string]any, error)
	writeFn        func(device tuya.Device, dataPoints map[string]any) error
	findFn         func(device tuya.Device, candidates []string) (string, error)
	readDevices    []tuya.Device
	writes         []writeCall
	findCandidates [][]string
	findCalls      int
}

func (c *fakeDeviceClient) ReadDataPoints(_ context.Context, device tuya.Device) (map[string]any, error) {
	c.mutex.Lock()
	c.readDevices = append(c.readDevices, device)
	c.mutex.Unlock()
	if c.readFn == nil {
		return map[string]any{}, nil
	}
	return c.readFn(device)
}

func (c *fakeDeviceClient) WriteDataPoints(_ context.Context, device tuya.Device, dataPoints map[string]any) error {
	c.mutex.Lock()
	c.writes = append(c.writes, writeCall{device: device, dataPoints: dataPoints})
	c.mutex.Unlock()
	if c.writeFn == nil {
		return nil
	}
	return c.writeFn(device, dataPoints)
}

func (c *fakeDeviceClient) FindDevice(_ context.Context, device tuya.Device, candidates []string) (string, error) {
	c.mutex.Lock()
	c.findCalls++
	c.findCandidates = append(c.findCandidates, candidates)
	c.mutex.Unlock()
	if c.findFn == nil {
		return "", tuya.ErrDeviceNotFound
	}
	return c.findFn(device, candidates)
}

type fakeClock struct {
	instant time.Time
}

func (c *fakeClock) now() time.Time { return c.instant }

func (c *fakeClock) advance(duration time.Duration) { c.instant = c.instant.Add(duration) }

type serviceFixture struct {
	service    *Service
	repository *fakeRepository
	client     *fakeDeviceClient
	cloud      *fakeCloudClient
	clock      *fakeClock
}

func newServiceFixture(t *testing.T) *serviceFixture {
	t.Helper()
	repository := &fakeRepository{}
	document, err := encodeSettings(settingsState{Address: testAddress, DeviceID: testDeviceID, LocalKey: testLocalKey})
	if err != nil {
		t.Fatal(err)
	}
	repository.document = document
	repository.isFound = true
	cloudDocument, err := encodeCloudSettings(cloudSettingsState{AccessID: testCloudAccessID, AccessSecret: testCloudAccessSecret, Region: "eu"})
	if err != nil {
		t.Fatal(err)
	}
	repository.cloudDocument = cloudDocument
	repository.isCloudFound = true

	client := &fakeDeviceClient{}
	clock := &fakeClock{instant: time.Date(2026, 11, 16, 10, 0, 0, 0, time.UTC)}
	cloud := &fakeCloudClient{layout: testMapLayout()}
	service := NewService(repository, client, func() ([]string, error) { return []string{"10.1.1.2", "10.1.1.3"}, nil }, cloud)
	service.now = clock.now
	return &serviceFixture{service: service, repository: repository, client: client, cloud: cloud, clock: clock}
}

func (f *serviceFixture) storedSettings(t *testing.T) settingsState {
	t.Helper()
	settings, err := decodeSettings(f.repository.document)
	if err != nil {
		t.Fatal(err)
	}
	return settings
}
