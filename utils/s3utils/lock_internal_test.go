package s3utils

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/config"
)

// lockS3 is an S3 that keeps objects by path and their ETags, and refuses a conditional write as S3 does, with 412,
// unless unconditional, as a store that does not know conditional writes
type lockS3 struct {
	sync.Mutex
	etags         map[string]string
	unconditional bool
}

func (f *lockS3) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.Lock()
	defer f.Unlock()
	current, found := f.etags[r.URL.Path]
	switch r.Method {
	case http.MethodPut:
		body, _ := io.ReadAll(r.Body)
		ifNone, ifMatch := r.Header.Get("If-None-Match"), r.Header.Get("If-Match")
		if !f.unconditional && ((ifNone == "*" && found) || (ifMatch != "" && ifMatch != current)) {
			w.WriteHeader(http.StatusPreconditionFailed)
			fmt.Fprint(w, `<Error><Code>PreconditionFailed</Code><Message>refused</Message></Error>`)
			return
		}
		f.etags[r.URL.Path] = fmt.Sprintf(`"%x"`, md5.Sum(body))
		w.Header().Set("ETag", f.etags[r.URL.Path])
	case http.MethodHead:
		if !found {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("ETag", current)
	case http.MethodDelete:
		delete(f.etags, r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}
}

// held is whether the fake has a lock
func (f *lockS3) held() bool {
	f.Lock()
	defer f.Unlock()
	return len(f.etags) > 0
}

// lockServers are an S3 and n servers that back up to the same host's folder of its bucket, with a lock written every
// 20 ms, stale after 100 ms, and looked at every 5 ms
func lockServers(t *testing.T, n int) (*lockS3, []*S3Client) {
	oldHeartbeat, oldStale, oldPoll := lockHeartbeat, lockStaleAfter, lockPoll
	lockHeartbeat, lockStaleAfter, lockPoll = 20*time.Millisecond, 100*time.Millisecond, 5*time.Millisecond
	t.Cleanup(func() { lockHeartbeat, lockStaleAfter, lockPoll = oldHeartbeat, oldStale, oldPoll })
	fake := &lockS3{etags: map[string]string{}}
	server := httptest.NewServer(fake)
	t.Cleanup(server.Close)
	clients := make([]*S3Client, n)
	for i := range clients {
		clients[i] = NewS3Client(&config.ServerConfig{S3URL: server.URL, S3Bucket: "bucket", HostName: "nighttroop.test",
			S3AccessKeyID: "key", S3AccessSecretKey: "secret"})
		clients[i].ServerConfig.Status.S3Working = true
		t.Cleanup(clients[i].Unlock)
	}
	return fake, clients
}

// locked starts c taking the lock, and gives a channel that gets its result
func locked(ctx context.Context, c *S3Client) <-chan error {
	done := make(chan error, 1)
	go func() { done <- c.Lock(ctx) }()
	return done
}

// A second server waits while the first holds the lock, however long, and takes it at once when the first releases it
func TestLockWaitsForTheServerThatHoldsIt(t *testing.T) {
	fake, servers := lockServers(t, 2)
	require.NoError(t, servers[0].Lock(t.Context()))

	second := locked(t.Context(), servers[1])
	select {
	case err := <-second:
		t.Fatalf("the second server took a lock that is held: %v", err)
	case <-time.After(3 * lockStaleAfter):
	}

	servers[0].Unlock()
	select {
	case err := <-second:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("the second server did not take the released lock")
	}
	require.True(t, fake.held())
	servers[1].Unlock()
	require.False(t, fake.held(), "a released lock is deleted")
}

// A lock whose server stopped without releasing it, as in a crash, is taken over once it is stale, and not before
func TestLockTakesOverAStaleLock(t *testing.T) {
	fake, servers := lockServers(t, 1)
	fake.etags["/bucket/"+servers[0].lockKey()] = `"crashed"`

	started := time.Now()
	require.NoError(t, servers[0].Lock(t.Context()))
	require.GreaterOrEqual(t, time.Since(started), lockStaleAfter)
	require.NotEqual(t, `"crashed"`, fake.etags["/bucket/"+servers[0].lockKey()])
}

// A server waiting for the lock stops waiting when it is told to stop
func TestLockStopsWaitingWhenCancelled(t *testing.T) {
	_, servers := lockServers(t, 2)
	require.NoError(t, servers[0].Lock(t.Context()))
	ctx, cancel := context.WithCancel(t.Context())
	second := locked(ctx, servers[1])
	cancel()
	require.ErrorIs(t, <-second, context.Canceled)
}

// A server whose lock another took backs up no more, and leaves the lock to that server when it stops
func TestLockLostStopsUploads(t *testing.T) {
	fake, servers := lockServers(t, 1)
	require.NoError(t, servers[0].Lock(t.Context()))
	require.NoError(t, servers[0].uploadStreamToS3(t.Context(), "nighttroop.test/sites/site.tar.enc", SiteRetention, strings.NewReader("backup")))

	fake.Lock()
	fake.etags["/bucket/"+servers[0].lockKey()] = `"another server"`
	fake.Unlock()
	require.Eventually(t, servers[0].lockLost.Load, time.Second, 5*time.Millisecond)
	err := servers[0].uploadStreamToS3(t.Context(), "nighttroop.test/sites/site.tar.enc", SiteRetention, strings.NewReader("backup"))
	require.ErrorIs(t, err, ErrLockLost)

	servers[0].Unlock()
	require.Equal(t, `"another server"`, fake.etags["/bucket/"+servers[0].lockKey()])
}

// With an S3 that does not refuse a conditional write there is no lock, and without S3 none is tried
func TestLockIsOffWithoutConditionalWritesOrS3(t *testing.T) {
	fake, servers := lockServers(t, 2)
	fake.unconditional = true
	require.NoError(t, servers[0].Lock(t.Context()))
	require.Nil(t, servers[0].unlock, "no lock is kept")
	require.NoError(t, servers[1].Lock(t.Context()), "so a second server starts")

	off := NewS3Client(&config.ServerConfig{})
	require.NoError(t, off.Lock(t.Context()))
	require.Nil(t, off.unlock)
}
