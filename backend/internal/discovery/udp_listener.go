package discovery

import (
	"encoding/json"
	"log"
	"net"
	"time"
)

const (
	DefaultUDPPort    = 19520
	discoveryMessage  = "KURANAS_DISCOVER"
	readDeadlineDelta = 1 * time.Second
)

type udpResponse struct {
	Service string `json:"service"`
	Port    int    `json:"port"`
	API     string `json:"api"`
}

type UDPListener struct {
	port     int
	httpPort int
	conn     *net.UDPConn
	done     chan struct{}
}

func NewUDPListener(port int, httpPort int) *UDPListener {
	return &UDPListener{
		port:     port,
		httpPort: httpPort,
		done:     make(chan struct{}),
	}
}

func (l *UDPListener) Start() error {
	addr := &net.UDPAddr{
		Port: l.port,
		IP:   net.IPv4zero,
	}

	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		return err
	}

	l.conn = conn
	go l.listen()

	log.Printf("[DISCOVERY] UDP listener started on port %d", l.port)
	return nil
}

func (l *UDPListener) Stop() {
	close(l.done)
	if l.conn != nil {
		l.conn.Close()
	}
	log.Println("[DISCOVERY] UDP listener stopped")
}

func (l *UDPListener) listen() {
	buf := make([]byte, 1024)

	for {
		select {
		case <-l.done:
			return
		default:
		}

		l.conn.SetReadDeadline(time.Now().Add(readDeadlineDelta))

		n, remoteAddr, err := l.conn.ReadFromUDP(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			select {
			case <-l.done:
				return
			default:
				log.Printf("[DISCOVERY] UDP read error: %v", err)
				continue
			}
		}

		message := string(buf[:n])
		if message != discoveryMessage {
			continue
		}

		log.Printf("[DISCOVERY] Discovery request from %s", remoteAddr.String())

		response := udpResponse{
			Service: "kuranas",
			Port:    l.httpPort,
			API:     "/api/v1",
		}

		data, err := json.Marshal(response)
		if err != nil {
			log.Printf("[DISCOVERY] Failed to marshal response: %v", err)
			continue
		}

		if _, err := l.conn.WriteToUDP(data, remoteAddr); err != nil {
			log.Printf("[DISCOVERY] Failed to send response: %v", err)
		}
	}
}
