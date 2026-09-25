package file

import (
	"net/netip"
	"testing"
)

func TestIsPrivateAddress(t *testing.T) {
	for _, address := range []string{
		"127.0.0.1",
		"10.0.0.5",
		"172.16.0.1",
		"192.168.1.1",
		"169.254.169.254",
		"100.64.0.1",
		"100.100.100.200",
		"100.127.255.254",
		"0.0.0.0",
		"::1",
		"fd00::1",
		"fe80::1",
		"::ffff:100.100.100.200",
	} {
		t.Run("refuses "+address, func(t *testing.T) {
			if !isPrivateAddress(netip.MustParseAddr(address).Unmap()) {
				t.Fatalf("expected %s to be private", address)
			}
		})
	}

	for _, address := range []string{
		"100.63.255.255",
		"100.128.0.0",
		"8.8.8.8",
		"2606:4700::1111",
	} {
		t.Run("accepts "+address, func(t *testing.T) {
			if isPrivateAddress(netip.MustParseAddr(address).Unmap()) {
				t.Fatalf("expected %s to be public", address)
			}
		})
	}
}
