package control

import (
	_ "embed"
)

//go:embed content.html
var ContentHTMX string

//go:embed htmx_tenant.html
var TenantHTMX string

type TenantHTMXParam struct {
	Subdomain string
	APIKey    string
}
