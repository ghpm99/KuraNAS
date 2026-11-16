package vacuum

import (
	"context"
	"sync"
	"testing"
	"time"

	"nas-go/api/pkg/tuya"
)

const (
	testDeviceID = "device-123"
	testLocalKey = "abcdefghijklmnop"
	testAddress  = "robot.local"
)

type fakeRepository struct {
	document    string
	isFound     bool
	getErr      error
	upsertErr   error
	upsertCalls int
}

func (r *fakeRepository) GetSettingsDocument() (string, bool, error) {
	return r.document, r.isFound, r.getErr
}

func (r *fakeRepository) UpsertSettingsDocument(document string) error {
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

	client := &fakeDeviceClient{}
	clock := &fakeClock{instant: time.Date(2026, 11, 16, 10, 0, 0, 0, time.UTC)}
	service := NewService(repository, client, func() ([]string, error) { return []string{"10.1.1.2", "10.1.1.3"}, nil })
	service.now = clock.now
	return &serviceFixture{service: service, repository: repository, client: client, clock: clock}
}

func (f *serviceFixture) storedSettings(t *testing.T) settingsState {
	t.Helper()
	settings, err := decodeSettings(f.repository.document)
	if err != nil {
		t.Fatal(err)
	}
	return settings
}
