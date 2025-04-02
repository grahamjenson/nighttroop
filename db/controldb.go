package db

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	mtsql "sqlitemt.t1d.nz/sql"
)

type ControlDB struct {
	*SQLDB

	// previously modified
	BackupInterval time.Duration
	SyncInterval   time.Duration

	S3Config *S3Config
	tenants  map[string]*TenantDB
	rwlock   sync.RWMutex
}

func InitControlDB(s3Config *S3Config) (*ControlDB, error) {
	// Create a new control db
	controlDB := &ControlDB{
		SQLDB: &SQLDB{
			Path: "data/db/control.sqlite",
		},
		BackupInterval: 5 * time.Second,
		SyncInterval:   5 * time.Second,
		S3Config:       s3Config,
		tenants:        make(map[string]*TenantDB),
		rwlock:         sync.RWMutex{},
	}

	return controlDB, nil
}

func (db *ControlDB) GetNTenantsDBs() int {
	l := 0
	db.rwlock.RLock()
	l = len(db.tenants)
	db.rwlock.RUnlock()
	return l
}

func (db *ControlDB) GetTenant(subdomain string) (*string, error) {
	// TODO: return the tenant db
	// Get tenant from db
	// If we have not seen this tenant in the last 10 mins update last_seen
	return nil, nil
}

func (db *ControlDB) CreateTenant(ctx context.Context) (string, error) {
	tenantID := NextTenantID()
	res, err := db.Exec(mtsql.ControlTenantInsert, tenantID)
	fmt.Println("CreateTenant res", res, err)
	if err != nil {
		return "", fmt.Errorf("cannot create tenant: %w", err)
	}

	// Create a new tenant db
	tenantDB, err := InitTenantDB(tenantID)
	if err != nil {
		return "", fmt.Errorf("cannot create tenant db: %w", err)
	}

	err = tenantDB.Open(mtsql.TenantCreate)
	if err != nil {
		return "", fmt.Errorf("cannot create tenant db: %w", err)
	}
	db.rwlock.Lock()
	db.tenants[tenantID] = tenantDB
	db.rwlock.Unlock()
	return tenantID, nil
}

func (db *ControlDB) Close() error {
	for _, tenant := range db.tenants {
		err := tenant.Close()
		if err != nil {
			fmt.Println("Error closing tenant db:", err)
			// We just ignore and keep on closing stuff
		}
	}
	return db.sqldb.Close()
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func NextTenantID() string {
	result := make([]byte, 12)
	for i := range result {
		result[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(result)
}
