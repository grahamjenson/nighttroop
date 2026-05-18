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
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

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

// EGV represents a Dexcom glucose reading.
type EGV struct {
	Value         int
	Trend         string
	TrendArrow    string
	SystemTime    time.Time
	DisplayTime   time.Time
	RealtimeValue int
}

// FetchLastHourEGVs fetches the last hour of EGV readings.
func (c *Client) FetchLastHourEGVs(ctx context.Context) ([]EGV, error) {
	if c.sessionID == "" {
		if err := c.Login(ctx); err != nil {
			return nil, fmt.Errorf("Dexcom Login Error")
		}
	}

	values := url.Values{}
	values.Set("sessionId", c.sessionID)
	values.Set("minutes", "60")
	values.Set("maxCount", "12")

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

	var raw []dexcomEGVResponse

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("error decoding Dexcom response: %w %s", err, resp.Body)
	}

	out := make([]EGV, 0, len(raw))

	for _, r := range raw {
		systemTime, _ := parseDexcomTime(r.ST)
		displayTime, _ := parseDexcomTime(r.DT)

		out = append(out, EGV{
			Value:         r.Value,
			Trend:         r.Trend,
			TrendArrow:    trendArrow(r.Trend),
			SystemTime:    systemTime,
			DisplayTime:   displayTime,
			RealtimeValue: r.RealtimeValue,
		})
	}

	return out, nil
}

type dexcomEGVResponse struct {
	WT            string `json:"WT"`
	ST            string `json:"ST"`
	DT            string `json:"DT"`
	Value         int    `json:"Value"`
	Trend         string `json:"Trend"`
	RealtimeValue int    `json:"RealtimeValue"`
}

// parseDexcomTime parses Dexcom's weird /Date(1234567890)/ format.
func parseDexcomTime(s string) (time.Time, error) {
	s = strings.TrimPrefix(s, "/Date(")
	s = strings.TrimSuffix(s, ")/")

	if idx := strings.IndexAny(s, "+-"); idx != -1 {
		s = s[:idx]
	}

	ms, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	return time.UnixMilli(ms), nil
}

func trendArrow(trend string) string {
	switch trend {
	case "DoubleUp":
		return "↑↑"
	case "SingleUp":
		return "↑"
	case "FortyFiveUp":
		return "↗"
	case "Flat":
		return "→"
	case "FortyFiveDown":
		return "↘"
	case "SingleDown":
		return "↓"
	case "DoubleDown":
		return "↓↓"
	default:
		return "?"
	}
}
