package block

import (
	"html/template"
	"strings"
	"time"
)

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
  {{- if .Timestamp -}}
    On <time datetime="{{ .Timestamp }}">{{ .FormattedDatetime }}</time>, {{ .Name }} said:
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
	Name              string
	FormattedDatetime string
	Timestamp         string
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
		params.Timestamp = b.Time.Format(time.RFC3339)

		if b.HasTime {
			params.FormattedDatetime = b.Time.Format("2 Jan 2006, 15:04 -07:00")
		} else {
			params.FormattedDatetime = b.Time.Format("2 Jan 2006")
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
