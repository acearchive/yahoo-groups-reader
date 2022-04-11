// Package flowed parses RFC 3676 text and renders it as HTML. The language used
// in comments in this package mirror language used in the RFC. There are a few
// places where we deliberately diverge from the spec to account for the fact
// that Yahoo Groups actually predates the RFC by several years.
package flowed

import (
	"bufio"
	"io"
)

type LineType string

const (
	LineTypeFlowed    LineType = "Flowed"
	LineTypeFixed     LineType = "Fixed"
	LineTypeSignature LineType = "Signature"
)

type TokenType string

const (
	TokenTypeStartParagraph TokenType = "StartParagraph"
	TokenTypeEndParagraph   TokenType = "EndParagraph"
	TokenTypeStartQuote     TokenType = "StartQuote"
	TokenTypeEndQuote       TokenType = "EndQuote"
	TokenTypeSignatureLine  TokenType = "SignatureLine"
	TokenTypeText           TokenType = "Text"
)

var (
	TokenStartParagraph = Token{kind: TokenTypeStartParagraph}
	TokenEndParagraph   = Token{kind: TokenTypeEndParagraph}
	TokenStartQuote     = Token{kind: TokenTypeStartQuote}
	TokenEndQuote       = Token{kind: TokenTypeEndQuote}
	TokenSignatureLine  = Token{kind: TokenTypeSignatureLine}
)

const (
	quoteChar     = '>'
	stuffChar     = ' '
	flowChar      = ' '
	signatureLine = "-- "
)

type Line struct {
	Kind       LineType
	QuoteDepth int
	Content    string
}

type Token struct {
	kind TokenType
	text string
}

func (t Token) Kind() TokenType {
	return t.kind
}

func (t Token) Text() string {
	return t.text
}

func NewTextToken(text string) Token {
	return Token{kind: TokenTypeText, text: text + "\n"}
}

func trimLineEnding(line []rune) []rune {
	// RFC 3676 assumes lines are CRLF-delimited, but we are supporting
	// Unix-style line-endings as well.
	if len(line) >= 1 && line[len(line)-1] == '\n' {
		return line[:len(line)-1]
	}

	if len(line) >= 2 && line[len(line)-2] == '\r' && line[len(line)-1] == '\n' {
		return line[:len(line)-2]
	}

	return line
}

func ParseLine(line string) Line {
	currentContent := trimLineEnding([]rune(line))

	switch {
	case len(currentContent) == 0:
		// Empty lines are considered fixed.
		return Line{
			Kind:    LineTypeFixed,
			Content: "",
		}
	case string(currentContent) == signatureLine:
		return Line{
			Kind:    LineTypeSignature,
			Content: string(currentContent),
		}
	}

	// If this is greater than 0, the line is quoted.
	quoteDepth := 0

	if currentContent[0] == quoteChar {
		// The line is quoted.
		currentIndex := 1
		quoteDepth++

	quoteDepthLoop:
		for {
			switch {
			case len(currentContent) > currentIndex && currentContent[currentIndex] == quoteChar:
				currentIndex += 1
				quoteDepth++
			case len(currentContent) > currentIndex+1 && currentContent[currentIndex] == ' ' && currentContent[currentIndex+1] == quoteChar:
				// According to the RFC, this should not be interpreted as a
				// nested quote, and should actually be interpreted as a leading
				// literal '> ' in the outer quote. However, we are diverging
				// from the spec here because Yahoo Groups seems to use this
				// syntax.
				currentIndex += 2
				quoteDepth++
			default:
				break quoteDepthLoop
			}
		}

		currentContent = currentContent[currentIndex:]
	}

	if len(currentContent) > 0 && currentContent[0] == stuffChar {
		// The line is stuffed.
		currentContent = currentContent[1:]
	}

	// We need to test for a signature line a second time after we delete quote
	// marks and stuffing. Note that a line that is space-stuffed but NOT quoted
	// cannot be a signature line.
	if quoteDepth > 0 && string(currentContent) == signatureLine {
		return Line{
			Kind:    LineTypeSignature,
			Content: string(currentContent),
		}
	}

	if len(currentContent) > 0 && currentContent[len(currentContent)-1] == flowChar {
		// The line is flowed.
		return Line{
			Kind:       LineTypeFlowed,
			QuoteDepth: quoteDepth,
			Content:    string(currentContent[:len(currentContent)-1]),
		}
	}

	// The line is fixed.
	return Line{
		Kind:       LineTypeFixed,
		QuoteDepth: quoteDepth,
		Content:    string(currentContent),
	}
}

func ParseLines(text io.Reader) ([]Line, error) {
	var lines []Line

	scanner := bufio.NewScanner(text)

	for scanner.Scan() {
		lines = append(lines, ParseLine(scanner.Text()))
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func Tokenize(lines []Line) []Token {
	var tokens []Token

	previousQuoteDepth := 0
	previousLineType := LineTypeFixed

	for _, line := range lines {
		textToken := NewTextToken(line.Content)

		switch {
		case line.Kind == LineTypeSignature:
			if previousLineType == LineTypeFlowed {
				tokens = append(tokens, TokenEndParagraph)
			}
			tokens = append(tokens, TokenSignatureLine)
		case line.QuoteDepth > previousQuoteDepth:
			for quoteIndex := previousQuoteDepth; quoteIndex < line.QuoteDepth; quoteIndex++ {
				tokens = append(tokens, TokenStartQuote)
			}

			tokens = append(tokens, TokenStartParagraph, textToken)

			if line.Kind == LineTypeFixed {
				tokens = append(tokens, TokenEndParagraph)
			}
		case line.QuoteDepth < previousQuoteDepth:
			// In a properly formatted message, quote blocks will always end in
			// a fixed line. However, the RFC allows for handling improperly
			// formatted messages by always ending a paragraph on a change in
			// quote depth, whether the final line is fixed or flowed.
			if previousLineType == LineTypeFlowed {
				tokens = append(tokens, TokenEndParagraph)
			}

			for quoteIndex := previousQuoteDepth; quoteIndex > line.QuoteDepth; quoteIndex-- {
				tokens = append(tokens, TokenEndQuote)
			}

			tokens = append(tokens, TokenStartParagraph, textToken)

			if line.Kind == LineTypeFixed {
				tokens = append(tokens, TokenEndParagraph)
			}
		case previousLineType == LineTypeFlowed && line.Kind == LineTypeFixed:
			tokens = append(tokens, textToken, TokenEndParagraph)
		case previousLineType == LineTypeFixed || previousLineType == LineTypeSignature:
			tokens = append(tokens, TokenStartParagraph, textToken)

			if line.Kind == LineTypeFixed {
				tokens = append(tokens, TokenEndParagraph)
			}
		default:
			tokens = append(tokens, textToken)
		}

		previousQuoteDepth = line.QuoteDepth
		previousLineType = line.Kind
	}

	if previousLineType == LineTypeFlowed {
		// Close the open paragraph.
		tokens = append(tokens, TokenEndParagraph)
	}

	// Close any open quote blocks.
	for quoteIndex := previousQuoteDepth; quoteIndex > 0; quoteIndex-- {
		tokens = append(tokens, TokenEndQuote)
	}

	return tokens
}
