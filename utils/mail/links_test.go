package mail

import (
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// An address is one email address as typed, of at most 254 characters
func TestAddress(t *testing.T) {
	for email, valid := range map[string]bool{
		"sam@example.com":                   true,
		strings.Repeat("a", 248) + "@x.com": true,
		strings.Repeat("a", 249) + "@x.com": false,
		"":                                  false,
		"sam":                               false,
		"Sam <sam@example.com>":             false,
		"sam@example.com, pat@example.com":  false,
	} {
		require.Equal(t, valid, Address(email), email)
	}
}

// A link works once, with its key, for its owner, until it is old or forgotten, and an owner waits to be sent another
func TestLinks(t *testing.T) {
	links, key := NewLinks(time.Hour, time.Minute), Key()
	require.Len(t, key, 64)
	require.NotEqual(t, key, Key())
	require.Zero(t, links.Wait("a"))

	links.Sent("a", key, "enc")
	require.InDelta(t, time.Minute, links.Wait("a"), float64(time.Second))
	require.Zero(t, links.Wait("b"), "each owner waits for their own")
	_, ok := links.Take("a", Key())
	require.False(t, ok, "another key")
	_, ok = links.Take("b", key)
	require.False(t, ok, "another owner")
	emailEnc, ok := links.Take("a", key)
	require.True(t, ok)
	require.Equal(t, "enc", emailEnc)
	_, ok = links.Take("a", key)
	require.False(t, ok, "used already")
	require.NotZero(t, links.Wait("a"), "using a link does not let another be sent at once")

	links.Sent("a", key, "enc")
	links.Forget("a")
	_, ok = links.Take("a", key)
	require.False(t, ok, "forgotten")
	require.NotZero(t, links.Wait("a"), "forgetting a link does not let another be sent at once")

	old := NewLinks(0, 0)
	old.Sent("a", key, "enc")
	_, ok = old.Take("a", key)
	require.False(t, ok, "old")
	require.Zero(t, old.Wait("a"))
}

// An owner's turn is taken in one step, before the email is sent: of the requests at once, one sends. An email that
// was not sent gives the turn back, with the link that was there.
func TestLinksClaim(t *testing.T) {
	links, key := NewLinks(time.Hour, time.Minute), Key()
	var start, done sync.WaitGroup
	var took atomic.Int32
	start.Add(1)
	for range 50 {
		done.Go(func() {
			start.Wait()
			if wait, _ := links.Claim("a"); wait == 0 {
				took.Add(1)
			}
		})
	}
	start.Done()
	done.Wait()
	require.EqualValues(t, 1, took.Load(), "requests at once")

	wait, notSent := links.Claim("b")
	require.Zero(t, wait)
	wait, again := links.Claim("b")
	require.InDelta(t, time.Minute, wait, float64(time.Second), "an email being sent counts")
	require.Nil(t, again)
	notSent()
	require.Zero(t, links.Wait("b"), "an email not sent does not count")

	links.Sent("b", key, "enc")
	wait, _ = links.Claim("b")
	require.InDelta(t, time.Minute, wait, float64(time.Second), "an email sent counts")

	old := NewLinks(time.Hour, 0)
	old.Sent("a", key, "enc")
	wait, notSent = old.Claim("a")
	require.Zero(t, wait)
	_, ok := old.Take("a", key)
	require.False(t, ok, "the link before does not work while another is sent")
	notSent()
	emailEnc, ok := old.Take("a", key)
	require.True(t, ok, "an email not sent puts back the link before")
	require.Equal(t, "enc", emailEnc)
}
