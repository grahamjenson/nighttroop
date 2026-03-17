package sql

//
// User
//

// devices
//go:generate go run ./gen/prepared.go  -file user/devices_get.sql -in "name string,token string" -out "name string,token string"
//go:generate go run ./gen/prepared.go  -file user/devices_create.sql -in "name string,token string" -out "name string,token string"
//go:generate go run ./gen/prepared.go  -file user/devices_all.sql  -out "name string,token string, LastSeenAt time.Time"

//
// Control
//

// admins
//go:generate go run ./gen/prepared.go  -file control/admins_count.sql  -out "count int"
//go:generate go run ./gen/prepared.go  -file control/admins_create.sql -in "username string,password string" -out "username string,password string"
//go:generate go run ./gen/prepared.go  -file control/admins_get.sql -in "username string" -out "username string,password string, created_at time.Time"

// users
//go:generate go run ./gen/prepared.go  -file control/users_all.sql  -out "Subdomain string, CreatedAt *time.Time, LastSeenAt *time.Time, DeletedAt *time.Time"
//go:generate go run ./gen/prepared.go  -file control/users_get_batch.sql  -in "batch string" -out "Subdomain string, LastSeenAt time.Time"
//go:generate go run ./gen/prepared.go  -file control/users_create.sql  -in "Subdomain string, BackupKeyHash string" -out "Subdomain string, LastSeenAt time.Time"
//go:generate go run ./gen/prepared.go  -file control/users_delete.sql  -in "Subdomain string" -out ""
//go:generate go run ./gen/prepared.go  -file control/users_get_inactive.sql -in "minutes float64"  -out "Subdomain string, LastSeenAt time.Time"
//go:generate go run ./gen/prepared.go  -file control/users_get.sql  -in "Subdomain string" -out "Subdomain string, LastSeenAt time.Time"
//go:generate go run ./gen/prepared.go  -file control/users_update_last_seen.sql  -in "Subdomain string, LastSeenAt time.Time" -out ""
//go:generate go run ./gen/prepared.go  -file control/users_truncate_all.sql  -in "" -out ""

//
// MIGRATIONS
//

// Control
//go:generate go run ./gen/migration.go -name control -file migrations/control_0001.sql -file migrations/control_0002.sql

// Users
//go:generate go run ./gen/migration.go -name user -file migrations/user_0001.sql
