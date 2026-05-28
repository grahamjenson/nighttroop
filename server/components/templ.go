package components

//go:generate templ generate -include-version=false

//go:generate tailwindcss -i ./css/input.css -o ../assets/static/style_gen.css --minify

// Static Assets
//go:generate sh -c "(cd ../assets && go run ./gen/static.go)"
