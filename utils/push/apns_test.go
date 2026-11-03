package push

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// verifyJWT checks an ES256 JWT is signed by the key, and is its header and claims
func verifyJWT(t *testing.T, jwt string, key *ecdsa.PublicKey) (map[string]any, map[string]any) {
	parts := strings.Split(jwt, ".")
	require.Len(t, parts, 3)
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	require.NoError(t, err)
	require.Len(t, signature, 64)
	digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	require.True(t, ecdsa.Verify(key, digest[:], new(big.Int).SetBytes(signature[:32]), new(big.Int).SetBytes(signature[32:])))
	decode := func(part string) map[string]any {
		b, err := base64.RawURLEncoding.DecodeString(part)
		require.NoError(t, err)
		m := map[string]any{}
		require.NoError(t, json.Unmarshal(b, &m))
		return m
	}
	return decode(parts[0]), decode(parts[1])
}

// fakeAPNs is Apple's push service over HTTP/2, answering each push with status and reason, and keeping what it got
func fakeAPNs(t *testing.T, status int, reason string) (*APNs, *[]*http.Request) {
	got := []*http.Request{}
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, 2, r.ProtoMajor, "Apple's push service needs HTTP/2")
		body, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(strings.NewReader(string(body)))
		got = append(got, r.Clone(r.Context()))
		w.WriteHeader(status)
		if reason != "" {
			_ = json.NewEncoder(w).Encode(map[string]string{"reason": reason})
		}
	}))
	server.EnableHTTP2 = true
	server.StartTLS()
	t.Cleanup(server.Close)
	return &APNs{Client: server.Client(), Host: server.URL}, &got
}

func TestAPNsSend(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	signer := Signer{ID: 1, TeamID: "TEAM123456", KeyID: "ABC123DEFG", Key: key}

	apns, got := fakeAPNs(t, http.StatusOK, "")
	require.NoError(t, apns.Send(t.Context(), signer, "a1b2", "com.TEAM123456.loopkit.Loop", true, []byte(`{"aps":{}}`)))
	require.Len(t, *got, 1)
	r := (*got)[0]
	require.Equal(t, "/3/device/a1b2", r.URL.Path)
	require.Equal(t, "com.TEAM123456.loopkit.Loop", r.Header.Get("apns-topic"))
	require.Equal(t, "alert", r.Header.Get("apns-push-type"))
	jwt, found := strings.CutPrefix(r.Header.Get("authorization"), "bearer ")
	require.True(t, found)
	header, claims := verifyJWT(t, jwt, &key.PublicKey)
	require.Equal(t, map[string]any{"alg": "ES256", "kid": "ABC123DEFG"}, header)
	require.Equal(t, "TEAM123456", claims["iss"])
	body, _ := io.ReadAll(r.Body)
	require.JSONEq(t, `{"aps":{}}`, string(body))

	// The provider token is used again until it is providerTokenAge old
	require.NoError(t, apns.Send(t.Context(), signer, "a1b2", "com.TEAM123456.loopkit.Loop", true, []byte(`{}`)))
	require.Equal(t, r.Header.Get("authorization"), (*got)[1].Header.Get("authorization"))
	again, err := apns.providerToken(signer, time.Now().Add(providerTokenAge))
	require.NoError(t, err)
	require.NotEqual(t, jwt, again)
}

// What Apple answers is a gone token, or an error with its reason
func TestAPNsRefusals(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	signer := Signer{ID: 2, TeamID: "TEAM123456", KeyID: "ABC123DEFG", Key: key}
	tests := []struct {
		name   string
		status int
		reason string
		gone   bool
		err    string
	}{
		{"an app uninstalled", http.StatusGone, "Unregistered", true, ""},
		{"a token of another app or server", http.StatusBadRequest, "BadDeviceToken", true, ""},
		{"a topic the key cannot push to", http.StatusBadRequest, "TopicDisallowed", false, "400 TopicDisallowed"},
		{"a provider token Apple refuses", http.StatusForbidden, "InvalidProviderToken", false, "403 InvalidProviderToken"},
		{"too many pushes", http.StatusTooManyRequests, "TooManyRequests", false, "429 TooManyRequests"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apns, _ := fakeAPNs(t, tt.status, tt.reason)
			err := apns.Send(t.Context(), signer, "a1b2", "com.TEAM123456.loopkit.Loop", false, []byte(`{}`))
			if tt.gone {
				require.ErrorIs(t, err, ErrTokenGone)
				return
			}
			require.ErrorContains(t, err, tt.err)
			require.NotErrorIs(t, err, ErrTokenGone)
			if tt.status == http.StatusForbidden {
				require.NotContains(t, apns.tokens, signer.ID, "a refused provider token is made again")
			}
		})
	}
}
