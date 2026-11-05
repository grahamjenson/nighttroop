package nighttroop

import _ "embed"

// OpenAPI is the API's OpenAPI schema, served at /openapi.json and shown as the API docs
//
//go:embed openapi.json
var OpenAPI []byte
