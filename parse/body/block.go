package body

import (
	"errors"
	"fmt"
	"regexp"
	"time"
)

var (
	ErrInvalidAttributionFormat = errors.New("invalid attributionFormat")
	ErrInvalidRegex             = errors.New("invalid regex")
)

const (
	attributionNameRegexPart                 = `(?:[^<>\s]|[^<>\s][^<>]*[^<>\s])`
	attributionEmailRegexPart                = `[^<>\s]+`
	attributionGroupEmailRegexPart           = `[^\s@]+@(?:yahoogroups\.com|y?\.{3})`
	attributionShortMonthRegexPart           = `(?:Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)`
	attributionShortWeekdayRegexPart         = `(?:Mon|Tue|Wed|Thu|Fri|Sat|Sun)`
	attributionTimeRegexPart                 = `\d{2}:\d{2}:\d{2}`
	attributionNumericTimezoneRegexPart      = `[-+]\d{4}`
	attributionAbbreviationTimezoneRegexPart = `\([A-Z]{2,}\)`
)

const (
	attributionDateFormat                 = "Mon, 2 Jan 2006"
	attributionNumericTimezoneFormat      = "Mon, 2 Jan 2006 15:04:05 -0700"
	attributionAbbreviationTimezoneFormat = "Mon, 2 Jan 2006 15:04:05 -0700 (MST)"
)

type attributionFormat string

const (
	attributionFormatName                         attributionFormat = "Name"
	attributionFormatNameDate                     attributionFormat = "NameDate"
	attributionFormatNameDateNumericTimezone      attributionFormat = "NameDateNumericTimezone"
	attributionFormatNameDateAbbreviationTimezone attributionFormat = "NameDateAbbreviationTimezone"
)

func (f attributionFormat) HasTime() bool {
	switch f {
	case attributionFormatNameDateNumericTimezone, attributionFormatNameDateAbbreviationTimezone:
		return true
	case attributionFormatName, attributionFormatNameDate:
		return false
	default:
		panic(fmt.Errorf("%w: %s", ErrInvalidAttributionFormat, f))
	}
}

func (f attributionFormat) DateFormat() *string {
	var format string

	switch f {
	case attributionFormatName:
		return nil
	case attributionFormatNameDate:
		format = attributionDateFormat
	case attributionFormatNameDateNumericTimezone:
		format = attributionNumericTimezoneFormat
	case attributionFormatNameDateAbbreviationTimezone:
		format = attributionAbbreviationTimezoneFormat
	default:
		panic(fmt.Errorf("%w: %s", ErrInvalidAttributionFormat, f))
	}

	return &format
}

func indicesForSubmatch(number int) []int {
	return []int{2 * number, 2*number + 1}
}

func (f attributionFormat) TimeIndices(match []int) []int {
	switch f {
	case attributionFormatName:
		return nil
	case attributionFormatNameDate, attributionFormatNameDateNumericTimezone, attributionFormatNameDateAbbreviationTimezone:
		return indicesForSubmatch(1)
	default:
		panic(fmt.Errorf("%w: %s", ErrInvalidAttributionFormat, f))
	}

}

func (f attributionFormat) NameIndices(match []int) []int {
	var firstNameSubmatchIndex int

	switch f {
	case attributionFormatName:
		firstNameSubmatchIndex = 1
	case attributionFormatNameDate, attributionFormatNameDateNumericTimezone, attributionFormatNameDateAbbreviationTimezone:
		firstNameSubmatchIndex = 2
	default:
		panic(fmt.Errorf("%w: %s", ErrInvalidAttributionFormat, f))
	}

	for i := firstNameSubmatchIndex; i <= attributionUserCapturingRegexPartNumSubmatches; i++ {
		// Try each capture group in `attributionUserCapturingRegexPart`
		// until we find the first one that matched.
		submatchIndices := indicesForSubmatch(i)
		startIndex, endIndex := submatchIndices[0], submatchIndices[1]
		if startIndex >= 0 && endIndex >= 0 {
			return []int{startIndex, endIndex}
		}
	}

	panic(ErrInvalidRegex)
}

type attributionRegex struct {
	Format attributionFormat
	Regex  *regexp.Regexp
}

const attributionUserCapturingRegexPartNumSubmatches = 4

var (
	dividerRegex                          = regexp.MustCompile(`(?m)^\s*[-_]{2,}\s*$`)
	fieldLabelRegex                       = regexp.MustCompile(`(?m)^\s*(From|Reply-To|To|Subject|Date|Sent|Message): +(\S)`)
	messageHeaderStartRegex               = regexp.MustCompile(`(?m)^\s*(?:-+ ?Original Message ?-+|\n)\s*\n\s*(From|Reply-To|To|Subject|Date|Sent|Message): +(\S)`)
	messageHeaderEndRegex                 = regexp.MustCompile(`(?m)^\s*\n`)
	attributionUserCapturingRegexPart     = fmt.Sprintf(`(?:"(%[1]s)"\s+<%[2]s>|(%[1]s)\s+<%[2]s>|<(%[2]s)>|(%[1]s))`, attributionNameRegexPart, attributionEmailRegexPart)
	dateRegexPart                         = fmt.Sprintf(`%s, \d{1,2} %s \d{4}`, attributionShortWeekdayRegexPart, attributionShortMonthRegexPart)
	timeWithNumericTimezoneRegexPart      = fmt.Sprintf(`%s %s %s`, dateRegexPart, attributionTimeRegexPart, attributionNumericTimezoneRegexPart)
	timeWithAbbreviationTimezoneRegexPart = fmt.Sprintf(`%s %s %s %s`, dateRegexPart, attributionTimeRegexPart, attributionNumericTimezoneRegexPart, attributionAbbreviationTimezoneRegexPart)
	attributionRegexes                    = []attributionRegex{
		{
			Format: attributionFormatName,
			Regex: regexp.MustCompile(fmt.Sprintf(
				`(?m)^--- In %s,\s+%s\s+wrote:$`,
				attributionGroupEmailRegexPart,
				attributionUserCapturingRegexPart,
			)),
		},
		{
			Format: attributionFormatName,
			Regex: regexp.MustCompile(fmt.Sprintf(
				`(?m)^---\s+%s\s+wrote:$`,
				attributionUserCapturingRegexPart,
			)),
		},
		{
			Format: attributionFormatNameDate,
			Regex: regexp.MustCompile(fmt.Sprintf(
				`(?m)^On\s+(%s),\s+%s\s+wrote:$`,
				dateRegexPart,
				attributionUserCapturingRegexPart,
			)),
		},
		{
			Format: attributionFormatNameDateNumericTimezone,
			Regex: regexp.MustCompile(fmt.Sprintf(
				`(?m)^On\s+(%s),\s+%s\s+wrote:$`,
				timeWithNumericTimezoneRegexPart,
				attributionUserCapturingRegexPart,
			)),
		},
		{
			Format: attributionFormatNameDateAbbreviationTimezone,
			Regex: regexp.MustCompile(fmt.Sprintf(
				`(?m)^On\s+(%s),\s+%s\s+wrote:$`,
				timeWithAbbreviationTimezoneRegexPart,
				attributionUserCapturingRegexPart,
			)),
		},
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
	Name    string
	Time    *time.Time
	HasTime bool
}

func (b *AttributionBlock) FromText(text string) (ok bool, before, after string) {
	for _, regex := range attributionRegexes {
		match := regex.Regex.FindStringSubmatchIndex(text)
		if match == nil {
			continue
		}

		matchStartIndex, matchEndIndex := match[0], match[1]
		nameIndices := regex.Format.NameIndices(match)

		b.Name = text[nameIndices[0]:nameIndices[1]]

		if dateFormat := regex.Format.DateFormat(); dateFormat != nil {
			timeIndices := regex.Format.TimeIndices(match)
			localTime, err := time.Parse(*dateFormat, text[timeIndices[0]:timeIndices[1]])
			if err != nil {
				continue
			}

			dateTime := localTime.UTC()
			b.Time = &dateTime
		}

		b.HasTime = regex.Format.HasTime()

		return true, text[:matchStartIndex], text[matchEndIndex:]
	}

	return false, "", ""
}
