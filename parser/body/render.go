package body

import (
	"html"
	"strings"
)

const IndentLen = 2

func IndentMultilineString(text string, indent int) string {
	var output strings.Builder

	lines := strings.Split(text, "\n")

	for _, line := range lines {
		for indentIndex := 0; indentIndex < indent; indentIndex++ {
			output.WriteString(" ")
		}

		output.WriteString(line)
		output.WriteString("\n")
	}

	return output.String()
}

func (StartParagraphToken) ToHtml() string {
	return "<p>"
}

func (EndParagraphToken) ToHtml() string {
	return "</p>"
}

func (StartQuoteToken) ToHtml() string {
	return "<blockquote>"
}

func (EndQuoteToken) ToHtml() string {
	return "</blockquote>"
}

func (b BlockToken) ToHtml() string {
	return b.Block.ToHtml()
}

func (t TextToken) ToHtml() string {
	return html.EscapeString(strings.TrimSpace(string(t)))
}

func Render(tokens []Token) string {
	var output strings.Builder

	indentLevel := 0

	writeToken := func(token Token) {
		output.WriteString(IndentMultilineString(token.ToHtml(), indentLevel*IndentLen))
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
