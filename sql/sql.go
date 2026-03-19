package sql

//
// User
//

// devices
//go:generate go run ./gen/prepared.go  -file user/devices_get.sql -in "name string,token string" -out "name string,token string"
//go:generate go run ./gen/prepared.go  -file user/devices_create.sql -in "name string,token string"
//go:generate go run ./gen/prepared.go  -file user/devices_all.sql  -out "name string,token string, LastSeenAt *time.Time"

//
// Control
//

// admins
//go:generate go run ./gen/prepared.go  -file control/admins_create.sql -in "name string,password string,role string,totp_secret string"
//go:generate go run ./gen/prepared.go  -file control/admins_get.sql -in "name string" -out "name string,password string,role string,totp_secret string,last_seen_at *time.Time,created_at *time.Time"

// users
//go:generate go run ./gen/prepared.go  -file control/users_all.sql  -out "Subdomain string, CreatedAt *time.Time, LastSeenAt *time.Time, DeletedAt *time.Time"
//go:generate go run ./gen/prepared.go  -file control/users_not_deleted.sql  -out "Subdomain string, CreatedAt *time.Time, LastSeenAt *time.Time, DeletedAt *time.Time"

//go:generate go run ./gen/prepared.go  -file control/users_get_batch.sql  -in "batch string" -out "Subdomain string, LastSeenAt *time.Time"
//go:generate go run ./gen/prepared.go  -file control/users_create.sql  -in "Subdomain string, BackupKeyHash string"
//go:generate go run ./gen/prepared.go  -file control/users_delete.sql  -in "Subdomain string"
//go:generate go run ./gen/prepared.go  -file control/users_get_inactive.sql -in "minutes float64"  -out "Subdomain string"
//go:generate go run ./gen/prepared.go  -file control/users_get.sql  -in "Subdomain string" -out "Subdomain string, CreatedAt *time.Time, LastSeenAt *time.Time"
//go:generate go run ./gen/prepared.go  -file control/users_update_last_seen.sql  -in "Subdomain string"
//go:generate go run ./gen/prepared.go  -file control/users_truncate_all.sql

//
// MIGRATIONS
//

// Control
//go:generate go run ./gen/migration.go -name control -file migrations/control_0001.sql
// Users
//go:generate go run ./gen/migration.go -name user -file migrations/user_0001.sql
