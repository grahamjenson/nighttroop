package httputils

import (
	"fmt"
	"net"
	"net/netip"
	"slices"
	"strings"

	"nighttroop.t1d.nz/utils/config"
)

var blockedHostList = []string{
	"127.0.0.1",
	"localhost",
	"169.254.169.254", // AWS metadata service, just in case
}

// blockedPrefixes are the ranges the internet cannot reach that netip.Addr has no method for
var blockedPrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),      // "this network", which Linux connects to as the local host
	netip.MustParsePrefix("100.64.0.0/10"),  // shared address space: carrier-grade NAT, Tailscale's tailnets and Alibaba Cloud's metadata
	netip.MustParsePrefix("192.0.0.0/24"),   // IETF protocol assignments
	netip.MustParsePrefix("198.18.0.0/15"),  // benchmarking
	netip.MustParsePrefix("240.0.0.0/4"),    // reserved, and the broadcast address
	netip.MustParsePrefix("64:ff9b:1::/48"), // local-use NAT64, only ever routed inside a network
}

// nat64Prefix is the well-known NAT64 prefix, which reaches the IPv4 address in its last 4 bytes
var nat64Prefix = netip.MustParsePrefix("64:ff9b::/96")

// sixToFourPrefix is 6to4, which reaches the IPv4 address in its bytes 2 to 5
var sixToFourPrefix = netip.MustParsePrefix("2002::/16")

// NormalizeHost is a URL's host as it is compared and saved: lowercase, without a fully qualified name's trailing dot
func NormalizeHost(host string) string {
	return strings.TrimSuffix(strings.ToLower(host), ".")
}

func IsNightscoutURLBlocked(host string, cfg *config.ServerConfig) error {
	if cfg.IsDev() {
		return nil
	}
	host = NormalizeHost(host)
	if host == "" {
		return fmt.Errorf("Nightscout URL must have a host.")
	}

	// Cant be the host or a subdomain of the host
	if host == cfg.HostName || strings.HasSuffix(host, "."+cfg.HostName) {
		return fmt.Errorf("Nightscout cannot be on this host.")
	}

	// Check if the URL is in the blocked list
	if slices.Contains(blockedHostList, host) {
		return fmt.Errorf("Nightscout cannot be on this host.")
	}

	// Literal IPs are checked directly; hostnames are resolved so a name that
	// points at a private or local address is rejected too.
	ips := []net.IP{}
	if ip := net.ParseIP(host); ip != nil {
		ips = append(ips, ip)
	} else if resolved, err := net.LookupIP(host); err == nil {
		ips = resolved
	}
	for _, ip := range ips {
		if IsBlockedIP(ip) {
			return fmt.Errorf("Nightscout cannot be on a private or local network.")
		}
	}

	return nil
}

// IsBlockedIP is whether the address is one the internet cannot reach: unspecified, loopback, private, link-local,
// multicast or reserved, or an IPv4-mapped, NAT64 or 6to4 address that reaches such an IPv4 address
func IsBlockedIP(ip net.IP) bool {
	addr, ok := netip.AddrFromSlice(ip)
	if !ok {
		return true
	}
	addr = addr.Unmap()
	bytes := addr.As16()
	if nat64Prefix.Contains(addr) && isBlockedAddr(netip.AddrFrom4([4]byte(bytes[12:16]))) {
		return true
	}
	if sixToFourPrefix.Contains(addr) && isBlockedAddr(netip.AddrFrom4([4]byte(bytes[2:6]))) {
		return true
	}
	return isBlockedAddr(addr)
}

func isBlockedAddr(addr netip.Addr) bool {
	if addr.IsUnspecified() || addr.IsLoopback() || addr.IsPrivate() || addr.IsLinkLocalUnicast() ||
		addr.IsMulticast() || addr.IsLinkLocalMulticast() || addr.IsInterfaceLocalMulticast() {
		return true
	}
	return slices.ContainsFunc(blockedPrefixes, func(prefix netip.Prefix) bool { return prefix.Contains(addr) })
}
