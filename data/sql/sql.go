package sql

// gesql will
// 1. generate all prepared statements
// 2. then it will generate all migrations
// 3. then it will extract all statments and reference in a single file for each folder

//go:generate go run ./gen/gensql.go  -folder controldb
//go:generate go run ./gen/gensql.go  -folder sitedb
//go:generate go run ./gen/gensql.go  -folder jobsdb
//go:generate go run ./gen/gensql.go  -folder logsdb
//go:generate go run ./gen/gensql.go  -folder livedb

//go:generate go run ./gen/getters.go -folder controldb
//go:generate go run ./gen/getters.go -folder sitedb
//go:generate go run ./gen/getters.go -folder jobsdb
//go:generate go run ./gen/getters.go -folder logsdb
//go:generate go run ./gen/getters.go -folder livedb

//go:generate goverter gen ./controldb
//go:generate goverter gen ./sitedb
//go:generate goverter gen ./jobsdb

//go:generate go run ./gen/check.go
