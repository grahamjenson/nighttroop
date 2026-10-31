// Package nighttroop is a client of Nighttroop's JSON API, generated from its OpenAPI schema, openapi.json, whose
// request and response bodies are JSON Schema. Add an endpoint to openapi.json and run go generate to use it.
package nighttroop

//go:generate go tool oapi-codegen -generate types,client -package nighttroop -o client.gen.go openapi.json
