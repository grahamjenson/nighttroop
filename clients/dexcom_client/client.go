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
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"nighttroop.t1d.nz/models/dexcommodels"
	"nighttroop.t1d.nz/utils/httputils"
)

// TestValidUsername is the only Dexcom username that logs in under go test,
// which never makes HTTP calls; any other username is rejected.
const TestValidUsername = "valid-dexcom-user"

// testSessionID is the session a test login gives; any other session goes over HTTP.
const testSessionID = "test-session"

const (
	usBaseURL        = "https://share2.dexcom.com"
	outsideUSBaseURL = "https://shareous1.dexcom.com"
	applicationID    = "d8665ade-9673-4e27-9ff6-92db4ce13d13"

	maxResponseBytes = 1 << 20 // a Dexcom response body larger than this is refused
	maxEGVs          = 1000    // Dexcom returns at most a few hundred EGVs
)

// Client is a Dexcom Share API client.
type Client struct {
	httpClient *http.Client
	baseURL    string

	username string
	password string

	sessionID string

	// Unreadable is told of each reading Dexcom sent that cannot be read, which is skipped; when nil it is logged
	Unreadable func(raw json.RawMessage, err error)
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
			Timeout: 10 * time.Second,
			// Never follow redirects, which could point at an internal host; the 3xx is a failure.
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		},
		baseURL:  baseURL,
		username: username,
		password: password,
	}
}

// responseError describes a non-200 response. The error is stored and shown as a
// sync error, so a refused redirect gives only its status and other bodies are cut short.
func responseError(what string, resp *http.Response) error {
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		return fmt.Errorf("%s failed: upstream redirected, status=%d", what, resp.StatusCode)
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 100))
	for _, code := range loginRefusedCodes {
		if bytes.Contains(body, []byte(`"Code":"`+code+`"`)) {
			return fmt.Errorf("%s failed: %w (%s)", what, ErrLoginRefused, code)
		}
	}
	return fmt.Errorf("%s failed: status=%d body=%q", what, resp.StatusCode, body)
}

// ErrLoginRefused is Dexcom not knowing the account's username or password, which its owner fixes on the Connectors page
var ErrLoginRefused = errors.New("Dexcom refused the username or password")

// loginRefusedCodes are the codes Dexcom answers a username or password it does not know with
var loginRefusedCodes = []string{"AccountPasswordInvalid", "SSO_AuthenticateAccountNotFound", "SSO_AuthenticatePasswordInvalid", "SSO_AuthenticateMaxAttemptsExceeed"}

// Login authenticates and stores the Dexcom session ID.
func (c *Client) Login(ctx context.Context) error {
	if testing.Testing() {
		if c.username != TestValidUsername {
			return fmt.Errorf("dexcom authentication failed: %w", ErrLoginRefused)
		}
		c.sessionID = testSessionID
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
	return c.postForID(ctx, "authenticate", "/ShareWebServices/Services/General/AuthenticatePublisherAccount",
		"account", map[string]string{"accountName": c.username})
}

func (c *Client) loginByAccountID(ctx context.Context, accountID string) (string, error) {
	return c.postForID(ctx, "login", "/ShareWebServices/Services/General/LoginPublisherAccountById",
		"session", map[string]string{"accountId": accountID})
}

// postForID logs in with the account and its password, and is the ID, named idName, Dexcom answers with in quotes
func (c *Client) postForID(ctx context.Context, what, path, idName string, account map[string]string) (string, error) {
	account["password"] = c.password
	account["applicationId"] = applicationID
	body, err := c.post(ctx, what, path, account)
	if err != nil {
		return "", err
	}
	id := strings.Trim(string(body), "\"")
	if id == "" {
		return "", fmt.Errorf("empty %s id", idName)
	}
	return id, nil
}

// post sends a POST to the Share path, with the JSON of reqBody unless it is nil, and is the body of
// its OK response, of at most maxResponseBytes
func (c *Client) post(ctx context.Context, what, path string, reqBody any) ([]byte, error) {
	var body io.Reader
	if reqBody != nil {
		bodyBytes, err := json.Marshal(reqBody)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(bodyBytes)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("error creating Dexcom request: %w", err)
	}
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Its URL has the session, which reads the account's glucose, in its query
		return nil, fmt.Errorf("error sending Dexcom request: %w", httputils.WithoutURL(err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, responseError(what, resp)
	}
	return httputils.ReadLimitedBody(resp.Body, maxResponseBytes)
}

// FetchEGVs fetches the last specified minutes of EGV readings.
func (c *Client) FetchEGVs(ctx context.Context, minutes int) ([]dexcommodels.EGV, error) {
	if c.sessionID == "" {
		if err := c.Login(ctx); err != nil {
			return nil, fmt.Errorf("Dexcom Login Error: %w", err)
		}
	}
	if testing.Testing() && c.sessionID == testSessionID {
		return nil, nil
	}

	values := url.Values{}
	values.Set("sessionId", c.sessionID)
	values.Set("minutes", fmt.Sprintf("%d", minutes))
	values.Set("maxCount", fmt.Sprintf("%d", minutes/5))

	body, err := c.post(ctx, "dexcom fetch", "/ShareWebServices/Services/Publisher/ReadPublisherLatestGlucoseValues?"+values.Encode(), nil)
	if err != nil {
		return nil, err
	}
	return decodeEGVs(body, c.Unreadable)
}

// decodeEGVs decodes the EGV array in a Dexcom response body, refusing more than maxEGVs. Each EGV that cannot be
// read is skipped, and unreadable is told of it, when it is set.
func decodeEGVs(data []byte, unreadable func(json.RawMessage, error)) ([]dexcommodels.EGV, error) {
	var items []json.RawMessage
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, fmt.Errorf("error decoding Dexcom response: %w", err)
	}
	if len(items) > maxEGVs {
		return nil, fmt.Errorf("dexcom returned %d EGVs, more than %d", len(items), maxEGVs)
	}

	out := make([]dexcommodels.EGV, 0, len(items))

	for _, raw := range items {
		var r dexcommodels.EGV
		if err := json.Unmarshal(raw, &r); err != nil {
			if unreadable == nil {
				slog.Error("Error unmarshalling Dexcom response", "error", err, "raw", string(raw))
			} else {
				unreadable(raw, err)
			}
			continue
		}

		out = append(out, r)
	}

	return out, nil
}
