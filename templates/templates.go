package templates

import (
	_ "embed"
	"html/template"

	"nighttroop.t1d.nz/templates/control"
	"nighttroop.t1d.nz/templates/user"
)

//go:embed base.html
var baseHTML string

var ControlHTMLTemplate *template.Template
var UserHTMLTemplate *template.Template

func CreateTemplates() error {
	var err error
	ControlHTMLTemplate, err = template.New("indexHTML").Parse(baseHTML + control.ContentHTMX + control.UserHTMX)
	if err != nil {
		return err
	}

	UserHTMLTemplate, err = template.New("indexHTML").Parse(baseHTML + user.ContentHTMX + user.GuessHTMX)
	if err != nil {
		return err
	}
	return nil
}
