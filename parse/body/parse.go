package body

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

var ErrInvalidTokenType = errors.New("invalid TokenType")

type TokenType string

const (
	TokenTypeStartParagraph TokenType = "StartParagraph"
	TokenTypeEndParagraph   TokenType = "EndParagraph"
	TokenTypeStartQuote     TokenType = "StartQuote"
	TokenTypeEndQuote       TokenType = "EndQuote"
	TokenTypeSignatureLine  TokenType = "SignatureLine"
	TokenTypeText           TokenType = "Text"
)

func (t TokenType) TagType() TagType {
	switch t {
	case TokenTypeStartParagraph:
		return TagTypeOpen
	case TokenTypeEndParagraph:
		return TagTypeClose
	case TokenTypeStartQuote:
		return TagTypeOpen
	case TokenTypeEndQuote:
		return TagTypeClose
	case TokenTypeSignatureLine:
		return TagTypeSelfClose
	case TokenTypeText:
		return TagTypeSelfClose
	default:
		panic(fmt.Errorf("%w: %v", ErrInvalidTokenType, t))
	}
}

type TagType string

const (
	TagTypeOpen      TagType = "Open"
	TagTypeClose     TagType = "Close"
	TagTypeSelfClose TagType = "SelfClose"
)

type Token struct {
	kind TokenType
	text string
}

var (
	TokenStartParagraph = Token{kind: TokenTypeStartParagraph}
	TokenEndParagraph   = Token{kind: TokenTypeEndParagraph}
	TokenStartQuote     = Token{kind: TokenTypeStartQuote}
	TokenEndQuote       = Token{kind: TokenTypeEndQuote}
	TokenSignatureLine  = Token{kind: TokenTypeSignatureLine}
)

func (t Token) Kind() TokenType {
	return t.kind
}

func (t Token) Text() string {
	return t.text
}

func NewTextToken(text string) Token {
	return Token{kind: TokenTypeText, text: text + "\n"}
}

type LineType string

const (
	LineTypeEmpty     LineType = "Empty"
	LineTypeSignature LineType = "Signature"
	LineTypeContent   LineType = "Content"
)

type Line struct {
	kind       LineType
	quoteDepth int
	content    string
}

func (l Line) Kind() LineType {
	return l.kind
}

func (l Line) QuoteDepth() int {
	return l.quoteDepth
}

func (l Line) Content() string {
	return l.content
}

const (
	signatureLine = "--"
	quoteChar     = ">"
)

func ParseLine(line string) Line {
	quoteDepth := 0
	content := line

	for strings.HasPrefix(content, quoteChar) {
		quoteDepth++
		content = strings.TrimPrefix(content, quoteChar)
		content = strings.TrimLeft(content, " ")
	}

	if len(strings.TrimSpace(content)) == 0 {
		return Line{
			kind:       LineTypeEmpty,
			quoteDepth: quoteDepth,
			content:    "",
		}
	}

	if strings.TrimRight(content, " ") == signatureLine {
		return Line{
			kind:       LineTypeSignature,
			quoteDepth: quoteDepth,
			content:    "",
		}
	}

	return Line{
		kind:       LineTypeContent,
		quoteDepth: quoteDepth,
		content:    content,
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
	previousLineType := LineTypeEmpty

	for _, line := range lines {
		switch {
		case line.Kind() == LineTypeSignature:
			tokens = append(tokens, TokenSignatureLine)
		case line.QuoteDepth() > previousQuoteDepth:
			if previousLineType == LineTypeContent {
				tokens = append(tokens, TokenEndParagraph)
			}

			for quoteIndex := previousQuoteDepth; quoteIndex < line.QuoteDepth(); quoteIndex++ {
				tokens = append(tokens, TokenStartQuote)
			}

			tokens = append(tokens, TokenStartParagraph, NewTextToken(line.Content()))
		case line.QuoteDepth() < previousQuoteDepth:
			if previousLineType == LineTypeContent {
				tokens = append(tokens, TokenEndParagraph)
			}

			for quoteIndex := previousQuoteDepth; quoteIndex > line.QuoteDepth(); quoteIndex-- {
				tokens = append(tokens, TokenEndQuote)
			}

			tokens = append(tokens, TokenStartParagraph, NewTextToken(line.Content()))
		case line.Kind() == LineTypeEmpty && previousLineType == LineTypeContent:
			tokens = append(tokens, TokenEndParagraph)
		case line.Kind() == LineTypeContent && previousLineType == LineTypeEmpty:
			tokens = append(tokens, TokenStartParagraph, NewTextToken(line.Content()))
		case line.Kind() == LineTypeContent:
			tokens = append(tokens, NewTextToken(line.Content()))
		}

		previousQuoteDepth = line.QuoteDepth()
		previousLineType = line.Kind()
	}

	return tokens
}
