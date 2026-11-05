package main

import (
	"bytes"
	"encoding/json"
	"html/template"
	"os"
	"reflect"
	"slices"
	"sort"
	"strings"

	"github.com/yuin/goldmark"
)

// schema is the part of an OpenAPI schema the API docs show
type schema struct {
	Ref         string             `json:"$ref"`
	Type        string             `json:"type"`
	Format      string             `json:"format"`
	Description string             `json:"description"`
	Enum        []any              `json:"enum"`
	Items       *schema            `json:"items"`
	Properties  map[string]*schema `json:"properties"`
	Required    []string           `json:"required"`
	AllOf       []*schema          `json:"allOf"`
	ReadOnly    bool               `json:"readOnly"`
}

type parameter struct {
	Ref         string  `json:"$ref"`
	Name        string  `json:"name"`
	In          string  `json:"in"`
	Required    bool    `json:"required"`
	Description string  `json:"description"`
	Schema      *schema `json:"schema"`
}

// body is a request body or a response
type body struct {
	Ref         string                              `json:"$ref"`
	Description string                              `json:"description"`
	Content     map[string]struct{ Schema *schema } `json:"content"`
}

type operation struct {
	OperationID string                `json:"operationId"`
	Summary     string                `json:"summary"`
	Description string                `json:"description"`
	Parameters  []*parameter          `json:"parameters"`
	RequestBody *body                 `json:"requestBody"`
	Responses   map[string]*body      `json:"responses"`
	Security    []map[string][]string `json:"security"`
}

type spec struct {
	Info     struct{ Title, Version, Description string }
	Security []map[string][]string `json:"security"`
	// Paths are each path's operations by method, and the parameters they all have
	Paths      map[string]map[string]json.RawMessage `json:"paths"`
	Components struct {
		SecuritySchemes map[string]struct{ Type, Scheme, In, Name, Description string } `json:"securitySchemes"`
		Parameters      map[string]*parameter                                           `json:"parameters"`
		RequestBodies   map[string]*body                                                `json:"requestBodies"`
		Responses       map[string]*body                                                `json:"responses"`
		Schemas         map[string]*schema                                              `json:"schemas"`
	}
}

var methods = []string{"get", "post", "put", "patch", "delete"}

// endpoint is one operation, with its path's parameters and its own, and how it is signed in to
type endpoint struct {
	*operation
	Method, Path string
	Params       []*parameter
	SignIn       []string
}

// refName is the last part of a $ref, the component it names
func refName(ref string) string { return ref[strings.LastIndex(ref, "/")+1:] }

// markdown is a description as HTML, as descriptions are CommonMark
func markdown(s string) template.HTML {
	var out bytes.Buffer
	if err := goldmark.Convert([]byte(s), &out); err != nil {
		panic(err)
	}
	return template.HTML(out.String())
}

// typeOf describes a schema's type, linking to the schema a $ref names
func typeOf(s *schema) template.HTML {
	switch {
	case s == nil:
		return ""
	case s.Ref != "":
		name := template.HTMLEscapeString(refName(s.Ref))
		return template.HTML(`<a href="#schema-` + name + `">` + name + `</a>`)
	case s.Type == "array":
		return "array of " + typeOf(s.Items)
	case len(s.Properties) > 0:
		fields := []string{}
		for _, name := range sortedKeys(s.Properties) {
			fields = append(fields, template.HTMLEscapeString(name)+": "+string(typeOf(s.Properties[name])))
		}
		return template.HTML("{" + strings.Join(fields, ", ") + "}")
	}
	t := template.HTMLEscapeString(s.Type)
	if s.Format != "" {
		t += " (" + template.HTMLEscapeString(s.Format) + ")"
	}
	if len(s.Enum) > 0 {
		values := []string{}
		for _, v := range s.Enum {
			values = append(values, template.HTMLEscapeString(strings.Trim(string(must(json.Marshal(v))), `"`)))
		}
		t += ", one of " + strings.Join(values, ", ")
	}
	return template.HTML(t)
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

// sortedKeys is a map's keys in order, so the docs are the same each time they are generated
func sortedKeys(m any) []string {
	keys := []string{}
	for _, k := range reflect.ValueOf(m).MapKeys() {
		keys = append(keys, k.String())
	}
	sort.Strings(keys)
	return keys
}

// apiDocs renders the OpenAPI schema as the API docs' HTML: the API's description, a contents list, each endpoint
// with its parameters, body, responses and how it is signed in to, then each schema, in a stable order
func apiDocs(path string) string {
	var api spec
	if err := json.Unmarshal(must(os.ReadFile(path)), &api); err != nil {
		panic(err)
	}
	c := api.Components
	param := func(p *parameter) *parameter {
		if p.Ref != "" {
			return c.Parameters[refName(p.Ref)]
		}
		return p
	}
	endpoints := []endpoint{}
	for _, path := range sortedKeys(api.Paths) {
		item, shared := api.Paths[path], []*parameter{}
		if raw, ok := item["parameters"]; ok {
			must(0, json.Unmarshal(raw, &shared))
		}
		for _, method := range methods {
			raw, ok := item[method]
			if !ok {
				continue
			}
			e := endpoint{operation: &operation{}, Method: strings.ToUpper(method), Path: path}
			must(0, json.Unmarshal(raw, e.operation))
			for _, p := range append(slices.Clone(shared), e.Parameters...) {
				e.Params = append(e.Params, param(p))
			}
			security := e.Security
			if security == nil {
				security = api.Security
			}
			for _, requirement := range security {
				for _, name := range sortedKeys(requirement) {
					scheme, signIn := c.SecuritySchemes[name], "a JWT as Authorization: Bearer"
					if scheme.Type == "apiKey" {
						signIn = "the " + scheme.Name + " " + scheme.In
						if d := scheme.Description; d != "" {
							signIn = strings.ToLower(d[:1]) + d[1:] + ", as " + signIn
						}
					}
					e.SignIn = append(e.SignIn, signIn)
				}
			}
			endpoints = append(endpoints, e)
		}
	}
	funcs := template.FuncMap{
		"markdown": markdown, "type": typeOf, "keys": sortedKeys,
		"body": func(b *body) *body {
			if b != nil && b.Ref != "" {
				if requestBody, ok := c.RequestBodies[refName(b.Ref)]; ok {
					return requestBody
				}
				return c.Responses[refName(b.Ref)]
			}
			return b
		},
		"required": func(s *schema, name string) bool { return slices.Contains(s.Required, name) },
	}
	page := template.Must(template.New("api").Funcs(funcs).Parse(apiDocsTemplate))
	var out bytes.Buffer
	data := struct {
		Info      any
		Endpoints []endpoint
		Schemas   map[string]*schema
	}{api.Info, endpoints, c.Schemas}
	if err := page.Execute(&out, data); err != nil {
		panic(err)
	}
	return out.String()
}

const apiDocsTemplate = `<h1>{{.Info.Title}} API <small>{{.Info.Version}}</small></h1>
{{markdown .Info.Description}}
<p><a id="api-docs--download" href="/openapi.json" download="openapi.json">Download the OpenAPI schema</a></p>
<h2>Endpoints</h2>
<ul>{{range .Endpoints}}
<li><a href="#{{.OperationID}}"><code>{{.Method}} {{.Path}}</code></a> {{.Summary}}</li>{{end}}
</ul>
{{range .Endpoints}}
<section id="{{.OperationID}}">
<h2><code>{{.Method}} {{.Path}}</code></h2>
<p>{{if .Summary}}{{.Summary}}{{else}}{{.OperationID}}{{end}}</p>
{{if .Description}}{{markdown .Description}}{{end}}
{{with .SignIn}}<p>Signed in with {{range $i, $s := .}}{{if $i}}, or {{end}}{{$s}}{{end}}</p>{{else}}<p>No sign in needed</p>{{end}}
{{with .Params}}<h3>Parameters</h3>
<table><tr><th>Name</th><th>In</th><th>Type</th><th>Description</th></tr>{{range .}}
<tr><td><code>{{.Name}}</code>{{if .Required}} (required){{end}}</td><td>{{.In}}</td><td>{{type .Schema}}</td><td>{{.Description}}</td></tr>{{end}}
</table>{{end}}
{{with body .RequestBody}}<h3>Body</h3>{{range $type, $content := .Content}}<p><code>{{$type}}</code>: {{type $content.Schema}}</p>{{end}}{{end}}
<h3>Responses</h3>
<table><tr><th>Status</th><th>Description</th><th>Body</th></tr>{{range $code, $response := .Responses}}{{with body $response}}
<tr><td>{{$code}}</td><td>{{.Description}}</td><td>{{range .Content}}{{type .Schema}}{{end}}</td></tr>{{end}}{{end}}
</table>
</section>{{end}}
<h2>Schemas</h2>
{{range $name := keys .Schemas}}{{$s := index $.Schemas $name}}
<section id="schema-{{$name}}">
<h3>{{$name}}</h3>
{{if $s.Description}}{{markdown $s.Description}}{{end}}
{{range $s.AllOf}}{{if .Ref}}<p>Has every field of {{type .}}</p>{{end}}{{end}}
{{if or $s.Properties $s.AllOf}}<table><tr><th>Field</th><th>Type</th><th>Description</th></tr>{{range $s.AllOf}}{{if not .Ref}}{{template "fields" .}}{{end}}{{end}}{{template "fields" $s}}
</table>{{else}}<p>{{type $s}}</p>{{end}}
</section>{{end}}
{{define "fields"}}{{$s := .}}{{range $field := keys .Properties}}{{$p := index $s.Properties $field}}
<tr><td><code>{{$field}}</code>{{if required $s $field}} (required){{end}}{{if $p.ReadOnly}} (read only){{end}}</td><td>{{type $p}}</td><td>{{$p.Description}}</td></tr>{{end}}{{end}}
`
