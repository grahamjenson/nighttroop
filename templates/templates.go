package templates

import (
	_ "embed"
	"html/template"
)

//go:embed index.html
var indexHTML string

//go:embed base.html
var baseHTML string

//go:embed tenant.html
var tenantHTML string

var IndexHTMLTemplate *template.Template

func CreateTemplates() error {
	var err error
	IndexHTMLTemplate, err = template.New("indexHTML").Parse(baseHTML + indexHTML + tenantHTML)
	if err != nil {
		return err
	}

	return nil
}
