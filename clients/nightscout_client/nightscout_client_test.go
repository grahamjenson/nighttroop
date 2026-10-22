package nightscout_client

import (
	"context"
	"net/http"
	"net/http/httptest"
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
