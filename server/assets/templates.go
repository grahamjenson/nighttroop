package templates

import (
	_ "embed"
)

// Public Templates
//go:generate go run ./gen/template.go -base public/base.html -file public/create_user.html
//go:generate go run ./gen/template.go -base public/created_user_info.htmx                     -in "url string, token string, backup_key string"
//go:generate go run ./gen/template.go -base public/create_user_rate_limited.htmx              -in "Seconds string"

// Control Templates
//go:generate go run ./gen/template.go -base control/base.html -file control/login.html
//go:generate go run ./gen/template.go -base control/base.html -file control/dashboard.html

// Static Assets

//go:generate go run ./gen/static.go -file favicon.ico -file htmx.js -file bootstrap.js -file bootstrap.css -file logo.png
