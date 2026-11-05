package dexcom_client

import (
	"bytes"
	"compress/gzip"
	"errors"
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

// Dexcom answers a username or password it does not know with a 500 and a code, which is the account owner's to fix,
// so it is ErrLoginRefused; any other failed login is not
func TestLoginRefusedIsAnAccountDexcomDoesNotKnow(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		refused bool
	}{
		{"a wrong password", `{"Code":"AccountPasswordInvalid","Message":"Publisher account password failed"}`, true},
		{"an unknown account", `{"Code":"SSO_AuthenticateAccountNotFound","Message":"Account not found"}`, true},
		{"a wrong password by SSO", `{"Code":"SSO_AuthenticatePasswordInvalid","Message":"Password not valid"}`, true},
		{"too many attempts", `{"Code":"SSO_AuthenticateMaxAttemptsExceeed","Message":"Max attempts"}`, true},
		{"Dexcom failing", `{"Code":"SSO_InternalError","Message":"Internal error"}`, false},
		{"no code", `Service Unavailable`, false},
	}
	for _, test := range tests {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(test.body))
		}))
		_, err := liveClient(srv.URL).authenticate(context.Background())
		srv.Close()
		require.Error(t, err, test.name)
		require.Equal(t, test.refused, errors.Is(err, ErrLoginRefused), test.name)
	}
}
