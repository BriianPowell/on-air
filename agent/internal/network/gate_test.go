package network

import (
	"net/netip"
	"testing"
)

func TestParseAllowedNetworksAcceptsCIDRAndExactIP(t *testing.T) {
	prefixes, err := parseAllowedNetworks([]string{
		"192.168.1.0/24",
		"10.10.20.30",
	})
	if err != nil {
		t.Fatal(err)
	}

	if !prefixes[0].Contains(netip.MustParseAddr("192.168.1.42")) {
		t.Fatalf("expected CIDR prefix to contain LAN address")
	}
	if !prefixes[1].Contains(netip.MustParseAddr("10.10.20.30")) {
		t.Fatalf("expected exact IP prefix to contain the configured IP")
	}
	if prefixes[1].Contains(netip.MustParseAddr("10.10.20.31")) {
		t.Fatalf("expected exact IP prefix not to contain adjacent IP")
	}
}

func TestParseAllowedNetworksRejectsInvalidEntry(t *testing.T) {
	if _, err := parseAllowedNetworks([]string{"home"}); err == nil {
		t.Fatal("expected invalid network entry to fail")
	}
}
