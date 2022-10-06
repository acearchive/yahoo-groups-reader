package render

import (
	_ "embed"
	"fmt"
	"github.com/Masterminds/sprig/v3"
	"html/template"
)

const (
	messageBodyIndent       = 18
	messageParentBodyIndent = 20
)

//go:embed template.html.tmpl
var templateString string

var Template *template.Template

func init() {
	templateFunctions := sprig.FuncMap()

	templateFunctions["comment"] = func(text string) template.HTML {
		return template.HTML(fmt.Sprintf("<!-- %s -->", text))
	}

	Template = template.Must(template.New("yahoo-groups-reader").Funcs(templateFunctions).Parse(templateString))
}
