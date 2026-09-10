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
	"fmt"
	"net/http"
	"strings"
	"time"
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

	// Parse into a map
	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
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

func (c *Client) refreshJWT() error {

	resp, err := c.httpClient.Get(jwtEndpoint(c.baseURL, c.token))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var jwtResp JWTAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&jwtResp); err != nil {
		return err
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
		return nil, fmt.Errorf("NSClient Error: %s: %w", reqURL, err)
	}

	if result, ok := resp["result"].([]any); ok {
		// Convert each item to a map[string]any
		converted := []map[string]any{}
		for _, item := range result {
			if m, ok := item.(map[string]any); ok {
				converted = append(converted, m)
			}
		}
		return converted, nil
	}

	return nil, fmt.Errorf("unexpected response format for URL: %s %s", reqURL, resp)

}
