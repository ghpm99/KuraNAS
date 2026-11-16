package tuya

import (
	"context"
	"errors"
	"net"
	"reflect"
	"strings"
	"testing"
	"time"
)

const testKey = "0123456789abcdef"

func newTestDevice(address string) Device {
	return Device{Address: address, DeviceID: "bf1234", LocalKey: testKey}
}

func newVacuumStatus() map[string]any {
	return map[string]any{"101": false, "105": "fullcharge", "106": float64(100)}
}

func TestReadDataPointsReturnsDeviceStatus(t *testing.T) {
	fake := &fakeDevice{deviceKey: []byte(testKey), dataPoints: newVacuumStatus()}
	address := startFakeDevice(t, fake)

	dataPoints, err := NewClient(time.Second).ReadDataPoints(context.Background(), newTestDevice(address))

	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(dataPoints, newVacuumStatus()) {
		t.Fatalf("unexpected data points: %v", dataPoints)
	}
	request := <-fake.requests
	if request.command != 0x0a || request.decryptedFields["devId"] != "bf1234" || request.hasVersionTag {
		t.Fatalf("unexpected request: %+v", request)
	}
}

func TestWriteDataPointsSendsExactlyTheGivenDataPoints(t *testing.T) {
	fake := &fakeDevice{deviceKey: []byte(testKey)}
	address := startFakeDevice(t, fake)

	err := NewClient(time.Second).WriteDataPoints(context.Background(), newTestDevice(address), map[string]any{"115": "gettime"})

	if err != nil {
		t.Fatal(err)
	}
	request := <-fake.requests
	if request.command != 0x07 || !request.hasVersionTag {
		t.Fatalf("unexpected request: %+v", request)
	}
	if !reflect.DeepEqual(request.decryptedFields["dps"], map[string]any{"115": "gettime"}) {
		t.Fatalf("unexpected dps: %v", request.decryptedFields["dps"])
	}
}

func TestReadWithWrongKeyIsRejected(t *testing.T) {
	fake := &fakeDevice{deviceKey: []byte("fedcba9876543210"), dataPoints: newVacuumStatus()}
	address := startFakeDevice(t, fake)

	_, err := NewClient(time.Second).ReadDataPoints(context.Background(), newTestDevice(address))

	if !errors.Is(err, ErrKeyRejected) {
		t.Fatalf("expected ErrKeyRejected, got %v", err)
	}
}

func TestWriteWithWrongKeyIsRejected(t *testing.T) {
	fake := &fakeDevice{deviceKey: []byte("fedcba9876543210")}
	address := startFakeDevice(t, fake)

	err := NewClient(time.Second).WriteDataPoints(context.Background(), newTestDevice(address), map[string]any{"115": "gettime"})

	if !errors.Is(err, ErrKeyRejected) {
		t.Fatalf("expected ErrKeyRejected, got %v", err)
	}
}

func TestClosedPortIsUnreachable(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	listener.Close()

	_, err = NewClient(time.Second).ReadDataPoints(context.Background(), newTestDevice(address))

	if !errors.Is(err, ErrUnreachable) {
		t.Fatalf("expected ErrUnreachable, got %v", err)
	}
}

func TestSilentDeviceIsUnreachableWithinTimeout(t *testing.T) {
	address := startFakeDevice(t, &fakeDevice{deviceKey: []byte(testKey), isSilent: true})
	startedAt := time.Now()

	_, err := NewClient(200*time.Millisecond).ReadDataPoints(context.Background(), newTestDevice(address))

	if !errors.Is(err, ErrUnreachable) {
		t.Fatalf("expected ErrUnreachable, got %v", err)
	}
	if elapsed := time.Since(startedAt); elapsed > time.Second {
		t.Fatalf("took too long: %v", elapsed)
	}
}

func TestSilentDeviceOnWriteIsUnreachable(t *testing.T) {
	address := startFakeDevice(t, &fakeDevice{deviceKey: []byte(testKey), isSilent: true})

	err := NewClient(200*time.Millisecond).WriteDataPoints(context.Background(), newTestDevice(address), map[string]any{"1": true})

	if !errors.Is(err, ErrUnreachable) {
		t.Fatalf("expected ErrUnreachable, got %v", err)
	}
}

func TestReadSkipsEmptyFrameBeforeRealReply(t *testing.T) {
	fake := &fakeDevice{deviceKey: []byte(testKey), dataPoints: newVacuumStatus(), shouldSendEmptyFirst: true}
	address := startFakeDevice(t, fake)

	dataPoints, err := NewClient(time.Second).ReadDataPoints(context.Background(), newTestDevice(address))

	if err != nil || !reflect.DeepEqual(dataPoints, newVacuumStatus()) {
		t.Fatalf("unexpected outcome: %v %v", dataPoints, err)
	}
}

func TestReadAcceptsReplyWithVersionHeader(t *testing.T) {
	fake := &fakeDevice{deviceKey: []byte(testKey), dataPoints: newVacuumStatus(), shouldPrefixVersion: true}
	address := startFakeDevice(t, fake)

	dataPoints, err := NewClient(time.Second).ReadDataPoints(context.Background(), newTestDevice(address))

	if err != nil || !reflect.DeepEqual(dataPoints, newVacuumStatus()) {
		t.Fatalf("unexpected outcome: %v %v", dataPoints, err)
	}
}

func TestReadFailsWhenReplyHasNoDataPoints(t *testing.T) {
	fake := &fakeDevice{deviceKey: []byte(testKey)}
	address := startFakeDevice(t, fake)

	_, err := NewClient(time.Second).ReadDataPoints(context.Background(), newTestDevice(address))

	if !errors.Is(err, ErrKeyRejected) {
		t.Fatalf("expected ErrKeyRejected, got %v", err)
	}
}

func TestInvalidDeviceNeverOpensConnection(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	connectionOpened := make(chan struct{}, 1)
	go func() {
		if connection, err := listener.Accept(); err == nil {
			connection.Close()
			connectionOpened <- struct{}{}
		}
	}()
	address := listener.Addr().String()
	invalidDevices := []Device{
		{Address: address, DeviceID: "bf1234", LocalKey: "short"},
		{Address: address, DeviceID: "", LocalKey: testKey},
		{Address: "", DeviceID: "bf1234", LocalKey: testKey},
	}

	for _, device := range invalidDevices {
		client := NewClient(time.Second)
		if _, err := client.ReadDataPoints(context.Background(), device); !errors.Is(err, ErrInvalidDevice) {
			t.Fatalf("read: expected ErrInvalidDevice, got %v", err)
		}
		if err := client.WriteDataPoints(context.Background(), device, nil); !errors.Is(err, ErrInvalidDevice) {
			t.Fatalf("write: expected ErrInvalidDevice, got %v", err)
		}
	}
	select {
	case <-connectionOpened:
		t.Fatal("a connection was opened for an invalid device")
	case <-time.After(100 * time.Millisecond):
	}
}

func TestNormalizeAddressDefaultsToPort6668(t *testing.T) {
	cases := map[string]string{
		"192.168.0.10":      "192.168.0.10:6668",
		"192.168.0.10:7000": "192.168.0.10:7000",
		"host.local":        "host.local:6668",
		"::1":               "[::1]:6668",
	}
	for address, expected := range cases {
		if normalized := normalizeAddress(address); normalized != expected {
			t.Errorf("normalizeAddress(%q) = %q, want %q", address, normalized, expected)
		}
	}
}

func TestCancelledContextInterruptsCall(t *testing.T) {
	address := startFakeDevice(t, &fakeDevice{deviceKey: []byte(testKey), isSilent: true})
	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(100*time.Millisecond, cancel)
	startedAt := time.Now()

	_, err := NewClient(10*time.Second).ReadDataPoints(ctx, newTestDevice(address))

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if elapsed := time.Since(startedAt); elapsed > time.Second {
		t.Fatalf("cancellation took too long: %v", elapsed)
	}
}

func TestKeyNeverAppearsInErrors(t *testing.T) {
	_, err := NewClient(time.Second).ReadDataPoints(context.Background(), Device{Address: "x", DeviceID: "y", LocalKey: "secret"})

	if err == nil || strings.Contains(err.Error(), "secret") {
		t.Fatalf("unexpected error: %v", err)
	}
}
