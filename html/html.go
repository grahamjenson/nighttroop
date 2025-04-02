package html

import (
	_ "embed"
	"html/template"
)

//go:embed index.html.tpl
var indexHTML string

var IndexHTMLTemplate *template.Template

func CreateTemplates() error {
	var err error
	IndexHTMLTemplate, err = template.New("indexHTML").Parse(indexHTML)
	if err != nil {
		return err
	}

	return nil
}
