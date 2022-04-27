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

const attributionTemplateString = `
<div class="inline-quote-attribution">
  <span class="inline-icon" aria-hidden="true">
    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" fill="currentColor" class="bi bi-quote" viewBox="0 0 16 16">
      <path d="M12 12a1 1 0 0 0 1-1V8.558a1 1 0 0 0-1-1h-1.388c0-.351.021-.703.062-1.054.062-.372.166-.703.31-.992.145-.29.331-.517.559-.683.227-.186.516-.279.868-.279V3c-.579 0-1.085.124-1.52.372a3.322 3.322 0 0 0-1.085.992 4.92 4.92 0 0 0-.62 1.458A7.712 7.712 0 0 0 9 7.558V11a1 1 0 0 0 1 1h2Zm-6 0a1 1 0 0 0 1-1V8.558a1 1 0 0 0-1-1H4.612c0-.351.021-.703.062-1.054.062-.372.166-.703.31-.992.145-.29.331-.517.559-.683.227-.186.516-.279.868-.279V3c-.579 0-1.085.124-1.52.372a3.322 3.322 0 0 0-1.085.992 4.92 4.92 0 0 0-.62 1.458A7.712 7.712 0 0 0 3 7.558V11a1 1 0 0 0 1 1h2Z"/>
    </svg>
  </span>
  {{ if (and .Date .Time) -}}
    On {{ .Date }} at {{ .Time }} {{ if .HasTimeZone }}UTC{{ else }}<em>(local time)</em>{{ end }}, {{ .Name }} said:
  {{- else if .Date -}}
    On {{ .Date }}, {{ .Name }} said:
  {{- else -}}
    {{ .Name }} said:
  {{- end }}
</div>
`

var attributionTemplate = template.Must(template.New("inline-quote-attribution").Parse(attributionTemplateString))

type messageHeaderTemplateParams struct {
	Fields []Field
}

type attributionTemplateParams struct {
	Name        string
	Date        string
	Time        string
	HasTimeZone bool
}

func (b *MessageHeaderBlock) ToHtml() string {
	params := messageHeaderTemplateParams{Fields: *b}

	var output strings.Builder

	if err := messageHeaderTemplate.Execute(&output, params); err != nil {
		panic(err)
	}

	return output.String()
}

func (b *DividerBlock) ToHtml() string {
	return "<hr>\n"
}

func (b *AttributionBlock) ToHtml() string {
	params := attributionTemplateParams{Name: b.Name}

	if !b.Time.IsZero() {
		params.Date = b.Time.Format("2 January 2006")

		if b.HasTime {
			params.Time = b.Time.Format("15:04")
			params.HasTimeZone = b.HasTimeZone
		}
	}

	var output strings.Builder

	if err := attributionTemplate.Execute(&output, params); err != nil {
		panic(err)
	}

	return output.String()
}

func (b *HardBreakBlock) ToHtml() string {
	return ""
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

	var tokenizer Tokenizer

	return Render(tokenizer.Tokenize(lines)), nil
}
