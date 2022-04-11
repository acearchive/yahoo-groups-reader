// Package flowed parses RFC 3676 text and renders it as HTML. The language used
// in comments in this package mirror language used in the RFC.
package flowed

import (
	"bufio"
	"io"
)

type lineType int

const (
	lineTypeFlowed lineType = iota
	lineTypeFixed
	lineTypeSignature
)

type tokenType int

const (
	tokenTypeStartParagraph tokenType = iota
	tokenTypeEndParagraph
	tokenTypeStartQuote
	tokenTypeEndQuote
	tokenTypeSignatureLine
	tokenTypeText
)

var (
	tokenStartParagraph = token{kind: tokenTypeStartParagraph}
	tokenEndParagraph   = token{kind: tokenTypeEndParagraph}
	tokenStartQuote     = token{kind: tokenTypeStartQuote}
	tokenEndQuote       = token{kind: tokenTypeEndQuote}
	tokenSignatureLine  = token{kind: tokenTypeSignatureLine}
)

const (
	quoteChar     = '>'
	stuffChar     = ' '
	flowChar      = ' '
	signatureLine = "-- "
)

type classifiedLine struct {
	kind       lineType
	quoteDepth int
	content    string
}

type token struct {
	kind tokenType
	text string
}

func trimLineEnding(line []rune) []rune {
	// RFC 3676 assumes lines are CRLF-delimited, but we are supporting
	// Unix-style line-endings as well.
	if line[len(line)-1] == '\n' {
		return line[:len(line)-1]
	}

	if line[len(line)-2] == '\r' && line[len(line)-1] == '\n' {
		return line[:len(line)-2]
	}

	return line
}

func classifyLine(line string) classifiedLine {
	currentContent := trimLineEnding([]rune(line))

	switch {
	case len(currentContent) == 0:
		// Empty lines are considered fixed.
		return classifiedLine{
			kind:    lineTypeFixed,
			content: "",
		}
	case string(currentContent) == signatureLine:
		return classifiedLine{
			kind:    lineTypeSignature,
			content: string(currentContent),
		}
	}

	// If this is greater than 0, the line is quoted.
	quoteDepth := 0

	if currentContent[0] == quoteChar {
		// The line is quoted.
		currentIndex := 1
		quoteDepth++

		for {
			switch {
			case currentContent[currentIndex] == quoteChar:
				currentIndex += 1
				quoteDepth++
			case currentContent[currentIndex] == ' ' && currentContent[currentIndex+1] == quoteChar:
				currentIndex += 2
				quoteDepth++
			default:
				break
			}
		}
	}

	if currentContent[0] == stuffChar {
		// The line is stuffed.
		currentContent = currentContent[1:]
	}

	// We need to test for a signature line a second time after we delete quote
	// marks and stuffing. Note that a line that is space-stuffed but NOT quoted
	// cannot be a signature line.
	if quoteDepth > 0 && string(currentContent) == signatureLine {
		return classifiedLine{
			kind:    lineTypeSignature,
			content: string(currentContent),
		}
	}

	if currentContent[len(currentContent)-1] == flowChar {
		// The line is flowed.
		return classifiedLine{
			kind:       lineTypeFlowed,
			quoteDepth: quoteDepth,
			content:    string(currentContent[:len(currentContent)-1]),
		}
	}

	// The line is fixed.
	return classifiedLine{
		kind:       lineTypeFixed,
		quoteDepth: quoteDepth,
		content:    string(currentContent),
	}
}

func classifyLines(text io.Reader) []classifiedLine {
	var classified []classifiedLine

	scanner := bufio.NewScanner(text)

	for scanner.Scan() {
		classified = append(classified, classifyLine(scanner.Text()))
	}

	return classified
}

func tokenize(lines []classifiedLine) []token {
	var tokens []token

	previousQuoteDepth := 0
	previousLineType := lineTypeFixed

	for _, line := range lines {
		textToken := token{kind: tokenTypeText, text: line.content + "\n"}

		switch {
		case line.kind == lineTypeSignature:
			tokens = append(tokens, tokenSignatureLine)
		case line.quoteDepth > previousQuoteDepth:
			for quoteIndex := previousQuoteDepth; quoteIndex < line.quoteDepth; quoteIndex++ {
				tokens = append(tokens, tokenStartQuote)
			}

			tokens = append(tokens, tokenStartParagraph, textToken)

			if line.kind == lineTypeFixed {
				tokens = append(tokens, tokenEndParagraph)
			}
		case line.quoteDepth < previousQuoteDepth:
			// In a properly formatted message, quote blocks will always end in a
			// fixed line. However, we don't verify that the final line is a
			// fixed line because the RFC allows for handling improperly formatted
			// messages by always ending a paragraph on a change in quote depth,
			// whether the final line is fixed or flowed.
			tokens = append(tokens, tokenEndParagraph)

			for quoteIndex := previousQuoteDepth; quoteIndex > line.quoteDepth; quoteIndex-- {
				tokens = append(tokens, tokenEndQuote)
			}

			tokens = append(tokens, tokenStartParagraph, textToken)

			if line.kind == lineTypeFixed {
				tokens = append(tokens, tokenEndParagraph)
			}
		case previousLineType == lineTypeFlowed && line.kind == lineTypeFixed:
			tokens = append(tokens, textToken, tokenEndParagraph)
		case previousLineType == lineTypeFixed:
			tokens = append(tokens, tokenStartParagraph, textToken)

			if line.kind == lineTypeFixed {
				tokens = append(tokens, tokenEndParagraph)
			}
		default:
			tokens = append(tokens, textToken)
		}

		previousQuoteDepth = line.quoteDepth
		previousLineType = line.kind
	}

	if previousLineType == lineTypeFlowed {
		// Close the open paragraph.
		tokens = append(tokens, tokenEndParagraph)
	}

	// Close any open quote blocks.
	for quoteIndex := previousQuoteDepth; quoteIndex > 0; quoteIndex-- {
		tokens = append(tokens, tokenEndQuote)
	}

	return tokens
}

// FlowedTextToHtml converts plain text as it appears in message bodies into
// HTML. This process largely follows RFC 3676, with some allowances for the
// fact that Yahoo Groups actually predates that spec.
func FlowedTextToHtml() (string, error) {
	return "", nil
}
