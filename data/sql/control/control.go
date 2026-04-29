package control

////////////
// Control
//////////////

// admins
//go:generate go run ../gen/prepared.go  -file admins/create.sql          -in "Name string,Password string,Role string,TotpSecret string"
//go:generate go run ../gen/prepared.go  -file admins/get.sql             -in "Name string" -out "Name string,Password string,Role string,TotpSecret string,LastSeenAt *time.Time,CreatedAt *time.Time"

// users
//go:generate go run ../gen/prepared.go  -file users/all.sql               -out "Subdomain string, CreatedAt *time.Time, LastSeenAt *time.Time, DeletedAt *time.Time"
//go:generate go run ../gen/prepared.go  -file users/not_deleted.sql       -out "Subdomain string, CreatedAt *time.Time, LastSeenAt *time.Time, DeletedAt *time.Time"

//go:generate go run ../gen/prepared.go  -file users/get_batch.sql         -in "batch string" -out "Subdomain string, LastSeenAt *time.Time"
//go:generate go run ../gen/prepared.go  -file users/create.sql            -in "Subdomain string, BackupKeyHash string"
//go:generate go run ../gen/prepared.go  -file users/delete.sql            -in "SubdomainsJSONList string"

//go:generate go run ../gen/prepared.go  -file users/get_inactive.sql      -in "lastSeenMinutes float64, lastWriteMinutes float64"  -out "Subdomain string"
//go:generate go run ../gen/prepared.go  -file users/get_never_seen.sql    -in "lastSeenMinutes float64, lastWriteMinutes float64"  -out "Subdomain string"

//go:generate go run ../gen/prepared.go  -file users/get.sql               -in "Subdomain string" -out "Subdomain string, CreatedAt *time.Time, LastSeenAt *time.Time"
//go:generate go run ../gen/prepared.go  -file users/update_last_seen.sql  -in "Subdomain string"
//go:generate go run ../gen/prepared.go  -file users/update_last_write.sql -in "Subdomain string"

//go:generate go run ../gen/prepared.go  -file test/test_users_truncate_all.sql

//////////////
// MIGRATIONS
///////////////

// Control
//go:generate go run ../gen/migration.go -name control -file migrations/control_0001.sql

//////////////
// Combined
//////////////
