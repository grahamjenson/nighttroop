package templates

import (
	_ "embed"
)

// Public Templates
//go:generate go run ./gen/template.go -base public/base.html -file public/create_user.html
//go:generate go run ./gen/template.go -base public/created_user_info.htmx                     -in "url string, password string, totp_secret string, backup_key string"
//go:generate go run ./gen/template.go -base public/create_user_rate_limited.htmx              -in "seconds string"

// Control Templates
//go:generate go run ./gen/template.go -base control/base.html -file control/login.html
//go:generate go run ./gen/template.go -base control/base.html -file control/dashboard.html

// User Templates
//go:generate go run ./gen/template.go -base user/base.html -file user/login.html             -in "Subdomain string"
//go:generate go run ./gen/template.go -base user/login_totp.htmx
//go:generate go run ./gen/template.go -base user/base.html -file user/dashboard.html         -in "Subdomain string"

// Static Assets

//go:generate go run ./gen/static.go
