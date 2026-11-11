package push

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"nighttroop.t1d.nz/utils/httputils"
)

// providerTokenAge is how long a provider token is used for: Apple refuses one older than an hour, and one made again
// within 20 minutes
const providerTokenAge = 40 * time.Minute

// ErrTokenGone is a push to a device token the push service says is no longer the app's, so is deleted
var ErrTokenGone = errors.New("the push token is no longer registered")

// Signer is what signs an APNs credential's provider tokens: its team ID, key ID and .p8 key. ID tells credentials
// apart, so each one's token is made again only as it ages.
type Signer struct {
	ID     int64
	TeamID string
	KeyID  string
	Key    *ecdsa.PrivateKey
}

// APNs sends pushes through Apple's push service, over HTTP/2, which Go's client uses with Apple's servers
type APNs struct {
	Client *http.Client
	// Host is where pushes go, Apple's production or sandbox server when empty, as the device token's build needs
	Host string

	mu     sync.Mutex
	tokens map[int64]providerToken
}

type providerToken struct {
	jwt string
	at  time.Time
}

// DefaultAPNs sends the server's APNs pushes
var DefaultAPNs = &APNs{Client: &http.Client{Timeout: 10 * time.Second}}

// Send pushes the JSON payload to the app with the bundle ID (its topic) on the device with the token. A device token
// that is no longer the app's is ErrTokenGone; any other refusal says Apple's reason.
func (a *APNs) Send(ctx context.Context, signer Signer, deviceToken string, topic string, production bool, payload []byte) error {
	jwt, err := a.providerToken(signer, time.Now())
	if err != nil {
		return err
	}
	host := a.Host
	if host == "" {
		host = "https://api.sandbox.push.apple.com"
		if production {
			host = "https://api.push.apple.com"
		}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, host+"/3/device/"+deviceToken, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("authorization", "bearer "+jwt)
	req.Header.Set("apns-topic", topic)
	req.Header.Set("apns-push-type", "alert")
	resp, err := a.Client.Do(req)
	if err != nil {
		// Its URL has the device token, which is never shown
		return fmt.Errorf("cannot reach Apple's push service: %w", httputils.WithoutURL(err))
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	var refused struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(io.LimitReader(resp.Body, 4096)).Decode(&refused)
	if resp.StatusCode == http.StatusGone || refused.Reason == "BadDeviceToken" || refused.Reason == "Unregistered" {
		return ErrTokenGone
	}
	if resp.StatusCode == http.StatusForbidden {
		// A provider token Apple refuses is made again next time, as the key may have been replaced
		a.mu.Lock()
		delete(a.tokens, signer.ID)
		a.mu.Unlock()
	}
	return fmt.Errorf("Apple refused the push: %d %s", resp.StatusCode, refused.Reason)
}

// providerToken is the signer's JWT that authorizes pushes, made again once it is providerTokenAge old
func (a *APNs) providerToken(signer Signer, now time.Time) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if token, ok := a.tokens[signer.ID]; ok && now.Sub(token.at) < providerTokenAge {
		return token.jwt, nil
	}
	header, _ := json.Marshal(map[string]string{"alg": "ES256", "kid": signer.KeyID})
	claims, _ := json.Marshal(map[string]any{"iss": signer.TeamID, "iat": now.Unix()})
	unsigned := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(claims)
	digest := sha256.Sum256([]byte(unsigned))
	r, s, err := ecdsa.Sign(rand.Reader, signer.Key, digest[:])
	if err != nil {
		return "", fmt.Errorf("cannot sign an APNs provider token: %w", err)
	}
	// ES256 is R then S, each 32 bytes
	signature := make([]byte, 64)
	r.FillBytes(signature[:32])
	s.FillBytes(signature[32:])
	jwt := unsigned + "." + base64.RawURLEncoding.EncodeToString(signature)
	if a.tokens == nil {
		a.tokens = map[int64]providerToken{}
	}
	a.tokens[signer.ID] = providerToken{jwt: jwt, at: now}
	return jwt, nil
}
