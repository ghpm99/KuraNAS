package tuya

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"testing"
	"time"
)

func startClosedPortAddress(t *testing.T) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	listener.Close()
	return address
}

func TestFindDeviceReturnsTheCandidateThatAcceptsTheKey(t *testing.T) {
	rightAddress := startFakeDevice(t, &fakeDevice{deviceKey: []byte(testKey), dataPoints: newVacuumStatus()})
	wrongKeyAddress := startFakeDevice(t, &fakeDevice{deviceKey: []byte("fedcba9876543210"), dataPoints: newVacuumStatus()})
	silentAddress := startFakeDevice(t, &fakeDevice{deviceKey: []byte(testKey), isSilent: true})
	candidates := []string{startClosedPortAddress(t), wrongKeyAddress, silentAddress, rightAddress}

	address, err := NewClient(time.Second).FindDevice(context.Background(), newTestDevice("ignored"), candidates)

	if err != nil {
		t.Fatal(err)
	}
	if address != rightAddress {
		t.Fatalf("expected %s, got %s", rightAddress, address)
	}
}

func TestFindDeviceWithoutAcceptingCandidateIsNotFound(t *testing.T) {
	wrongKeyAddress := startFakeDevice(t, &fakeDevice{deviceKey: []byte("fedcba9876543210"), dataPoints: newVacuumStatus()})

	_, err := NewClient(time.Second).FindDevice(context.Background(), newTestDevice(""), []string{startClosedPortAddress(t), wrongKeyAddress})

	if !errors.Is(err, ErrDeviceNotFound) {
		t.Fatalf("expected ErrDeviceNotFound, got %v", err)
	}
}

func TestFindDeviceWithoutCandidatesIsNotFound(t *testing.T) {
	_, err := NewClient(time.Second).FindDevice(context.Background(), newTestDevice(""), nil)

	if !errors.Is(err, ErrDeviceNotFound) {
		t.Fatalf("expected ErrDeviceNotFound, got %v", err)
	}
}

func TestFindDeviceRejectsInvalidDeviceWithoutConnecting(t *testing.T) {
	fake := &fakeDevice{deviceKey: []byte(testKey), dataPoints: newVacuumStatus()}
	address := startFakeDevice(t, fake)

	_, err := NewClient(time.Second).FindDevice(context.Background(), Device{DeviceID: "bf1234", LocalKey: "short"}, []string{address})

	if !errors.Is(err, ErrInvalidDevice) {
		t.Fatalf("expected ErrInvalidDevice, got %v", err)
	}
	if len(fake.requests) != 0 {
		t.Fatal("no connection was expected")
	}
}

func TestFindDeviceEndsNearTheContextDeadline(t *testing.T) {
	var candidates []string
	for range 3 {
		candidates = append(candidates, startFakeDevice(t, &fakeDevice{deviceKey: []byte(testKey), isSilent: true}))
	}
	searchContext, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	startedAt := time.Now()

	_, err := NewClient(5*time.Second).FindDevice(searchContext, newTestDevice(""), candidates)

	if !errors.Is(err, ErrDeviceNotFound) {
		t.Fatalf("expected ErrDeviceNotFound, got %v", err)
	}
	if elapsed := time.Since(startedAt); elapsed > time.Second {
		t.Fatalf("search took %v", elapsed)
	}
}

func TestFindDeviceWithMoreCandidatesThanConcurrentProbes(t *testing.T) {
	rightAddress := startFakeDevice(t, &fakeDevice{deviceKey: []byte(testKey), dataPoints: newVacuumStatus()})
	closedAddress := startClosedPortAddress(t)
	var candidates []string
	for range maxConcurrentProbes + 10 {
		candidates = append(candidates, closedAddress)
	}
	candidates = append(candidates, rightAddress)

	address, err := NewClient(time.Second).FindDevice(context.Background(), newTestDevice(""), candidates)

	if err != nil || address != rightAddress {
		t.Fatalf("got %q, %v", address, err)
	}
}

func hostsOf(t *testing.T, prefixText string) []string {
	t.Helper()
	return candidateAddressesFromPrefixes([]netip.Prefix{netip.MustParsePrefix(prefixText)})
}

func TestCandidatesOfPrivateSlash24ExcludeNetworkBroadcastAndSelf(t *testing.T) {
	candidates := hostsOf(t, "192.168.18.7/24")

	if len(candidates) != 253 || candidates[0] != "192.168.18.1" || candidates[len(candidates)-1] != "192.168.18.254" {
		t.Fatalf("unexpected candidates: %d", len(candidates))
	}
	for _, candidate := range candidates {
		if candidate == "192.168.18.7" || candidate == "192.168.18.0" || candidate == "192.168.18.255" {
			t.Fatalf("unexpected candidate %s", candidate)
		}
	}
}

func TestCandidatesOfWideNetworkAreReducedToTheInterfaceSlash24(t *testing.T) {
	candidates := hostsOf(t, "172.21.13.211/20")

	if len(candidates) != 253 || candidates[0] != "172.21.13.1" || candidates[len(candidates)-1] != "172.21.13.254" {
		t.Fatalf("unexpected candidates: %d %v", len(candidates), candidates[:1])
	}
}

func TestCandidatesOfSmallerNetworkKeepItsOwnRange(t *testing.T) {
	candidates := hostsOf(t, "10.0.0.5/29")

	if len(candidates) != 5 || candidates[0] != "10.0.0.1" || candidates[4] != "10.0.0.6" {
		t.Fatalf("unexpected candidates: %v", candidates)
	}
}

func TestCandidatesIgnoreNonPrivateLoopbackIpv6AndHostPrefixes(t *testing.T) {
	for _, prefixText := range []string{"8.8.8.8/24", "127.0.0.1/8", "fd00::1/64", "192.168.1.1/32"} {
		if candidates := hostsOf(t, prefixText); len(candidates) != 0 {
			t.Fatalf("%s produced %d candidates", prefixText, len(candidates))
		}
	}
}

func TestCandidatesOfTwoPrefixesAreUnitedWithoutDuplicates(t *testing.T) {
	prefixes := []netip.Prefix{netip.MustParsePrefix("192.168.18.7/24"), netip.MustParsePrefix("192.168.18.9/24"), netip.MustParsePrefix("10.1.1.1/24")}

	candidates := candidateAddressesFromPrefixes(prefixes)

	if len(candidates) != 253+253+1 {
		t.Fatalf("unexpected count %d", len(candidates))
	}
}

func TestLocalNetworkCandidatesRuns(t *testing.T) {
	if _, err := LocalNetworkCandidates(); err != nil {
		t.Fatal(err)
	}
}
