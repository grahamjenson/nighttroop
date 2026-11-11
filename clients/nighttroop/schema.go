package nighttroop

import (
	"bytes"
	_ "embed"
)

// OpenAPI is the API's OpenAPI schema, served at /openapi.json and shown as the API docs
//
//go:embed openapi.json
var OpenAPI []byte

// SchemaHost is the host the schema's servers are on, as each server serving it is on its own
const SchemaHost = "nighttroop.example"

// OpenAPIOn is the schema with its servers on a server's host, from its linkURL, e.g. a site's Nightscout host
// https://{site_id}.ns.staging.nighttroop.nz, so the API docs show where each endpoint is
func OpenAPIOn(linkURL func(subdomain string) string) []byte {
	schema := OpenAPI
	for _, subdomain := range []string{"", "{site_id}", "{site_id}.ns"} {
		host := SchemaHost
		if subdomain != "" {
			host = subdomain + "." + host
		}
		schema = bytes.ReplaceAll(schema, []byte(`"url": "https://`+host+`"`), []byte(`"url": "`+linkURL(subdomain)+`"`))
	}
	return schema
}
