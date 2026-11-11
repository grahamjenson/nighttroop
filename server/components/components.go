// Package components contains the http html components
package components

// templ is the go.mod tool, so its generated code always matches the templ library it is built with
//go:generate go tool templ generate -include-version=false

//go:generate tailwindcss -i ./css/input.css -o ../assets/static/style_gen.css

// Static Assets
//go:generate sh -c "(cd ../assets && go run ./gen/docs)"
//go:generate sh -c "(cd ../assets && go run ./gen/static.go)"
