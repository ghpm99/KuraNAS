package tuya

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"sync"
	"testing"
	"time"
)

const listenerTestKey = "0123456789abcdef"

type fakePushDevice struct {
	listener   net.Listener
	heartbeats chan struct{}
	connection chan net.Conn
}

func startFakePushDevice(t *testing.T) *fakePushDevice {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	device := &fakePushDevice{listener: listener, heartbeats: make(chan struct{}, 16), connection: make(chan net.Conn, 1)}
	t.Cleanup(func() { listener.Close() })
	go func() {
		connection, err := listener.Accept()
		if err != nil {
			return
		}
		device.connection <- connection
		for {
			command, _, err := readRawRequest(connection)
			if err != nil {
				return
			}
			if command == 0x09 {
				device.heartbeats <- struct{}{}
			}
		}
	}()
	return device
}

func (device *fakePushDevice) push(t *testing.T, command uint32, dataPoints map[string]any, shouldPrefixVersion bool) {
	t.Helper()
	plain, err := json.Marshal(map[string]any{"dps": dataPoints})
	if err != nil {
		t.Fatal(err)
	}
	payload := aesEncrypt([]byte(listenerTestKey), plain)
	if shouldPrefixVersion {
		payload = append(append([]byte("3.3"), make([]byte, 12)...), payload...)
	}
	device.send(t, rawFrame(command, payload))
}

func (device *fakePushDevice) send(t *testing.T, bytes []byte) {
	t.Helper()
	connection := <-device.connection
	device.connection <- connection
	if _, err := connection.Write(bytes); err != nil {
		t.Fatal(err)
	}
}

func listenerTestDevice(address string) Device {
	return Device{Address: address, DeviceID: "device-1", LocalKey: listenerTestKey}
}

type pushCollector struct {
	mutex    sync.Mutex
	received []map[string]any
	arrived  chan struct{}
}

func newPushCollector() *pushCollector {
	return &pushCollector{arrived: make(chan struct{}, 16)}
}

func (collector *pushCollector) collect(dataPoints map[string]any) {
	collector.mutex.Lock()
	collector.received = append(collector.received, dataPoints)
	collector.mutex.Unlock()
	collector.arrived <- struct{}{}
}

func (collector *pushCollector) waitForPush(t *testing.T) {
	t.Helper()
	select {
	case <-collector.arrived:
	case <-time.After(3 * time.Second):
		t.Fatal("no push delivered")
	}
}

func startListening(client *Client, device Device, collector *pushCollector) (cancel context.CancelFunc, finished chan error) {
	ctx, cancel := context.WithCancel(context.Background())
	finished = make(chan error, 1)
	go func() { finished <- client.Listen(ctx, device, collector.collect) }()
	return cancel, finished
}

func TestListenDeliversStatusPushesAndIgnoresOtherFrames(t *testing.T) {
	fake := startFakePushDevice(t)
	collector := newPushCollector()
	cancel, finished := startListening(NewClient(time.Second), listenerTestDevice(fake.listener.Addr().String()), collector)
	defer cancel()

	fake.send(t, rawFrame(0x09, nil))
	fake.send(t, rawFrame(0x08, nil))
	fake.push(t, 0x0a, map[string]any{"1": "ignored command"}, false)
	fake.push(t, 0x08, map[string]any{"123": "payload"}, true)
	collector.waitForPush(t)
	fake.push(t, 0x08, map[string]any{"106": 80}, false)
	collector.waitForPush(t)

	if len(collector.received) != 2 || collector.received[0]["123"] != "payload" || collector.received[1]["106"] != float64(80) {
		t.Fatalf("unexpected pushes: %v", collector.received)
	}
	cancel()
	if err := <-finished; !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
}

func TestListenSendsHeartbeats(t *testing.T) {
	fake := startFakePushDevice(t)
	client := NewClient(time.Second)
	client.heartbeatInterval = 20 * time.Millisecond
	cancel, _ := startListening(client, listenerTestDevice(fake.listener.Addr().String()), newPushCollector())
	defer cancel()

	for beat := 0; beat < 2; beat++ {
		select {
		case <-fake.heartbeats:
		case <-time.After(3 * time.Second):
			t.Fatal("no heartbeat received")
		}
	}
}

func TestListenReturnsUnreachableWhenTheConnectionDrops(t *testing.T) {
	fake := startFakePushDevice(t)
	cancel, finished := startListening(NewClient(time.Second), listenerTestDevice(fake.listener.Addr().String()), newPushCollector())
	defer cancel()

	connection := <-fake.connection
	connection.Close()

	if err := <-finished; !errors.Is(err, ErrUnreachable) {
		t.Fatalf("got %v", err)
	}
}

func TestListenReturnsUnreachableWhenTheDeviceStaysSilent(t *testing.T) {
	fake := startFakePushDevice(t)
	client := NewClient(time.Second)
	client.heartbeatInterval = 20 * time.Millisecond
	cancel, finished := startListening(client, listenerTestDevice(fake.listener.Addr().String()), newPushCollector())
	defer cancel()

	if err := <-finished; !errors.Is(err, ErrUnreachable) {
		t.Fatalf("got %v", err)
	}
}

func TestListenReturnsKeyRejectedForUndecryptablePush(t *testing.T) {
	fake := startFakePushDevice(t)
	cancel, finished := startListening(NewClient(time.Second), listenerTestDevice(fake.listener.Addr().String()), newPushCollector())
	defer cancel()

	fake.send(t, rawFrame(0x08, []byte("not a multiple of sixteen")))

	if err := <-finished; !errors.Is(err, ErrKeyRejected) {
		t.Fatalf("got %v", err)
	}
}

func TestListenSkipsPushesWithoutDataPoints(t *testing.T) {
	fake := startFakePushDevice(t)
	collector := newPushCollector()
	cancel, _ := startListening(NewClient(time.Second), listenerTestDevice(fake.listener.Addr().String()), collector)
	defer cancel()

	fake.send(t, rawFrame(0x08, aesEncrypt([]byte(listenerTestKey), []byte(`{"other":1}`))))
	fake.push(t, 0x08, map[string]any{"106": 1}, false)
	collector.waitForPush(t)

	if len(collector.received) != 1 {
		t.Fatalf("unexpected pushes: %v", collector.received)
	}
}

func TestListenReturnsInvalidFrameFailure(t *testing.T) {
	fake := startFakePushDevice(t)
	cancel, finished := startListening(NewClient(time.Second), listenerTestDevice(fake.listener.Addr().String()), newPushCollector())
	defer cancel()

	fake.send(t, make([]byte, 32))

	if err := <-finished; err == nil || errors.Is(err, ErrUnreachable) {
		t.Fatalf("got %v", err)
	}
}

func TestListenRejectsInvalidDeviceAndUnreachableAddress(t *testing.T) {
	client := NewClient(200 * time.Millisecond)
	if err := client.Listen(context.Background(), Device{}, nil); !errors.Is(err, ErrInvalidDevice) {
		t.Fatalf("got %v", err)
	}

	closed, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := closed.Addr().String()
	closed.Close()
	if err := client.Listen(context.Background(), listenerTestDevice(address), nil); !errors.Is(err, ErrUnreachable) {
		t.Fatalf("got %v", err)
	}
}
