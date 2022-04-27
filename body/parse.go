package body

import (
	"bufio"
	block2 "github.com/acearchive/yg-render/block"
	"io"
	"strings"
)

type TagType string

const (
	TagTypeOpen      TagType = "Open"
	TagTypeClose     TagType = "Close"
	TagTypeSelfClose TagType = "SelfClose"
)

type Token interface {
	TagType() TagType
	ToHtml() string
}

type StartParagraphToken struct{}

func (StartParagraphToken) TagType() TagType {
	return TagTypeOpen
}

type EndParagraphToken struct{}

func (EndParagraphToken) TagType() TagType {
	return TagTypeClose
}

type StartQuoteToken struct{}

func (StartQuoteToken) TagType() TagType {
	return TagTypeOpen
}

type EndQuoteToken struct{}

func (EndQuoteToken) TagType() TagType {
	return TagTypeClose
}

type BlockToken struct {
	block2.Block
}

func (BlockToken) TagType() TagType {
	return TagTypeSelfClose
}

type TextToken string

func (TextToken) TagType() TagType {
	return TagTypeSelfClose
}

type Line struct {
	QuoteDepth int
	Content    string
}

func (l Line) IsEmpty() bool {
	return len(l.Content) == 0
}

const (
	whitespaceChars = " \t"
	quoteChar       = ">"
)

func TrimSpaceStart(text string) string {
	return strings.TrimLeft(text, whitespaceChars)
}

func ParseLine(line string) Line {
	quoteDepth := 0
	content := TrimSpaceStart(line)

	for strings.HasPrefix(content, quoteChar) {
		quoteDepth++
		content = strings.TrimPrefix(content, quoteChar)
		content = TrimSpaceStart(content)
	}

	if len(strings.TrimSpace(content)) == 0 {
		return Line{
			Content:    "",
			QuoteDepth: quoteDepth,
		}
	}

	return Line{
		Content:    content,
		QuoteDepth: quoteDepth,
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

type Tokenizer struct {
	previousLine      Line
	currentQuoteDepth int
}

func (t *Tokenizer) reset() {
	t.previousLine = Line{Content: "", QuoteDepth: 0}
	t.currentQuoteDepth = 0
}

func (t *Tokenizer) tokenizeLineWithoutBlocks(line Line) []Token {
	var tokens []Token

	switch {
	case line.QuoteDepth > t.currentQuoteDepth:
		if !t.previousLine.IsEmpty() {
			tokens = append(tokens, EndParagraphToken{})
		}

		for quoteIndex := t.currentQuoteDepth; quoteIndex < line.QuoteDepth; quoteIndex++ {
			tokens = append(tokens, StartQuoteToken{})
		}

		if !line.IsEmpty() {
			tokens = append(tokens, StartParagraphToken{}, TextToken(line.Content))
		}

		t.currentQuoteDepth = line.QuoteDepth
	case line.QuoteDepth < t.currentQuoteDepth && line.IsEmpty():
		if !t.previousLine.IsEmpty() {
			tokens = append(tokens, EndParagraphToken{})
		}

		for quoteIndex := t.currentQuoteDepth; quoteIndex > line.QuoteDepth; quoteIndex-- {
			tokens = append(tokens, EndQuoteToken{})
		}

		t.currentQuoteDepth = line.QuoteDepth
	case line.IsEmpty() && !t.previousLine.IsEmpty():
		tokens = append(tokens, EndParagraphToken{})
	case !line.IsEmpty():
		if t.previousLine.IsEmpty() {
			tokens = append(tokens, StartParagraphToken{})
		}

		tokens = append(tokens, TextToken(line.Content))
	}

	t.previousLine = line

	return tokens
}

func (t *Tokenizer) TokenizeLines(lines []Line) []Token {
	t.reset()

	var tokens []Token

	for _, line := range lines {
		tokens = append(tokens, t.tokenizeLineWithoutBlocks(line)...)
	}

	return parseBlocks(tokens)
}

func (t *Tokenizer) Tokenize(body io.Reader) ([]Token, error) {
	lines, err := ParseLines(body)
	if err != nil {
		return nil, err
	}

	return t.TokenizeLines(lines), nil
}

func createEmptyBlocks() []block2.Block {
	return []block2.Block{
		&block2.HardBreakBlock{},
		&block2.DividerBlock{},
		&block2.MessageHeaderBlock{},
		&block2.AttributionBlock{},
	}
}

func findBlocksInParagraph(text string) []Token {
	for _, newBlock := range createEmptyBlocks() {
		if ok, before, after := newBlock.FromText(text); ok {
			beforeBlocks := findBlocksInParagraph(before)
			afterBlocks := findBlocksInParagraph(after)

			output := make([]Token, 0, len(beforeBlocks)+len(afterBlocks)+1)
			output = append(output, beforeBlocks...)
			output = append(output, BlockToken{newBlock})
			output = append(output, afterBlocks...)

			return output
		}
	}

	if len(strings.TrimSpace(text)) == 0 {
		return []Token{}
	}

	return []Token{
		StartParagraphToken{},
		TextToken(text),
		EndParagraphToken{},
	}
}

func parseBlocks(tokens []Token) []Token {
	output := make([]Token, 0, len(tokens))

	var currentParagraph strings.Builder

	for _, token := range tokens {
		switch concrete := token.(type) {
		case StartParagraphToken:
			currentParagraph.Reset()
		case EndParagraphToken:
			output = append(output, findBlocksInParagraph(currentParagraph.String())...)
		case TextToken:
			currentParagraph.WriteString(string(concrete))
			currentParagraph.WriteString("\n")
		default:
			output = append(output, token)
		}
	}

	return output
}
