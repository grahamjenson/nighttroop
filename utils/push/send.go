package push

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"time"

	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/security"
)

// ErrNoCredential is a push to an app that neither its site nor the server has a credential for
var ErrNoCredential = errors.New("there is no push credential for the app")

// Alert is a notification shown on the device
type Alert struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	// Data is given to the app with the alert, at the top of an APNs push, as Loop reads a remote command from it.
	// A push with data wakes the app to act on it, and shows even in a Focus mode, as Nightscout's to Loop does.
	Data map[string]any `json:"-"`
}

// Send shows the alert on the registration's device, with the credential PushCredentialFor finds for its app: the
// site's own, else the server's. It records the push, with its error, if it had one, and deletes a token the push
// service says is gone.
func Send(ctx context.Context, q *controldb.Queries, sec *security.SecurityConfig, registration *controldb.PushRegistration, alert Alert) error {
	credential, err := q.PushCredentialFor(ctx, controldb.PushCredentialForParams{SiteID: &registration.SiteID,
		Provider: registration.Provider, App: string(registration.App), BundleID: registration.BundleID})
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%w %s", ErrNoCredential, registration.App)
	}
	if err != nil {
		return err
	}
	err = send(ctx, sec, credential, registration, alert)
	countPush(time.Now(), err)
	if errors.Is(err, ErrTokenGone) {
		if deleteErr := q.DeletePushRegistrationByTokenHash(ctx, controldb.DeletePushRegistrationByTokenHashParams{SiteID: registration.SiteID, TokenHash: registration.TokenHash}); deleteErr != nil {
			return errors.Join(err, deleteErr)
		}
		return err
	}
	var lastError *string
	if err != nil {
		lastError = new(err.Error()[:min(len(err.Error()), 256)])
	}
	if recordErr := q.RecordPush(ctx, controldb.RecordPushParams{ID: registration.ID, SiteID: registration.SiteID, TokenHash: registration.TokenHash, LastError: lastError}); recordErr != nil {
		return errors.Join(err, recordErr)
	}
	return err
}

// send pushes the alert with the credential, its secret and the registration's token decrypted
func send(ctx context.Context, sec *security.SecurityConfig, credential *controldb.PushCredential, registration *controldb.PushRegistration, alert Alert) error {
	if credential.SecretEnc == nil || (credential.Provider == consts.PushProviderAPNs && credential.KeyID == nil) {
		return ErrNoCredential
	}
	// A credential for every site, with no site of its own, is encrypted for none
	credentialSite := ""
	if credential.SiteID != nil {
		credentialSite = *credential.SiteID
	}
	secret, err := sec.DecryptFor(security.PurposePushSecret, credentialSite, *credential.SecretEnc)
	if err != nil {
		return fmt.Errorf("cannot decrypt the push credential: %w", err)
	}
	token, err := sec.DecryptFor(security.PurposePushToken, registration.SiteID, registration.TokenEnc)
	if err != nil {
		return fmt.Errorf("cannot decrypt the push token: %w", err)
	}
	if credential.Provider == consts.PushProviderFCM {
		message, _ := json.Marshal(map[string]any{"message": map[string]any{"token": token, "notification": alert}})
		return DefaultFCM.Send(ctx, credential.ID, secret, credential.TeamOrProjectID, message)
	}
	if credential.Provider != consts.PushProviderAPNs {
		return fmt.Errorf("pushes through %s are not sent", credential.Provider)
	}
	key, err := APNsKey(secret)
	if err != nil {
		return err
	}
	aps := map[string]any{"alert": alert, "sound": "default"}
	if alert.Data != nil {
		aps["content-available"], aps["interruption-level"] = 1, "time-sensitive"
	}
	body := maps.Clone(alert.Data)
	if body == nil {
		body = map[string]any{}
	}
	body["aps"] = aps
	payload, _ := json.Marshal(body)
	signer := Signer{ID: credential.ID, TeamID: credential.TeamOrProjectID, KeyID: *credential.KeyID, Key: key}
	return DefaultAPNs.Send(ctx, signer, token, registration.BundleID, registration.Production, payload)
}
