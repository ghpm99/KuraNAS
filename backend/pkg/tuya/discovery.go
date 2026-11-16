package tuya

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"sync"
)

const (
	maxConcurrentProbes  = 256
	smallestSearchedBits = 24
	largestSearchedBits  = 30
	ipv4AddressBits      = 32
)

// LocalNetworkCandidates lists every host of the private IPv4 networks (10/8, 172.16/12, 192.168/16)
// of the active, non-loopback interfaces, excluding the server's own addresses. A network wider than
// /24 is reduced to the /24 that contains the interface address.
func LocalNetworkCandidates() ([]string, error) {
	prefixes, err := activeInterfacePrefixes()
	if err != nil {
		return nil, fmt.Errorf("tuya: cannot list network interfaces: %w", err)
	}
	return candidateAddressesFromPrefixes(prefixes), nil
}

// FindDevice returns the address (host, without port) of the first candidate that accepts the
// DeviceID and LocalKey of device; the Address field of device is ignored. Candidates are probed
// concurrently and returned exactly as given. It returns ErrInvalidDevice before any connection when
// the id or key is invalid, and ErrDeviceNotFound when no candidate accepts them, including when ctx ends first.
func (client *Client) FindDevice(ctx context.Context, device Device, candidateAddresses []string) (string, error) {
	if err := validateCredentials(device); err != nil {
		return "", err
	}
	if len(candidateAddresses) == 0 {
		return "", ErrDeviceNotFound
	}
	searchContext, stopSearch := context.WithCancel(ctx)
	defer stopSearch()

	pendingAddresses := make(chan string)
	acceptedAddress := make(chan string, 1)
	var runningProbes sync.WaitGroup
	for range min(maxConcurrentProbes, len(candidateAddresses)) {
		runningProbes.Add(1)
		go func() {
			defer runningProbes.Done()
			client.probeAddresses(searchContext, stopSearch, device, pendingAddresses, acceptedAddress)
		}()
	}
	feedAddresses(searchContext, candidateAddresses, pendingAddresses)
	runningProbes.Wait()

	select {
	case address := <-acceptedAddress:
		return address, nil
	default:
		return "", ErrDeviceNotFound
	}
}

func feedAddresses(searchContext context.Context, candidateAddresses []string, pendingAddresses chan<- string) {
	defer close(pendingAddresses)
	for _, address := range candidateAddresses {
		select {
		case pendingAddresses <- address:
		case <-searchContext.Done():
			return
		}
	}
}

func (client *Client) probeAddresses(searchContext context.Context, stopSearch context.CancelFunc, device Device, pendingAddresses <-chan string, acceptedAddress chan<- string) {
	for address := range pendingAddresses {
		if searchContext.Err() != nil {
			return
		}
		if !client.acceptsCredentials(searchContext, device, address) {
			continue
		}
		select {
		case acceptedAddress <- address:
		default:
		}
		stopSearch()
		return
	}
}

func (client *Client) acceptsCredentials(searchContext context.Context, device Device, address string) bool {
	device.Address = address
	_, err := client.ReadDataPoints(searchContext, device)
	return err == nil
}

func activeInterfacePrefixes() ([]netip.Prefix, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	var prefixes []netip.Prefix
	for _, networkInterface := range interfaces {
		if networkInterface.Flags&net.FlagUp == 0 || networkInterface.Flags&net.FlagLoopback != 0 {
			continue
		}
		interfaceAddresses, err := networkInterface.Addrs()
		if err != nil {
			continue
		}
		prefixes = append(prefixes, prefixesOfAddresses(interfaceAddresses)...)
	}
	return prefixes, nil
}

func prefixesOfAddresses(interfaceAddresses []net.Addr) []netip.Prefix {
	var prefixes []netip.Prefix
	for _, interfaceAddress := range interfaceAddresses {
		network, ok := interfaceAddress.(*net.IPNet)
		if !ok {
			continue
		}
		address, ok := netip.AddrFromSlice(network.IP)
		if !ok {
			continue
		}
		maskBits, _ := network.Mask.Size()
		prefixes = append(prefixes, netip.PrefixFrom(address.Unmap(), maskBits))
	}
	return prefixes
}

func candidateAddressesFromPrefixes(prefixes []netip.Prefix) []string {
	seenAddresses := make(map[netip.Addr]bool)
	var candidateAddresses []string
	for _, prefix := range prefixes {
		for _, host := range searchableHosts(prefix) {
			if seenAddresses[host] {
				continue
			}
			seenAddresses[host] = true
			candidateAddresses = append(candidateAddresses, host.String())
		}
	}
	return candidateAddresses
}

func searchableHosts(prefix netip.Prefix) []netip.Addr {
	interfaceAddress := prefix.Addr()
	if !interfaceAddress.Is4() || !interfaceAddress.IsPrivate() || prefix.Bits() > largestSearchedBits {
		return nil
	}
	searchedBits := max(prefix.Bits(), smallestSearchedBits)
	network := netip.PrefixFrom(interfaceAddress, searchedBits).Masked()
	hostCount := 1<<(ipv4AddressBits-searchedBits) - 2
	hosts := make([]netip.Addr, 0, hostCount)
	host := network.Addr().Next()
	for range hostCount {
		if host != interfaceAddress {
			hosts = append(hosts, host)
		}
		host = host.Next()
	}
	return hosts
}
