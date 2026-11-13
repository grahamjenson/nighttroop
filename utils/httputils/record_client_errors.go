package httputils

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	lru "github.com/hashicorp/golang-lru/v2"
	"golang.org/x/time/rate"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/utils/aslog"
)

const (
	// clientErrorBodyBytes is how much of a refused request's response is kept to say why
	clientErrorBodyBytes = 1024
	// clientErrorTextLength is the most of a reason or User-Agent recorded
	clientErrorTextLength = 200
	// clientErrorEvery is how often the same client error of a site is recorded, so a client sending the same bad
	// request again and again cannot fill the errors or Sentry
	clientErrorEvery = time.Minute
	// clientErrorsKept is how many recent client errors, and sites, are remembered for those limits
	clientErrorsKept = 10000
	// clientErrorsPerMinute is the most client errors of one site recorded a minute, however they differ, as a reason
	// can name a field the client sent
	clientErrorsPerMinute = 20
)

// clientErrorStatuses are the client errors recorded: a request refused for what it sent, which shows what a client
// gets wrong. Missing credentials or permissions, a record not found and rate limits are left out, being expected.
var clientErrorStatuses = []int{http.StatusBadRequest, http.StatusMethodNotAllowed, http.StatusRequestEntityTooLarge,
	http.StatusUnprocessableEntity}

// errorBody keeps the start of a response refused with a client error, to say why. It is after compression, so it
// reads the body as it was written.
type errorBody struct {
	status func() int
	body   []byte
}

func (b *errorBody) Write(p []byte) (int, error) {
	if slices.Contains(clientErrorStatuses, b.status()) {
		b.body = append(b.body, p[:min(len(p), clientErrorBodyBytes-len(b.body))]...)
	}
	return len(p), nil
}

type leftOutKey struct{}

// SetLeftOut keeps why a request that was taken left out part of what it sent, such as a batch's entries it could not
// store, or a document's fields, so RecordClientErrors records it as a client error, though its status is not one. Each
// reason is added to those kept, once, so a request records one client error with them all.
func SetLeftOut(ctx context.Context, reason string) {
	if kept, ok := ctx.Value(leftOutKey{}).(*string); ok && !strings.Contains(*kept, reason) {
		*kept = strings.TrimPrefix(*kept+"; "+reason, "; ")
	}
}

// RecordClientErrors records each request refused for what it sent, such as a 400, or taken but for a part it left
// out, as SetLeftOut said, in the logs DB's errors, with its site when it has one, and in Sentry as a warning: its
// route's pattern, why it was refused, as its response said, the identity that sent it, its User-Agent and what it
// sent, its credentials redacted. Why is what the handler said with SetLeftOut or SetErrorReason, when it did. The same
// error of a site is recorded at most every clientErrorEvery, and at most clientErrorsPerMinute of a site's.
func RecordClientErrors(logs *db.LogsDB) func(http.Handler) http.Handler {
	recent, _ := lru.New[string, time.Time](clientErrorsKept)
	perSite, _ := lru.New[string, *rate.Limiter](clientErrorsKept)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			body := &errorBody{status: ww.Status}
			ww.Tee(body)
			r, reason := withErrorReason(r)
			leftOut := new(string)
			r = r.WithContext(context.WithValue(r.Context(), leftOutKey{}, leftOut))
			next.ServeHTTP(ww, r)
			if *leftOut == "" && !slices.Contains(clientErrorStatuses, ww.Status()) {
				return
			}
			attributes := db.ClientErrorAttributes{Status: ww.Status(), Method: r.Method, Reason: clip(cmp.Or(*leftOut, *reason, errorReason(body.body))),
				RequestID: middleware.GetReqID(r.Context()), Client: clip(r.UserAgent()), RequestBody: ErrorRequestBody(r.Context())}
			if routeContext := chi.RouteContext(r.Context()); routeContext != nil {
				attributes.Route = routeContext.RoutePattern()
			}
			var siteID *string
			site := ""
			if siteContext := GetSiteContext(r.Context()); siteContext != nil && siteContext.SiteModel != nil {
				siteID, site = &siteContext.SiteModel.ID, siteContext.SiteModel.ID
				if identity := siteContext.IdentityModel; identity != nil {
					attributes.Identity, attributes.Role = identity.Name, string(identity.Role)
				}
			}
			key := fmt.Sprint(site, attributes.Status, attributes.Method, attributes.Route, attributes.Reason)
			if last, ok := recent.Get(key); ok && time.Since(last) < clientErrorEvery {
				return
			}
			limiter, ok := perSite.Get(site)
			if !ok {
				limiter = rate.NewLimiter(rate.Every(time.Minute/clientErrorsPerMinute), clientErrorsPerMinute)
				perSite.Add(site, limiter)
			}
			if !limiter.Allow() {
				return
			}
			recent.Add(key, time.Now())
			logs.ClientError(context.WithoutCancel(r.Context()), siteID, attributes)
			reportClientError(r, attributes)
		})
	}
}

// errorReason is why a response refused a request: its JSON's message, or its text
func errorReason(body []byte) string {
	var reply struct {
		Message string `json:"message"`
	}
	if json.Unmarshal(body, &reply) == nil {
		return clip(reply.Message)
	}
	return clip(string(body))
}

// clip is text trimmed to at most clientErrorTextLength
func clip(text string) string {
	text = strings.TrimSpace(text)
	if len(text) > clientErrorTextLength {
		return text[:clientErrorTextLength]
	}
	return text
}

// reportClientError sends a client error to Sentry, when it is set up, as a warning grouped by its status, route and
// reason, tagged with the identity's role and the client, and with its request's site and identity, as an error is
func reportClientError(r *http.Request, attributes db.ClientErrorAttributes) {
	hub := sentry.GetHubFromContext(r.Context())
	if hub == nil {
		return
	}
	hub.WithScope(func(scope *sentry.Scope) {
		scope.SetLevel(sentry.LevelWarning)
		scope.SetContext("extra", aslog.SentryContext(r.Context()))
		scope.SetTags(map[string]string{"kind": "client_error", "status": strconv.Itoa(attributes.Status), "route": attributes.Route,
			"role": attributes.Role, "client": attributes.Client})
		scope.SetFingerprint([]string{"client_error", strconv.Itoa(attributes.Status), attributes.Method, attributes.Route, attributes.Reason})
		hub.CaptureMessage(fmt.Sprintf("%d %s %s: %s", attributes.Status, attributes.Method, attributes.Route, attributes.Reason))
	})
}
