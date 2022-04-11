package flowed

import (
	"html"
	"io"
	"strings"
)

const IndentPrefix = "  "

func stripEmptyTags(tokens []Token) []Token {
	strippedTokens := make([]Token, 0, len(tokens))

	for i := 0; i < len(tokens); i++ {
		switch {
		case i+1 < len(tokens) && tokens[i].Kind().TagType() == TagTypeOpen && tokens[i+1].Kind().TagType() == TagTypeClose:
			// An open tag immediately followed by a close tag.
			i += 1
		case i+2 < len(tokens) && tokens[i].Kind().TagType() == TagTypeOpen && tokens[i+1].Kind() == TokenTypeText && strings.TrimSpace(tokens[i+1].Text()) == "" && tokens[i+2].Kind().TagType() == TagTypeClose:
			// An open tag followed by text that's just whitespace and then close tag.
			i += 2
		default:
			strippedTokens = append(strippedTokens, tokens[i])
		}
	}

	return strippedTokens
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

	for _, token := range stripEmptyTags(tokens) {
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
