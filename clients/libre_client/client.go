// Package libre_client is a minimal LibreLinkUp (FreeStyle Libre follower) client, using the endpoints
// open-source clients such as nightscout-librelink-up use. It reads the first followed patient's graph.
package libre_client

import (
	"bytes"
	"cmp"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"nighttroop.t1d.nz/models/libremodels"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
)

// TestValidEmail is the only LibreLinkUp email that logs in under go test,
// which never makes HTTP calls; any other email is rejected.
const TestValidEmail = "valid-libre-user@example.com"

const (
	maxResponseBytes = 1 << 20 // a LibreLinkUp response body larger than this is refused
	maxReadings      = 1000    // the graph has about 12 hours of readings, every 15 minutes
)

// regionURL is the API host of one of consts.LibreRegions; no other host is ever called
func regionURL(region string) (string, error) {
	if !consts.IsLibreRegion(region) {
		return "", fmt.Errorf("unknown LibreLinkUp region %q", region)
	}
	if region == consts.LibreRegionGlobal {
		return "https://api.libreview.io", nil
	}
	return "https://api-" + strings.ToLower(region) + ".libreview.io", nil
}

// Client is a LibreLinkUp client, which keeps its login until it expires.
type Client struct {
	httpClient *http.Client
	urlFor     func(region string) (string, error)
	// offline is set under go test, so logging in needs no HTTP calls
	offline bool

	email    string
	password string
	baseURL  string

	token     string
	accountID string
	expires   time.Time

	// Unreadable is told of each reading Libre sent that cannot be read, which is skipped; when nil it is logged
	Unreadable func(raw json.RawMessage, err error)
}

// NewClient creates a client for the region, one of consts.LibreRegions.
func NewClient(email, password, region string) (*Client, error) {
	baseURL, err := regionURL(region)
	if err != nil {
		return nil, err
	}
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
			// Never follow redirects, which could point at an internal host; the 3xx is a failure.
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		},
		urlFor:   regionURL,
		offline:  testing.Testing(),
		email:    email,
		password: password,
		baseURL:  baseURL,
	}, nil
}

// do sends a request with the headers the LibreLinkUp app sends and decodes the JSON response into out.
// Go asks for and unzips gzip itself, and the limit is on the unzipped body.
func (c *Client) do(ctx context.Context, method, path string, body any, out any) error {
	var reqBody io.Reader
	if body != nil {
		b, _ := json.Marshal(body) // a map of strings always marshals
		reqBody = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return err
	}
	req.Header.Set("product", "llu.android")
	req.Header.Set("version", "4.16.0")
	req.Header.Set("cache-control", "no-cache")
	req.Header.Set("content-type", "application/json")
	if c.token != "" {
		req.Header.Set("authorization", "Bearer "+c.token)
		req.Header.Set("account-id", c.accountID)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Its URL has the followed patient's ID
		return httputils.WithoutURL(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		c.token = "" // log in again next time
	}
	if resp.StatusCode != http.StatusOK {
		// The error is stored and shown as a sync error, so only the status is given, and the path's last part, as the
		// graph's path has the followed patient's ID
		return fmt.Errorf("libre %s failed: status=%d", path[strings.LastIndex(path, "/")+1:], resp.StatusCode)
	}
	data, err := httputils.ReadLimitedBody(resp.Body, maxResponseBytes)
	if err == nil {
		err = json.Unmarshal(data, out)
	}
	if err != nil {
		return fmt.Errorf("error decoding libre response: %w", err)
	}
	return nil
}

type loginResponse struct {
	Status int `json:"status"`
	Data   struct {
		Redirect bool   `json:"redirect"`
		Region   string `json:"region"`
		Step     struct {
			Type string `json:"type"`
		} `json:"step"`
		User struct {
			ID string `json:"id"`
		} `json:"user"`
		AuthTicket struct {
			Token   string `json:"token"`
			Expires int64  `json:"expires"`
		} `json:"authTicket"`
	} `json:"data"`
}

// stepActions are what LibreLinkUp asks to be done in its app before it lets a login through
var stepActions = map[string]string{"tou": "accept the terms of use", "pp": "accept the privacy policy", "verifyEmail": "verify your email"}

// Login logs in, moving to the account's region when LibreLinkUp says it is in another one.
func (c *Client) Login(ctx context.Context) error {
	if c.offline {
		if c.email != TestValidEmail {
			return fmt.Errorf("libre authentication failed: invalid email or password")
		}
		c.token, c.expires = "test-token", time.Now().Add(time.Hour)
		return nil
	}
	c.token = ""
	// At most one move to another region, so a login never bounces between hosts
	for range 2 {
		var resp loginResponse
		if err := c.do(ctx, http.MethodPost, "/llu/auth/login", map[string]string{"email": c.email, "password": c.password}, &resp); err != nil {
			return err
		}
		d := resp.Data
		if d.Redirect {
			// Only a region in consts.LibreRegions is followed, never a host from the response
			baseURL, err := c.urlFor(strings.ToUpper(d.Region))
			if err != nil {
				return fmt.Errorf("libre login redirected: %w", err)
			}
			c.baseURL = baseURL
			continue
		}
		if action, ok := stepActions[d.Step.Type]; ok || resp.Status == 4 {
			return fmt.Errorf("libre login needs you to %s in the LibreLinkUp app", cmp.Or(action, stepActions["tou"]))
		}
		if resp.Status != 0 || d.AuthTicket.Token == "" || d.User.ID == "" {
			return fmt.Errorf("libre authentication failed: invalid email or password")
		}
		sum := sha256.Sum256([]byte(d.User.ID))
		c.token, c.accountID, c.expires = d.AuthTicket.Token, hex.EncodeToString(sum[:]), time.Unix(d.AuthTicket.Expires, 0)
		return nil
	}
	return fmt.Errorf("libre login redirected more than once")
}

type connectionsResponse struct {
	Data []struct {
		PatientID string `json:"patientId"`
	} `json:"data"`
}

type graphResponse struct {
	Data struct {
		Connection struct {
			GlucoseMeasurement json.RawMessage `json:"glucoseMeasurement"`
		} `json:"connection"`
		GraphData []json.RawMessage `json:"graphData"`
	} `json:"data"`
}

// FetchReadings reads the first followed patient's graph readings, oldest first, then their latest one.
// Only the latest has a trend.
func (c *Client) FetchReadings(ctx context.Context) ([]libremodels.Reading, error) {
	if c.token == "" || time.Now().After(c.expires.Add(-time.Minute)) {
		if err := c.Login(ctx); err != nil {
			return nil, err
		}
	}
	if c.offline {
		return nil, nil
	}

	var connections connectionsResponse
	if err := c.do(ctx, http.MethodGet, "/llu/connections", nil, &connections); err != nil {
		return nil, err
	}
	if len(connections.Data) == 0 || connections.Data[0].PatientID == "" {
		return nil, fmt.Errorf("libre account follows no one: accept a LibreLinkUp invite in the app first")
	}

	var graph graphResponse
	if err := c.do(ctx, http.MethodGet, "/llu/connections/"+url.PathEscape(connections.Data[0].PatientID)+"/graph", nil, &graph); err != nil {
		return nil, err
	}
	if len(graph.Data.GraphData) > maxReadings {
		return nil, fmt.Errorf("libre returned %d readings, more than %d", len(graph.Data.GraphData), maxReadings)
	}

	readings := make([]libremodels.Reading, 0, len(graph.Data.GraphData)+1)
	for i, raw := range append(graph.Data.GraphData, graph.Data.Connection.GlucoseMeasurement) {
		var r libremodels.Reading
		if err := json.Unmarshal(raw, &r); err != nil {
			if c.Unreadable == nil {
				slog.Error("Error unmarshalling libre reading", "error", err)
			} else {
				c.Unreadable(raw, err)
			}
			continue
		}
		if i < len(graph.Data.GraphData) {
			r.TrendArrow = nil // only the latest measurement's trend is kept
		}
		readings = append(readings, r)
	}
	return readings, nil
}
