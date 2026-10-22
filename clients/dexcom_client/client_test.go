package dexcom_client

import (
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
