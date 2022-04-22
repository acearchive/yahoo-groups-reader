package body

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

type TagType string

const (
	TagTypeOpen      TagType = "Open"
	TagTypeClose     TagType = "Close"
	TagTypeSelfClose TagType = "SelfClose"
)

type Block interface {
	ToHtml() string
}

type Field struct {
	Name  string
	Value string
}

type MessageHeaderBlock []Field

type DividerBlock struct{}

type AttributionBlock struct {
	Name string
}

type Token interface {
	IsToken()
	TagType() TagType
	ToHtml() string
}

type StartParagraphToken struct{}

func (StartParagraphToken) IsToken() {}

func (StartParagraphToken) TagType() TagType {
	return TagTypeOpen
}

type EndParagraphToken struct{}

func (EndParagraphToken) IsToken() {}

func (EndParagraphToken) TagType() TagType {
	return TagTypeClose
}

type StartQuoteToken struct{}

func (StartQuoteToken) IsToken() {}

func (StartQuoteToken) TagType() TagType {
	return TagTypeOpen
}

type EndQuoteToken struct{}

func (EndQuoteToken) IsToken() {}

func (EndQuoteToken) TagType() TagType {
	return TagTypeClose
}

type BlockToken struct {
	Block
}

func (BlockToken) IsToken() {}

func (BlockToken) TagType() TagType {
	return TagTypeSelfClose
}

type TextToken string

func (TextToken) IsToken() {}

func (TextToken) TagType() TagType {
	return TagTypeSelfClose
}

type Line struct {
	QuoteDepth int
	Content    LineContent
}

func (l Line) IsEmpty() bool {
	_, ok := l.Content.(EmptyLineContent)
	return ok
}

func (l Line) IsText() bool {
	_, ok := l.Content.(TextLineContent)
	return ok
}

func (l Line) IsDivider() bool {
	_, ok := l.Content.(DividerLineContent)
	return ok
}

func (l Line) IsMessageHeader() bool {
	_, ok := l.Content.(MessageHeaderLineContent)
	return ok
}

func (l Line) IsField() bool {
	_, ok := l.Content.(FieldLineContent)
	return ok
}

func (l Line) IsAttribution() bool {
	_, ok := l.Content.(AttributionLineContent)
	return ok
}

type LineContent interface {
	IsLineContent()
}

type EmptyLineContent struct{}

func (EmptyLineContent) IsLineContent() {}

type TextLineContent string

func (TextLineContent) IsLineContent() {}

type DividerLineContent struct{}

func (DividerLineContent) IsLineContent() {}

type MessageHeaderLineContent struct{}

func (MessageHeaderLineContent) IsLineContent() {}

type FieldLineContent struct {
	Field
	Text string
}

func (FieldLineContent) IsLineContent() {}

type AttributionLineContent string

func (AttributionLineContent) IsLineContent() {}

const (
	whitespaceChars = " \t"
	quoteChar       = ">"
)

func TrimSpaceStart(text string) string {
	return strings.TrimLeft(text, whitespaceChars)
}

func TrimSpaceEnd(text string) string {
	return strings.TrimRight(text, whitespaceChars)
}

var (
	attributionLineRegex = regexp.MustCompile(`^--- In [^\s@]+@(?:yahoogroups\.com|y?\.{3}),`)
	attributionNameRegex = regexp.MustCompile(`^--- In [^\s@]+@(?:yahoogroups\.com|y?\.{3}), ([^<>]+)(?: <[^<>\s]+>)? wrote:$`)
	dividerRegex         = regexp.MustCompile(`^[-_]{2,}$`)
	fieldRegex           = regexp.MustCompile(`^(From|Reply-To|To|Subject|Date|Sent|Message): (\S.*)$`)
	messageHeaderRegex   = regexp.MustCompile(`^-+ ?Original Message ?-+$`)
)

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
			Content:    EmptyLineContent{},
			QuoteDepth: quoteDepth,
		}
	}

	if dividerRegex.MatchString(TrimSpaceEnd(content)) {
		return Line{
			Content:    DividerLineContent{},
			QuoteDepth: quoteDepth,
		}
	}

	if messageHeaderRegex.MatchString(TrimSpaceEnd(content)) {
		return Line{
			Content:    MessageHeaderLineContent{},
			QuoteDepth: quoteDepth,
		}
	}

	if matches := fieldRegex.FindStringSubmatch(content); matches != nil {
		return Line{
			Content: FieldLineContent{
				Field: Field{
					Name:  matches[1],
					Value: matches[2],
				},
				Text: content,
			},
			QuoteDepth: quoteDepth,
		}
	}

	if attributionLineRegex.MatchString(TrimSpaceEnd(content)) {
		return Line{
			Content:    AttributionLineContent(content),
			QuoteDepth: quoteDepth,
		}
	}

	return Line{
		Content:    TextLineContent(content),
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
	currentHeader     MessageHeaderBlock
	currentQuoteDepth int
	tokens            []Token
}

func (t *Tokenizer) reset() {
	t.previousLine = Line{
		Content:    EmptyLineContent{},
		QuoteDepth: 0,
	}
	t.currentHeader = nil
	t.currentQuoteDepth = 0
	t.tokens = nil
}

func (t *Tokenizer) addTokens(tokens ...Token) {
	t.tokens = append(t.tokens, tokens...)
}

func (t *Tokenizer) nextLine(line Line) {
	t.previousLine = line
}

func (t *Tokenizer) tokenizeSimpleLine(line Line) {
	switch {
	case line.QuoteDepth > t.currentQuoteDepth:
		if t.previousLine.IsText() {
			t.addTokens(EndParagraphToken{})
		}

		for quoteIndex := t.currentQuoteDepth; quoteIndex < line.QuoteDepth; quoteIndex++ {
			t.addTokens(StartQuoteToken{})
		}

		switch content := line.Content.(type) {
		case DividerLineContent:
			t.addTokens(BlockToken{DividerBlock{}})
		case TextLineContent:
			t.addTokens(StartParagraphToken{}, TextToken(content))
		case FieldLineContent:
			t.currentHeader = append(t.currentHeader, content.Field)
		}

		t.currentQuoteDepth = line.QuoteDepth
	case line.QuoteDepth < t.currentQuoteDepth && line.IsEmpty():
		if t.previousLine.IsText() {
			t.addTokens(EndParagraphToken{})
		}

		for quoteIndex := t.currentQuoteDepth; quoteIndex > line.QuoteDepth; quoteIndex-- {
			t.addTokens(EndQuoteToken{})
		}

		t.currentQuoteDepth = line.QuoteDepth
	case line.IsAttribution():
		if t.previousLine.IsText() {
			t.addTokens(EndParagraphToken{})
		}
	case line.IsDivider():
		if t.previousLine.IsText() {
			t.addTokens(EndParagraphToken{})
		}

		t.addTokens(BlockToken{DividerBlock{}})
	case (line.IsEmpty() || line.IsMessageHeader()) && t.previousLine.IsText():
		t.addTokens(EndParagraphToken{})
	case line.IsText():
		if !t.previousLine.IsText() {
			t.addTokens(StartParagraphToken{})
		}

		t.addTokens(TextToken(line.Content.(TextLineContent)))
	}

	t.nextLine(line)
}

func (t *Tokenizer) tokenizeLineForHeader(line Line) bool {
	if t.currentHeader == nil {
		return false
	}

	if line.QuoteDepth == t.currentQuoteDepth {
		switch lineContent := line.Content.(type) {
		case FieldLineContent:
			t.currentHeader = append(t.currentHeader, lineContent.Field)
			t.nextLine(line)
			return true
		case TextLineContent:
			// This is a continuation line for the previous field.
			t.currentHeader[len(t.currentHeader)-1].Value += string(lineContent)
			t.nextLine(line)
			return true
		}
	}

	t.addTokens(BlockToken{t.currentHeader})
	t.currentHeader = nil

	t.tokenizeSimpleLine(line)

	return true
}

func (t *Tokenizer) tokenizeLineForAttribution(line Line) bool {
	attributionContent, isAttribution := t.previousLine.Content.(AttributionLineContent)
	if !isAttribution {
		return false
	}

	if matches := attributionNameRegex.FindStringSubmatch(string(attributionContent)); matches != nil {
		name := strings.TrimSuffix(strings.TrimPrefix(matches[1], "\""), "\"")
		t.addTokens(BlockToken{AttributionBlock{Name: name}})
		t.nextLine(Line{
			Content:    EmptyLineContent{},
			QuoteDepth: line.QuoteDepth,
		})
	} else {
		if textContent, isText := line.Content.(TextLineContent); line.QuoteDepth <= t.currentQuoteDepth && isText {
			t.nextLine(Line{
				Content:    AttributionLineContent(fmt.Sprintf("%s %s", TrimSpaceEnd(string(attributionContent)), textContent)),
				QuoteDepth: line.QuoteDepth,
			})
			return true
		}

		t.tokenizeSimpleLine(Line{
			Content:    TextLineContent(attributionContent),
			QuoteDepth: line.QuoteDepth,
		})
	}

	t.tokenizeSimpleLine(line)

	return true
}

func (t *Tokenizer) tokenizeLineForField(line Line) bool {
	if content, isField := line.Content.(FieldLineContent); isField {
		if line.QuoteDepth == t.currentQuoteDepth && (t.previousLine.IsEmpty() || t.previousLine.IsMessageHeader()) {
			t.currentHeader = append(t.currentHeader, content.Field)
			t.nextLine(Line{QuoteDepth: line.QuoteDepth, Content: content})
			return true
		}

		t.tokenizeSimpleLine(Line{
			Content:    TextLineContent(content.Text),
			QuoteDepth: line.QuoteDepth,
		})

		return true
	}

	return false
}

func (t *Tokenizer) tokenizeLine(line Line) {
	if t.tokenizeLineForHeader(line) {
		return
	}

	if t.tokenizeLineForAttribution(line) {
		return
	}

	if t.tokenizeLineForField(line) {
		return
	}

	t.tokenizeSimpleLine(line)
}

func (t *Tokenizer) Tokenize(lines []Line) []Token {
	t.reset()

	for _, line := range lines {
		t.tokenizeLine(line)
	}

	return t.tokens
}
