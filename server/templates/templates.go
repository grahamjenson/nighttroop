package templates

import (
	_ "embed"
)

//go:generate go run ./gen/template.go -folder control
//go:generate go run ./gen/template.go -folder user
