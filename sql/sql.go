package sql

import _ "embed"

///
//  Control
///

//go:embed control_create.sql
var ControlCreate string

//go:embed control_tenant_insert.sql
var ControlTenantInsert string

//go:embed control_tenant_delete.sql
var ControlTenantDelete string

///
//  User
///

//go:embed tenant_create.sql
var TenantCreate string

//go:embed tenant_scores_insert.sql
var TenantScoresInsert string

//go:embed tenant_scores_read.sql
var TenantRead string

//go:embed tenant_scores_sum.sql
var TenantSum string

//go:embed tenant_create_api_key.sql
var TenantCreateAPIKey string
