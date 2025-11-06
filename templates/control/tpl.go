package control

import (
	_ "embed"
)

//go:embed content.html
var ContentHTMX string

//go:embed htmx_tenant.html
var UserHTMX string

type UserHTMXParam struct {
	Subdomain string
	APIKey    string
}
