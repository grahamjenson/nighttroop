// Package authclients is where each identity recently authenticated from, kept in memory only, so a site's admin can
// see an identity shared by several apps or devices. It is a hint, not a record: it is lost on restart, and the least
// recently used identities are dropped.
package authclients

import (
	"net"
	"net/netip"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
)

const (
	// maxClients is how many clients of each identity are kept, the most recently seen
	maxClients = 5
	// maxUserAgent is the longest User-Agent kept
	maxUserAgent = 128
	// SharedWithin is how recently two apps must both have authenticated as one identity for it to be shared
	SharedWithin = 24 * time.Hour
)

// Client is an app, by its User-Agent, on a network that an identity authenticated from
type Client struct {
	UserAgent string `json:"user_agent"`
	// Network is the IP's /24, or /48 for IPv6, so the address itself is never kept
	Network string `json:"network"`
	LastAt  int64  `json:"last_at"`
}

// Recent is the clients each identity recently authenticated from, keyed by site and identity
type Recent struct {
	mu      sync.Mutex
	clients *lru.Cache[string, []Client]
}

// New keeps the clients of at most size identities
func New(size int) *Recent {
	// It only fails for a size under 1
	clients, _ := lru.New[string, []Client](size)
	return &Recent{clients: clients}
}

// Add records an identity authenticating now from the User-Agent and IP
func (r *Recent) Add(key string, userAgent string, ip string, now time.Time) {
	if len(userAgent) > maxUserAgent {
		userAgent = userAgent[:maxUserAgent]
	}
	client := Client{UserAgent: userAgent, Network: network(ip), LastAt: now.UnixMilli()}
	r.mu.Lock()
	defer r.mu.Unlock()
	// The client moves to the front, replacing itself, and the oldest are dropped
	old, _ := r.clients.Get(key)
	clients := []Client{client}
	for _, c := range old {
		if (c.UserAgent != client.UserAgent || c.Network != client.Network) && len(clients) < maxClients {
			clients = append(clients, c)
		}
	}
	r.clients.Add(key, clients)
}

// Get is the clients the identity recently authenticated from, newest first
func (r *Recent) Get(key string) []Client {
	clients, _ := r.clients.Get(key)
	return clients
}

// Shared is whether more than one app authenticated as the identity within SharedWithin of now. One app on another
// network, as a phone moving from wifi to mobile data, is not.
func Shared(clients []Client, now time.Time) bool {
	since := now.Add(-SharedWithin).UnixMilli()
	apps := map[string]bool{}
	for _, c := range clients {
		if c.LastAt >= since {
			apps[c.UserAgent] = true
		}
	}
	return len(apps) > 1
}

// network is the /24 or /48 the IP is in, or "" when it is not an IP
func network(ip string) string {
	if host, _, err := net.SplitHostPort(ip); err == nil {
		ip = host
	}
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return ""
	}
	bits := 48
	if addr.Unmap().Is4() {
		addr, bits = addr.Unmap(), 24
	}
	prefix, _ := addr.Prefix(bits)
	return prefix.String()
}
