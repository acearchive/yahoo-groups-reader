package body

import (
	"html"
	"io"
	"strings"
)

const IndentPrefix = "  "

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
