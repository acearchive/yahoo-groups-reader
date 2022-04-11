package flowed

import (
	"html"
	"io"
	"strings"
)

const IndentPrefix = "  "

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
		switch token.Kind() {
		case TokenTypeStartParagraph:
			node = "<p>\n"
		case TokenTypeEndParagraph:
			node = "</p>\n"
		case TokenTypeStartQuote:
			node = "<blockquote>\n"
		case TokenTypeEndQuote:
			node = "</blockquote>\n"
		case TokenTypeSignatureLine:
			node = "<hr>\n"
		case TokenTypeText:
			node = html.EscapeString(token.Text())
		}

		switch token.Kind().TagType() {
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
