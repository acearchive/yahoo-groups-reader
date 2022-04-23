package body

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var ErrInvalidAttributionFormat = errors.New("invalid attributionFormat")

const (
	attributionEmailRegexPart      = `<[^<>\s]+>`
	attributionNameRegexPart       = `(?:[^<>\s]|[^<>\s][^<>]*[^<>\s])`
	attributionGroupEmailRegexPart = `[^\s@]+@(?:yahoogroups\.com|y?\.{3})`
	shortMonthRegexPart            = `(?:Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)`
	shortWeekDayRegexPart          = `(?:Mon|Tue|Wed|Thu|Fri|Sat|Sun)`
	timeRegexPart                  = `\d{2}:\d{2}:\d{2}`
	numericTimeZoneRegexPart       = `[-+]\d{4}`
	abbreviationTimeZoneRegexPart  = `\([A-Z]{2,}\)`
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

func (f attributionFormat) TimeIndices(match []int) []int {
	switch f {
	case attributionFormatName:
		return nil
	case attributionFormatNameDate, attributionFormatNameDateNumericTimezone, attributionFormatNameDateAbbreviationTimezone:
		return []int{match[2], match[3]}
	default:
		panic(fmt.Errorf("%w: %s", ErrInvalidAttributionFormat, f))
	}

}

func (f attributionFormat) NameIndices(match []int) []int {
	switch f {
	case attributionFormatName:
		return []int{match[2], match[3]}
	case attributionFormatNameDate, attributionFormatNameDateNumericTimezone, attributionFormatNameDateAbbreviationTimezone:
		return []int{match[4], match[5]}
	default:
		panic(fmt.Errorf("%w: %s", ErrInvalidAttributionFormat, f))
	}
}

type attributionRegex struct {
	Format attributionFormat
	Regex  *regexp.Regexp
}

var (
	dividerRegex                          = regexp.MustCompile(`(?m)^\s*[-_]{2,}\s*$`)
	fieldLabelRegex                       = regexp.MustCompile(`(?m)^\s*(From|Reply-To|To|Subject|Date|Sent|Message): +(\S)`)
	messageHeaderStartRegex               = regexp.MustCompile(`(?m)^\s*(?:-+ ?Original Message ?-+|\n)\s*\n\s*(From|Reply-To|To|Subject|Date|Sent|Message): +(\S)`)
	messageHeaderEndRegex                 = regexp.MustCompile(`(?m)^\s*\n`)
	dateRegexPart                         = fmt.Sprintf(`%s, \d{1,2} %s \d{4}`, shortWeekDayRegexPart, shortMonthRegexPart)
	timeWithNumericTimezoneRegexPart      = fmt.Sprintf(`%s %s %s`, dateRegexPart, timeRegexPart, numericTimeZoneRegexPart)
	timeWithAbbreviationTimezoneRegexPart = fmt.Sprintf(`%s %s %s %s`, dateRegexPart, timeRegexPart, numericTimeZoneRegexPart, abbreviationTimeZoneRegexPart)
	attributionRegexes                    = []attributionRegex{
		{
			Format: attributionFormatName,
			Regex: regexp.MustCompile(fmt.Sprintf(
				`(?m)^--- In %s,\s+(%s)(?:\s+%s)?\s+wrote:$`,
				attributionGroupEmailRegexPart,
				attributionNameRegexPart,
				attributionEmailRegexPart,
			)),
		},
		{
			Format: attributionFormatName,
			Regex: regexp.MustCompile(fmt.Sprintf(
				`(?m)^---\s+(%s)(?:\s+%s)?\s+wrote:$`,
				attributionNameRegexPart,
				attributionEmailRegexPart,
			)),
		},
		{
			Format: attributionFormatNameDate,
			Regex: regexp.MustCompile(fmt.Sprintf(
				`(?m)^On\s+(%s),\s+(%s)(?:\s+%s)?\s+wrote:$`,
				dateRegexPart,
				attributionNameRegexPart,
				attributionEmailRegexPart,
			)),
		},
		{
			Format: attributionFormatNameDateNumericTimezone,
			Regex: regexp.MustCompile(fmt.Sprintf(
				`(?m)^On\s+(%s),\s+(%s)(?:\s+%s)?\s+wrote:$`,
				timeWithNumericTimezoneRegexPart,
				attributionNameRegexPart,
				attributionEmailRegexPart,
			)),
		},
		{
			Format: attributionFormatNameDateAbbreviationTimezone,
			Regex: regexp.MustCompile(fmt.Sprintf(
				`(?m)^On\s+(%s),\s+(%s)(?:\s+%s)?\s+wrote:$`,
				timeWithAbbreviationTimezoneRegexPart,
				attributionNameRegexPart,
				attributionEmailRegexPart,
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

		b.Name = strings.TrimSuffix(strings.TrimPrefix(text[nameIndices[0]:nameIndices[1]], "\""), "\"")

		if dateFormat := regex.Format.DateFormat(); dateFormat != nil {
			timeIndices := regex.Format.TimeIndices(match)
			datetime, err := time.Parse(*dateFormat, text[timeIndices[0]:timeIndices[1]])
			if err != nil {
				continue
			}

			b.Time = &datetime
		}

		b.HasTime = regex.Format.HasTime()

		return true, text[:matchStartIndex], text[matchEndIndex:]
	}

	return false, "", ""
}
