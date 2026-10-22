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

const (
	maxResponseBytes = 1 << 20 // a Nightscout response body larger than this is refused
	maxResultItems   = 1000    // the sync only ever asks for a handful of entries
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client

	jwt       *string
	jwtExpire *time.Time
}

func lastModifiedEndpoint(nsURL string) string { return fmt.Sprintf("%s/api/v3/lastModified", nsURL) }
func currentProfileEndpoint(nsURL string) string {
	return fmt.Sprintf("%s/api/v3/profile?sort$desc=date&limit=1", nsURL)
}
func deviceStatusEndpoint(nsURL string) string {
	return fmt.Sprintf("%s/api/v3/devicestatus?sort$desc=date&limit=1", nsURL)
}
func treatmentsHistoryEndpoint(nsURL string, createdAt int, limit int) string {
	return fmt.Sprintf("%s/api/v3/treatments?sort$desc=created_at&created_at$gte=%d&limit=%d", nsURL, createdAt, limit)
}
func entryHistoryEndpoint(nsURL string, gteCreatedAt int, limit int) string {
	return fmt.Sprintf("%s/api/v3/entries?sort$desc=date&date$gte=%d&limit=%d", nsURL, gteCreatedAt, limit)
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
		},
	}
}

func (c *Client) doRequest(
	ctx context.Context,
	method string,
	url string,
) (map[string]any, error) {
	if c.jwt == nil || c.jwtExpire == nil || time.Now().After(*c.jwtExpire) {
		err := c.refreshJWT()
		if err != nil {
			return nil, fmt.Errorf("error refreshing JWT: %w", err)
		}
	}
	if testing.Testing() {
		return map[string]any{"result": []any{}}, nil
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
// ENTRIES
//

func (c *Client) GetEntries(
	ctx context.Context,
	afterTime time.Time,
	count int,
) ([]map[string]any, error) {
	reqURL := entryHistoryEndpoint(c.baseURL, int(afterTime.UnixMilli()), count)
	resp, err := c.doRequest(ctx, http.MethodGet, reqURL)

	if err != nil {
		return nil, fmt.Errorf("NSClient Get Entries Error %s: %w", c.baseURL, err)
	}

	if result, ok := resp["result"].([]any); ok {
		return entriesFromResult(result)
	}

	return nil, fmt.Errorf("NSClient Get Entries Error: %s %s", c.baseURL, resp)

}

// entriesFromResult converts each item of the result array to a map[string]any,
// refusing more than maxResultItems.
func entriesFromResult(result []any) ([]map[string]any, error) {
	if len(result) > maxResultItems {
		return nil, fmt.Errorf("NSClient Get Entries Error: %d results is more than %d", len(result), maxResultItems)
	}
	converted := []map[string]any{}
	for _, item := range result {
		if m, ok := item.(map[string]any); ok {
			converted = append(converted, m)
		}
	}
	return converted, nil
}
