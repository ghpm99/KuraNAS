package discovery

import (
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/mdns"
)

const DefaultServiceName = "_kuranas._tcp"

type MdnsRegistrar struct {
	serviceName string
	port        int
	server      *mdns.Server
}

func NewMdnsRegistrar(serviceName string, port int) *MdnsRegistrar {
	return &MdnsRegistrar{
		serviceName: serviceName,
		port:        port,
	}
}

func (r *MdnsRegistrar) Start() error {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "kuranas"
	}

	info := []string{fmt.Sprintf("api=/api/v1")}

	service, err := mdns.NewMDNSService(
		hostname,
		r.serviceName,
		"",
		"",
		r.port,
		nil,
		info,
	)
	if err != nil {
		return fmt.Errorf("failed to create mDNS service: %w", err)
	}

	server, err := mdns.NewServer(&mdns.Config{Zone: service})
	if err != nil {
		return fmt.Errorf("failed to start mDNS server: %w", err)
	}

	r.server = server

	log.Printf("[DISCOVERY] mDNS service registered: %s on port %d", r.serviceName, r.port)
	return nil
}

func (r *MdnsRegistrar) Stop() {
	if r.server != nil {
		r.server.Shutdown()
		log.Println("[DISCOVERY] mDNS service stopped")
	}
}
