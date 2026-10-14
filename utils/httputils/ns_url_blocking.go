package httputils

import (
	"fmt"
	"net"
	"slices"
	"strings"

	"nighttroop.t1d.nz/utils/config"
)

var blockedHostList = []string{
	"127.0.0.1",
	"localhost",
	"169.254.169.254", // AWS metadata service, just in case
}

var blockedIPCidrs = []string{

	"127.0.0.0/8",
	"10.0.0.0/8",
	"172.16.0.0/12",
	"192.168.0.0/16",
	"169.254.0.0/16",
	"::1/128",
	"fc00::/7",
	"fe80::/10",
}

func IsNightscoutURLBlocked(host string, cfg *config.ServerConfig) error {
	if cfg.IsDev() {
		return nil
	}
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
		if isBlockedIP(ip) {
			return fmt.Errorf("Nightscout cannot be on a private or local network.")
		}
	}

	return nil
}

func isBlockedIP(ip net.IP) bool {
	if ip.IsUnspecified() {
		return true
	}
	for _, cidr := range blockedIPCidrs {
		if _, network, err := net.ParseCIDR(cidr); err == nil && network.Contains(ip) {
			return true
		}
	}
	return false
}
