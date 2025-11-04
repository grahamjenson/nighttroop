package db

import (
	"fmt"
	"sort"
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

func (tp *TenantPool) allTenants() []*TenantDB {
	tenants := make([]*TenantDB, 0)
	tp.rwLock.RLock()
	for _, tenant := range tp.tenants {
		tenants = append(tenants, tenant)
	}
	tp.rwLock.RUnlock()
	sort.Slice(tenants, func(i, j int) bool {
		return tenants[i].Subdomain < tenants[j].Subdomain
	})
	return tenants
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

func (tp *TenantPool) CreateTenant(subdomain string) (*TenantDB, string, error) {
	var err error
	tenantDB := NewTenantDB(subdomain)
	if tenantDB.DBFileExists() {
		return nil, "", fmt.Errorf("tenant db file already exists")
	}

	// Create the DB file
	err = tenantDB.Open(&mtsql.TenantCreate)
	if err != nil {
		return nil, "", fmt.Errorf("cannot create tenant db: %w", err)
	}

	apiKey, err := tenantDB.CreateAPIToken()
	if err != nil {
		tenantDB.Close()
		return nil, "", fmt.Errorf("cannot create tenant db: %w", err)
	}

	// Add the tenant to the pool
	tp.rwLock.Lock()
	tp.tenants[subdomain] = tenantDB
	tp.rwLock.Unlock()

	return tenantDB, apiKey, nil
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

func (tp *TenantPool) backup(s3Client *S3Client) error {
	tenants := tp.allTenants()
	if len(tenants) == 0 {
		return nil
	}
	defer trackTime(time.Now(), "backup tenant pool "+string(tp.char))

	var err error
	paths := []string{}

	for _, tenant := range tenants {
		// Have to add it anyway because we need all paths
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

func (tp *TenantPool) updateModified() (bool, error) {

	tenants := tp.allTenants()
	if len(tenants) == 0 {
		return false, nil
	}

	anyModified := false
	for _, tenant := range tenants {
		modified, err := tenant.UpdateModified()
		if err != nil {
			// TODO error group
			fmt.Println("cannot update tenant modified db:", tenant.Subdomain, err)
		}
		if modified {
			anyModified = true
		}
	}
	return anyModified, nil
}

// This method looks though all the ages of the DBs and removes tenants that are older than the max age from the circulation
// TODO It should then group these tenants up together, zip them and then upload them to S3 while logging their location for future hydration
func (tp *TenantPool) archive(s3Client *S3Client, maxTenantAge time.Duration) ([]*TenantDB, error) {
	removeList := []*TenantDB{}

	tenants := tp.allTenants()
	if len(tenants) == 0 {
		return removeList, nil
	}

	for _, tenant := range tenants {
		// Check if the tenant is older than the max age
		lm := time.Unix(tenant.lastModifiedAt, 0)
		if time.Since(lm) > maxTenantAge {
			// Archive the tenant
			removeList = append(removeList, tenant)
		}
	}
	tp.rwLock.Lock()
	defer tp.rwLock.Unlock()
	for _, tenant := range removeList {
		fmt.Println("Removing tenant", tenant.Subdomain)
		delete(tp.tenants, tenant.Subdomain)
		err := tenant.Close()
		if err != nil {
			fmt.Println("cannot close tenant db", err)
		}

		err = s3Client.SyncAllFilesToSingleLZ4(tenant.ArchivePath(), []string{tenant.Path})
		if err != nil {
			fmt.Println("cannot sync to s3 tenant archive", err)
		}
	}

	return removeList, nil
}
