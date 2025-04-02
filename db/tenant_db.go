package db

type TenantDB struct {
	*SQLDB
	subdomain string
}

func InitTenantDB(subdomain string) (*TenantDB, error) {
	tenantDB := &TenantDB{
		SQLDB: &SQLDB{Path: "data/db/tenant/" + subdomain + ".sqlite"},
	}

	return tenantDB, nil
}

func (db *TenantDB) ValidateToken(token string) error {
	// If we have not seen this token in the last 10 mins update last_seen
	return nil
}

func (db *TenantDB) GetScoresSum() (int, error) {
	return 0, nil
}
