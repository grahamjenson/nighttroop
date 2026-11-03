package push

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

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
	if errors.Is(err, ErrTokenGone) {
		if deleteErr := q.DeletePushRegistrationByTokenHash(ctx, controldb.DeletePushRegistrationByTokenHashParams{TokenHash: registration.TokenHash}); deleteErr != nil {
			return errors.Join(err, deleteErr)
		}
		return err
	}
	var lastError *string
	if err != nil {
		lastError = new(err.Error()[:min(len(err.Error()), 256)])
	}
	if recordErr := q.RecordPush(ctx, controldb.RecordPushParams{ID: registration.ID, LastError: lastError}); recordErr != nil {
		return errors.Join(err, recordErr)
	}
	return err
}

// send pushes the alert with the credential, its secret and the registration's token decrypted
func send(ctx context.Context, sec *security.SecurityConfig, credential *controldb.PushCredential, registration *controldb.PushRegistration, alert Alert) error {
	if credential.Provider != consts.PushProviderAPNs {
		return fmt.Errorf("pushes through %s are not sent yet", credential.Provider)
	}
	if credential.SecretEnc == nil || credential.KeyID == nil {
		return ErrNoCredential
	}
	p8, err := sec.Decrypt(*credential.SecretEnc)
	if err != nil {
		return fmt.Errorf("cannot decrypt the push credential: %w", err)
	}
	key, err := APNsKey(p8)
	if err != nil {
		return err
	}
	token, err := sec.Decrypt(registration.TokenEnc)
	if err != nil {
		return fmt.Errorf("cannot decrypt the push token: %w", err)
	}
	payload, _ := json.Marshal(map[string]any{"aps": map[string]any{"alert": alert, "sound": "default"}})
	signer := Signer{ID: credential.ID, TeamID: credential.TeamOrProjectID, KeyID: *credential.KeyID, Key: key}
	return DefaultAPNs.Send(ctx, signer, token, registration.BundleID, registration.Production, payload)
}
