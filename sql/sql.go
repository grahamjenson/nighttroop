package sql

//
// User
//

//go:generate go run ./gen/prepared.go  -file user/devices_get.sql -in "name string,token string" -out "name string,token string"
//go:generate go run ./gen/prepared.go  -file user/devices_create.sql -in "name string,token string" -out "name string,token string"
//go:generate go run ./gen/prepared.go  -file user/devices_all.sql  -out "name string,token string, last_seen_at time.Time"

//
// Control
//

//go:generate go run ./gen/prepared.go  -file control/admins_get.sql -in "username string" -out "username string,password string, created_at time.Time"
//go:generate go run ./gen/prepared.go  -file control/admins_create.sql -in "username string,password string" -out "username string,password string"
//go:generate go run ./gen/prepared.go  -file control/admins_count.sql  -out "count int"
