package nightscout_client

import (
	"bytes"
	"compress/gzip"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// liveClient is a client whose requests go over HTTP to the given server, which httptest runs on a local address.
func liveClient(baseURL string) *Client {
	return signedInClient(baseURL, true)
}

func signedInClient(baseURL string, allowLocal bool) *Client {
	c := NewClient(baseURL, "token", ClientOptions{AllowLocalAddresses: allowLocal})
	jwt, expire := "jwt", time.Now().Add(time.Hour)
	c.jwt, c.jwtExpire = &jwt, &expire
	return c
}

// A Nightscout host checked when saved can later resolve to an internal address, so the client checks as it connects
func TestRefusesToConnectToLocalAddresses(t *testing.T) {
	hits := 0
	internal := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":{"collections":{}}}`))
	}))
	defer internal.Close()

	_, err := signedInClient(internal.URL, false).GetLastModified(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "private or local network")
	require.Equal(t, 0, hits, "a local address must never be requested")

	for _, address := range []string{"127.0.0.1:443", "10.0.0.1:443", "169.254.169.254:80", "[::1]:443", "[fd00:ec2::254]:80"} {
		require.Error(t, refuseBlockedIPs("tcp", address, nil), address)
	}
	require.NoError(t, refuseBlockedIPs("tcp", "93.184.215.14:443", nil))
}

func TestGetDocumentsRefusesRedirect(t *testing.T) {
	secretHits := 0
	secret := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secretHits++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"secret":"xyz"}`))
	}))
	defer secret.Close()
	redirector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, secret.URL+"/latest/meta-data", http.StatusFound)
	}))
	defer redirector.Close()

	_, err := liveClient(redirector.URL).GetDocuments(context.Background(), "entries", time.Now(), time.Time{}, 10)
	require.Error(t, err)
	require.Contains(t, err.Error(), "302")
	require.NotContains(t, err.Error(), "xyz")
	require.Equal(t, 0, secretHits, "redirect target must never be requested")
}

// A request that fails is a sync error, stored, logged and sent to Sentry, so it never has the user's Nightscout URL
func TestRequestErrorsHideTheNightscoutURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/my-nightscout/api/v3/lastModified" {
			_, _ = w.Write([]byte(`{"result":{}}`))
			return
		}
		<-r.Context().Done()
	}))
	defer srv.Close()
	c := liveClient(srv.URL + "/my-nightscout")
	c.httpClient.Timeout = 50 * time.Millisecond

	_, documentsErr := c.GetDocuments(context.Background(), "entries", time.Now(), time.Time{}, 10)
	require.ErrorContains(t, documentsErr, "Client.Timeout exceeded")
	_, profileErr := c.GetCurrentProfile(context.Background())
	_, modifiedErr := c.GetLastModified(context.Background())
	for _, err := range []error{documentsErr, profileErr, modifiedErr} {
		require.Error(t, err)
		require.NotContains(t, err.Error(), "my-nightscout")
		require.NotContains(t, err.Error(), strings.TrimPrefix(srv.URL, "http://"))
	}
}

func TestGetDocumentsErrorOmitsBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"secret":"xyz"}`))
	}))
	defer srv.Close()

	_, err := liveClient(srv.URL).GetDocuments(context.Background(), "entries", time.Now(), time.Time{}, 10)
	require.Error(t, err)
	require.NotContains(t, err.Error(), "xyz")
}

func TestParseResultRefusesLargeBodies(t *testing.T) {
	// A few MB of spaces inside the object inflate well from a small gzip body
	big := `{"result":[` + strings.Repeat(" ", 3<<20) + `]}`
	var gz bytes.Buffer
	w := gzip.NewWriter(&gz)
	_, err := w.Write([]byte(big))
	require.NoError(t, err)
	require.NoError(t, w.Close())

	tests := []struct {
		name     string
		body     []byte
		encoding string
	}{
		{"plain", []byte(big), ""},
		{"gzip", gz.Bytes(), "gzip"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.encoding != "" {
					w.Header().Set("Content-Encoding", tt.encoding)
				}
				_, _ = w.Write(tt.body)
			}))
			defer server.Close()
			resp, err := http.Get(server.URL)
			require.NoError(t, err)
			defer resp.Body.Close()

			result, err := parseResult(resp.Body)
			require.Error(t, err)
			require.Nil(t, result)
		})
	}
}

func TestDocumentsFromResultRefusesTooManyDocuments(t *testing.T) {
	result := make([]any, 0, maxResultItems+1)
	for i := 0; i < maxResultItems; i++ {
		result = append(result, map[string]any{"sgv": 100})
	}
	entries, err := documentsFromResult(result)
	require.NoError(t, err)
	require.Len(t, entries, maxResultItems)

	entries, err = documentsFromResult(append(result, map[string]any{"sgv": 100}))
	require.Error(t, err)
	require.Nil(t, entries)
}

func TestGetLastModified(t *testing.T) {
	tests := []struct {
		name string
		body string
		want map[string]int64
	}{
		{"collections", `{"status":200,"result":{"srvDate":1700000000500,"collections":{"entries":1700000000000,"treatments":1690000000000,"settings":0}}}`, map[string]int64{"entries": 1700000000000, "treatments": 1690000000000, "settings": 0}},
		{"no collections", `{"status":200,"result":{"srvDate":1700000000500}}`, nil},
		{"no result", `{"status":200}`, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "/api/v3/lastModified", r.URL.Path)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			lastModified, err := liveClient(srv.URL).GetLastModified(context.Background())
			if tt.want == nil {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, lastModified)
		})
	}
}

func TestRequestsRefreshAJWTCloseToExpiring(t *testing.T) {
	tests := []struct {
		name      string
		expiresIn time.Duration
		refreshed bool
	}{
		{"an hour left", time.Hour, false},
		{"seconds left", 30 * time.Second, true},
		{"expired", -time.Minute, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hits := 0
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				hits++
				_, _ = w.Write([]byte(`{"result":{"collections":{"entries":1}}}`))
			}))
			defer srv.Close()
			c := liveClient(srv.URL)
			expire := time.Now().Add(tt.expiresIn)
			c.jwtExpire = &expire

			// Under go test a refresh only works for TestValidURL, so refreshing this server's JWT fails before any request
			_, err := c.GetLastModified(context.Background())
			if tt.refreshed {
				require.ErrorIs(t, err, ErrJWTNoHost)
				require.Equal(t, 0, hits)
			} else {
				require.NoError(t, err)
				require.Equal(t, 1, hits)
			}
		})
	}
}

// Nightscout gives its current profile as an object and another Nighttroop as a list of it, and one with none gives
// null, so a site with no profiles copies nothing
func TestGetCurrentProfile(t *testing.T) {
	for body, want := range map[string]map[string]any{
		`{"_id":"a","defaultProfile":"Default","store":{"Default":{}}}`:   {"_id": "a", "defaultProfile": "Default", "store": map[string]any{"Default": map[string]any{}}},
		`[{"_id":"b","defaultProfile":"Default","store":{"Default":{}}}]`: {"_id": "b", "defaultProfile": "Default", "store": map[string]any{"Default": map[string]any{}}},
		`[]`:   nil,
		`null`: nil,
		`{}`:   nil,
	} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/api/v1/profile/current", r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(body))
		}))
		profile, err := liveClient(srv.URL).GetCurrentProfile(context.Background())
		srv.Close()
		require.NoError(t, err, body)
		require.Equal(t, want, profile, body)
	}
}
