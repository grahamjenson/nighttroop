package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"nighttroop.t1d.nz/data/sql/user"

	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/can"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
)

// //
// TenentDB
// //

type UserDB struct {
	serverConfig config.ServerConfig

	SQLiteDB  *sql.DB
	Subdomain string

	// Prepared statements
	Queries *user.Queries
}

func NewUserDB(ctx context.Context, subdomain string, mustExist bool, cfg config.ServerConfig) (*UserDB, error) {
	path, _ := utils.UserDBPaths(subdomain)
	if mustExist && !utils.FileExists(path) {
		return nil, fmt.Errorf("The User DB %s Must exist and doesnt", subdomain)
	}

	sqldb, err := OpenSqliteDB(path, 100, cfg.IsTest())
	if err != nil {
		return nil, fmt.Errorf("cannot open user db: %s %w", subdomain, err)
	}

	// Create a new control db
	userDB := &UserDB{
		serverConfig: cfg,
		SQLiteDB:     sqldb,
		Subdomain:    subdomain,
	}

	err = user.MigrateUp(subdomain, sqldb)
	if err != nil {
		sqldb.Close()
		return userDB, fmt.Errorf("Migration Error Occured user db: %s %w", subdomain, err)
	}

	userDB.Queries, err = user.Prepare(context.TODO(), sqldb)
	if err != nil {
		sqldb.Close()
		return userDB, fmt.Errorf("cannot prepare statements for user db: %s %w", subdomain, err)
	}

	return userDB, nil
}

// WARNING THIS IS A BEST TRY FUNCTION ALL ERRORS ARE SWOLLOWED
func (db *UserDB) TryCloseAndDelete() {
	if db == nil {
		return
	}
	if db.SQLiteDB != nil {
		db.SQLiteDB.Close()
	}
	if db.Subdomain == "" {
		return
	}
	path, _ := utils.UserDBPaths(db.Subdomain)
	_ = utils.DeleteFile(path)
}

func (db *UserDB) Close() error {
	return db.SQLiteDB.Close()
}

func (db *UserDB) URL() string {
	return fmt.Sprintf("%s://%s.%s:%d", db.serverConfig.HostProtocol, db.Subdomain, db.serverConfig.HostName, db.serverConfig.HostPort)
}

func (db *UserDB) NSURL() string {
	return fmt.Sprintf("%s://ns.%s.%s:%d", db.serverConfig.HostProtocol, db.Subdomain, db.serverConfig.HostName, db.serverConfig.HostPort)
}

// ////
// Identities
// ////

func (db *UserDB) createDefaultUsers() (*IdentityValues, *IdentityValues, error) {
	// Create Default Admin User
	adminIdentity, err := db.CreateIdentity(consts.FIRST_ADMIN_NAME, consts.IdentityRoleAdmin)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to Create Default Admin User:	 %w", err)
	}
	if adminIdentity.TotpSecret == nil {
		return nil, nil, fmt.Errorf("Failed to Create Default Admin User: TOTP Secret is nil")
	}

	// Create Default PWD User (Able to write data on the ns API)

	pwdIdentity, err := db.CreateIdentity(consts.FIRST_PWD_NAME, consts.IdentityRoleDevicePWDPhone)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to Create Default PWD User:	 %w", err)
	}
	if pwdIdentity.NSToken == nil {
		return nil, nil, fmt.Errorf("Failed to Create Default PWD User: NSToken is nil")
	}

	////
	// Create Sync Identities (used to trail)
	////
	_, err = db.CreateIdentity(consts.DEXCOM_SYNC_IDENTITY_NAME, consts.IdentityRoleSync)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create dexcom sync identity: %w", err)
	}

	_, err = db.CreateIdentity(consts.NIGHTSCOUT_SYNC_IDENTITY_NAME, consts.IdentityRoleSync)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create nightscout sync identity: %w", err)
	}

	_, err = db.CreateIdentity(consts.LIBRE_SYNC_IDENTITY_NAME, consts.IdentityRoleSync)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create libre sync identity: %w", err)
	}

	return adminIdentity, pwdIdentity, nil
}

func (db *UserDB) createDefaultSettings() error {
	// TODO: we need to enum all allowed settings
	_, err := user.UnitsSetting.UpsertValue(context.TODO(), db.Queries, user.UnitsSettingValue{Units: consts.UNITS_MG_DL})
	if err != nil {
		return fmt.Errorf("cannot set default units setting: %w", err)
	}
	_, err = user.TimezoneSetting.UpsertValue(context.TODO(), db.Queries, user.TimezoneSettingValue{TZ: *time.UTC})
	if err != nil {
		return fmt.Errorf("cannot set default timezone setting: %w", err)
	}
	return nil
}

type IdentityValues struct {
	Password   string
	TotpSecret *string
	NSToken    *string
}

func (db *UserDB) CreateIdentity(name string, role consts.IdentityRole) (*IdentityValues, error) {

	security := db.serverConfig.Security
	hostname := db.serverConfig.HostName
	password, hashedPassword, err := security.GeneratePassword()

	if err != nil {
		return nil, fmt.Errorf("cannot generate user password: %w", err)
	}

	var totpSecretPtr *string
	encTOTPSecret := ""
	if can.MustLoginWithTotp(role, false) {

		totpSecret, encTOTPSecretRet, err := security.TOTPGenerate(fmt.Sprintf("%s@%s.%s", name, db.Subdomain, hostname))
		if err != nil {
			return nil, fmt.Errorf("cannot generate user TOTP: %w", err)
		}

		totpSecretPtr = &totpSecret
		encTOTPSecret = encTOTPSecretRet
	}

	err = db.Queries.CreateIdentity(context.TODO(), user.CreateIdentityParams{
		Name:       name,
		Password:   hashedPassword,
		TotpSecret: encTOTPSecret,
		Role:       role,
	})

	if err != nil {
		return nil, fmt.Errorf("cannot create user identity: %w", err)
	}
	var nsToken *string
	if can.HaveNightscoutToken(role) {
		token, err := security.CreateNightscoutTokenForIdentity(db.Subdomain, name, string(role), password)
		if err != nil {
			return nil, fmt.Errorf("cannot create nightscout token for identity: %w", err)
		}
		nsToken = &token
	}

	return &IdentityValues{
		Password:   password,
		TotpSecret: totpSecretPtr,
		NSToken:    nsToken,
	}, nil
}

func (db *UserDB) UpdateIdentityDateFlags(identity *user.Identity, updateLastSeen bool, updateLastWrite bool) error {

	// Only update if not seen in 10 minutes
	lastWriteMoreThan10MinutesAgo := identity.LastWriteAt == nil || time.Now().After(identity.LastWriteAtDate().Add(-10*time.Minute))
	lastSeenMoreThan10MinutesAgo := identity.LastSeenAt == nil || time.Now().After(identity.LastSeenAtDate().Add(-10*time.Minute))

	now := time.Now().UnixMilli()
	var lastSeenAt *int64
	var lastWriteAt *int64
	if updateLastSeen && lastSeenMoreThan10MinutesAgo {
		lastSeenAt = &now
	}
	if updateLastWrite && lastWriteMoreThan10MinutesAgo {
		lastWriteAt = &now
	}
	db.Queries.UpdateIdentityDateFlags(context.TODO(), lastSeenAt, lastWriteAt, identity.ID)

	return nil
}
