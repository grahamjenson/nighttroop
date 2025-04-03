package db

import (
	"context"
	"errors"
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

	ctx    context.Context
	cancel func()
	wg     sync.WaitGroup
}

func NewControlDB(s3Config *S3Config) (*ControlDB, error) {
	// Create a new control db
	controlDB := &ControlDB{
		SQLDB: &SQLDB{
			Path:       "data/db/control.sqlite",
			BackupPath: "data/backup/control.sqlite",
		},
		BackupInterval: 5 * time.Second,
		SyncInterval:   5 * time.Second,
		S3Config:       s3Config,
		tenants:        make(map[string]*TenantDB),
		rwlock:         sync.RWMutex{},
	}
	controlDB.ctx, controlDB.cancel = context.WithCancel(context.Background())

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
	tenantDB, err := NewTenantDB(tenantID)
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

func (db *ControlDB) GetTenantDBsSnapshot() []*TenantDB {
	db.rwlock.RLock()
	defer db.rwlock.RUnlock()

	// Create a snapshot of the tenant databases
	tenantDBs := make([]*TenantDB, 0, len(db.tenants))
	for _, tenant := range db.tenants {
		tenantDBs = append(tenantDBs, tenant)
	}
	return tenantDBs
}

func (db *ControlDB) Close() error {
	// Keep it locked so no new tenants can be created
	db.rwlock.Lock()
	for _, tenant := range db.tenants {
		err := tenant.Close()
		if err != nil {
			fmt.Println("Error closing tenant db:", err)
			// We just ignore and keep on closing stuff
		}
	}
	db.rwlock.Unlock()

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

func (db ControlDB) Restore() error {
	return nil
}

func (db *ControlDB) StartBackupLoop() {
	// TODO sync and backup are backwards
	// start a go loop that checks for changes in the sqlite dbs and then copies them to the backup db path
	if db.SyncInterval > 0 {
		db.wg.Add(1)
		go func() { defer db.wg.Done(); db.syncMonitor() }()
	}

	// start a loop that goes over changed dbs and back them up to db
	if db.BackupInterval > 0 {
		db.wg.Add(1)
		go func() { defer db.wg.Done(); db.backupMonitor() }()
	}

}

func (db *ControlDB) syncMonitor() {
	syncTicker := time.NewTicker(db.SyncInterval)
	defer syncTicker.Stop()

	for {
		// Wait for ticker or context close.
		select {
		case <-db.ctx.Done():
			return
		case <-syncTicker.C:
		}
		fmt.Println("Sync loop")
		// Sync the database to the shadow WAL.
		if err := db.sync(); err != nil && !errors.Is(err, context.Canceled) {
			fmt.Println("sync error", "error", err)
		}
	}
}

func (db *ControlDB) backupMonitor() {
	backupTicker := time.NewTicker(db.BackupInterval)
	defer backupTicker.Stop()

	for {
		// Wait for ticker or context close.
		select {
		case <-db.ctx.Done():
			return
		case <-backupTicker.C:
		}

		fmt.Println("Backup loop")
		// Sync the database to the shadow WAL.
		if err := db.backup(); err != nil && !errors.Is(err, context.Canceled) {
			fmt.Println("backup error", "error", err)
		}
	}
}

func (db *ControlDB) sync() error {
	defer trackTime(time.Now(), "Sync")
	var err error
	err = db.UpdateModified()
	tenants := db.GetTenantDBsSnapshot()
	if err != nil {
		return fmt.Errorf("cannot update last modified: %w", err)
	}

	if db.Modified {
		err = db.SqliteBackup()
		if err != nil {
			return fmt.Errorf("cannot backup control db: %w", err)
		}
	}

	for _, tenant := range tenants {
		err = tenant.UpdateModified()
		if err != nil {
			return fmt.Errorf("cannot update last modified: %w", err)
		}
		if tenant.Modified {
			err = tenant.SqliteBackup()
			if err != nil {
				return fmt.Errorf("cannot backup tenant db: %w", err)
			}
		}
	}

	// Copy the control db to the backup location
	return nil
}

func (db *ControlDB) backup() error {
	defer trackTime(time.Now(), "backup")
	return nil
}

func (db *ControlDB) StopBackupLoop() {
	// This will send a cancel signal to context
	db.cancel()
	// This will wait
	db.wg.Wait()
}

func (db *ControlDB) ForceBackup() {
	// Go through every tenant and force a WAL dump with `PRAGMA wal_checkpoint(FULL);`
	// Then check to see if they have changed (not sure if the wal_checkpoint itself causes modified date to change)
}
