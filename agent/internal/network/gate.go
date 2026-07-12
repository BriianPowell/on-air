package network

import (
	"fmt"
	"net"
	"net/netip"
)

type Gate struct {
	allowedNetworks []string
}

func NewGate(allowedNetworks []string) *Gate {
	return &Gate{allowedNetworks: allowedNetworks}
}

func (g *Gate) Allowed() (bool, error) {
	if g == nil || len(g.allowedNetworks) == 0 {
		return true, nil
	}

	prefixes, err := parseAllowedNetworks(g.allowedNetworks)
	if err != nil {
		return false, err
	}

	addrs, err := interfaceAddrs()
	if err != nil {
		return false, err
	}

	for _, addr := range addrs {
		for _, prefix := range prefixes {
			if prefix.Contains(addr) {
				return true, nil
			}
		}
	}
	return false, nil
}

func parseAllowedNetworks(values []string) ([]netip.Prefix, error) {
	prefixes := make([]netip.Prefix, 0, len(values))
	for _, value := range values {
		prefix, err := parseNetwork(value)
		if err != nil {
			return nil, fmt.Errorf("parse allowed network %q: %w", value, err)
		}
		prefixes = append(prefixes, prefix)
	}
	return prefixes, nil
}

func parseNetwork(value string) (netip.Prefix, error) {
	prefix, err := netip.ParsePrefix(value)
	if err == nil {
		return prefix.Masked(), nil
	}

	addr, addrErr := netip.ParseAddr(value)
	if addrErr != nil {
		return netip.Prefix{}, err
	}
	return netip.PrefixFrom(addr, addr.BitLen()), nil
}

func interfaceAddrs() ([]netip.Addr, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("list network interfaces: %w", err)
	}

	var addrs []netip.Addr
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		ifaceAddrs, err := iface.Addrs()
		if err != nil {
			return nil, fmt.Errorf("list addresses for %s: %w", iface.Name, err)
		}

		for _, ifaceAddr := range ifaceAddrs {
			addr, ok := addrFromInterfaceAddr(ifaceAddr)
			if ok {
				addrs = append(addrs, addr)
			}
		}
	}
	return addrs, nil
}

func addrFromInterfaceAddr(ifaceAddr net.Addr) (netip.Addr, bool) {
	switch value := ifaceAddr.(type) {
	case *net.IPNet:
		addr, ok := netip.AddrFromSlice(value.IP)
		if !ok {
			return netip.Addr{}, false
		}
		return addr.Unmap(), true
	case *net.IPAddr:
		addr, ok := netip.AddrFromSlice(value.IP)
		if !ok {
			return netip.Addr{}, false
		}
		return addr.Unmap(), true
	default:
		return netip.Addr{}, false
	}
}
