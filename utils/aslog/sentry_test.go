package aslog_test

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi/v5"
	slogsentry "github.com/samber/slog-sentry/v2"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/aslog"
)

// sentEvents is a Sentry transport keeping the events it is given rather than sending them
type sentEvents struct {
	mu     sync.Mutex
	events []*sentry.Event
}

func (s *sentEvents) Flush(time.Duration) bool              { return true }
func (s *sentEvents) FlushWithContext(context.Context) bool { return true }
func (s *sentEvents) Configure(sentry.ClientOptions)        {}
func (s *sentEvents) Close()                                {}
func (s *sentEvents) SendEvent(event *sentry.Event) {
	s.mu.Lock()
	s.events = append(s.events, event)
	s.mu.Unlock()
}
func (s *sentEvents) sent() []*sentry.Event { s.mu.Lock(); defer s.mu.Unlock(); return s.events }

// An error logged in a request reaches Sentry with its site, identity, request and route as tags, its identity as its
// user, and the stack it was logged from, so its issue can be searched by them and shows where it happened
func TestErrorsReachSentrySearchableByTheirContext(t *testing.T) {
	transport := &sentEvents{}
	client, err := sentry.NewClient(sentry.ClientOptions{Dsn: "https://public@sentry.example.invalid/1", Transport: transport,
		BeforeSend: func(event *sentry.Event, _ *sentry.EventHint) *sentry.Event { return aslog.SentryEvent(event) }})
	require.NoError(t, err)
	hub := sentry.NewHub(client, sentry.NewScope())
	logger := slog.New(aslog.ContextHandler{Handler: slogsentry.Option{Level: slog.LevelError, Hub: hub,
		AttrFromContext: []func(context.Context) []slog.Attr{aslog.SentryRoute}}.NewSentryHandler()})

	router := chi.NewRouter()
	router.Get("/data/{collection}", func(w http.ResponseWriter, r *http.Request) {
		// As the request logger and sign in add them
		ctx := aslog.NewContext(r.Context(), slog.String("request_id", "req-1"), slog.String("site_id", "abcdefgh"))
		aslog.AddAttrs(ctx, slog.Int64("identity_id", 7), slog.String("identity_name", "carer"), slog.String("role", "user"))
		logger.ErrorContext(ctx, "Cannot read the collection", "error", errors.New("disk on fire"))
		logger.ErrorContext(ctx, "Something is wrong")
	})
	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/data/glucose", nil))

	events := transport.sent()
	require.Len(t, events, 2)
	for _, event := range events {
		for key, want := range map[string]string{"site_id": "abcdefgh", "identity_id": "7", "role": "user", "request_id": "req-1",
			"method": "GET", "route": "/data/{collection}"} {
			require.Equal(t, want, event.Tags[key], key)
		}
		require.Equal(t, "abcdefgh/7", event.User.ID)
		require.Equal(t, "carer", event.User.Username)
		require.Equal(t, "GET /data/{collection}", event.Transaction)
	}

	// Each has the stack it was logged from, ending in the handler that logged it, without slog's frames
	withError, withoutError := events[0].Exception[len(events[0].Exception)-1].Stacktrace, events[1].Threads[0].Stacktrace
	for _, stack := range []*sentry.Stacktrace{withError, withoutError} {
		require.NotNil(t, stack)
		last := stack.Frames[len(stack.Frames)-1]
		require.True(t, strings.HasPrefix(last.Function, "TestErrorsReachSentrySearchableByTheirContext"), last.Function)
		for _, frame := range stack.Frames {
			require.NotEqual(t, "log/slog", frame.Module)
		}
	}
}

// A control identity's ID is told apart from a site's, and a warning, such as a client error's report, is given the
// context it is captured with but not a stack, as where it was captured says nothing of the client's mistake
func TestSentryEventOfAControlIdentityAndAWarning(t *testing.T) {
	ctx := aslog.NewContext(context.Background(), slog.Int64("identity_id", 2), slog.String("identity_name", "admin"))
	event := sentry.NewEvent()
	event.Level = sentry.LevelWarning
	event.Contexts["extra"] = aslog.SentryContext(ctx)
	event = aslog.SentryEvent(event)
	require.Equal(t, "control/2", event.User.ID)
	require.Equal(t, "admin", event.User.Username)
	require.Equal(t, "2", event.Tags["identity_id"])
	require.Empty(t, event.Threads)
	require.Empty(t, event.Tags["site_id"])
}
