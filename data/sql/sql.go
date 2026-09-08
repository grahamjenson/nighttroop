package sql

// gesql will
// 1. generate all prepared statements
// 2. then it will generate all migrations
// 3. then it will extract all statments and reference in a single file for each folder

//go:generate go build -o ../../bin/getters ./gen/getters.go
//go:generate go run ./gen/gensql.go  -folder control
//go:generate go run ./gen/gensql.go  -folder site
//go:generate go run ./gen/gensql.go  -folder jobs

//go:generate goverter gen ./control
//go:generate goverter gen ./site
//go:generate goverter gen ./jobs

//go:generate go run ./gen/check.go
