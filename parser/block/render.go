package block

import (
	_ "embed"
	"html/template"
	"strings"
	"time"
)

//go:embed header.html.tmpl
var messageHeaderTemplateString string

var messageHeaderTemplate = template.Must(template.New("header-block").Parse(messageHeaderTemplateString))

//go:embed attribution.html.tmpl
var attributionTemplateString string

var attributionTemplate = template.Must(template.New("attribution-block").Parse(attributionTemplateString))

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
