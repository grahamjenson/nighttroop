package api_test

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"maps"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"nighttroop.t1d.nz/server"
	"nighttroop.t1d.nz/tests/dsl/jsonapi"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/testutils"
)

// FuzzEndpoints replays every request the .http files make, to the Nightscout API, the site API and the control
// API, signed in as each one's admin, with its path, query, body, method and headers mangled in the ways most likely
// to reach a 500: nonsense ids, operators and numbers out of range, and bodies of the wrong type or encoding.
// A request may be refused or not found; it fails only if the server panics or answers 5xx.
func FuzzEndpoints(f *testing.F) {
	fuzz := newEndpointFuzzer(f)
	// Each endpoint as it is, then mangled a little and a lot
	for i := range fuzz.endpoints {
		for _, modifications := range []int{0, 3, 12} {
			f.Add(int64(i), modifications, i)
		}
	}
	f.Fuzz(func(t *testing.T, seed int64, modifications int, endpointIndex int) {
		if failure := fuzz.send(seed, modifications, endpointIndex); failure != "" {
			t.Error(failure)
		}
	})
}

// endpointFuzzer sends the .http files' endpoints, mangled, to a router signed in to each API as its admin
type endpointFuzzer struct {
	router    http.Handler
	endpoints []*endpoint
	hosts     map[endpointHost]struct{ host, auth string }
	// ids are those of records the endpoints made, so updates, reads and deletes reach records that exist
	ids []string
}

var idPattern = regexp.MustCompile(`"(?:_id|id|identifier)":\s*("[^"]{1,64}"|\d{1,19})`)

// makeRecords sends every endpoint as the .http files do, twice so the second reads what the first made, keeping the
// ids of the records they answer with
func (fuzz *endpointFuzzer) makeRecords() {
	seen := map[string]bool{}
	for range 2 {
		for i := range fuzz.endpoints {
			rr := fuzz.serve(rand.New(rand.NewSource(0)), i, 0)
			for _, match := range idPattern.FindAllStringSubmatch(rr.Body.String(), -1) {
				if id := strings.Trim(match[1], `"`); !seen[id] {
					seen[id] = true
					fuzz.ids = append(fuzz.ids, id)
				}
			}
		}
	}
}

func newEndpointFuzzer(tb testing.TB) *endpointFuzzer {
	config, controlDB, site := testutils.NewControlDBWithSite(tb)
	subdomain := site.SiteDB.Subdomain
	router := server.GetMainHandler(controlDB, config)

	siteJWT, _, err := config.Security.GenerateUserJWTToken(subdomain, consts.FIRST_ADMIN_NAME, string(consts.IdentityRoleAdmin))
	if err != nil {
		tb.Fatalf("failed to make the site's JWT: %v", err)
	}
	fuzz := &endpointFuzzer{router: router, endpoints: collectEndpoints(tb), hosts: map[endpointHost]struct{ host, auth string }{
		nightscoutHost: {subdomain + ".ns.localhost", nightscoutJWT(tb, router, subdomain, site.PwdNSToken)},
		siteHost:       {subdomain + ".localhost", siteJWT},
		controlHost:    {"control.localhost", controlLogin(tb, router)},
	}}
	// The mangled requests are meant to fail, so discard the logs they cause until this target ends
	testutils.UseLogHandler(tb, slog.DiscardHandler)
	fuzz.makeRecords()
	return fuzz
}

// send sends an endpoint mangled by seed, and is why it failed, or empty when it did not
func (fuzz *endpointFuzzer) send(seed int64, modifications int, endpointIndex int) string {
	rr := fuzz.serve(rand.New(rand.NewSource(seed)), endpointIndex, modifications)
	// Not implemented is a deliberate answer, for what the site does not do yet
	if rr.Code >= 500 && rr.Code != http.StatusNotImplemented {
		return fmt.Sprintf("server error %d: %s\nrequest: %s", rr.Code, rr.Body.String(), rr.Header().Get(sentHeader))
	}
	return ""
}

// sentHeader holds what was sent on the recorded response, to show when it fails
const sentHeader = "X-Fuzz-Sent"

// serve sends the endpoint mangled by r and is its response
func (fuzz *endpointFuzzer) serve(r *rand.Rand, endpointIndex int, modifications int) *httptest.ResponseRecorder {
	endpoint := fuzz.endpoints[(endpointIndex%len(fuzz.endpoints)+len(fuzz.endpoints))%len(fuzz.endpoints)]
	req, body := endpoint.mangle(r, modifications, fuzz.ids)
	req.Host = fuzz.hosts[endpoint.host].host
	req.Header.Set("Authorization", "Bearer "+fuzz.auth(r, endpoint.host, modifications))

	rr := httptest.NewRecorder()
	fuzz.router.ServeHTTP(rr, req)
	rr.Header().Set(sentHeader, fmt.Sprintf("%s %s %s?%s %s", req.Method, req.Host, req.URL.RawPath, req.URL.RawQuery, body))
	return rr
}

// auth is the token a request signs in with: its API's admin's, or when mangled, none, nonsense, or another API's
func (fuzz *endpointFuzzer) auth(r *rand.Rand, host endpointHost, modifications int) string {
	if modifications == 0 {
		return fuzz.hosts[host].auth
	}
	switch r.Intn(10) {
	case 0:
		return ""
	case 1:
		return nastyValue(r)
	case 2:
		return fuzz.hosts[endpointHost(r.Intn(3))].auth
	}
	return fuzz.hosts[host].auth
}

// endpointHost is which API an endpoint is on
type endpointHost int

const (
	nightscoutHost endpointHost = iota
	siteHost
	controlHost
)

// endpoint is a request from a .http file: its method, its path split into segments with template segments marked
// by an empty string, its query, its headers and its body
type endpoint struct {
	host     endpointHost
	method   string
	segments []string
	query    url.Values
	headers  http.Header
	body     *string
}

// skippedPaths are signing in and out, which would change who the other requests run as. Making sites is on the
// public site and the control API's site actions are not reads, which are both skipped too, as they could disable
// the test's own site.
var skippedPaths = regexp.MustCompile(`^(/login.*|/logout)$`)

var templated = regexp.MustCompile(`\{\{[^}]*\}\}`)

// collectEndpoints is every distinct method and path of the .http files' requests
func collectEndpoints(tb testing.TB) []*endpoint {
	endpoints := []*endpoint{}
	seen := map[string]bool{}
	err := fs.WalkDir(embeddedFiles, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".http" {
			return err
		}
		requestFile, err := jsonapi.ParseRequestFile(embeddedFiles, path)
		if err != nil {
			return nil
		}
		for _, request := range requestFile.Requests {
			e := parseEndpoint(request)
			if e == nil {
				continue
			}
			key := e.method + " " + fmt.Sprint(e.host) + strings.Join(e.segments, "/")
			if !seen[key] {
				seen[key] = true
				endpoints = append(endpoints, e)
			}
		}
		return nil
	})
	if err != nil {
		tb.Fatalf("failed walking embedded files: %v", err)
	}
	if len(endpoints) == 0 {
		tb.Fatalf("no endpoints found")
	}
	return endpoints
}

// parseEndpoint is a request's endpoint, or nil when it is skipped or not a request
func parseEndpoint(request *jsonapi.Request) *endpoint {
	if request.RunBackgroundJobs || request.Method == "" {
		return nil
	}
	rest, ok := strings.CutPrefix(request.URL, "http://")
	if !ok {
		return nil
	}
	hostPart, pathAndQuery, _ := strings.Cut(rest, "/")
	path, rawQuery, _ := strings.Cut("/"+pathAndQuery, "?")
	if skippedPaths.MatchString(path) {
		return nil
	}
	e := &endpoint{method: request.Method, headers: request.Headers.Clone(), body: request.Body}
	switch {
	case strings.HasPrefix(hostPart, "control."):
		e.host = controlHost
		if e.method != http.MethodGet {
			return nil
		}
	case templated.MatchString(hostPart) && strings.HasPrefix(path, "/api/"):
		e.host = nightscoutHost
	case templated.MatchString(hostPart):
		e.host = siteHost
	default:
		return nil // the public site, which has no data to break
	}
	for _, segment := range strings.Split(strings.TrimPrefix(path, "/"), "/") {
		if templated.MatchString(segment) {
			segment = ""
		}
		e.segments = append(e.segments, segment)
	}
	e.query, _ = url.ParseQuery(templated.ReplaceAllString(rawQuery, "1"))
	e.query.Del("token")
	if e.headers == nil {
		e.headers = http.Header{}
	}
	e.headers.Del("Authorization")
	return e
}

// mangle is the endpoint's request with as many changes as modifications, chosen by r, and the body it sends.
// A template's id is a record's from ids, or nonsense.
func (e *endpoint) mangle(r *rand.Rand, modifications int, ids []string) (*http.Request, string) {
	method := e.method
	if modifications > 0 && r.Intn(8) == 0 {
		method = []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete}[r.Intn(5)]
	}

	segments := slices.Clone(e.segments)
	for i, segment := range segments {
		switch {
		case segment == "" && len(ids) > 0 && r.Intn(2) == 0:
			segments[i] = ids[r.Intn(len(ids))]
		case segment == "" || (modifications > 0 && r.Intn(6) == 0):
			segments[i] = nastyValue(r)
		}
	}

	query := url.Values{}
	for key, values := range e.query {
		query[key] = slices.Clone(values)
	}
	for range modifications {
		switch r.Intn(4) {
		case 0:
			query.Add(queryKey(r), nastyValue(r))
		case 1:
			if keys := slices.Sorted(maps.Keys(query)); len(keys) > 0 {
				query.Set(keys[r.Intn(len(keys))], nastyValue(r))
			}
		}
	}
	rawQuery := query.Encode()
	if modifications > 0 && r.Intn(10) == 0 {
		rawQuery += "&" + []string{"find[[[=1", "%zz=%", "a=%", "find[date][$gte]=%ff", "limit=1&limit=2", "&&&=", "find[]=1"}[r.Intn(7)]
	}

	body, contentType := e.mangleBody(r, modifications)
	req := httptest.NewRequest(method, "/", strings.NewReader(body))
	escaped := make([]string, len(segments))
	for i, segment := range segments {
		escaped[i] = url.PathEscape(segment)
	}
	req.URL.Path = "/" + strings.Join(segments, "/")
	req.URL.RawPath = "/" + strings.Join(escaped, "/")
	req.URL.RawQuery = rawQuery
	req.RequestURI = req.URL.RequestURI()
	for key, values := range e.headers {
		req.Header[key] = slices.Clone(values)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if e.host == siteHost && modifications > 0 {
		switch r.Intn(4) {
		case 0:
			req.Header.Set("Accept", "text/html")
		case 1:
			req.Header.Set("HX-Request", "true")
		}
	}
	return req, body
}

// mangleBody is the endpoint's body, mangled, and the content type to send it as, when it is not the request's own
func (e *endpoint) mangleBody(r *rand.Rand, modifications int) (string, string) {
	body := ""
	if e.body != nil {
		body = templated.ReplaceAllString(*e.body, "1")
	}
	if modifications == 0 {
		return body, ""
	}
	switch r.Intn(10) {
	case 0:
		return "", ""
	case 1:
		return randomString(r), ""
	case 2:
		return `{"date": 1e309, "sgv": "`, "application/json"
	case 3:
		// A form of the body's fields, as the site's forms send
		var fields map[string]any
		json.Unmarshal([]byte(body), &fields)
		form := url.Values{}
		for key, value := range fields {
			form.Set(key, fmt.Sprint(value))
		}
		form.Set(queryKey(r), nastyValue(r))
		return form.Encode(), "application/x-www-form-urlencoded"
	}
	var parsed any
	if json.Unmarshal([]byte(body), &parsed) != nil {
		parsed = map[string]any{}
	}
	if fields, ok := parsed.(map[string]any); ok && len(fields) > 0 && r.Intn(3) == 0 {
		// One field extreme in an otherwise good body, so it gets past the checks of the others to be stored
		key := slices.Sorted(maps.Keys(fields))[r.Intn(len(fields))]
		fields[key] = extremeValue(r)
		marshalled, _ := json.Marshal(fields)
		return string(marshalled), ""
	}
	switch value := parsed.(type) {
	case map[string]any:
		mutated := mutateBody(value, r, modifications)
		if r.Intn(4) == 0 {
			// Many of them, as the Nightscout API takes lists
			marshalled, _ := json.Marshal([]any{mutated, mutateBody(value, r, modifications), randomJSONValue(r)})
			return string(marshalled), ""
		}
		marshalled, _ := json.Marshal(mutated)
		return string(marshalled), ""
	default:
		mutated, _ := mutateJSONValue(value, r, modifications)
		marshalled, _ := json.Marshal(mutated)
		return string(marshalled), ""
	}
}

// queryFields and queryOps make the filters the Nightscout API reads, as v1's find[field][$op] and v3's field$op
var (
	queryFields = []string{"date", "created_at", "dateString", "sgv", "type", "eventType", "device", "carbs", "insulin", "duration", "_id", "identifier", "srvModified", "srvCreated", "mills", "app", "category", "activityType", "name", "enteredBy", "notes", "units"}
	queryOps    = []string{"", "$eq", "$ne", "$gt", "$gte", "$lt", "$lte", "$in", "$nin", "$re", "$regex", "$exists", "$where"}
	queryKeys   = []string{"limit", "skip", "count", "sort", "sort$desc", "fields", "hours", "order", "search", "type", "device", "timezone", "language", "envelope", "now", "since", "lastModified", "find", "find[]", "token", "secret"}
)

func queryKey(r *rand.Rand) string {
	field, op := queryFields[r.Intn(len(queryFields))], queryOps[r.Intn(len(queryOps))]
	switch r.Intn(4) {
	case 0:
		return queryKeys[r.Intn(len(queryKeys))]
	case 1:
		if op == "" {
			return "find[" + field + "]"
		}
		return "find[" + field + "][" + op + "]"
	case 2:
		return field + op
	}
	return randomKey(r)
}

// nastyValues are values most likely to break an id, a number, a date, a filter or a sort
var nastyValues = []string{
	"", " ", "0", "-1", "1", "-0", "1.5", "1e309", "-1e309", "1e-400", "NaN", "Infinity", "-Infinity",
	"9223372036854775807", "9223372036854775808", "-9223372036854775808", "-9223372036854775809", "99999999999999999999999999",
	"0x10", "007", "+1", "1,2", "true", "null", "undefined", "[]", "{}", `{"$gt":""}`, "[1,2]",
	"2026-10-29", "2026-13-45T25:61:61Z", "0000-00-00", "1970-01-01T00:00:00Z", "9999-12-31T23:59:59Z", "-292277022657-01-27T08:29:52Z",
	"%", "%00", "\x00", "..", "../../etc/passwd", "/", "\\", `"`, "'", "' OR '1'='1", "; DROP TABLE entries;--",
	"(", ")", "[", "]", "*", ".*", "(a+)+$", "\\", "$where", "$gt", "sgv", "_id", "-date", "date,sgv", "a.b.c",
	"😀", "日本語", "‮", strings.Repeat("a", 300), strings.Repeat("9", 400),
	"123e4567-e89b-12d3-a456-426614174000", "5f8d0d55b54764421b7156c9", "ffffffffffffffffffffffff",
}

// extremeValue is a JSON value at or past the edge of what a field takes: too long, too big, too small or empty
func extremeValue(r *rand.Rand) any {
	return []any{
		strings.Repeat("x", 5000), strings.Repeat("😀", 2000), "", " ", "\x00", nastyValue(r),
		1e19, -1e19, 1e300, -1e300, 9223372036854775807, -9223372036854775808, 0, -1, 0.5, -0.5,
		[]any{}, map[string]any{}, nil, true, []any{strings.Repeat("x", 5000)}, map[string]any{"$gt": ""},
	}[r.Intn(22)]
}

func nastyValue(r *rand.Rand) string {
	if r.Intn(5) == 0 {
		return randomString(r)
	}
	return nastyValues[r.Intn(len(nastyValues))]
}

// nightscoutJWT is the JWT the Nightscout API gives for the site's PWD phone token, as its clients get one
func nightscoutJWT(tb testing.TB, router http.Handler, subdomain, token string) string {
	req := httptest.NewRequest(http.MethodGet, "http://"+subdomain+".ns.localhost/api/v2/authorization/request/"+token, nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	var auth struct {
		Token string `json:"token"`
	}
	if rr.Code != http.StatusOK || json.Unmarshal(rr.Body.Bytes(), &auth) != nil || auth.Token == "" {
		tb.Fatalf("Nightscout authorization failed: %d %s", rr.Code, rr.Body.String())
	}
	return auth.Token
}

// controlLogin signs in as the control API's admin, as the .http files do, and is its JWT
func controlLogin(tb testing.TB, router http.Handler) string {
	req := httptest.NewRequest(http.MethodPost, "http://control.localhost/login", strings.NewReader(`{"name": "admin", "password": "password", "totp": "123123"}`))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	var login struct {
		Token string `json:"token"`
	}
	if rr.Code != http.StatusOK || json.Unmarshal(rr.Body.Bytes(), &login) != nil || login.Token == "" {
		tb.Fatalf("control login failed: %d %s", rr.Code, rr.Body.String())
	}
	return login.Token
}
