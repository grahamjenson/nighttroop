package db

import "fmt"

type TenantDB struct {
	*SQLDB
	Subdomain string
}

func NewTenantDB(subdomain string) *TenantDB {
	firstChar := string(subdomain[0])
	topPath := fmt.Sprintf("/tenant/%s/%s.sqlite", firstChar, subdomain)
	tenantDB := &TenantDB{
		SQLDB: &SQLDB{
			Path:       "data/db" + topPath,
			BackupPath: "data/backup" + topPath,
			CacheSize:  200,
		},
		Subdomain: subdomain,
	}

	return tenantDB
}

func (db *TenantDB) ValidateToken(token string) error {
	// If we have not seen this token in the last 10 mins update last_seen
	return nil
}

func (db *TenantDB) GetScoresSum() (int, error) {
	return 0, nil
}
