package flowed

import (
	"io"
	"strings"
)

const IndentPrefix = "  "

type IndentMode int

const (
	IndentModeIndent IndentMode = iota
	IndentModeDedent
	IndentModeSame
)

func Render(tokens []Token) string {
	var output strings.Builder

	indentLevel := 0
	mode := IndentModeSame
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
			mode = IndentModeIndent
		case TokenTypeEndParagraph:
			node = "</p>\n"
			mode = IndentModeDedent
		case TokenTypeStartQuote:
			node = "<blockquote>\n"
			mode = IndentModeIndent
		case TokenTypeEndQuote:
			node = "</blockquote>\n"
			mode = IndentModeDedent
		case TokenTypeSignatureLine:
			node = "<hr>\n"
			mode = IndentModeSame
		case TokenTypeText:
			node = token.Text()
			mode = IndentModeSame
		}

		switch mode {
		case IndentModeIndent:
			writeToken()
			indentLevel++
		case IndentModeDedent:
			indentLevel--
			writeToken()
		case IndentModeSame:
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
