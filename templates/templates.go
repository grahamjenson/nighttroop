package templates

import (
	_ "embed"
	"html/template"

	"sqlitemt.t1d.nz/templates/control"
	"sqlitemt.t1d.nz/templates/tenant"
)

//go:embed base.html
var baseHTML string

var ControlHTMLTemplate *template.Template
var TenantHTMLTemplate *template.Template

func CreateTemplates() error {
	var err error
	ControlHTMLTemplate, err = template.New("indexHTML").Parse(baseHTML + control.ContentHTMX + control.TenantHTMX)
	if err != nil {
		return err
	}

	TenantHTMLTemplate, err = template.New("indexHTML").Parse(baseHTML + tenant.ContentHTMX + tenant.GuessHTMX)
	if err != nil {
		return err
	}
	return nil
}
