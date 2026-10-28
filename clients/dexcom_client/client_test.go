package dexcom_client

import (
	"bytes"
	"compress/gzip"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// liveClient is a logged-in client whose requests go over HTTP to the given server.
func liveClient(baseURL string) *Client {
	c := NewClient(TestValidUsername, "password", false)
	c.baseURL = baseURL
	c.sessionID = "live-session"
	return c
}

func TestFetchEGVsRefusesRedirect(t *testing.T) {
	secretHits := 0
	secret := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secretHits++
		_, _ = w.Write([]byte(`[{"secret":"xyz"}]`))
	}))
	defer secret.Close()
	redirector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, secret.URL+"/latest/meta-data", http.StatusFound)
	}))
	defer redirector.Close()

	_, err := liveClient(redirector.URL).FetchEGVs(context.Background(), 60)
	require.Error(t, err)
	require.Contains(t, err.Error(), "302")
	require.NotContains(t, err.Error(), "xyz")
	require.NotContains(t, err.Error(), secret.URL)
	require.Equal(t, 0, secretHits, "redirect target must never be requested")
}

func TestFetchEGVsErrorBodyIsBounded(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(strings.Repeat("a", 500) + "xyz"))
	}))
	defer srv.Close()

	_, err := liveClient(srv.URL).FetchEGVs(context.Background(), 60)
	require.Error(t, err)
	require.Contains(t, err.Error(), "500")
	require.NotContains(t, err.Error(), "xyz")
	require.Less(t, len(err.Error()), 200)
}

func TestFetchEGVsRefusesLargeBodies(t *testing.T) {
	// A few MB of spaces inside the array inflate well from a small gzip body
	big := `[{"Value":100}` + strings.Repeat(" ", 3<<20) + `]`
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

			egvs, err := liveClient(server.URL).FetchEGVs(context.Background(), 60)
			require.Error(t, err)
			require.Nil(t, egvs)
		})
	}
}

func TestDecodeEGVsRefusesTooManyEGVs(t *testing.T) {
	egv := `{"WT":"Date(1700000000000)","ST":"Date(1700000000000)","DT":"Date(1700000000000)","Value":100,"Trend":"Flat"}`
	ok := "[" + strings.Repeat(egv+",", maxEGVs-1) + egv + "]"
	egvs, err := decodeEGVs([]byte(ok))
	require.NoError(t, err)
	require.Len(t, egvs, maxEGVs)

	tooMany := "[" + strings.Repeat(egv+",", maxEGVs) + egv + "]"
	egvs, err = decodeEGVs([]byte(tooMany))
	require.Error(t, err)
	require.Nil(t, egvs)
}
