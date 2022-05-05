package render

import (
	_ "embed"
	"html/template"
)

const (
	messageBodyIndent       = 18
	messageParentBodyIndent = 20
)

//go:embed template.html.tmpl
var templateString string

var Template = template.Must(template.New("yg-render").Parse(templateString))
