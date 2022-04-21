package body

import (
	"html"
	"io"
	"strings"
)

const IndentPrefix = "  "

func (b MessageHeaderBlock) ToHtml(indent string) string {
	// TODO: Implement
	return ""
}

func (b SignatureLineBlock) ToHtml(indent string) string {
	return "<hr>\n"
}

func (StartParagraphToken) ToHtml(indent string) string {
	return "<p>\n"
}

func (EndParagraphToken) ToHtml(indent string) string {
	return "</p>\n"
}

func (StartQuoteToken) ToHtml(indent string) string {
	return "<blockquote>\n"
}

func (EndQuoteToken) ToHtml(indent string) string {
	return "</blockquote>\n"
}

func (b BlockToken) ToHtml(indent string) string {
	return b.ToHtml(indent)
}

func (t TextToken) ToHtml(indent string) string {
	return html.EscapeString(string(t))
}

func Render(tokens []Token) string {
	var output strings.Builder

	indentLevel := 0
	node := ""

	writeToken := func() {
		for space := indentLevel; space > 0; space-- {
			output.WriteString(IndentPrefix)
		}
		output.WriteString(node)
	}

	for _, token := range tokens {
		node = token.ToHtml(IndentPrefix)

		switch token.TagType() {
		case TagTypeOpen:
			writeToken()
			indentLevel++
		case TagTypeClose:
			indentLevel--
			writeToken()
		case TagTypeSelfClose:
			writeToken()
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
