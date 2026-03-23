package templates

import (
	_ "embed"
)

// Public Templates
//go:generate go run ./gen/template.go -name public_index -file base.html -file public/index.html
//go:generate go run ./gen/template.go -name public_create_user -in "url string, token string, backup_key string" -file public/create_user.htmx

// Control Templates
//go:generate go run ./gen/template.go -name control_index -file base.html -file control/index.html

// Static Assets

//go:generate go run ./gen/static.go -file favicon.ico -file htmx.js -file bootstrap.js -file bootstrap.css -file logo.png
