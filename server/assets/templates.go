package templates

import (
	_ "embed"
)

// Public Templates
//go:generate go run ./gen/template.go -name public_create_user -base base.html -file public/create_user.html
//go:generate go run ./gen/template.go -name public_created_user_info -in "url string, token string, backup_key string" -base public/created_user_info.htmx

// Control Templates
//go:generate go run ./gen/template.go -name control_login -base base.html -file control/login.html
//go:generate go run ./gen/template.go -name control_dashboard  -base base.html -file control/dashboard.html

// Static Assets

//go:generate go run ./gen/static.go -file favicon.ico -file htmx.js -file bootstrap.js -file bootstrap.css -file logo.png
