package db

import (
	"context"
	"database/sql"
	"fmt"

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

	sqldb, err := OpenSqliteDB(path, 100)
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

// ////
// Identities
// ////

func (db *UserDB) GetIdentitySplat(ctx context.Context, name string) (identityName string, password string, totpSecret string, role consts.IdentityRole, err error) {
	identity, err := db.Queries.GetIdentity(ctx, name)
	if err != nil {
		return "", "", "", "", fmt.Errorf("identity not found: %w", err)
	}

	return identity.Name, identity.Password, identity.TotpSecret, identity.Role, nil
}

func (db *UserDB) createDefaultUsers() (string, string, error) {

	adminPassword, totpSecretPtr, err := db.CreateIdentity(consts.FIRST_ADMIN_NAME, consts.IdentityRoleAdmin)
	if err != nil {
		return "", "", fmt.Errorf("Failed to Create Default Admin User:	 %w", err)
	}
	if totpSecretPtr == nil {
		return "", "", fmt.Errorf("Failed to Create Default Admin User: TOTP Secret is nil")
	}
	totpSecret := *totpSecretPtr

	_, _, err = db.CreateIdentity(consts.DEXCOM_SYNC_IDENTITY_NAME, consts.IdentityRoleSync)
	if err != nil {
		return "", "", fmt.Errorf("cannot create dexcom sync identity: %w", err)
	}

	_, _, err = db.CreateIdentity(consts.NIGHTSCOUT_SYNC_IDENTITY_NAME, consts.IdentityRoleSync)
	if err != nil {
		return "", "", fmt.Errorf("cannot create nightscout sync identity: %w", err)
	}

	_, _, err = db.CreateIdentity(consts.LIBRE_SYNC_IDENTITY_NAME, consts.IdentityRoleSync)
	if err != nil {
		return "", "", fmt.Errorf("cannot create libre sync identity: %w", err)
	}

	return adminPassword, totpSecret, nil
}

func (db *UserDB) CreateIdentity(name string, role consts.IdentityRole) (string, *string, error) {

	security := db.serverConfig.Security
	hostname := db.serverConfig.HostName
	password, hashedPassword, err := security.GeneratePassword()

	if err != nil {
		return "", nil, fmt.Errorf("cannot generate user password: %w", err)
	}

	var totpSecretPtr *string
	encTOTPSecret := ""
	if can.MustLoginWithTotp(role, false) {

		totpSecret, encTOTPSecretRet, err := security.TOTPGenerate(fmt.Sprintf("%s@%s.%s", name, db.Subdomain, hostname))
		if err != nil {
			return "", nil, fmt.Errorf("cannot generate user TOTP: %w", err)
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
		return "", nil, fmt.Errorf("cannot create user identity: %w", err)
	}

	return password, totpSecretPtr, nil
}
