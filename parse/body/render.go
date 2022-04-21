package body

import (
	"html"
	"html/template"
	"io"
	"strings"
)

const IndentPrefix = "  "

const messageHeaderTemplateString = `
<div class="inline-message-header">
  <dl class="field-list">
  {{ range .Fields -}}
    <dt>{{ .Name }}</dt>
    <dd>{{ .Value }}</dd>
  {{ end }}
  </dl>
</div>
`

var messageHeaderTemplate = template.Must(template.New("inline-message-header").Parse(messageHeaderTemplateString))

type messageHeaderTemplateField struct {
	Name  string
	Value string
}

type messageHeaderTemplateParams struct {
	Fields []messageHeaderTemplateField
}

func (b MessageHeaderBlock) ToHtml() string {
	params := messageHeaderTemplateParams{
		Fields: make([]messageHeaderTemplateField, 0, len(b)),
	}

	for name, value := range b {
		params.Fields = append(params.Fields, messageHeaderTemplateField{Name: name, Value: value})
	}

	var output strings.Builder

	if err := messageHeaderTemplate.Execute(&output, params); err != nil {
		panic(err)
	}

	return output.String()
}

func (b SignatureLineBlock) ToHtml() string {
	return "<hr>\n"
}

func (StartParagraphToken) ToHtml() string {
	return "<p>\n"
}

func (EndParagraphToken) ToHtml() string {
	return "</p>\n"
}

func (StartQuoteToken) ToHtml() string {
	return "<blockquote>\n"
}

func (EndQuoteToken) ToHtml() string {
	return "</blockquote>\n"
}

func (b BlockToken) ToHtml() string {
	return b.Block.ToHtml()
}

func (t TextToken) ToHtml() string {
	return html.EscapeString(string(t))
}

func Render(tokens []Token) string {
	var output strings.Builder

	indentLevel := 0

	writeToken := func(token Token) {
		for space := indentLevel; space > 0; space-- {
			output.WriteString(IndentPrefix)
		}
		output.WriteString(token.ToHtml())
	}

	for _, token := range tokens {
		switch token.TagType() {
		case TagTypeOpen:
			writeToken(token)
			indentLevel++
		case TagTypeClose:
			indentLevel--
			writeToken(token)
		case TagTypeSelfClose:
			writeToken(token)
		}
	}

	return output.String()
}

func ToHtml(text io.Reader) (string, error) {
	lines, err := ParseLines(text)
	if err != nil {
		return "", err
	}

	return Render(Tokenize(lines)), nil
}
