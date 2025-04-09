package db

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"runtime"
	"sort"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	mtsql "sqlitemt.t1d.nz/sql"
)

type ControlDB struct {
	*SQLDB

	// previously modified
	ControlBackupInterval time.Duration
	PoolBackupInterval    time.Duration

	tenantPools map[byte]*TenantPool

	ctx    context.Context
	cancel func()
	wg     sync.WaitGroup

	S3Client *S3Client
}

func NewControlDB(endpoint, bucket string) (*ControlDB, error) {
	var err error

	// Create a new control db
	controlDB := &ControlDB{
		SQLDB: &SQLDB{
			Path:       "data/db/control.sqlite",
			BackupPath: "data/backup/control.sqlite",
			Modified:   false,
		},
		ControlBackupInterval: 10 * time.Second,
		PoolBackupInterval:    60 * time.Second,
	}
	controlDB.ctx, controlDB.cancel = context.WithCancel(context.Background())

	controlDB.S3Client, err = NewS3Client(endpoint, bucket)
	if err != nil {
		return nil, fmt.Errorf("cannot create s3 client: %w", err)
	}

	// Create a pool for each starting character of a tenant subdomain
	// This map will never be added or removed from so no need for a lock
	controlDB.tenantPools = map[byte]*TenantPool{}
	for _, c := range tenantCharSet {
		controlDB.tenantPools[byte(c)] = NewTenantPool(byte(c))
	}

	return controlDB, nil
}

func (db *ControlDB) TryRestoreDBFile() error {
	// If it already exists dont download
	if db.DBFileExists() {
		return nil
	}
	os.MkdirAll("data/backup", os.ModePerm)
	os.MkdirAll("data/db", os.ModePerm)
	found, err := db.S3Client.DownloadFile(db.BackupPath, db.Path)
	if !found && err == nil {
		fmt.Println("No control db backup found")
		return nil
	}
	// TODO: check to see the error is not a 404
	if err != nil {
		return fmt.Errorf("cannot download control db: %w", err)
	}

	return nil
}

func (db *ControlDB) RestoreAndOpenAllTenants() error {
	defer trackTime(time.Now(), "RestoreAndOpenAllTenants")
	g, _ := errgroup.WithContext(context.Background())
	limit := int(math.Max(1, float64(runtime.NumCPU()-1)))
	g.SetLimit(limit)
	fmt.Println("RestoreAndOpenAllTenants running", len(db.tenantPools), limit)
	for _, pool := range db.allPools() {

		g.Go(func() error {
			tenantSubdomains, err := db.GetTenantPoolTennantsFromDB(pool.char)
			fmt.Println("TenantSubdomains", pool.char, tenantSubdomains)
			if err != nil {
				return fmt.Errorf("cannot get tenant pool tenants: %w", err)
			}
			err = pool.RestoreAndOpenAllTenants(db.S3Client, tenantSubdomains)
			if err != nil {
				return fmt.Errorf("cannot restore tenant pool: %w", err)
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		fmt.Printf("One of the RestoreAndOpenAllTenants  tasks failed: %v\n", err)
		return fmt.Errorf("cannot restore and open all tenants: %w", err)
	} else {
		fmt.Println("All RestoreAndOpenAllTenants tasks succeeded!")
	}

	return nil
}

func (db *ControlDB) GetTenantPoolTennantsFromDB(char byte) ([]string, error) {
	tenants := []string{}
	res, err := db.Query("SELECT subdomain FROM tenants WHERE pool = ? ORDER BY subdomain DESC", int(char))
	if err != nil {
		return tenants, fmt.Errorf("cannot query tenants: %w", err)
	}

	defer res.Close()

	for res.Next() {
		var subdomain string
		err = res.Scan(&subdomain)
		if err != nil {
			return tenants, fmt.Errorf("cannot scan tenant: %w", err)
		}
		tenants = append(tenants, subdomain)
	}
	err = res.Err()
	if err != nil {
		return tenants, fmt.Errorf("rows error: %w", err)
	}

	return tenants, nil
}

func (db *ControlDB) GetNTenantsDBs() int {
	total := 0
	for _, pool := range db.tenantPools {
		total += pool.GetNTenantsDBs()
	}
	return total
}

func (db *ControlDB) GetTenant(subdomain string, updateLastSeen bool) (*TenantDB, error) {
	if subdomain == "" {
		return nil, fmt.Errorf("subdomain cannot be empty")
	}

	fmt.Println("GetTenant", subdomain)

	// See if the tenant is already there
	tp := db.tenantPools[subdomain[0]]
	tenant := tp.GetTenant(subdomain)
	tenantInDB, err := db.TenantInDB(subdomain, updateLastSeen)

	fmt.Println("TenantInDB", tenantInDB, err)
	if err != nil {
		fmt.Println("Error getting tenant from DB", err)
		return nil, fmt.Errorf("cannot get tenant from DB: %w", err)
	}

	// Absolute Tenant not in the ControlDB we dont return
	if !tenantInDB {
		fmt.Println("Tenant not found in DB")
		return nil, fmt.Errorf("tenant not found in DB")
	}

	if tenant == nil {
		fmt.Println("Tenant not found in map")
		return tenant, nil
	}

	return tenant, nil
}

func (db *ControlDB) TenantInDB(subdomain string, updateLastSeen bool) (bool, error) {
	// Check if tenant exists in the db
	rows, err := db.Query("SELECT last_seen_at FROM tenants WHERE subdomain = ? LIMIT 1", subdomain)
	if err != nil {
		return false, fmt.Errorf("cannot get tenant: %w", err)
	}
	if rows == nil {
		return false, fmt.Errorf("tenant not found")
	}

	defer rows.Close()
	err = rows.Err()
	if err != nil {
		return false, fmt.Errorf("rows error: %w", err)
	}

	rows.Next()

	var last_seen_at string
	err = rows.Scan(&last_seen_at)
	if err != nil {
		return false, fmt.Errorf("cannot scan tenant: %w", err)
	}
	fmt.Println("Found Tenant: ", last_seen_at)

	if updateLastSeen {
		err = db.TryUpdateLastSeenAt(subdomain, last_seen_at)
		// Not going to break because this errors
		if err != nil {
			fmt.Println("Error updating last seen at", err)
		}
	}

	return true, nil
}

func (db *ControlDB) TryUpdateLastSeenAt(subdomain string, last_seen_at string) error {

	last_seen_at_time, err := time.Parse(time.RFC3339, last_seen_at)
	if err != nil {
		fmt.Println("Error parsing last seen at", err)
		return fmt.Errorf("cannot parse last seen at: %w", err)

	}

	// Only update if the last seen at is older than 10 minutes
	if last_seen_at_time.Add(10 * time.Minute).After(time.Now()) {
		return nil
	}

	// Try update lsat seen at, I am not going to throw here if not updated
	fmt.Println("Trying to updating last seen at")
	_, err = db.Exec("UPDATE tenants SET last_seen_at = ? WHERE subdomain = ?", time.Now().Format(time.RFC3339), subdomain)
	if err != nil {
		fmt.Println("Error updating last seen at", err)
		return fmt.Errorf("cannot update last seen at: %w", err)
	}
	fmt.Println("Last seen at updated")
	return nil
}

func (db *ControlDB) CreateTenant(ctx context.Context) (string, error) {
	tenantID := NextTenantID()
	_, err := db.Exec(mtsql.ControlTenantInsert, tenantID, int(tenantID[0]))
	if err != nil {
		return "", fmt.Errorf("cannot create tenant: %w", err)
	}

	tp := db.tenantPools[tenantID[0]]
	_, err = tp.CreateTenant(tenantID)
	if err != nil {
		return "", fmt.Errorf("cannot create tenant: %w", err)
	}
	return tenantID, nil
}

func (db *ControlDB) Close() error {
	// Keep it locked so no new tenants can be created
	for _, pool := range db.allPools() {
		pool.Close()
	}

	return db.sqldb.Close()
}

func (db *ControlDB) StartBackupLoops() {
	// CotrolDB backup and Sync Loop
	if db.ControlBackupInterval > 0 {
		db.wg.Add(1)
		go func() {
			defer db.wg.Done()
			db.backupLoop()
		}()
	}

	if db.PoolBackupInterval > 0 {
		db.wg.Add(1)
		// The goal here is over the PoolBackupInterval we will backup all pools
		// But we will stagger them so that we do not swamp the CPU or network
		staggeredInterval := db.PoolBackupInterval / time.Duration(len(db.tenantPools))
		go func() {
			defer db.wg.Done()
			db.tpBackupLoop(staggeredInterval)
		}()
	}

}

func (db *ControlDB) allPools() []*TenantPool {
	var allPools []*TenantPool
	for _, pool := range db.tenantPools {
		allPools = append(allPools, pool)
	}

	sort.Slice(allPools, func(i, j int) bool {
		return allPools[i].char < allPools[j].char
	})
	return allPools
}

func (db *ControlDB) tpBackupLoop(interval time.Duration) {
	syncTicker := time.NewTicker(interval)
	defer syncTicker.Stop()

	allPools := db.allPools()

	index := 0
	for {
		// Wait for ticker or context close.
		select {
		case <-db.ctx.Done():
			return
		case <-syncTicker.C:
		}
		tp := allPools[index]
		fmt.Println("Backup Tenant loop")
		// Backup DB to S3
		if err := tp.backup(db.S3Client); err != nil && !errors.Is(err, context.Canceled) {
			fmt.Println("Backup error", "error", err)
		}
		index = (index + 1) % len(allPools)
	}
}

func (db *ControlDB) backupLoop() {
	syncTicker := time.NewTicker(db.ControlBackupInterval)
	defer syncTicker.Stop()

	for {
		// Wait for ticker or context close.
		select {
		case <-db.ctx.Done():
			return
		case <-syncTicker.C:
		}

		// Backup DB to S3
		if err := db.backup(); err != nil && !errors.Is(err, context.Canceled) {
			fmt.Println("Backup error", "error", err)
		}

	}

}

func (db *ControlDB) backup() error {
	defer trackTime(time.Now(), "control backup")

	var err error
	db.UpdateModified()

	if db.Modified {
		err = db.SqliteBackup()
		if err != nil {
			return fmt.Errorf("cannot backup control db: %w", err)
		}

		err = db.S3Client.SyncAllFiles([]string{db.BackupPath})
		if err != nil {
			return fmt.Errorf("cannot sync control db: %w", err)
		}
	}

	// Copy the control db to the backup location
	return nil
}

func (db *ControlDB) StopBackupLoop() {
	// This will send a cancel signal to context
	db.cancel()
	// This will wait
	db.wg.Wait()
}

func (db *ControlDB) ForceBackupSync() error {
	defer trackTime(time.Now(), "ForceBackupSync")
	// Go through every tenant and force a WAL dump with `PRAGMA wal_checkpoint(FULL);`
	// Then check to see if they have changed (not sure if the wal_checkpoint itself causes modified date to change)
	err := db.backup()
	if err != nil {
		fmt.Println("Error forcing backup:", err)
		return fmt.Errorf("cannot force backup: %w", err)
	}

	g, _ := errgroup.WithContext(context.Background())

	// Limit the number of concurrent goroutines to the number of CPU cores - 1
	limit := int(math.Max(1, float64(runtime.NumCPU()-1)))
	g.SetLimit(limit)
	fmt.Println("ForceBackupSync running", len(db.tenantPools), limit)
	for _, pool := range db.allPools() {

		g.Go(func() error {
			return pool.backup(db.S3Client)
		})
	}

	if err := g.Wait(); err != nil {
		fmt.Printf("One of the ForceBackupSync tasks failed: %v\n", err)
		return fmt.Errorf("cannot sync all tenants: %w", err)
	} else {
		fmt.Println("All ForceBackupSync tasks succeeded!")
	}

	return nil
}
