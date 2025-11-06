package tenant

import (
	_ "embed"
)

//go:embed content.html
var ContentHTMX string

//go:embed htmx_guess.html
var GuessHTMX string

type GuessHTMXParam struct {
	Subdomain string
	APIKey    string
}
