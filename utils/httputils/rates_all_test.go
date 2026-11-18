package httputils_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
	"nighttroop.t1d.nz/utils/httputils"
)

// A request is let through only when every limiter has a token, and one any limiter refuses spends none of the
// others', so an IP over its own site create limit cannot empty the server's
func TestWithinRateLimitAllSpendsNoneUnlessAllAllow(t *testing.T) {
	server, ip := rate.NewLimiter(rate.Limit(1.0/60), 20), rate.NewLimiter(rate.Limit(1.0/600), 1)

	allowed, wait := httputils.WithinRateLimitAll(server, ip)
	require.True(t, allowed)
	require.Zero(t, wait)
	require.InDelta(t, 19, server.Tokens(), 0.1)
	require.InDelta(t, 0, ip.Tokens(), 0.1)

	// The IP has none left, so it is refused until its next token, and the server keeps its own
	for range 10 {
		allowed, wait = httputils.WithinRateLimitAll(server, ip)
		require.False(t, allowed)
		require.InDelta(t, 600, wait, 1)
	}
	require.InDelta(t, 19, server.Tokens(), 0.1)
	require.InDelta(t, 0, ip.Tokens(), 0.1)

	// A limiter with no burst never lets one through, and spends none of the others' either
	allowed, wait = httputils.WithinRateLimitAll(server, rate.NewLimiter(1, 0))
	require.False(t, allowed)
	require.Zero(t, wait)
	require.InDelta(t, 19, server.Tokens(), 0.1)

	// One limiter alone spends its token when it allows, as before
	single := rate.NewLimiter(rate.Every(time.Minute), 2)
	allowed, _ = httputils.WithinRateLimitAll(single)
	require.True(t, allowed)
	require.InDelta(t, 1, single.Tokens(), 0.1)
}
