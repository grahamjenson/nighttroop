package sql

// gesql will
// 1. generate all prepared statements
// 2. then it will generate all migrations
// 3. then it will extract all statments and reference in a single file for each folder

//go:generate go build -o ../../bin/getters ./gen/getters.go
//go:generate go run ./gen/gensql.go  -folder controldb
//go:generate go run ./gen/gensql.go  -folder sitedb
//go:generate go run ./gen/gensql.go  -folder jobsdb

//go:generate goverter gen ./controldb
//go:generate goverter gen ./sitedb
//go:generate goverter gen ./jobsdb

//go:generate go run ./gen/check.go
