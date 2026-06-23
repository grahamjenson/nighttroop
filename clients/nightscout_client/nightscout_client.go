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
	"net/url"
	"strings"
	"time"

	"nighttroop.t1d.nz/nightscout"
)

type Client struct {
	baseURL    string
	apiSecret  string
	httpClient *http.Client
}

func NewClient(baseURL, apiSecret string) *Client {
	baseURL = strings.TrimRight(baseURL, "/")

	return &Client{
		baseURL:   baseURL,
		apiSecret: apiSecret,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) newRequest(
	ctx context.Context,
	method string,
	path string,
	query url.Values,
) (*http.Request, error) {
	u := c.baseURL + path

	if query != nil {
		u += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, u, nil)
	if err != nil {
		return nil, err
	}

	if c.apiSecret != "" {
		req.Header.Set("api-secret", c.apiSecret)
	}

	req.Header.Set("Accept", "application/json")

	return req, nil
}

//
// ENTRIES
//

func (c *Client) GetEntries(
	ctx context.Context,
	count int,
) ([]nightscout.Entry, error) {
	q := url.Values{}
	q.Set("count", fmt.Sprintf("%d", count))

	req, err := c.newRequest(
		ctx,
		http.MethodGet,
		"/api/v1/entries.json",
		q,
	)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"nightscout entries failed: %s",
			resp.Status,
		)
	}

	var items []json.RawMessage

	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, err
	}

	var entries []nightscout.Entry
	for _, raw := range items {
		var entry nightscout.Entry
		if err := json.Unmarshal(raw, &entry); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

//
// TREATMENTS
//

type Treatment struct {
	ID        string     `json:"_id"`
	EventType string     `json:"eventType"`
	CreatedAt *time.Time `json:"created_at"`
	EnteredBy string     `json:"enteredBy"`

	Insulin float64 `json:"insulin"`
	Carbs   float64 `json:"carbs"`

	Absolute float64 `json:"absolute"`
	Rate     float64 `json:"rate"`
	Duration float64 `json:"duration"`

	Notes   string  `json:"notes"`
	Glucose float64 `json:"glucose"`
	Units   string  `json:"units"`
}

func (c *Client) GetTreatments(
	ctx context.Context,
	count int,
) ([]Treatment, error) {
	q := url.Values{}
	q.Set("count", fmt.Sprintf("%d", count))

	req, err := c.newRequest(
		ctx,
		http.MethodGet,
		"/api/v1/treatments.json",
		q,
	)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"nightscout treatments failed: %s",
			resp.Status,
		)
	}

	var treatments []Treatment

	if err := json.NewDecoder(resp.Body).Decode(&treatments); err != nil {
		return nil, err
	}

	return treatments, nil
}
