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

////
// Tenant Pool
////

type TenantPool struct {
	tenants map[string]*TenantDB
	rwLock  sync.RWMutex

	S3Path string
	Path   string
	char   byte
}

func NewTenantPool(char byte) *TenantPool {
	return &TenantPool{
		char:    char,
		tenants: make(map[string]*TenantDB),
		S3Path:  "tenantpool/" + string(char) + ".tar.lz4",
		Path:    "data/db/tenant/" + string(char),
	}
}

func (tp *TenantPool) GetNTenantsDBs() int {
	l := 0
	tp.rwLock.RLock()
	l = len(tp.tenants)
	tp.rwLock.RUnlock()
	return l
}

func (tp *TenantPool) GetTenant(subdomain string) *TenantDB {
	tp.rwLock.RLock()
	tenant := tp.tenants[subdomain]
	tp.rwLock.RUnlock()

	return tenant
}

func (tp *TenantPool) RestoreAndOpenAllTenants(s3Client *S3Client, subdomains []string) error {
	if len(subdomains) == 0 {
		return nil
	}

	var err error
	// Check to see if tenants are already on disk
	for _, subdomain := range subdomains {
		_, err := tp.OpenTenant(subdomain)
		if err != nil {
			fmt.Printf("Error opening tenant db %v %v %v\n", string(tp.char), subdomain, err)
			return fmt.Errorf("cannot open tenant db: %w", err)
		}
		// We ignore if the DB was not found to be created
	}

	restored := tp.GetNTenantsDBs()
	if restored > 0 {
		if restored != len(subdomains) {
			fmt.Printf("Not all tenants were restored for Tennant Pool %v %v %v\n", string(tp.char), restored, len(subdomains))
			return fmt.Errorf("not all tenants were restored")
		}
		// If they should all be on disk if any are
		return nil
	}

	// Now we need to see if we want to restore from S3
	fmt.Println("Trying to restore tenant from backup for pool", string(tp.char))

	found, err := s3Client.DownloadAndExtractLZ4(tp.S3Path, tp.Path)
	fmt.Println("Finished s3 download backup for pool", string(tp.char))
	if !found && err == nil {
		fmt.Println("No tenant backup found")
		return nil
	}
	if err != nil {
		// TODO dont fail here if just not found
		return fmt.Errorf("cannot download tenant db: %w", err)
	}

	for _, subdomain := range subdomains {

		// At this point in time we are ignoring errors
		tenant, err := tp.OpenTenant(subdomain)
		if err != nil {
			fmt.Printf("Error opening tenant db %v %v %v\n", string(tp.char), subdomain, err)
			return fmt.Errorf("cannot open tenant db: %w", err)
		}
		if tenant == nil {
			fmt.Printf("Tenant db %v %v not found\n", string(tp.char), subdomain)
			return fmt.Errorf("cannot open tenant db: %w", err)
		}
	}

	return nil

}

func (tp *TenantPool) CreateTenant(subdomain string) (*TenantDB, error) {
	var err error
	tenantDB := NewTenantDB(subdomain)
	if tenantDB.DBFileExists() {
		return nil, fmt.Errorf("tenant db file already exists")
	}

	// Create the DB file
	err = tenantDB.Open(&mtsql.TenantCreate)
	if err != nil {
		return nil, fmt.Errorf("cannot create tenant db: %w", err)
	}

	// Add the tenant to the pool
	tp.rwLock.Lock()
	tp.tenants[subdomain] = tenantDB
	tp.rwLock.Unlock()

	return tenantDB, nil
}

func (tp *TenantPool) OpenTenant(subdomain string) (*TenantDB, error) {
	tenantDB := NewTenantDB(subdomain)
	if tenantDB.DBFileExists() {
		// If the DB file is there, write the tenant
		err := tenantDB.Open(nil)
		if err != nil {
			return nil, fmt.Errorf("cannot open tenant db: %w", err)
		}
		// Tenant db opened and ready to use
		tp.rwLock.Lock()
		tp.tenants[subdomain] = tenantDB
		tp.rwLock.Unlock()
		return tenantDB, nil
	}
	return nil, nil
}

func (tp *TenantPool) Close() error {
	tp.rwLock.Lock()
	defer tp.rwLock.Unlock()

	for _, tenant := range tp.tenants {
		err := tenant.Close()
		if err != nil {
			fmt.Println("cannot close tenant db", err)
		}
	}

	return nil
}

func (tp *TenantPool) BackupLoop(interval time.Duration, ctx context.Context, s3Client *S3Client) {
	syncTicker := time.NewTicker(interval)
	defer syncTicker.Stop()

	for {
		// Wait for ticker or context close.
		select {
		case <-ctx.Done():
			return
		case <-syncTicker.C:
		}

		fmt.Println("Backup Tenant loop")
		// Backup DB to S3
		if err := tp.backup(s3Client); err != nil && !errors.Is(err, context.Canceled) {
			fmt.Println("Backup error", "error", err)
		}

	}
}

func (tp *TenantPool) backup(s3Client *S3Client) error {
	defer trackTime(time.Now(), "backup tenant pool")

	tp.rwLock.RLock()
	defer tp.rwLock.RUnlock()

	var err error
	paths := []string{}
	for _, tenant := range tp.tenants {
		err = tenant.SqliteBackup()
		if err != nil {
			return fmt.Errorf("cannot backup tenant db: %w %v", err, tenant.Subdomain)
		}

		paths = append(paths, tenant.BackupPath)

	}
	if len(paths) == 0 {
		fmt.Println("No tenants to backup")
		return nil
	}
	err = s3Client.SyncAllFilesToSingleLZ4(tp.S3Path, paths)
	if err != nil {
		return fmt.Errorf("cannot sync tenant pool: %w %v", err, string(tp.char))
	}

	// Copy the control db to the backup location
	return nil
}

// //
// TenentDB
// //
type TenantDB struct {
	*SQLDB
	Subdomain string
	char      byte
}

func NewTenantDB(subdomain string) *TenantDB {
	char := subdomain[0]
	topPath := fmt.Sprintf("/tenant/%s/%s.sqlite", string(char), subdomain)
	tenantDB := &TenantDB{
		SQLDB: &SQLDB{
			Path:       "data/db" + topPath,
			BackupPath: "data/backup" + topPath,
			CacheSize:  500,
		},
		Subdomain: subdomain,
		char:      char,
	}

	return tenantDB
}

func (db *TenantDB) ValidateToken(token string) error {
	// If we have not seen this token in the last 10 mins update last_seen
	return nil
}

func (db *TenantDB) GetScoresSum() (int, error) {
	rows, err := db.Query("SELECT SUM(score) AS score FROM scores")
	if err != nil {
		return 0, fmt.Errorf("cannot get scores sum: %w", err)
	}
	if err = rows.Err(); err != nil {
		return 0, fmt.Errorf("cannot get scores sum: %w", err)
	}
	defer rows.Close()

	scoreRow := rows.Next()
	if !scoreRow {
		return 0, fmt.Errorf("cannot get scores sum: %w", err)
	}
	var score int
	err = rows.Scan(&score)
	if err != nil {
		return 0, fmt.Errorf("cannot get scores sum: %w", err)
	}

	return score, nil
}

// //
// UTILS
// //

const tenantCharSet = "abcdefghijklmnopqrstuvwxyz0123456789"

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func NextTenantID() string {
	result := make([]byte, 12)
	for i := range result {
		result[i] = tenantCharSet[seededRand.Intn(len(tenantCharSet))]
	}
	return string(result)
}
