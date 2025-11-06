package sql

import _ "embed"

///
//  Control
///

//go:embed control_create.sql
var ControlCreate string

//go:embed control_user_insert.sql
var ControlUserInsert string

//go:embed control_user_delete.sql
var ControlUserDelete string

///
//  User
///

//go:embed user_create.sql
var UserCreate string

//go:embed user_create_api_key.sql
var UserCreateAPIKey string
