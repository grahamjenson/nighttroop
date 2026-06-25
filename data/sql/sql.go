package sql

// gesql will
// 1. generate all prepared statements
// 2. then it will generate all migrations
// 3. then it will extract all statments and reference in a single file for each folder

//go:generate go build -o ../../bin/getters ./gen/getters.go
//go:generate go run ./gen/gensql.go  -folder control
//go:generate go run ./gen/gensql.go  -folder user

//go:generate goverter gen ./user

//go:generate go run ./gen/check.go
