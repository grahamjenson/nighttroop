package httputils

import (
	"fmt"
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

	// Cant be the host or a subdomain of the host
	if strings.HasSuffix(host, "."+cfg.HostName) {
		return fmt.Errorf("Nightscout cannot be on this host.")
	}

	// Check if the URL is in the blocked list
	for _, blocked := range blockedHostList {
		if host == blocked {
			return fmt.Errorf("Nightscout cannot be on this host.")
		}
	}

	return nil
}
