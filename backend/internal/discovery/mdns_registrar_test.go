package discovery

import (
	"testing"
	"time"

	"github.com/hashicorp/mdns"
)

func TestMdnsRegistrarStartAndStop(t *testing.T) {
	registrar := NewMdnsRegistrar("_kuranas_test._tcp", 8000)

	if err := registrar.Start(); err != nil {
		t.Fatalf("failed to start mDNS registrar: %v", err)
	}

	registrar.Stop()
}

func TestMdnsRegistrarIsDiscoverable(t *testing.T) {
	registrar := NewMdnsRegistrar("_kuranas_test._tcp", 8000)

	if err := registrar.Start(); err != nil {
		t.Fatalf("failed to start mDNS registrar: %v", err)
	}
	defer registrar.Stop()

	entriesCh := make(chan *mdns.ServiceEntry, 1)

	go func() {
		params := mdns.DefaultParams("_kuranas_test._tcp")
		params.Entries = entriesCh
		params.Timeout = 2 * time.Second
		mdns.Query(params)
		close(entriesCh)
	}()

	found := false
	for entry := range entriesCh {
		if entry.Port == 8000 {
			found = true
			break
		}
	}

	if !found {
		t.Error("mDNS service was not discoverable")
	}
}

func TestMdnsRegistrarStopWithoutStart(t *testing.T) {
	registrar := NewMdnsRegistrar("_kuranas_test._tcp", 8000)
	registrar.Stop()
}
