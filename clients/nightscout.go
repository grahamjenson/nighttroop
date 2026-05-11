// Nightscout client
package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

// -----------------------------
// Types
// -----------------------------

// Generic raw JSON results returned for entries/treatments etc.
// Caller may unmarshal into concrete structs as needed.
type RawJSON = json.RawMessage

type NSEntries RawJSON
type NSTreatments RawJSON
type NSProfile struct {
	// minimal - expand to match NightScout fields you care about
	Name string `json:"name,omitempty"`
}
type NSProperties RawJSON
type NSDeviceStatus RawJSON

type jwtResponse struct {
	Token string `json:"token"`
	Exp   int64  `json:"exp"` // seconds since epoch
	// other fields ignored
}

// NSClient is the Go equivalent of the Dart singleton client.
// Use NewClient(...) to create an instance. It's safe to reuse.
type NSClient struct {
	BaseURL string
	Token   string

	// JWT state
	jwt       string
	jwtExpire time.Time
	jwtURL    string
	jwtToken  string

	http *http.Client

	// metrics/state
	totalErrors int64 // atomic
	lastFetched time.Time
}

// NewClient returns a configured NSClient. If httpClient is nil, a default with 10s timeout is used.
func NewClient(baseURL, token string, httpClient *http.Client) *NSClient {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	// Ensure baseURL has no trailing slash for consistent endpoint building
	base := strings.TrimRight(baseURL, "/")
	return &NSClient{
		BaseURL: base,
		Token:   token,
		http:    httpClient,
	}
}

// -----------------------------
// Helpers: endpoints + formatting
// -----------------------------

func (c *NSClient) testEndpoint() string {
	// Dart: '$nsURL/api/v2/authorization/request/token=$token'
	return fmt.Sprintf("%s/api/v2/authorization/request/token=%s", c.BaseURL, url.QueryEscape(c.Token))
}

func (c *NSClient) entryEndpoint(gteCreatedAt time.Time) string {
	// Dart uses millisecondsSinceEpoch query param
	ms := strconv.FormatInt(gteCreatedAt.UnixNano()/1e6, 10)
	return fmt.Sprintf("%s/api/v1/entries.json?token=%s&find[date][$gte]=%s&count=1000",
		c.BaseURL, url.QueryEscape(c.Token), ms)
}

func toNSDateString(dt time.Time) string {
	// Dart formatted YYYY-MM-DDTHH:MM (no seconds)
	utc := dt.UTC()
	return fmt.Sprintf("%04d-%02d-%02dT%02d:%02d",
		utc.Year(), utc.Month(), utc.Day(), utc.Hour(), utc.Minute())
}

func (c *NSClient) treatmentsEndpoint(gteCreatedAt time.Time) string {
	qs := toNSDateString(gteCreatedAt)
	return fmt.Sprintf("%s/api/v1/treatments.json?token=%s&find[created_at][$gte]=%s&count=1000",
		c.BaseURL, url.QueryEscape(c.Token), url.QueryEscape(qs))
}

func (c *NSClient) treatmentsProfileSwitch(count int) string {
	return fmt.Sprintf("%s/api/v1/treatments.json?token=%s&find[eventType]=Profile+Switch&count=%d",
		c.BaseURL, url.QueryEscape(c.Token), count)
}

func (c *NSClient) treatmentsNote(count int) string {
	return fmt.Sprintf("%s/api/v1/treatments.json?token=%s&find[eventType]=Note&count=%d",
		c.BaseURL, url.QueryEscape(c.Token), count)
}

func (c *NSClient) currentProfileEndpoint() string {
	return fmt.Sprintf("%s/api/v1/profile/current.json?token=%s", c.BaseURL, url.QueryEscape(c.Token))
}

func (c *NSClient) propertiesEndpoint() string {
	return fmt.Sprintf("%s/api/v2/properties.json?token=%s", c.BaseURL, url.QueryEscape(c.Token))
}

func (c *NSClient) deviceStatusEndpoint() string {
	return fmt.Sprintf("%s/api/v1/devicestatus.json?token=%s&count=1", c.BaseURL, url.QueryEscape(c.Token))
}

// v3 treatments endpoint (JWT used in Authorization header)
func (c *NSClient) v3TreatmentsEndpoint() string {
	return fmt.Sprintf("%s/api/v3/treatments", c.BaseURL)
}

func (c *NSClient) jwtEndpoint() string {
	return fmt.Sprintf("%s/api/v2/authorization/request/token=%s", c.BaseURL, url.QueryEscape(c.Token))
}

// -----------------------------
// Internal utility methods
// -----------------------------

func (c *NSClient) incTotalErrors() {
	atomic.AddInt64(&c.totalErrors, 1)
}

func (c *NSClient) setLastFetched(t time.Time) {
	c.lastFetched = t
}

// doRequest wraps HTTP calls and returns response body (and error).
// It does NOT close resp.Body for you.
func (c *NSClient) doRequest(ctx context.Context, method, urlStr string, headers map[string]string, body []byte) (*http.Response, []byte, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, urlStr, bodyReader)
	if err != nil {
		c.incTotalErrors()
		return nil, nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		c.incTotalErrors()
		return nil, nil, err
	}
	respBody, err := io.ReadAll(resp.Body)
	// always close here
	resp.Body.Close()
	if err != nil {
		c.incTotalErrors()
		return resp, nil, err
	}
	// if server error code
	if resp.StatusCode == 500 {
		c.incTotalErrors()
		return resp, respBody, fmt.Errorf("server error 500: %s", string(respBody))
	}
	// non 200/201 -> error
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		c.incTotalErrors()
		// attempt to parse message/description from JSON
		var j map[string]any
		_ = json.Unmarshal(respBody, &j)
		var msg string
		if j != nil {
			if m, ok := j["message"].(string); ok {
				if d, ok := j["description"].(string); ok {
					msg = fmt.Sprintf("%s - %s", m, d)
				} else {
					msg = m
				}
			}
		}
		if msg == "" {
			msg = string(respBody)
		}
		return resp, respBody, fmt.Errorf("non-200 response %d: %s", resp.StatusCode, msg)
	}
	// success
	c.setLastFetched(time.Now())
	return resp, respBody, nil
}

// -----------------------------
// Public API methods
// -----------------------------

// TestURL checks the NS URL and token quickly (3s timeout) — returns nil on success.
func (c *NSClient) TestURL(ctx context.Context) error {
	if c.BaseURL == "" || len(c.BaseURL) < 8 {
		return errors.New("NightScout URL not valid")
	}
	if c.Token == "" || len(c.Token) < 10 {
		return errors.New("NightScout Token not valid")
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	_, _, err := c.doRequest(ctx, http.MethodGet, c.testEndpoint(), nil, nil)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || strings.Contains(err.Error(), "Client.Timeout") {
			return fmt.Errorf("timeout: could not connect to NightScout in 3 seconds: %w", err)
		}
		return fmt.Errorf("error: %w", err)
	}
	return nil
}

// ShouldRefreshJWT returns true if JWT is missing, expired, or URL/token changed.
func (c *NSClient) ShouldRefreshJWT() bool {
	if c.jwt == "" || c.jwtExpire.IsZero() {
		return true
	}
	if c.jwtURL != c.BaseURL || c.jwtToken != c.Token {
		return true
	}
	// if jwtExpire before now
	if time.Now().After(c.jwtExpire) {
		return true
	}
	return false
}

// RefreshJWT refreshes the JWT and sets expiry (with 5 minute wiggle).
func (c *NSClient) RefreshJWT(ctx context.Context) error {
	if c.BaseURL == "" || c.Token == "" {
		return errors.New("no NS URL or token")
	}
	if !c.ShouldRefreshJWT() {
		return nil
	}
	ep := c.jwtEndpoint()
	_, body, err := c.doRequest(ctx, http.MethodGet, ep, nil, nil)
	if err != nil {
		return fmt.Errorf("refresh jwt request failed: %w", err)
	}
	var jr jwtResponse
	if err := json.Unmarshal(body, &jr); err != nil {
		return fmt.Errorf("failed to parse jwt response: %w", err)
	}
	if jr.Token == "" || jr.Exp == 0 {
		return errors.New("invalid jwt response")
	}
	// subtract 300 seconds (5 min) wiggle
	expiry := time.Unix(jr.Exp, 0).Add(-300 * time.Second)
	c.jwt = jr.Token
	c.jwtExpire = expiry
	c.jwtURL = c.BaseURL
	c.jwtToken = c.Token
	return nil
}

// GetEntries returns raw JSON for entries since gteCreateAt
func (c *NSClient) GetEntries(ctx context.Context, gteCreateAt time.Time) (NSEntries, error) {
	if c.BaseURL == "" || c.Token == "" {
		return nil, errors.New("no NS URL or token")
	}
	ep := c.entryEndpoint(gteCreateAt)
	_, body, err := c.doRequest(ctx, http.MethodGet, ep, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("get entries failed: %w", err)
	}
	return NSEntries(body), nil
}

// GetTreatments returns raw JSON
func (c *NSClient) GetTreatments(ctx context.Context, gteCreateAt time.Time) (NSTreatments, error) {
	if c.BaseURL == "" || c.Token == "" {
		return nil, errors.New("no NS URL or token")
	}
	ep := c.treatmentsEndpoint(gteCreateAt)
	_, body, err := c.doRequest(ctx, http.MethodGet, ep, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("get treatments failed: %w", err)
	}
	return NSTreatments(body), nil
}

// GetProfileSwitches
func (c *NSClient) GetProfileSwitches(ctx context.Context, count int) (NSTreatments, error) {
	if c.BaseURL == "" || c.Token == "" {
		return nil, errors.New("no NS URL or token")
	}
	ep := c.treatmentsProfileSwitch(count)
	_, body, err := c.doRequest(ctx, http.MethodGet, ep, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("get profile switches failed: %w", err)
	}
	return NSTreatments(body), nil
}

// GetNotes
func (c *NSClient) GetNotes(ctx context.Context, count int) (NSTreatments, error) {
	if c.BaseURL == "" || c.Token == "" {
		return nil, errors.New("no NS URL or token")
	}
	ep := c.treatmentsNote(count)
	_, body, err := c.doRequest(ctx, http.MethodGet, ep, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("get notes failed: %w", err)
	}
	return NSTreatments(body), nil
}

// GetCurrentProfiles returns the current profile and list of other profiles and the raw body
// The shape of response varies; we return raw JSON in addition to attempting to parse current profile.
func (c *NSClient) GetCurrentProfiles(ctx context.Context) (*NSProfile, []NSProfile, RawJSON, error) {
	if c.BaseURL == "" || c.Token == "" {
		return nil, nil, nil, errors.New("no NS URL or token")
	}
	ep := c.currentProfileEndpoint()
	_, body, err := c.doRequest(ctx, http.MethodGet, ep, nil, nil)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get current profiles failed: %w", err)
	}
	// Attempt to parse a minimal structure; keep raw body as well
	var parsed struct {
		Current  *NSProfile  `json:"current,omitempty"`
		Profiles []NSProfile `json:"profiles,omitempty"`
	}
	_ = json.Unmarshal(body, &parsed)
	return parsed.Current, parsed.Profiles, RawJSON(body), nil
}

// GetProperties
func (c *NSClient) GetProperties(ctx context.Context) (*NSProperties, error) {
	if c.BaseURL == "" || c.Token == "" {
		return nil, errors.New("no NS URL or token")
	}
	ep := c.propertiesEndpoint()
	_, body, err := c.doRequest(ctx, http.MethodGet, ep, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("get properties failed: %w", err)
	}
	r := NSProperties(body)
	return &r, nil
}

// GetDeviceStatus
func (c *NSClient) GetDeviceStatus(ctx context.Context) (*NSDeviceStatus, error) {
	if c.BaseURL == "" || c.Token == "" {
		return nil, errors.New("no NS URL or token")
	}
	ep := c.deviceStatusEndpoint()
	_, body, err := c.doRequest(ctx, http.MethodGet, ep, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("get device status failed: %w", err)
	}
	r := NSDeviceStatus(body)
	return &r, nil
}

// -----------------------------
// Diagnostics getters
// -----------------------------

func (c *NSClient) TotalErrors() int64 {
	return atomic.LoadInt64(&c.totalErrors)
}

func (c *NSClient) LastFetched() time.Time {
	return c.lastFetched
}
