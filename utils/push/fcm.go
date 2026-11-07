package push

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/jwt"
)

// FCM sends pushes through Firebase Cloud Messaging's HTTP v1 API, with an OAuth token for each service account
type FCM struct {
	Client *http.Client
	// Host is where pushes go, Firebase's when empty
	Host string
	// TokenURL is where a service account's signed JWT gets an OAuth token, Google's when empty. The token_uri in its
	// JSON is never used, so no site can have the server send requests where it likes.
	TokenURL string

	mu      sync.Mutex
	sources map[int64]oauth2.TokenSource
}

// DefaultFCM sends the server's FCM pushes
var DefaultFCM = &FCM{Client: &http.Client{Timeout: 10 * time.Second}}

// Send pushes the JSON message (`{"message":{"token":...}}`) through the project, with the service account's JSON of
// the credential with the ID, which keeps each one's token apart. A registration token that is no longer the app's is
// ErrTokenGone; any other refusal says Firebase's reason.
func (f *FCM) Send(ctx context.Context, credentialID int64, serviceAccount string, project string, message []byte) error {
	source, err := f.tokenSource(credentialID, serviceAccount)
	if err != nil {
		return err
	}
	token, err := source.Token()
	if refused, ok := errors.AsType[*oauth2.RetrieveError](err); ok {
		var body struct {
			Error string `json:"error"`
		}
		_ = json.Unmarshal(refused.Body, &body)
		return fmt.Errorf("Google refused the service account: %d %s", refused.Response.StatusCode, body.Error)
	}
	if err != nil {
		return fmt.Errorf("cannot get the service account's OAuth token: %w", err)
	}
	host := cmp.Or(f.Host, "https://fcm.googleapis.com")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, host+"/v1/projects/"+url.PathEscape(project)+"/messages:send", bytes.NewReader(message))
	if err != nil {
		return err
	}
	token.SetAuthHeader(req)
	req.Header.Set("content-type", "application/json")
	resp, err := f.Client.Do(req)
	if err != nil {
		return fmt.Errorf("cannot reach Firebase's push service: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	var refused struct {
		Error struct {
			Status  string `json:"status"`
			Details []struct {
				ErrorCode string `json:"errorCode"`
			} `json:"details"`
		} `json:"error"`
	}
	_ = json.NewDecoder(io.LimitReader(resp.Body, 4096)).Decode(&refused)
	reason := refused.Error.Status
	for _, detail := range refused.Error.Details {
		reason = cmp.Or(detail.ErrorCode, reason)
	}
	if reason == "UNREGISTERED" {
		return ErrTokenGone
	}
	if resp.StatusCode == http.StatusUnauthorized {
		// An OAuth token Firebase refuses is got again next time
		f.mu.Lock()
		delete(f.sources, credentialID)
		f.mu.Unlock()
	}
	return fmt.Errorf("Firebase refused the push: %d %s", resp.StatusCode, reason)
}

// tokenSource is the service account's OAuth tokens to send pushes with, each used until five minutes before it expires
func (f *FCM) tokenSource(credentialID int64, serviceAccount string) (oauth2.TokenSource, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if source, ok := f.sources[credentialID]; ok {
		return source, nil
	}
	var account struct {
		ClientEmail  string `json:"client_email"`
		PrivateKeyID string `json:"private_key_id"`
		PrivateKey   string `json:"private_key"`
	}
	if json.Unmarshal([]byte(serviceAccount), &account) != nil {
		return nil, fmt.Errorf("the secret is not a service account's JSON")
	}
	config := &jwt.Config{Email: account.ClientEmail, PrivateKey: []byte(account.PrivateKey), PrivateKeyID: account.PrivateKeyID,
		Scopes: []string{"https://www.googleapis.com/auth/firebase.messaging"}, TokenURL: cmp.Or(f.TokenURL, "https://oauth2.googleapis.com/token")}
	// Tokens are got outside any one push's context, as they are kept for later ones, with the client's timeout
	source := oauth2.ReuseTokenSourceWithExpiry(nil, config.TokenSource(context.WithValue(context.Background(), oauth2.HTTPClient, f.Client)), 5*time.Minute)
	if f.sources == nil {
		f.sources = map[int64]oauth2.TokenSource{}
	}
	f.sources[credentialID] = source
	return source, nil
}
