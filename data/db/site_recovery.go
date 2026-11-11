package db

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/security"
)

// ErrBadBackupKey is a backup key that is not the site's, or was already used
var ErrBadBackupKey = errors.New("bad backup key")

// RecoveredAdmin is what a recovery gives the site's owner: the admin's new credentials and the next backup key
type RecoveredAdmin struct {
	Name       string `json:"name"`
	Password   string `json:"password"`
	TotpSecret string `json:"totp_secret"`
	BackupKey  string `json:"backup_key"`
}

// RecoverSiteAdmin proves ownership of a site with its backup key, the one secret only its creator was shown. The
// key works once: it is replaced by a new one, and the site's admin gets a new password and TOTP, or is made again
// if it is gone, so whoever had the old credentials or key is shut out.
func (db *ControlDB) RecoverSiteAdmin(ctx context.Context, siteID string, backupKey string) (*RecoveredAdmin, error) {
	site, err := db.Queries.GetSite(ctx, controldb.GetSiteParams{ID: siteID})
	if err != nil {
		return nil, err
	}
	// A key is its site's subdomain then a secret, so a key for another site is refused before the costly check
	if !strings.HasPrefix(backupKey, siteID+":") || !db.ServerConfig.Security.VerifyPassword(backupKey, site.BackupKeyHash) {
		return nil, ErrBadBackupKey
	}

	newKey, newKeyHash, err := security.SecureCredentialsWithPrefix(siteID+":", db.ServerConfig.GetBcryptCost())
	if err != nil {
		return nil, fmt.Errorf("cannot generate backup key: %w", err)
	}
	password, passwordHash, err := db.ServerConfig.Security.GeneratePassword()
	if err != nil {
		return nil, fmt.Errorf("cannot generate password: %w", err)
	}
	name := consts.FIRST_ADMIN_NAME
	totpSecret, totpSecretEnc, err := db.ServerConfig.Security.TOTPGenerate(siteID, fmt.Sprintf("%s@%s.%s", name, siteID, db.ServerConfig.HostName))
	if err != nil {
		return nil, fmt.Errorf("cannot generate TOTP: %w", err)
	}
	tx, err := db.SQLiteDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	q := db.Queries.WithTx(tx)

	rotated, err := q.RotateSiteBackupKey(ctx, controldb.RotateSiteBackupKeyParams{NewHash: newKeyHash, ID: siteID, OldHash: site.BackupKeyHash})
	if err != nil {
		return nil, err
	}
	if n, err := rotated.RowsAffected(); err != nil || n != 1 {
		return nil, ErrBadBackupKey // another use of the same key got there first
	}
	reset, err := q.ResetSiteAdminCredentials(ctx, controldb.ResetSiteAdminCredentialsParams{PasswordHash: passwordHash, TotpSecretEnc: totpSecretEnc, SiteID: siteID, Name: name})
	if err == nil {
		err = q.DeleteSiteAdminPasskeys(ctx, controldb.DeleteSiteAdminPasskeysParams{SiteID: siteID, Name: name})
	}
	if err != nil {
		return nil, err
	}
	if n, err := reset.RowsAffected(); err != nil || n == 0 {
		// The admin was deleted, so it is made again, with credentials of its own
		created, err := db.CreateSiteIdentity(ctx, q, siteID, name, consts.IdentityRoleAdmin, nil)
		if err != nil {
			return nil, fmt.Errorf("cannot make the admin again: %w", err)
		}
		password, totpSecret = created.Password, *created.TotpSecret
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	db.Logs.SiteAdminRecoveredEvent(ctx, siteID)
	db.Logs.IdentityCredentialsChangedEvent(ctx, siteID, name, "password_and_totp", "backup key")
	return &RecoveredAdmin{Name: name, Password: password, TotpSecret: totpSecret, BackupKey: newKey}, nil
}
