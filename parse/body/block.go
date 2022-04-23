package body

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

const (
	attributionEmailRegexPart      = `<[^<>\s]+>`
	attributionNameRegexPart       = `(?:[^<>\s]|[^<>\s][^<>]*[^<>\s])`
	attributionGroupEmailRegexPart = `[^\s@]+@(?:yahoogroups\.com|y?\.{3})`
)

var (
	dividerRegex            = regexp.MustCompile(`(?m)^\s*[-_]{2,}\s*$`)
	fieldLabelRegex         = regexp.MustCompile(`(?m)^\s*(From|Reply-To|To|Subject|Date|Sent|Message): +(\S)`)
	messageHeaderStartRegex = regexp.MustCompile(`(?m)^\s*(?:-+ ?Original Message ?-+|\n)\s*\n\s*(From|Reply-To|To|Subject|Date|Sent|Message): +(\S)`)
	messageHeaderEndRegex   = regexp.MustCompile(`(?m)^\s*\n`)
	attributionNameRegexes  = []*regexp.Regexp{
		regexp.MustCompile(fmt.Sprintf(
			`(?m)^--- In %s,\s+(%s)(?:\s+%s)?\s+wrote:$`,
			attributionGroupEmailRegexPart,
			attributionNameRegexPart,
			attributionEmailRegexPart,
		)),
		regexp.MustCompile(fmt.Sprintf(
			`(?m)^---\s+(%s)(?:\s+%s)?\s+wrote:$`,
			attributionNameRegexPart,
			attributionEmailRegexPart,
		)),
	}
)

type Block interface {
	ToHtml() string
	FromText(text string) (ok bool, before, after string)
}

type Field struct {
	Name  string
	Value string
}

type MessageHeaderBlock []Field

type messageHeaderFieldPosition struct {
	LabelStartIndex int
	LabelEndIndex   int
	ValueStartIndex int
}

func (b *MessageHeaderBlock) FromText(text string) (ok bool, before, after string) {
	var fieldPositions []messageHeaderFieldPosition

	remaining := text
	currentIndex := 0
	absoluteFieldListEndIndex := len(text)

	if match := messageHeaderStartRegex.FindStringSubmatchIndex(remaining); match != nil {
		position := messageHeaderFieldPosition{
			LabelStartIndex: match[2],
			LabelEndIndex:   match[3],
			ValueStartIndex: match[4],
		}
		fieldPositions = append(fieldPositions, position)

		currentIndex += position.LabelEndIndex
		remaining = remaining[position.LabelEndIndex:]

		matchStartIndex := match[0]
		before = text[:matchStartIndex]
	} else {
		return false, "", ""
	}

	if match := messageHeaderEndRegex.FindStringIndex(remaining); match != nil {
		relativeStartIndex, relativeEndIndex := match[0], match[1]
		absoluteStartIndex, absoluteEndIndex := currentIndex+relativeStartIndex, currentIndex+relativeEndIndex

		remaining = remaining[:relativeStartIndex]
		absoluteFieldListEndIndex = absoluteStartIndex
		after = text[absoluteEndIndex:]
	}

	for {
		match := fieldLabelRegex.FindStringSubmatchIndex(remaining)
		if match == nil {
			break
		}

		relativeFieldStartIndex, relativeFieldEndIndex, relativeValueStartIndex := match[2], match[3], match[4]

		position := messageHeaderFieldPosition{
			LabelStartIndex: currentIndex + relativeFieldStartIndex,
			LabelEndIndex:   currentIndex + relativeFieldEndIndex,
			ValueStartIndex: currentIndex + relativeValueStartIndex,
		}
		fieldPositions = append(fieldPositions, position)

		currentIndex += relativeFieldEndIndex
		remaining = remaining[relativeFieldEndIndex:]
	}

	for i, position := range fieldPositions {
		if i+1 < len(fieldPositions) {
			nextPosition := fieldPositions[i+1]

			*b = append(*b, Field{
				Name:  text[position.LabelStartIndex:position.LabelEndIndex],
				Value: text[position.ValueStartIndex:nextPosition.LabelStartIndex],
			})
		} else {
			*b = append(*b, Field{
				Name:  text[position.LabelStartIndex:position.LabelEndIndex],
				Value: text[position.ValueStartIndex:absoluteFieldListEndIndex],
			})
		}
	}

	return true, before, after
}

type DividerBlock struct{}

func (DividerBlock) FromText(text string) (ok bool, before, after string) {
	match := dividerRegex.FindStringIndex(text)
	if match == nil {
		return false, "", ""
	}

	matchStartIndex, matchEndIndex := match[0], match[1]

	return true, text[:matchStartIndex], text[matchEndIndex:]
}

type AttributionBlock struct {
	Name        string
	Time        *time.Time
	IncludeTime bool
}

func (b *AttributionBlock) FromText(text string) (ok bool, before, after string) {
	for _, regex := range attributionNameRegexes {
		match := regex.FindStringSubmatchIndex(text)
		if match == nil {
			continue
		}

		matchStartIndex, matchEndIndex := match[0], match[1]
		nameStartIndex, nameEndIndex := match[2], match[3]

		b.Name = strings.TrimSuffix(strings.TrimPrefix(text[nameStartIndex:nameEndIndex], "\""), "\"")

		return true, text[:matchStartIndex], text[matchEndIndex:]
	}

	return false, "", ""
}
