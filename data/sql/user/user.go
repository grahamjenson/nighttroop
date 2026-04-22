package sql

//////////////
// User
///////////////

// devices
//go:generate go run ../gen/prepared.go  -file devices/devices_get.sql -in "Name string" -out "Name string, Password string, TotpSecret string, Role string, CreatedAt *time.Time, LastSeenAt *time.Time"
//go:generate go run ../gen/prepared.go  -file devices/devices_create.sql -in "Name string, Password string, TotpSecret string, Role string"
//go:generate go run ../gen/prepared.go  -file devices/devices_all.sql  -out "Name string, Password string, TotpSecret string, Role string, CreatedAt *time.Time, LastSeenAt *time.Time"
//go:generate go run ../gen/prepared.go  -file devices/devices_update_last_seen_at.sql  -in "Name string"

//////////////
// MIGRATIONS
///////////////

// Users
//go:generate go run ../gen/migration.go -name user -file migrations/user_0001.sql

//////////////
// Combined
//////////////
