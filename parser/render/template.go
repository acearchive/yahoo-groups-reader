package render

import (
	_ "embed"
	"github.com/Masterminds/sprig/v3"
	"html/template"
)

const (
	messageBodyIndent       = 18
	messageParentBodyIndent = 20
)

//go:embed template.html.tmpl
var templateString string

var Template = template.Must(template.New("yg-render").Funcs(sprig.FuncMap()).Parse(templateString))
