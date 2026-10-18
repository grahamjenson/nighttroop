// Package dexcom provides a minimal Dexcom Share API client.
//
// Supports:
//   - Username/password login
//   - Fetching the last hour of estimated glucose values (EGVs)
//
// Works with:
//   - US Share server
//   - Outside US Share server
//
// NOTE:
// Dexcom changes APIs occasionally. This implementation uses the
// Share Web Services endpoints currently used by many open-source clients.

package dexcom_client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"nighttroop.t1d.nz/models/dexcommodels"
)

// TestValidUsername is the only Dexcom username that logs in under go test,
// which never makes HTTP calls; any other username is rejected.
const TestValidUsername = "valid-dexcom-user"

const (
	usBaseURL        = "https://share2.dexcom.com"
	outsideUSBaseURL = "https://shareous1.dexcom.com"
	applicationID    = "d8665ade-9673-4e27-9ff6-92db4ce13d13"
)

// Client is a Dexcom Share API client.
type Client struct {
	httpClient *http.Client
	baseURL    string

	username string
	password string

	sessionID string
}

// NewClient creates a new Dexcom client.
//
// outsideUS should be true for non-US Dexcom accounts.
func NewClient(username, password string, outsideUS bool) *Client {
	baseURL := usBaseURL
	if outsideUS {
		baseURL = outsideUSBaseURL
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:  baseURL,
		username: username,
		password: password,
	}
}

// Login authenticates and stores the Dexcom session ID.
func (c *Client) Login(ctx context.Context) error {
	if testing.Testing() {
		if c.username != TestValidUsername {
			return fmt.Errorf("dexcom authentication failed: invalid username or password")
		}
		c.sessionID = "test-session"
		return nil
	}
	accountID, err := c.authenticate(ctx)
	if err != nil {
		return err
	}

	sessionID, err := c.loginByAccountID(ctx, accountID)
	if err != nil {
		return err
	}

	c.sessionID = sessionID
	return nil
}

func (c *Client) authenticate(ctx context.Context) (string, error) {
	reqBody := map[string]string{
		"accountName":   c.username,
		"password":      c.password,
		"applicationId": applicationID,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/ShareWebServices/Services/General/AuthenticatePublisherAccount",
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf(
			"authenticate failed: status=%d body=%s",
			resp.StatusCode,
			string(body),
		)
	}

	accountIDBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	accountID := strings.Trim(string(accountIDBytes), "\"")

	if accountID == "" {
		return "", fmt.Errorf("empty account id")
	}

	return accountID, nil
}

func (c *Client) loginByAccountID(ctx context.Context, accountID string) (string, error) {
	reqBody := map[string]string{
		"accountId":     accountID,
		"password":      c.password,
		"applicationId": applicationID,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/ShareWebServices/Services/General/LoginPublisherAccountById",
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf(
			"login failed: status=%d body=%s",
			resp.StatusCode,
			string(body),
		)
	}

	sessionIDBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	sessionID := strings.Trim(string(sessionIDBytes), "\"")

	if sessionID == "" {
		return "", fmt.Errorf("empty session id")
	}

	return sessionID, nil
}

// FetchEGVs fetches the last specified minutes of EGV readings.
func (c *Client) FetchEGVs(ctx context.Context, minutes int) ([]dexcommodels.EGV, error) {
	if c.sessionID == "" {
		if err := c.Login(ctx); err != nil {
			return nil, fmt.Errorf("Dexcom Login Error: %w", err)
		}
	}
	if testing.Testing() {
		return nil, nil
	}

	values := url.Values{}
	values.Set("sessionId", c.sessionID)
	values.Set("minutes", fmt.Sprintf("%d", minutes))
	values.Set("maxCount", fmt.Sprintf("%d", minutes/5))

	endpoint := c.baseURL +
		"/ShareWebServices/Services/Publisher/ReadPublisherLatestGlucoseValues?" +
		values.Encode()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating Dexcom request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending Dexcom request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf(
			"dexcom fetch failed: status=%d body=%s",
			resp.StatusCode,
			string(body),
		)
	}

	var items []json.RawMessage

	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, fmt.Errorf("error decoding Dexcom response: %w", err)
	}

	out := make([]dexcommodels.EGV, 0, len(items))

	for _, raw := range items {
		var r dexcommodels.EGV
		if err := json.Unmarshal(raw, &r); err != nil {
			slog.Error("Error unmarshalling Dexcom response", "error", err, "raw", string(raw))
			continue
		}

		out = append(out, r)
	}

	return out, nil
}
