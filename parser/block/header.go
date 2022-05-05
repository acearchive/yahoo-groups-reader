package block

import (
	"fmt"
	"regexp"
	"strings"
)

const fieldNameRegexPart = `From|Reply-To|To|Subject|Date|Sent|Message`

var (
	messageHeaderBannerRegexPart = fmt.Sprintf(`%[1]s-+ ?Original Message ?-+%[1]s`, nonNewlineWhitespaceRegexPart)
	fieldLabelRegex              = regexp.MustCompile(fmt.Sprintf(`(?m)^%s(%s): +(\S)`, nonNewlineWhitespaceRegexPart, fieldNameRegexPart))
	messageHeaderStartRegex      = regexp.MustCompile(fmt.Sprintf(`(?:^%[2]s\n|^%[1]s\n?|\n%[1]s(?:%[2]s)?\n)%[1]s(%[3]s): +(\S)`, nonNewlineWhitespaceRegexPart, messageHeaderBannerRegexPart, fieldNameRegexPart))
	messageHeaderEndRegex        = regexp.MustCompile(fmt.Sprintf(`(?m)^%s\n`, nonNewlineWhitespaceRegexPart))
)

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
		var nextField Field
		if i+1 < len(fieldPositions) {
			nextPosition := fieldPositions[i+1]

			nextField.Name = text[position.LabelStartIndex:position.LabelEndIndex]
			nextField.Value = text[position.ValueStartIndex:nextPosition.LabelStartIndex]
		} else {
			nextField.Name = text[position.LabelStartIndex:position.LabelEndIndex]
			nextField.Value = text[position.ValueStartIndex:absoluteFieldListEndIndex]
		}

		nextField.Name = strings.TrimSpace(nextField.Name)
		nextField.Value = strings.TrimSpace(nextField.Value)

		*b = append(*b, nextField)
	}

	return true, before, after
}
