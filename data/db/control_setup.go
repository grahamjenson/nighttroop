package db

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"strings"
	"sync"

	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/consts"
	nterrors "nighttroop.t1d.nz/utils/errors"
	"nighttroop.t1d.nz/utils/security"
)

// A new server has no control admin, so it is set up first: until the admin is made, every page goes to the control
// site's setup page, which asks for a one-time code the server prints to its logs, so only whoever runs it can.

// ErrBadSetup is a setup code that is wrong, or a TOTP code that is not the shown secret's
var ErrBadSetup = errors.New("bad setup code or TOTP")

// SetupAdmin is the first control admin's credentials, shown on the setup page and saved once its TOTP is confirmed
type SetupAdmin struct {
	Name          string
	Password      string
	TotpSecret    string
	passwordHash  string
	totpSecretEnc string
}

type setup struct {
	mu      sync.Mutex
	code    string // empty once the server is set up
	pending *SetupAdmin
}

// startSetupIfNoAdmin gives a server without a control admin its setup code
func (db *ControlDB) startSetupIfNoAdmin(ctx context.Context) error {
	admins, err := db.Queries.CountControlAdmins(ctx)
	if err != nil || admins > 0 {
		return err
	}
	db.setup.mu.Lock()
	defer db.setup.mu.Unlock()
	db.setup.code = security.SecureRandomString(16, security.API_KEY_CHAR_SET)
	return nil
}

// SetupCode is the code the setup page asks for, or empty once the server is set up
func (db *ControlDB) SetupCode() string {
	db.setup.mu.Lock()
	defer db.setup.mu.Unlock()
	return db.setup.code
}

func (db *ControlDB) SetupDone() bool {
	return db.SetupCode() == ""
}

// BeginSetup makes the admin's credentials to show, given the setup code; asking again replaces them
func (db *ControlDB) BeginSetup(code string) (*SetupAdmin, error) {
	db.setup.mu.Lock()
	defer db.setup.mu.Unlock()
	if !db.setupCodeMatches(code) {
		return nil, ErrBadSetup
	}
	password, passwordHash, err := db.ServerConfig.Security.GeneratePassword()
	if err != nil {
		return nil, err
	}
	name := consts.FIRST_ADMIN_NAME
	totpSecret, totpSecretEnc, err := db.ServerConfig.Security.TOTPGenerate(fmt.Sprintf("%s@%s", name, consts.CONTROL_IDENTITY_SUBDOMAIN))
	if err != nil {
		return nil, err
	}
	db.setup.pending = &SetupAdmin{Name: name, Password: password, TotpSecret: totpSecret, passwordHash: passwordHash, totpSecretEnc: totpSecretEnc}
	return db.setup.pending, nil
}

// CompleteSetup saves the shown admin once a TOTP code from its secret proves it was saved, ending setup
func (db *ControlDB) CompleteSetup(ctx context.Context, code string, totp string) error {
	db.setup.mu.Lock()
	defer db.setup.mu.Unlock()
	pending := db.setup.pending
	if !db.setupCodeMatches(code) || pending == nil || !db.ServerConfig.Security.TOTPValidate(totp, pending.TotpSecret) {
		return ErrBadSetup
	}
	err := db.Queries.CreateControlIdentity(ctx, controldb.CreateControlIdentityParams{Name: pending.Name,
		PasswordHash: pending.passwordHash, TotpSecretEnc: pending.totpSecretEnc, Role: consts.ControlRoleAdmin})
	if err != nil {
		return fmt.Errorf("cannot create control admin: %w", err)
	}
	db.setup.code, db.setup.pending = "", nil
	return nil
}

func (db *ControlDB) setupCodeMatches(code string) bool {
	return db.setup.code != "" && subtle.ConstantTimeCompare([]byte(code), []byte(db.setup.code)) == 1
}

// SeedDevAdmin sets up a development or test server with an admin of the development password and TOTP
func (db *ControlDB) SeedDevAdmin(ctx context.Context) error {
	_, err := db.CreateControlIdentity(ctx, consts.FIRST_ADMIN_NAME, consts.ControlRoleAdmin, nil)
	return db.endSeededSetup(err)
}

// SeedStagingAdmin sets up a staging server with an admin of its configured password and TOTP secret, so a staging
// server whose data is wiped needs no setting up again. An admin it already has is kept as it is.
func (db *ControlDB) SeedStagingAdmin(ctx context.Context) error {
	cfg := db.ServerConfig
	totpSecretEnc, err := cfg.Security.Encrypt(strings.ToUpper(cfg.StagingAdminTOTPSecret))
	if err != nil {
		return fmt.Errorf("cannot encrypt the staging admin's TOTP secret: %w", err)
	}
	err = db.Queries.CreateControlIdentity(ctx, controldb.CreateControlIdentityParams{Name: consts.FIRST_ADMIN_NAME,
		PasswordHash: security.HashPassword(cfg.StagingAdminPassword, cfg.GetBcryptCost()), TotpSecretEnc: totpSecretEnc,
		Role: consts.ControlRoleAdmin})
	return db.endSeededSetup(err)
}

// endSeededSetup ends setup once a seeded admin is made, or was already there
func (db *ControlDB) endSeededSetup(err error) error {
	if err != nil && !nterrors.IsSQLUniqueViolationError(err) {
		return err
	}
	db.setup.mu.Lock()
	defer db.setup.mu.Unlock()
	db.setup.code, db.setup.pending = "", nil
	return nil
}
