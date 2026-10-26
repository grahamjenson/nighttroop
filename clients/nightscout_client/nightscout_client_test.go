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

// liveClient is a client whose requests go over HTTP to the given server.
func liveClient(baseURL string) *Client {
	c := NewClient(baseURL, "token")
	jwt, expire := "jwt", time.Now().Add(time.Hour)
	c.jwt, c.jwtExpire = &jwt, &expire
	return c
}

func TestGetEntriesRefusesRedirect(t *testing.T) {
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

	_, err := liveClient(redirector.URL).GetEntries(context.Background(), time.Now(), 10)
	require.Error(t, err)
	require.Contains(t, err.Error(), "302")
	require.NotContains(t, err.Error(), "xyz")
	require.Equal(t, 0, secretHits, "redirect target must never be requested")
}

func TestGetEntriesErrorOmitsBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"secret":"xyz"}`))
	}))
	defer srv.Close()

	_, err := liveClient(srv.URL).GetEntries(context.Background(), time.Now(), 10)
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

func TestEntriesFromResultRefusesTooManyEntries(t *testing.T) {
	result := make([]any, 0, maxResultItems+1)
	for i := 0; i < maxResultItems; i++ {
		result = append(result, map[string]any{"sgv": 100})
	}
	entries, err := entriesFromResult(result)
	require.NoError(t, err)
	require.Len(t, entries, maxResultItems)

	entries, err = entriesFromResult(append(result, map[string]any{"sgv": 100}))
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
