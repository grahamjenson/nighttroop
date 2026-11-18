package s3utils

import (
	"context"
	"crypto/md5"
	"errors"
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
	etags, bodies map[string]string
	unconditional bool
	// unanswered is how many writes it takes and then answers 500, and busy how many it refuses with 409
	unanswered, busy int
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
		if f.busy > 0 {
			f.busy--
			w.WriteHeader(http.StatusConflict)
			fmt.Fprint(w, `<Error><Code>ConditionalRequestConflict</Code><Message>busy</Message></Error>`)
			return
		}
		f.etags[r.URL.Path], f.bodies[r.URL.Path] = fmt.Sprintf(`"%x"`, md5.Sum(body)), string(body)
		if f.unanswered > 0 {
			f.unanswered--
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("ETag", f.etags[r.URL.Path])
	case http.MethodHead, http.MethodGet:
		if !found {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("ETag", current)
		fmt.Fprint(w, f.bodies[r.URL.Path])
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
	fake := &lockS3{etags: map[string]string{}, bodies: map[string]string{}}
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
	fake.etags["/bucket/"+servers[0].lockKey()], fake.bodies["/bucket/"+servers[0].lockKey()] = `"another server"`, "another server's"
	fake.Unlock()
	select {
	case <-servers[0].Lost():
	case <-time.After(time.Second):
		t.Fatal("the server is not told its lock was taken")
	}
	require.True(t, servers[0].lockLost.Load())
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

// A server that started without S3 takes the lock only once S3 has worked three tests in a row, so one that comes and
// goes is not taken as back, and waits for a server that holds the lock
func TestLockWhenWorking(t *testing.T) {
	_, servers := lockServers(t, 2)
	oldEvery := workingEvery
	workingEvery = time.Millisecond
	t.Cleanup(func() { workingEvery = oldEvery })

	// Two passes, a failure, then it works: five more tests, three of them in a row
	results, tests := []bool{true, true, false, true, true, true}, 0
	flaky := func(context.Context) error {
		tests++
		if !results[tests-1] {
			return errors.New("refused")
		}
		return nil
	}
	require.True(t, servers[0].lockWhenWorking(t.Context(), flaky))
	require.Equal(t, 6, tests)
	require.NotNil(t, servers[0].unlock, "it holds the lock")

	// While that server holds the lock, another waits, however long S3 has worked
	ctx, cancel := context.WithTimeout(t.Context(), 3*lockStaleAfter)
	defer cancel()
	require.False(t, servers[1].lockWhenWorking(ctx, func(context.Context) error { return nil }))
}

// A server keeps its lock when S3 takes a write of it and never answers, or refuses one as another write is under
// way: neither is another server taking it, so the server is not told its lock is lost
func TestLockIsKeptThroughS3Faults(t *testing.T) {
	for name, fault := range map[string]func(*lockS3){
		// The SDK tries the write again, which S3 refuses, as the first try changed the lock
		"a write S3 never answers":   func(f *lockS3) { f.unanswered = 1 },
		"a write S3 refuses as busy": func(f *lockS3) { f.busy = 2 },
	} {
		t.Run(name, func(t *testing.T) {
			fake, servers := lockServers(t, 1)
			require.NoError(t, servers[0].Lock(t.Context()))
			written := func() string {
				fake.Lock()
				defer fake.Unlock()
				return fake.etags["/bucket/"+servers[0].lockKey()]
			}
			fake.Lock()
			fault(fake)
			fake.Unlock()
			// Past the fault, and the SDK's wait before it tries again, it writes its lock again, twice
			for range 2 {
				before := written()
				require.Eventually(t, func() bool {
					fake.Lock()
					faults := fake.unanswered + fake.busy
					fake.Unlock()
					return faults == 0 && written() != before
				}, 30*time.Second, 5*time.Millisecond, "it writes its lock again")
			}
			require.False(t, servers[0].lockLost.Load())
		})
	}
}
