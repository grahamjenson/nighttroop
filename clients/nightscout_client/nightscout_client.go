package nightscout_client

// Package nightscout provides a small Nightscout API v1 client.
//
// Supports:
//   - Fetching entries
//   - Fetching treatments
//
// Compatible with:
//   /api/v1/entries.json
//   /api/v1/treatments.json

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"nighttroop.t1d.nz/utils/httputils"
)

// TestValidURL is the only Nightscout URL that authenticates under go test, which
// never makes HTTP calls; any other URL fails as if the host does not exist.
const TestValidURL = "http://valid.nightscout.test"

// TestLastModified is when TestValidURL says each collection it has was last modified
const TestLastModified int64 = 1_700_000_000_000

// TestDocuments are the documents TestValidURL returns for each collection; tests set them
var TestDocuments = map[string][]any{}

const (
	maxResponseBytes = 1 << 20 // a Nightscout response body larger than this is refused
	maxResultItems   = 1000    // the sync only ever asks for a handful of documents
	// A JWT this close to expiring is refreshed first, so a kept client never sends an expired one
	jwtRefreshMargin = time.Minute
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client

	jwt       *string
	jwtExpire *time.Time
}

func documentsEndpoint(nsURL string, collection string, gteDate int64, lteDate int64, limit int) string {
	endpoint := fmt.Sprintf("%s/api/v3/%s?sort$desc=date&date$gte=%d&limit=%d", nsURL, collection, gteDate, limit)
	if lteDate > 0 {
		endpoint += fmt.Sprintf("&date$lte=%d", lteDate)
	}
	return endpoint
}
func lastModifiedEndpoint(nsURL string) string {
	return fmt.Sprintf("%s/api/v3/lastModified", nsURL)
}
func jwtEndpoint(nsURL, token string) string {
	return fmt.Sprintf("%s/api/v2/authorization/request/token=%s", nsURL, token)
}

func NewClient(baseURL, token string) *Client {
	baseURL = strings.TrimRight(baseURL, "/")

	return &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			// Never follow redirects: the sync URL is only checked when saved, so a
			// redirect could point at an internal host. The 3xx is handled as a failure.
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		},
	}
}

func (c *Client) doRequest(
	ctx context.Context,
	method string,
	url string,
) (map[string]any, error) {
	if c.jwt == nil || c.jwtExpire == nil || time.Now().Add(jwtRefreshMargin).After(*c.jwtExpire) {
		err := c.refreshJWT()
		if err != nil {
			return nil, fmt.Errorf("error refreshing JWT: %w", err)
		}
	}
	// Only TestValidURL is faked: a test that sets a JWT and an httptest URL goes over HTTP.
	if testing.Testing() && c.baseURL == TestValidURL {
		if url == lastModifiedEndpoint(c.baseURL) {
			return map[string]any{"result": map[string]any{"collections": map[string]any{"entries": float64(TestLastModified), "treatments": float64(TestLastModified), "devicestatus": float64(TestLastModified)}}}, nil
		}
		collection, _, _ := strings.Cut(strings.TrimPrefix(url, c.baseURL+"/api/v3/"), "?")
		return map[string]any{"result": append([]any{}, TestDocuments[collection]...)}, nil
	}

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}
	// Just in case it is the API secret not a token
	req.Header.Set("api-secret", c.token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", *c.jwt))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Reset if unauthorized
	if resp.StatusCode == http.StatusUnauthorized {
		c.jwt = nil
		c.jwtExpire = nil
		return nil, fmt.Errorf(
			"nightscout request unauthorized: %s",
			resp.Status,
		)
	}

	// Handle other non-OK status codes
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"nightscout request failed: %s",
			resp.Status,
		)
	}

	return parseResult(resp.Body)
}

// parseResult decodes a Nightscout response body of at most maxResponseBytes into a map.
func parseResult(body io.Reader) (map[string]any, error) {
	data, err := httputils.ReadLimitedBody(body, maxResponseBytes)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

type JWTAuthResponse struct {
	Token            string     `json:"token"`
	Sub              string     `json:"sub"`
	PermissionGroups [][]string `json:"permissionGroups"`
	IAT              int64      `json:"iat"`
	Exp              int64      `json:"exp"`
}

var (
	ErrJWTRequest         = errors.New("JWT request failed")
	ErrJWTNoHost          = errors.New("JWT host unavailable")
	ErrJWTTimeout         = errors.New("JWT request timed out")
	ErrJWTServer          = errors.New("JWT server returned an error")
	ErrJWTInvalidResponse = errors.New("invalid JWT response")
)

// Verify checks the URL and token work by requesting a JWT.
func (c *Client) Verify() error {
	return c.refreshJWT()
}

func (c *Client) refreshJWT() error {
	if testing.Testing() {
		if c.baseURL != TestValidURL {
			return ErrJWTNoHost
		}
		jwt, expire := "test-jwt", time.Now().Add(time.Hour)
		c.jwt, c.jwtExpire = &jwt, &expire
		return nil
	}
	resp, err := c.httpClient.Get(jwtEndpoint(c.baseURL, c.token))
	if err != nil {
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return ErrJWTTimeout
		}

		var dnsErr *net.DNSError
		if errors.As(err, &dnsErr) {
			return ErrJWTNoHost
		}

		return ErrJWTRequest
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%w: HTTP %d", ErrJWTServer, resp.StatusCode)
	}

	body, err := httputils.ReadLimitedBody(resp.Body, maxResponseBytes)
	if err != nil {
		return ErrJWTInvalidResponse
	}
	var jwtResp JWTAuthResponse
	if err := json.Unmarshal(body, &jwtResp); err != nil {
		return ErrJWTInvalidResponse
	}

	c.jwt = &jwtResp.Token
	expire := time.Unix(jwtResp.Exp, 0)
	c.jwtExpire = &expire

	return nil
}

//
// LAST MODIFIED
//

// GetLastModified returns when each collection was last modified on the Nightscout, in
// milliseconds, so a sync only reads the collections that have changed since it last did.
func (c *Client) GetLastModified(ctx context.Context) (map[string]int64, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, lastModifiedEndpoint(c.baseURL))
	if err != nil {
		return nil, fmt.Errorf("NSClient Get Last Modified Error %s: %w", c.baseURL, err)
	}
	result, _ := resp["result"].(map[string]any)
	collections, ok := result["collections"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("NSClient Get Last Modified Error: %s returned no collections", c.baseURL)
	}
	lastModified := map[string]int64{}
	for collection, modified := range collections {
		if ms, ok := modified.(float64); ok {
			lastModified[collection] = int64(ms)
		}
	}
	return lastModified, nil
}

//
// DOCUMENTS
//

// GetDocuments returns up to count of a collection's newest documents, such as entries or treatments, dated from
// afterTime, and up to beforeTime unless it is zero
func (c *Client) GetDocuments(
	ctx context.Context,
	collection string,
	afterTime time.Time,
	beforeTime time.Time,
	count int,
) ([]map[string]any, error) {
	before := int64(0)
	if !beforeTime.IsZero() {
		before = beforeTime.UnixMilli()
	}
	reqURL := documentsEndpoint(c.baseURL, collection, afterTime.UnixMilli(), before, count)
	resp, err := c.doRequest(ctx, http.MethodGet, reqURL)

	if err != nil {
		return nil, fmt.Errorf("NSClient Get %s Error %s: %w", collection, c.baseURL, err)
	}

	if result, ok := resp["result"].([]any); ok {
		return documentsFromResult(result)
	}

	// The body is never put in the error: it is stored and shown as a sync error.
	return nil, fmt.Errorf("NSClient Get %s Error: %s returned no result array", collection, c.baseURL)

}

// documentsFromResult converts each item of the result array to a map[string]any,
// refusing more than maxResultItems.
func documentsFromResult(result []any) ([]map[string]any, error) {
	if len(result) > maxResultItems {
		return nil, fmt.Errorf("NSClient Get Documents Error: %d results is more than %d", len(result), maxResultItems)
	}
	converted := []map[string]any{}
	for _, item := range result {
		if m, ok := item.(map[string]any); ok {
			converted = append(converted, m)
		}
	}
	return converted, nil
}
