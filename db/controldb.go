package db

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	mtsql "sqlitemt.t1d.nz/sql"
)

type ControlDB struct {
	*SQLDB

	// previously modified
	BackupInterval time.Duration
	SyncInterval   time.Duration

	tenants map[string]*TenantDB
	rwlock  sync.RWMutex

	ctx    context.Context
	cancel func()
	wg     sync.WaitGroup

	S3Client *S3Client
}

func NewControlDB(s3URL string) (*ControlDB, error) {
	var err error

	// Create a new control db
	controlDB := &ControlDB{
		SQLDB: &SQLDB{
			Path:       "data/db/control.sqlite",
			BackupPath: "data/backup/control.sqlite",
		},
		BackupInterval: 10 * time.Second,
		SyncInterval:   30 * time.Second,

		tenants: make(map[string]*TenantDB),
		rwlock:  sync.RWMutex{},
	}
	controlDB.ctx, controlDB.cancel = context.WithCancel(context.Background())

	controlDB.S3Client, err = NewS3Client(s3URL)
	if err != nil {
		return nil, fmt.Errorf("cannot create s3 client: %w", err)
	}

	return controlDB, nil
}

func (db *ControlDB) TryRestoreDBFile() error {

	err := db.S3Client.DownloadFile(db.BackupPath, db.Path)
	// TODO: check to see the error is not a 404
	if err != nil {
		return fmt.Errorf("cannot download control db: %w", err)
	}

	return nil
}

func (db *ControlDB) TryRestoreTenantDBFile(tenant *TenantDB) error {
	// Check if the tenant db file exists
	_, err := os.Stat(tenant.Path)
	if err == nil {
		// File exists, no need to download
		fmt.Println("Tenant db file exists, no need to download")
		return nil
	} else if !os.IsNotExist(err) {
		// The error should be not exists
		// Some other error occurred
		return fmt.Errorf("cannot stat tenant db: %w", err)
	}

	fmt.Println("Attempting to download tenant db file")
	err = db.S3Client.DownloadFile(tenant.BackupPath, tenant.Path)
	// TODO: check to see the error is not a 404
	if err != nil {
		return fmt.Errorf("cannot download tenant db: %w", err)
	}

	return nil
}

func (db *ControlDB) GetNTenantsDBs() int {
	l := 0
	db.rwlock.RLock()
	l = len(db.tenants)
	db.rwlock.RUnlock()
	return l
}

func (db *ControlDB) GetTenant(subdomain string) (*TenantDB, error) {
	fmt.Println("GetTenant", subdomain)
	// See if the tenant is already there
	db.rwlock.RLock()
	tenant, ok := db.tenants[subdomain]
	db.rwlock.RUnlock()
	if ok {
		fmt.Println("Found tenant in map")
		return tenant, nil
	}

	// Check if tenant exists in the db
	fmt.Println("Tenant not found in map, checking DB")
	lastSeenAt, err := db.TenantInDB(subdomain)
	if err != nil {
		fmt.Println("Error getting tenant from DB", err)
		return nil, fmt.Errorf("cannot get tenant from DB: %w", err)
	}
	if lastSeenAt == nil {
		fmt.Println("Tenant not found in DB")
		return nil, fmt.Errorf("tenant not found in DB")
	}

	tenantDB := NewTenantDB(subdomain)

	fmt.Println("Trying to restore tenant from backup")
	err = db.TryRestoreTenantDBFile(tenantDB)
	if err != nil {
		return nil, fmt.Errorf("cannot restore tenant db: %w", err)
	}
	fmt.Println("Tenant db Restored and Opening")
	err = tenantDB.Open(mtsql.TenantCreate)
	if err != nil {
		return nil, fmt.Errorf("cannot open tenant db: %w", err)
	}
	fmt.Println("Tenant db opened adding to map")

	// Add the tenant to the map
	db.rwlock.Lock()
	db.tenants[subdomain] = tenantDB
	db.rwlock.Unlock()

	// Update the last seen at
	fmt.Println("Updating last seen at")
	_, err = db.Exec("UPDATE tenants SET last_seen_at = ? WHERE subdomain = ?", time.Now().Format(time.RFC3339), subdomain)
	if err != nil {
		fmt.Println("Error updating last seen at", err)
		return nil, fmt.Errorf("cannot update last seen at: %w", err)
	}
	fmt.Println("Last seen at updated")
	return tenantDB, nil
}

func (db *ControlDB) TenantInDB(subdomain string) (*string, error) {
	// Check if tenant exists in the db
	rows, err := db.Query("SELECT last_seen_at FROM tenants WHERE subdomain = ? LIMIT 1", subdomain)
	if err != nil {
		return nil, fmt.Errorf("cannot get tenant: %w", err)
	}
	if rows == nil {
		return nil, fmt.Errorf("tenant not found")
	}

	defer rows.Close()
	rows.Next()

	var last_seen_at string
	err = rows.Scan(&last_seen_at)
	if err != nil {
		return nil, fmt.Errorf("cannot scan tenant: %w", err)
	}
	fmt.Println("Found Tenant: ", last_seen_at)

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return &last_seen_at, nil
}

func (db *ControlDB) CreateTenant(ctx context.Context) (string, error) {
	tenantID := NextTenantID()
	_, err := db.Exec(mtsql.ControlTenantInsert, tenantID)
	if err != nil {
		return "", fmt.Errorf("cannot create tenant: %w", err)
	}

	// Create a new tenant db
	tenantDB := NewTenantDB(tenantID)

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

func (db *ControlDB) StartBackupLoop() {
	// Backup loops over all dbs and backup to the backup folder
	if db.BackupInterval > 0 {
		db.wg.Add(1)
		go func() { defer db.wg.Done(); db.backupMonitor() }()
	}

	// Sync loops over all
	if db.SyncInterval > 0 {
		db.wg.Add(1)
		go func() { defer db.wg.Done(); db.syncMonitor() }()
	}

	// start a loop that goes over changed dbs and back them up to db

}

func (db *ControlDB) backupMonitor() {
	syncTicker := time.NewTicker(db.BackupInterval)
	defer syncTicker.Stop()

	for {
		// Wait for ticker or context close.
		select {
		case <-db.ctx.Done():
			return
		case <-syncTicker.C:
		}
		fmt.Println("Backup loop")
		// Sync the database to the shadow WAL.
		if err := db.backup(); err != nil && !errors.Is(err, context.Canceled) {
			fmt.Println("Backup error", "error", err)
		}
	}
}

func (db *ControlDB) syncMonitor() {
	backupTicker := time.NewTicker(db.SyncInterval)
	defer backupTicker.Stop()

	for {
		// Wait for ticker or context close.
		select {
		case <-db.ctx.Done():
			return
		case <-backupTicker.C:
		}

		fmt.Println("Sync loop")
		// Sync the database to the shadow WAL.
		if err := db.sync(); err != nil && !errors.Is(err, context.Canceled) {
			fmt.Println("Sync error", "error", err)
		}
	}
}

func (db *ControlDB) backup() error {
	defer trackTime(time.Now(), "backup")
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
		db.shouldBeSynced = true
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
			tenant.shouldBeSynced = true
		}
	}

	// Copy the control db to the backup location
	return nil
}

// This function will loop over all files in the /data/backup directory and upload to the s3 bucket
func (db *ControlDB) sync() error {
	defer trackTime(time.Now(), "sync")
	// Upload the control db to the s3 bucket
	tosync := []*SQLDB{}
	paths := []string{}
	if db.shouldBeSynced {
		tosync = append(tosync, db.SQLDB)
		paths = append(paths, db.BackupPath)
	}

	tenants := db.GetTenantDBsSnapshot()
	for _, t := range tenants {
		if t.shouldBeSynced {
			tosync = append(tosync, t.SQLDB)
			paths = append(paths, t.BackupPath)
		}
	}

	fmt.Println("Syncing", len(tosync), "files to S3")

	err := db.S3Client.SyncAllFiles(paths)
	if err != nil {
		return fmt.Errorf("cannot sync files to s3: %w", err)
	}

	for _, ss := range tosync {
		ss.shouldBeSynced = false
	}

	return nil
}

func (db *ControlDB) StopBackupLoop() {
	// This will send a cancel signal to context
	db.cancel()
	// This will wait
	db.wg.Wait()
}

func (db *ControlDB) ForceBackupSync() error {
	// Go through every tenant and force a WAL dump with `PRAGMA wal_checkpoint(FULL);`
	// Then check to see if they have changed (not sure if the wal_checkpoint itself causes modified date to change)
	err := db.backup()
	if err != nil {
		fmt.Println("Error forcing backup:", err)
		return fmt.Errorf("cannot force backup: %w", err)
	}

	err = db.sync()
	if err != nil {
		fmt.Println("Error forcing sync:", err)
		return fmt.Errorf("cannot force sync: %w", err)
	}

	return nil
}
