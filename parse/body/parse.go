package body

import (
	"bufio"
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

var (
	attributionLineRegex = regexp.MustCompile(`^--- In [^\s@]+@(?:yahoogroups\.com|y?\.{3}),`)
	attributionNameRegex = regexp.MustCompile(`^--- In [^\s@]+@(?:yahoogroups\.com|y?\.{3}), ([^<>]+)(?: <[^<>\s]+>)? wrote:$`)
	dividerRegex         = regexp.MustCompile(`^[-_]{2,}$`)
	fieldRegex           = regexp.MustCompile(`^(From|Reply-To|To|Subject|Date|Sent|Message): (\S.*)$`)
	messageHeaderRegex   = regexp.MustCompile(`^-+ ?Original Message ?-+$`)
)

func ParseLine(line string) Line {
	quoteDepth := 0
	content := strings.TrimLeft(line, whitespaceChars)

	for strings.HasPrefix(content, quoteChar) {
		quoteDepth++
		content = strings.TrimPrefix(content, quoteChar)
		content = strings.TrimLeft(content, whitespaceChars)
	}

	if len(strings.TrimSpace(content)) == 0 {
		return Line{
			Content:    EmptyLineContent{},
			QuoteDepth: quoteDepth,
		}
	}

	if dividerRegex.MatchString(strings.TrimRight(content, whitespaceChars)) {
		return Line{
			Content:    DividerLineContent{},
			QuoteDepth: quoteDepth,
		}
	}

	if messageHeaderRegex.MatchString(strings.TrimRight(content, whitespaceChars)) {
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

	if attributionLineRegex.MatchString(strings.TrimRight(content, whitespaceChars)) {
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

func Tokenize(lines []Line) []Token {
	var tokens []Token

	currentQuoteDepth := 0

	var (
		previousLine       Line
		currentHeader      MessageHeaderBlock
		currentAttribution string
	)

	for _, line := range lines {
		if len(currentHeader) > 0 {
			if line.QuoteDepth == currentQuoteDepth {
				switch lineContent := line.Content.(type) {
				case FieldLineContent:
					currentHeader = append(currentHeader, lineContent.Field)
					previousLine = line
					continue
				case TextLineContent:
					// This is a continuation line for the previous field.
					currentHeader[len(currentHeader)-1].Value += string(lineContent)
					previousLine = line
					continue
				}
			}

			tokens = append(tokens, BlockToken{currentHeader})
			currentHeader = nil
		}

		if currentAttribution != "" {
			if matches := attributionNameRegex.FindStringSubmatch(currentAttribution); matches != nil {
				name := strings.TrimSuffix(strings.TrimPrefix(matches[1], "\""), "\"")
				tokens = append(tokens, BlockToken{AttributionBlock{Name: name}})
				currentAttribution = ""
			} else {
				if textContent, isText := line.Content.(TextLineContent); line.QuoteDepth <= currentQuoteDepth && isText {
					currentAttribution += string(textContent)
					previousLine = line
					continue
				}

				line = Line{
					Content:    TextLineContent(currentAttribution),
					QuoteDepth: line.QuoteDepth,
				}
				currentAttribution = ""
			}
		}

		if fieldContent, isField := line.Content.(FieldLineContent); isField {
			if line.QuoteDepth == currentQuoteDepth && (previousLine.IsEmpty() || previousLine.IsMessageHeader()) {
				currentHeader = append(currentHeader, fieldContent.Field)
				previousLine = line
				continue
			}

			line = Line{
				Content:    TextLineContent(fieldContent.Text),
				QuoteDepth: line.QuoteDepth,
			}
		}

		switch {
		case line.QuoteDepth > currentQuoteDepth:
			if previousLine.IsText() {
				tokens = append(tokens, EndParagraphToken{})
			}

			for quoteIndex := currentQuoteDepth; quoteIndex < line.QuoteDepth; quoteIndex++ {
				tokens = append(tokens, StartQuoteToken{})
			}

			switch content := line.Content.(type) {
			case DividerLineContent:
				tokens = append(tokens, BlockToken{DividerBlock{}})
			case TextLineContent:
				tokens = append(tokens, StartParagraphToken{}, TextToken(content))
			case FieldLineContent:
				currentHeader = append(currentHeader, content.Field)
			case AttributionLineContent:
				currentAttribution = string(content)
			}

			currentQuoteDepth = line.QuoteDepth
		case line.QuoteDepth < currentQuoteDepth && line.IsEmpty():
			if previousLine.IsText() {
				tokens = append(tokens, EndParagraphToken{})
			}

			for quoteIndex := currentQuoteDepth; quoteIndex > line.QuoteDepth; quoteIndex-- {
				tokens = append(tokens, EndQuoteToken{})
			}

			currentQuoteDepth = line.QuoteDepth
		case line.IsAttribution():
			if previousLine.IsText() {
				tokens = append(tokens, EndParagraphToken{})
			}

			currentAttribution = string(line.Content.(AttributionLineContent))
		case line.IsDivider():
			if previousLine.IsText() {
				tokens = append(tokens, EndParagraphToken{})
			}

			tokens = append(tokens, BlockToken{DividerBlock{}})
		case (line.IsEmpty() || line.IsMessageHeader()) && previousLine.IsText():
			tokens = append(tokens, EndParagraphToken{})
		case line.IsText():
			if !previousLine.IsText() {
				tokens = append(tokens, StartParagraphToken{})
			}

			tokens = append(tokens, TextToken(line.Content.(TextLineContent)))
		}

		previousLine = line
	}

	return tokens
}
