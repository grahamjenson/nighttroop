package aslog

import (
	"cmp"
	"context"
	"log/slog"
	"slices"
	"sync"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/utils/metrics"
)

// The most keys counted apart for each reason, past which they are counted together as other, so a sweep of every
// subdomain is held in fixed memory; the most of a key kept; and how many of the most refused keys a summary lists
const (
	maxRefusalKeys     = 100
	maxRefusalKeyBytes = 64
	topRefusalKeys     = 10
)

// Refusals sums the requests refused before they are signed in, such as for a site that does not exist, a missing
// token or a rate limit, so a stranger sending many cannot fill the logs DB with them. Flush logs a summary of each.
type Refusals struct {
	mu      sync.Mutex
	reasons map[refusal]*refused
}

// refusal is why a request was refused, the debug log's message, and the status it was answered with
type refusal struct {
	reason string
	status int
}

// refused is how many requests a refusal had, how many for each key and how many for keys past maxRefusalKeys
type refused struct {
	total, other int
	keys         map[string]int
}

// KeyCount is a key's count in a summary
type KeyCount struct {
	Key   string `json:"key"`
	Count int    `json:"count"`
}

// RefusedRequests sums the server's refused requests, which the background jobs flush each minute
var RefusedRequests = &Refusals{}

// Refuse logs a refused request at debug, which the logs DB does not keep, and counts it by its message, status and
// key: a missing site's subdomain, a host, or else the route's pattern, never an IP address, token or path
func Refuse(ctx context.Context, status int, key, msg string, args ...any) {
	slog.DebugContext(ctx, msg, args...)
	if routeContext := chi.RouteContext(ctx); key == "" && routeContext != nil {
		key = routeContext.RoutePattern()
	}
	RefusedRequests.Add(msg, status, key)
	metrics.Refused(msg, status)
}

// Add counts a request refused for reason with status, by key
func (r *Refusals) Add(reason string, status int, key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.reasons == nil {
		r.reasons = map[refusal]*refused{}
	}
	counts := r.reasons[refusal{reason, status}]
	if counts == nil {
		counts = &refused{keys: map[string]int{}}
		r.reasons[refusal{reason, status}] = counts
	}
	counts.total++
	key = truncate(key, maxRefusalKeyBytes)
	switch {
	case key == "":
	case counts.keys[key] > 0 || len(counts.keys) < maxRefusalKeys:
		counts.keys[key]++
	default:
		counts.other++
	}
}

// Flush logs one summary at info of each reason with requests refused since the last flush, with its most refused
// keys, and starts counting again
func (r *Refusals) Flush(ctx context.Context) {
	r.mu.Lock()
	reasons := r.reasons
	r.reasons = nil
	r.mu.Unlock()
	for k, counts := range reasons {
		slog.InfoContext(ctx, "Unauthenticated requests refused", "reason", k.reason, "status", k.status, "total", counts.total,
			"keys", len(counts.keys), "other", counts.other, "top", topKeys(counts.keys))
	}
}

// topKeys are the topRefusalKeys most counted keys, most first, then by key
func topKeys(keys map[string]int) []KeyCount {
	top := []KeyCount{}
	for key, count := range keys {
		top = append(top, KeyCount{key, count})
	}
	slices.SortFunc(top, func(a, b KeyCount) int { return cmp.Or(cmp.Compare(b.Count, a.Count), cmp.Compare(a.Key, b.Key)) })
	return top[:min(len(top), topRefusalKeys)]
}
