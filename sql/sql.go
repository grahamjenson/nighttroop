package sql

//
// User
//

// devices
//go:generate go run ./gen/prepared.go  -file user/devices_get.sql -in "name string,token string" -out "name string,token string"
//go:generate go run ./gen/prepared.go  -file user/devices_create.sql -in "name string,token string" -out "name string,token string"
//go:generate go run ./gen/prepared.go  -file user/devices_all.sql  -out "name string,token string, last_seen_at time.Time"

//
// Control
//

// admins
//go:generate go run ./gen/prepared.go  -file control/admins_count.sql  -out "count int"
//go:generate go run ./gen/prepared.go  -file control/admins_create.sql -in "username string,password string" -out "username string,password string"
//go:generate go run ./gen/prepared.go  -file control/admins_get.sql -in "username string" -out "username string,password string, created_at time.Time"

// users
//go:generate go run ./gen/prepared.go  -file control/users_all.sql  -out "name string,token string, last_seen_at time.Time"
//go:generate go run ./gen/prepared.go  -file control/users_create.sql  -in "name string" -out "name string,token string, last_seen_at time.Time"
//go:generate go run ./gen/prepared.go  -file control/users_delete.sql  -in "name string" -out ""
//go:generate go run ./gen/prepared.go  -file control/users_find_unused_users.sql  -out "name string,token string, last_seen_at time.Time"
//go:generate go run ./gen/prepared.go  -file control/users_get.sql  -in "name string" -out "name string,token string, last_seen_at time.Time"
//go:generate go run ./gen/prepared.go  -file control/users_update_last_seen.sql  -in "name string" -out ""

//
// MIGRATIONS
//

// Control
//go:generate go run ./gen/migration.go -name control -file migrations/control_0001.sql -file migrations/control_0002.sql

// Users
//go:generate go run ./gen/migration.go -name user -file migrations/user_0001.sql
