package templates

import (
	_ "embed"
)

//go:generate go run ./gen/template.go -folder control
//go:generate go run ./gen/template.go -folder user
//go:generate go run ./gen/template.go -folder public

// Static Assets

//go:generate go run ./gen/static.go
