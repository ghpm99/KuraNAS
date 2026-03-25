package discovery

import (
	"encoding/json"
	"net"
	"testing"
	"time"
)

func TestUDPListenerRespondsToDiscovery(t *testing.T) {
	listener := NewUDPListener(0, 8000)

	if err := listener.Start(); err != nil {
		t.Fatalf("failed to start listener: %v", err)
	}
	defer listener.Stop()

	listenerPort := listener.conn.LocalAddr().(*net.UDPAddr).Port

	conn, err := net.DialUDP("udp4", nil, &net.UDPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: listenerPort,
	})
	if err != nil {
		t.Fatalf("failed to dial UDP: %v", err)
	}
	defer conn.Close()

	_, err = conn.Write([]byte(discoveryMessage))
	if err != nil {
		t.Fatalf("failed to send discovery message: %v", err)
	}

	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}

	var response udpResponse
	if err := json.Unmarshal(buf[:n], &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if response.Service != "kuranas" {
		t.Errorf("expected service 'kuranas', got '%s'", response.Service)
	}

	if response.Port != 8000 {
		t.Errorf("expected port 8000, got %d", response.Port)
	}

	if response.API != "/api/v1" {
		t.Errorf("expected api '/api/v1', got '%s'", response.API)
	}
}

func TestUDPListenerIgnoresInvalidMessage(t *testing.T) {
	listener := NewUDPListener(0, 8000)

	if err := listener.Start(); err != nil {
		t.Fatalf("failed to start listener: %v", err)
	}
	defer listener.Stop()

	listenerPort := listener.conn.LocalAddr().(*net.UDPAddr).Port

	conn, err := net.DialUDP("udp4", nil, &net.UDPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: listenerPort,
	})
	if err != nil {
		t.Fatalf("failed to dial UDP: %v", err)
	}
	defer conn.Close()

	_, err = conn.Write([]byte("INVALID_MESSAGE"))
	if err != nil {
		t.Fatalf("failed to send message: %v", err)
	}

	conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	buf := make([]byte, 1024)
	_, err = conn.Read(buf)
	if err == nil {
		t.Error("expected no response for invalid message, but got one")
	}
}

func TestUDPListenerStopsCleanly(t *testing.T) {
	listener := NewUDPListener(0, 8000)

	if err := listener.Start(); err != nil {
		t.Fatalf("failed to start listener: %v", err)
	}

	listener.Stop()

	time.Sleep(100 * time.Millisecond)
}
