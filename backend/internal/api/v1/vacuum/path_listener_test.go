package vacuum

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"nas-go/api/pkg/tuya"
)

type fakePushClient struct {
	mutex    sync.Mutex
	devices  []tuya.Device
	attempts chan struct{}
	behavior func(ctx context.Context, attempt int, onDataPoints func(map[string]any)) error
}

func newFakePushClient(behavior func(ctx context.Context, attempt int, onDataPoints func(map[string]any)) error) *fakePushClient {
	return &fakePushClient{attempts: make(chan struct{}, 32), behavior: behavior}
}

func (c *fakePushClient) Listen(ctx context.Context, device tuya.Device, onDataPoints func(map[string]any)) error {
	c.mutex.Lock()
	c.devices = append(c.devices, device)
	attempt := len(c.devices)
	c.mutex.Unlock()
	c.attempts <- struct{}{}
	return c.behavior(ctx, attempt, onDataPoints)
}

func (c *fakePushClient) attemptCount() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return len(c.devices)
}

func (c *fakePushClient) addressOfAttempt(attempt int) string {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.devices[attempt-1].Address
}

func (c *fakePushClient) waitForAttempt(t *testing.T) {
	t.Helper()
	select {
	case <-c.attempts:
	case <-time.After(3 * time.Second):
		t.Fatal("the listener did not connect")
	}
}

func fastListener(fixture *serviceFixture, client PushListenerClient) *PathListener {
	listener := NewPathListener(fixture.service, client)
	listener.initialBackoff = 5 * time.Millisecond
	listener.maxBackoff = 20 * time.Millisecond
	return listener
}

func waitUntil(t *testing.T, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for !condition() {
		if time.Now().After(deadline) {
			t.Fatal("condition not reached")
		}
		time.Sleep(2 * time.Millisecond)
	}
}

func TestPathListenerFeedsPushesIntoThePathAndStopsCleanly(t *testing.T) {
	fixture := newServiceFixture(t)
	client := newFakePushClient(func(ctx context.Context, _ int, onDataPoints func(map[string]any)) error {
		onDataPoints(map[string]any{dpPath: encodedPathReport(21011, 3, 0, [][2]int{{1, 2}})})
		<-ctx.Done()
		return ctx.Err()
	})
	listener := fastListener(fixture, client)

	listener.Start()
	listener.Start()
	waitUntil(t, func() bool { return len(fixture.service.GetPath().Points) == 1 })
	listener.Stop()
	listener.Stop()

	if client.attemptCount() != 1 || client.addressOfAttempt(1) != testAddress {
		t.Fatalf("unexpected listen target: %+v", client.devices)
	}
}

func TestPathListenerReconnectsWithBackoffAfterTheConnectionDrops(t *testing.T) {
	fixture := newServiceFixture(t)
	client := newFakePushClient(func(ctx context.Context, attempt int, _ func(map[string]any)) error {
		if attempt < 3 {
			return tuya.ErrUnreachable
		}
		<-ctx.Done()
		return ctx.Err()
	})
	listener := fastListener(fixture, client)

	listener.Start()
	defer listener.Stop()
	for attempt := 0; attempt < 3; attempt++ {
		client.waitForAttempt(t)
	}
}

func TestPathListenerRereadsTheStoredAddressOnEveryReconnection(t *testing.T) {
	fixture := newServiceFixture(t)
	client := newFakePushClient(func(ctx context.Context, attempt int, _ func(map[string]any)) error {
		if attempt == 1 {
			return tuya.ErrUnreachable
		}
		<-ctx.Done()
		return ctx.Err()
	})
	listener := fastListener(fixture, client)
	listener.initialBackoff = 200 * time.Millisecond

	listener.Start()
	defer listener.Stop()
	client.waitForAttempt(t)
	if err := fixture.service.storeAddress("10.5.5.5"); err != nil {
		t.Fatal(err)
	}
	client.waitForAttempt(t)

	if address := client.addressOfAttempt(2); address != "10.5.5.5" {
		t.Fatalf("second attempt must use the new address, got %q", address)
	}
}

func TestPathListenerWaitsWhileTheRobotIsNotConfiguredAndRestartConnects(t *testing.T) {
	fixture := newServiceFixture(t)
	fixture.repository.isFound = false
	client := newFakePushClient(func(ctx context.Context, _ int, _ func(map[string]any)) error {
		<-ctx.Done()
		return ctx.Err()
	})
	listener := fastListener(fixture, client)
	listener.initialBackoff = time.Hour
	listener.maxBackoff = time.Hour

	listener.Start()
	defer listener.Stop()
	time.Sleep(30 * time.Millisecond)
	if len(client.attempts) != 0 {
		t.Fatal("an unconfigured robot must not be contacted")
	}

	document, _ := encodeSettings(settingsState{Address: testAddress, DeviceID: testDeviceID, LocalKey: testLocalKey})
	fixture.repository.storeSettingsConcurrently(document)
	listener.Restart()
	client.waitForAttempt(t)
}

func TestPathListenerDoesNotConnectWithoutAnAddress(t *testing.T) {
	fixture := newServiceFixture(t)
	document, _ := encodeSettings(settingsState{DeviceID: testDeviceID, LocalKey: testLocalKey})
	fixture.repository.document = document
	client := newFakePushClient(func(ctx context.Context, _ int, _ func(map[string]any)) error { return nil })
	listener := fastListener(fixture, client)

	listener.Start()
	time.Sleep(30 * time.Millisecond)
	listener.Stop()

	if client.attemptCount() != 0 {
		t.Fatal("without an address there is nothing to connect to")
	}
}

func TestPathListenerSurvivesSettingsReadFailures(t *testing.T) {
	fixture := newServiceFixture(t)
	fixture.repository.getErr = errors.New("db down")
	client := newFakePushClient(func(ctx context.Context, _ int, _ func(map[string]any)) error { return nil })
	listener := fastListener(fixture, client)

	listener.Start()
	time.Sleep(30 * time.Millisecond)
	listener.Stop()

	if client.attemptCount() != 0 {
		t.Fatal("unreadable settings must not produce a connection")
	}
}

func TestRestartDropsTheCurrentConnectionAndReconnectsRightAway(t *testing.T) {
	fixture := newServiceFixture(t)
	client := newFakePushClient(func(ctx context.Context, _ int, _ func(map[string]any)) error {
		<-ctx.Done()
		return ctx.Err()
	})
	listener := fastListener(fixture, client)
	listener.initialBackoff = time.Hour
	listener.maxBackoff = time.Hour

	listener.Start()
	defer listener.Stop()
	client.waitForAttempt(t)
	listener.Restart()
	client.waitForAttempt(t)
}

func TestLongLivedConnectionsResetTheBackoff(t *testing.T) {
	fixture := newServiceFixture(t)
	client := newFakePushClient(func(ctx context.Context, attempt int, _ func(map[string]any)) error {
		if attempt == 1 {
			time.Sleep(25 * time.Millisecond)
			return tuya.ErrUnreachable
		}
		<-ctx.Done()
		return ctx.Err()
	})
	listener := fastListener(fixture, client)

	listener.Start()
	defer listener.Stop()
	client.waitForAttempt(t)
	client.waitForAttempt(t)
}
