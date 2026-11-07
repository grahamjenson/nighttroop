package push

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// fakeGoogle is Google's OAuth token endpoint, its tokens lasting expiresIn seconds, and Firebase's send API,
// answering each push with status and the FCM error code, keeping the token requests' forms and the pushes it got
type fakeGoogle struct {
	expiresIn int
	status    int
	code      string
	tokens    []map[string][]string
	pushes    []*http.Request
}

func newFakeGoogle(t *testing.T) (*fakeGoogle, *FCM) {
	fake := &fakeGoogle{expiresIn: 3600, status: http.StatusOK}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		if r.URL.Path == "/token" {
			require.NoError(t, r.ParseForm())
			fake.tokens = append(fake.tokens, r.PostForm)
			w.Header().Set("content-type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "access-" + strconv.Itoa(len(fake.tokens)),
				"expires_in": fake.expiresIn, "token_type": "Bearer"})
			return
		}
		body, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(strings.NewReader(string(body)))
		fake.pushes = append(fake.pushes, r.Clone(r.Context()))
		w.WriteHeader(fake.status)
		if fake.status != http.StatusOK {
			_ = json.NewEncoder(w).Encode(map[string]any{"error": map[string]any{"code": fake.status, "status": "NOT_FOUND",
				"details": []map[string]string{{"@type": "type.googleapis.com/google.firebase.fcm.v1.FcmError", "errorCode": fake.code}}}})
		}
	}))
	t.Cleanup(server.Close)
	return fake, &FCM{Client: server.Client(), Host: server.URL, TokenURL: server.URL + "/token"}
}

// newServiceAccount is a service account's JSON, with a token_uri that is never used, and its key
func newServiceAccount(t *testing.T) (string, *rsa.PrivateKey) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	account, _ := json.Marshal(map[string]string{"type": "service_account", "project_id": "t1d-carer", "private_key_id": "key1",
		"client_email": "push@t1d-carer.iam.gserviceaccount.com", "private_key": pkcs8(t, key), "token_uri": "http://127.0.0.1:1/token"})
	return string(account), key
}

// The JWT asks for a token to send pushes with, signed by the service account's key, and the token sends the message
// to the project; it is used again until five minutes before it expires
func TestFCMSend(t *testing.T) {
	fake, fcm := newFakeGoogle(t)
	account, key := newServiceAccount(t)
	message := `{"message":{"token":"abc","notification":{"title":"Nighttroop","body":"Test"}}}`
	require.NoError(t, fcm.Send(t.Context(), 1, account, "t1d-carer", []byte(message)))

	require.Len(t, fake.tokens, 1)
	require.Equal(t, "urn:ietf:params:oauth:grant-type:jwt-bearer", fake.tokens[0]["grant_type"][0])
	parts := strings.Split(fake.tokens[0]["assertion"][0], ".")
	require.Len(t, parts, 3)
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	require.NoError(t, err)
	digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	require.NoError(t, rsa.VerifyPKCS1v15(&key.PublicKey, crypto.SHA256, digest[:], signature))
	decode := func(part string) map[string]any {
		b, err := base64.RawURLEncoding.DecodeString(part)
		require.NoError(t, err)
		m := map[string]any{}
		require.NoError(t, json.Unmarshal(b, &m))
		return m
	}
	require.Equal(t, map[string]any{"alg": "RS256", "typ": "JWT", "kid": "key1"}, decode(parts[0]))
	claims := decode(parts[1])
	require.Equal(t, "push@t1d-carer.iam.gserviceaccount.com", claims["iss"])
	require.Equal(t, "https://www.googleapis.com/auth/firebase.messaging", claims["scope"])
	require.Equal(t, fcm.TokenURL, claims["aud"])
	require.Equal(t, float64(3600), claims["exp"].(float64)-claims["iat"].(float64))

	require.Len(t, fake.pushes, 1)
	r := fake.pushes[0]
	require.Equal(t, "/v1/projects/t1d-carer/messages:send", r.URL.Path)
	require.Equal(t, "Bearer access-1", r.Header.Get("authorization"))
	require.Equal(t, "application/json", r.Header.Get("content-type"))
	body, _ := io.ReadAll(r.Body)
	require.JSONEq(t, message, string(body))

	// The token is used again by each push
	require.NoError(t, fcm.Send(t.Context(), 1, account, "t1d-carer", []byte(message)))
	require.Len(t, fake.tokens, 1)
	require.Equal(t, "Bearer access-1", fake.pushes[1].Header.Get("authorization"))

	// Another service account gets its own
	other, _ := newServiceAccount(t)
	require.NoError(t, fcm.Send(t.Context(), 2, other, "other-project", []byte(message)))
	require.Len(t, fake.tokens, 2)
	require.Equal(t, "/v1/projects/other-project/messages:send", fake.pushes[2].URL.Path)
	require.Equal(t, "Bearer access-2", fake.pushes[2].Header.Get("authorization"))

	// A token within five minutes of expiring is got again
	fake.expiresIn = 299
	expiring, _ := newServiceAccount(t)
	require.NoError(t, fcm.Send(t.Context(), 3, expiring, "t1d-carer", []byte(message)))
	require.NoError(t, fcm.Send(t.Context(), 3, expiring, "t1d-carer", []byte(message)))
	require.Len(t, fake.tokens, 4)
	require.Equal(t, "Bearer access-4", fake.pushes[4].Header.Get("authorization"))
}

// What Firebase answers is a gone token, or an error with its reason
func TestFCMRefusals(t *testing.T) {
	tests := []struct {
		name   string
		status int
		code   string
		gone   bool
		err    string
	}{
		{"an app uninstalled", http.StatusNotFound, "UNREGISTERED", true, ""},
		{"a bad message", http.StatusBadRequest, "INVALID_ARGUMENT", false, "400 INVALID_ARGUMENT"},
		{"a token of another project", http.StatusForbidden, "SENDER_ID_MISMATCH", false, "403 SENDER_ID_MISMATCH"},
		{"an OAuth token Firebase refuses", http.StatusUnauthorized, "THIRD_PARTY_AUTH_ERROR", false, "401 THIRD_PARTY_AUTH_ERROR"},
		{"too many pushes", http.StatusTooManyRequests, "QUOTA_EXCEEDED", false, "429 QUOTA_EXCEEDED"},
		{"no error code", http.StatusInternalServerError, "", false, "500 NOT_FOUND"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake, fcm := newFakeGoogle(t)
			fake.status, fake.code = tt.status, tt.code
			account, _ := newServiceAccount(t)
			err := fcm.Send(t.Context(), 1, account, "t1d-carer", []byte(`{}`))
			if tt.gone {
				require.ErrorIs(t, err, ErrTokenGone)
				return
			}
			require.ErrorContains(t, err, tt.err)
			require.NotErrorIs(t, err, ErrTokenGone)
			if tt.status != http.StatusUnauthorized {
				require.Contains(t, fcm.sources, int64(1))
				return
			}
			require.NotContains(t, fcm.sources, int64(1), "a refused OAuth token is got again")
			_ = fcm.Send(t.Context(), 1, account, "t1d-carer", []byte(`{}`))
			require.Len(t, fake.tokens, 2)
		})
	}
}

// A service account Google refuses says Google's error code, without its description, the JWT or the key, and no
// push is sent; it is tried again with the next push
func TestFCMTokenRefused(t *testing.T) {
	var tokens, pushes int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/token" {
			pushes++
			return
		}
		tokens++
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid_grant", "error_description": "Invalid JWT Signature."})
	}))
	t.Cleanup(server.Close)
	fcm := &FCM{Client: server.Client(), Host: server.URL, TokenURL: server.URL + "/token"}
	account, _ := newServiceAccount(t)
	err := fcm.Send(t.Context(), 1, account, "t1d-carer", []byte(`{}`))
	require.EqualError(t, err, "Google refused the service account: 400 invalid_grant")
	require.Zero(t, pushes)
	require.Error(t, fcm.Send(t.Context(), 1, account, "t1d-carer", []byte(`{}`)))
	require.Equal(t, 2, tokens)

	server.Close()
	err = fcm.Send(t.Context(), 1, account, "t1d-carer", []byte(`{}`))
	require.ErrorContains(t, err, "cannot get the service account's OAuth token")
	require.NotContains(t, err.Error(), "PRIVATE KEY")

	err = fcm.Send(t.Context(), 2, `{"private_key":"secret`, "t1d-carer", []byte(`{}`))
	require.EqualError(t, err, "the secret is not a service account's JSON")
}
