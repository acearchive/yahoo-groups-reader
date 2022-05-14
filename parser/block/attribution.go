package block

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var (
	ErrInvalidDateFormat       = errors.New("invalid date format")
	ErrInvalidTimeFormat       = errors.New("invalid time format")
	ErrInvalidNameFormat       = errors.New("invalid name format")
	ErrInvalidCaptureKind      = errors.New("invalid capture kind")
	ErrNoMatchingCaptureGroups = errors.New("match has no matching capture groups")
)

const (
	attributionNameRegexPart       = `(?:[^<>,"\s]|[^<>,"\s][^<>,"]*[^<>,"\s])`
	attributionEmailRegexPart      = `[^<>@\s]+@[^<>@\s]*`
	attributionGroupEmailRegexPart = `[^\s@]+@(?:yahoogroups\.com|y?\.{3})`
	shortMonthRegexPart            = `(?:Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)`
	shortWeekdayRegexPart          = `(?:Mon|Tue|Wed|Thu|Fri|Sat|Sun)`
)

type regexMatcher interface {
	Regex() *regexp.Regexp
}

func joinMatchers(matchers []regexMatcher) string {
	regexParts := make([]string, len(matchers))

	for i, matcher := range matchers {
		regexParts[i] = matcher.Regex().String()
	}

	return fmt.Sprintf("(?:%s)", strings.Join(regexParts, "|"))
}

func joinNameFormats(formats []nameFormat) string {
	matchers := make([]regexMatcher, len(formats))

	for i, format := range formats {
		matchers[i] = format
	}

	return joinMatchers(matchers)
}

func joinDateFormats(formats []dateFormat) string {
	matchers := make([]regexMatcher, len(formats))

	for i, format := range formats {
		matchers[i] = format
	}

	return joinMatchers(matchers)
}

func joinTimeFormats(formats []timeFormat) string {
	matchers := make([]regexMatcher, len(formats))

	for i, format := range formats {
		matchers[i] = format
	}

	return joinMatchers(matchers)
}

type nameFormat string

const (
	nameFormatName                     = "Name"
	nameFormatEmail                    = "Email"
	nameFormatNameEmail                = "NameEmail"
	nameFormatQuotedName               = "QuotedName"
	nameFormatQuotedNameEmail          = "QuotedNameEmail"
	nameFormatQuotedNameDuplicateEmail = "QuotedNameDuplicateEmail"
)

func allNameFormats() []nameFormat {
	return []nameFormat{
		nameFormatQuotedNameDuplicateEmail,
		nameFormatQuotedNameEmail,
		nameFormatNameEmail,
		nameFormatEmail,
		nameFormatQuotedName,
		nameFormatName,
	}
}

func allEmailNameFormats() []nameFormat {
	return []nameFormat{
		nameFormatQuotedNameDuplicateEmail,
		nameFormatQuotedNameEmail,
		nameFormatNameEmail,
		nameFormatEmail,
	}
}

func (f nameFormat) Regex() *regexp.Regexp {
	switch f {
	case nameFormatName:
		return regexp.MustCompile(fmt.Sprintf(`(%s)`, attributionNameRegexPart))
	case nameFormatEmail:
		return regexp.MustCompile(fmt.Sprintf(`<(%s)>`, attributionEmailRegexPart))
	case nameFormatNameEmail:
		return regexp.MustCompile(fmt.Sprintf(`(%s)\s+<%s>`, attributionNameRegexPart, attributionEmailRegexPart))
	case nameFormatQuotedName:
		return regexp.MustCompile(fmt.Sprintf(`"(%s)"`, attributionNameRegexPart))
	case nameFormatQuotedNameEmail:
		return regexp.MustCompile(fmt.Sprintf(`"(%s)"\s+<%s>`, attributionNameRegexPart, attributionEmailRegexPart))
	case nameFormatQuotedNameDuplicateEmail:
		return regexp.MustCompile(fmt.Sprintf(`"(%[1]s)\s+<%[2]s>"\s+<%[2]s>`, attributionNameRegexPart, attributionEmailRegexPart))
	default:
		panic(fmt.Errorf("%w: %s", ErrInvalidNameFormat, f))
	}
}

type dateFormat string

const (
	dateFormatShort                    = "Short"
	dateFormatShortWeekday             = "ShortWeekday"
	dateFormatShortPadded              = "ShortPadded"
	dateFormatShortPaddedWeekday       = "ShortPaddedWeekday"
	dateFormatShortYearMonthDay        = "ShortYearMonthDay"
	dateFormatShortYearMonthDayWeekday = "ShortYearMonthDayWeekday"
	dateFormatLongDayMonthYear         = "LongDayMonthYear"
	dateFormatLongMonthDayYear         = "LongMonthDayYear"
	dateFormatLongDayMonthYearWeekday  = "LongDayMonthYearWeekday"
	dateFormatLongMonthDayYearWeekday  = "LongMonthDayYearWeekday"
)

func allDateFormats() []dateFormat {
	return []dateFormat{
		dateFormatLongDayMonthYearWeekday,
		dateFormatLongMonthDayYearWeekday,
		dateFormatLongDayMonthYear,
		dateFormatLongMonthDayYear,
		dateFormatShortYearMonthDayWeekday,
		dateFormatShortYearMonthDay,
		dateFormatShortPaddedWeekday,
		dateFormatShortWeekday,
		dateFormatShortPadded,
		dateFormatShort,
	}
}

func (f dateFormat) FormatString() string {
	switch f {
	case dateFormatShort:
		return "1/2/06"
	case dateFormatShortWeekday:
		return "Mon, 1/2/06"
	case dateFormatShortPadded:
		return "01/02/06"
	case dateFormatShortPaddedWeekday:
		return "Mon, 01/02/06"
	case dateFormatShortYearMonthDay:
		return "2006-01-02"
	case dateFormatShortYearMonthDayWeekday:
		return "Mon, 2006-01-02"
	case dateFormatLongDayMonthYear:
		return "2 Jan 2006"
	case dateFormatLongMonthDayYear:
		return "Jan 2, 2006"
	case dateFormatLongDayMonthYearWeekday:
		return "Mon, 2 Jan 2006"
	case dateFormatLongMonthDayYearWeekday:
		return "Mon, Jan 2, 2006"
	default:
		panic(fmt.Errorf("%w: %s", ErrInvalidDateFormat, f))
	}
}

func (f dateFormat) Regex() *regexp.Regexp {
	switch f {
	case dateFormatShort:
		return regexp.MustCompile(`(\d{1,2}/\d{1,2}/\d{2})`)
	case dateFormatShortWeekday:
		return regexp.MustCompile(fmt.Sprintf(`(%s, \d{1,2}/\d{1,2}/\d{2})`, shortWeekdayRegexPart))
	case dateFormatShortPadded:
		return regexp.MustCompile(`(\d{2}/\d{2}/\d{2})`)
	case dateFormatShortPaddedWeekday:
		return regexp.MustCompile(fmt.Sprintf(`(%s, \d{2}/\d{2}/\d{2})`, shortWeekdayRegexPart))
	case dateFormatShortYearMonthDay:
		return regexp.MustCompile(`(\d{4}-\d{2}-\d{2})`)
	case dateFormatShortYearMonthDayWeekday:
		return regexp.MustCompile(fmt.Sprintf(`(%s, \d{4}-\d{2}-\d{2})`, shortWeekdayRegexPart))
	case dateFormatLongDayMonthYear:
		return regexp.MustCompile(fmt.Sprintf(`(\d{1,2} %s \d{4})`, shortMonthRegexPart))
	case dateFormatLongMonthDayYear:
		return regexp.MustCompile(fmt.Sprintf(`(%s \d{1,2}, \d{4})`, shortMonthRegexPart))
	case dateFormatLongDayMonthYearWeekday:
		return regexp.MustCompile(fmt.Sprintf(`(%s, \d{1,2} %s \d{4})`, shortWeekdayRegexPart, shortMonthRegexPart))
	case dateFormatLongMonthDayYearWeekday:
		return regexp.MustCompile(fmt.Sprintf(`(%s, %s \d{1,2}, \d{4})`, shortWeekdayRegexPart, shortMonthRegexPart))
	default:
		panic(fmt.Errorf("%w: %s", ErrInvalidDateFormat, f))
	}
}

type timeFormat string

const (
	timeFormatShort12Hr  = "Short12Hr"
	timeFormatShort24Hr  = "Short24Hr"
	timeFormatLong       = "Long"
	timeFormatLongTzName = "LongTzName"
)

func allTimeFormats() []timeFormat {
	return []timeFormat{
		timeFormatLongTzName,
		timeFormatLong,
		timeFormatShort12Hr,
		timeFormatShort24Hr,
	}
}

func (f timeFormat) FormatString() string {
	switch f {
	case timeFormatShort12Hr:
		return "3:04 PM"
	case timeFormatShort24Hr:
		return "15:04"
	case timeFormatLong:
		return "15:04:05 -0700"
	case timeFormatLongTzName:
		return "15:04:05 -0700 (MST)"
	default:
		panic(fmt.Errorf("%w: %s", ErrInvalidTimeFormat, f))
	}
}

func (f timeFormat) Regex() *regexp.Regexp {
	switch f {
	case timeFormatShort12Hr:
		return regexp.MustCompile(`(\d{1,2}:\d{2} (?:AM|PM))`)
	case timeFormatShort24Hr:
		return regexp.MustCompile(`(\d{1,2}:\d{2})`)
	case timeFormatLong:
		return regexp.MustCompile(`(\d{2}:\d{2}:\d{2} [+-]\d{4})`)
	case timeFormatLongTzName:
		return regexp.MustCompile(`(\d{2}:\d{2}:\d{2} [+-]\d{4} \([A-Z]{2,5}\))`)
	default:
		panic(fmt.Errorf("%w: %s", ErrInvalidTimeFormat, f))
	}
}

func (f timeFormat) HasTimeZone() bool {
	switch f {
	case timeFormatShort12Hr, timeFormatShort24Hr:
		return false
	case timeFormatLong, timeFormatLongTzName:
		return true
	default:
		panic(fmt.Errorf("%w: %s", ErrInvalidTimeFormat, f))
	}
}

type attributionRegexPart interface {
	IsAttributionRegexPart()
}

type attributionRegexCapture string

const (
	attributionRegexCaptureName attributionRegexCapture = "Name"
	attributionRegexCaptureDate attributionRegexCapture = "Date"
	attributionRegexCaptureTime attributionRegexCapture = "Time"
)

func (attributionRegexCapture) IsAttributionRegexPart() {}

type attributionRegexLiteral string

func (attributionRegexLiteral) IsAttributionRegexPart() {}

type attributionRegex struct {
	Template    string
	Parts       []attributionRegexPart
	NameFormats []nameFormat
	DateFormats []dateFormat
	TimeFormats []timeFormat
	regex       *regexp.Regexp
}

func (r *attributionRegex) HasDate() bool {
	return len(r.DateFormats) > 0
}

func (r *attributionRegex) HasTime() bool {
	return len(r.TimeFormats) > 0
}

func (r *attributionRegex) Regex() *regexp.Regexp {
	if r.regex != nil {
		return r.regex
	}

	formatArgs := make([]interface{}, len(r.Parts))

	for partIndex, part := range r.Parts {
		switch concretePart := part.(type) {
		case attributionRegexCapture:
			switch concretePart {
			case attributionRegexCaptureName:
				formatArgs[partIndex] = joinNameFormats(r.NameFormats)
			case attributionRegexCaptureDate:
				formatArgs[partIndex] = joinDateFormats(r.DateFormats)
			case attributionRegexCaptureTime:
				formatArgs[partIndex] = joinTimeFormats(r.TimeFormats)
			}
		case attributionRegexLiteral:
			formatArgs[partIndex] = string(concretePart)
		}
	}

	r.regex = regexp.MustCompile(fmt.Sprintf(r.Template, formatArgs...))

	return r.regex
}

func (r *attributionRegex) matchersOfKind(kind attributionRegexCapture) []regexMatcher {
	var matchers []regexMatcher

	switch kind {
	case attributionRegexCaptureName:
		for _, format := range r.NameFormats {
			matchers = append(matchers, format)
		}
	case attributionRegexCaptureDate:
		for _, format := range r.DateFormats {
			matchers = append(matchers, format)
		}
	case attributionRegexCaptureTime:
		for _, format := range r.TimeFormats {
			matchers = append(matchers, format)
		}
	default:
		panic(fmt.Errorf("%w: %s", ErrInvalidCaptureKind, kind))
	}

	return matchers
}

func indicesForCaptureGroup(match []int, number int) (start, end int) {
	return match[2*number], match[2*number+1]
}

func (r *attributionRegex) MatchIndices(match []int, kind attributionRegexCapture) (start, end int, matcher regexMatcher) {
	precedingCaptureGroups := 0

	for _, part := range r.Parts {
		concretePart, isCapture := part.(attributionRegexCapture)

		if !isCapture {
			continue
		} else if concretePart == kind {
			break
		}

		precedingCaptureGroups += len(r.matchersOfKind(concretePart))
	}

	for i, matcher := range r.matchersOfKind(kind) {
		captureGroupNumber := 1 + precedingCaptureGroups + i
		startIndex, endIndex := indicesForCaptureGroup(match, captureGroupNumber)
		if startIndex >= 0 && endIndex >= 0 {
			return startIndex, endIndex, matcher
		}
	}

	panic(fmt.Errorf("%w of kind %s", ErrNoMatchingCaptureGroups, kind))
}

func (r *attributionRegex) NameIndices(match []int) (start, end int, format nameFormat) {
	start, end, matcher := r.MatchIndices(match, attributionRegexCaptureName)

	return start, end, matcher.(nameFormat)
}

func (r *attributionRegex) DateIndices(match []int) (start, end int, format dateFormat) {
	start, end, matcher := r.MatchIndices(match, attributionRegexCaptureDate)

	return start, end, matcher.(dateFormat)
}

func (r *attributionRegex) TimeIndices(match []int) (start, end int, format timeFormat) {
	start, end, matcher := r.MatchIndices(match, attributionRegexCaptureTime)

	return start, end, matcher.(timeFormat)
}

var attributionRegexes = []attributionRegex{
	{
		Template: `(?m)^%[1]s(?:-{2,3}\s+)?On\s+%[2]s\s+(?:at\s+)?%[3]s,?\s+%[4]s\s+wrote:\s+`,
		Parts: []attributionRegexPart{
			attributionRegexLiteral(nonNewlineWhitespaceRegexPart),
			attributionRegexCaptureDate,
			attributionRegexCaptureTime,
			attributionRegexCaptureName,
		},
		NameFormats: allNameFormats(),
		DateFormats: allDateFormats(),
		TimeFormats: allTimeFormats(),
	},
	{
		Template: `(?m)^%[1]s(?:-{2,3}\s+)?On\s+%[2]s,?\s+%[3]s\s+wrote:\s+`,
		Parts: []attributionRegexPart{
			attributionRegexLiteral(nonNewlineWhitespaceRegexPart),
			attributionRegexCaptureDate,
			attributionRegexCaptureName,
		},
		NameFormats: allNameFormats(),
		DateFormats: allDateFormats(),
		TimeFormats: nil,
	},
	{
		Template: `(?m)^%[1]s(?:-{2,3}\s+)?In\s+%[2]s,\s+%[3]s\s+wrote:\s+`,
		Parts: []attributionRegexPart{
			attributionRegexLiteral(nonNewlineWhitespaceRegexPart),
			attributionRegexLiteral(attributionGroupEmailRegexPart),
			attributionRegexCaptureName,
		},
		NameFormats: allNameFormats(),
		DateFormats: nil,
		TimeFormats: nil,
	},
	{
		Template: `(?m)^%[1]s-{2,3}\s+%[2]s\s+wrote:\s+`,
		Parts: []attributionRegexPart{
			attributionRegexLiteral(nonNewlineWhitespaceRegexPart),
			attributionRegexCaptureName,
		},
		NameFormats: allNameFormats(),
		DateFormats: nil,
		TimeFormats: nil,
	},
	{
		Template: `(?m)^%[1]s%[2]s%[1]swrote:\s+`,
		Parts: []attributionRegexPart{
			attributionRegexLiteral(nonNewlineWhitespaceRegexPart),
			attributionRegexCaptureName,
		},
		// We only allow name formats that include an email address to reduce
		// the likelihood of false positive matches on this pattern.
		NameFormats: allEmailNameFormats(),
		DateFormats: nil,
		TimeFormats: nil,
	},
}

type AttributionBlock struct {
	Name    string
	Time    time.Time
	HasTime bool
}

func (b *AttributionBlock) FromText(text string) (ok bool, before, after string) {
	for i := range attributionRegexes {
		regex := &attributionRegexes[i]

		match := regex.Regex().FindStringSubmatchIndex(text)
		if match == nil {
			continue
		}

		matchStartIndex, matchEndIndex := match[0], match[1]

		nameStartIndex, nameEndIndex, _ := regex.NameIndices(match)
		b.Name = text[nameStartIndex:nameEndIndex]

		var err error

		if regex.HasDate() {
			dateStartIndex, dateEndIndex, matchedDateFormat := regex.DateIndices(match)
			b.Time, err = time.Parse(matchedDateFormat.FormatString(), text[dateStartIndex:dateEndIndex])
			if err != nil {
				continue
			}
		}

		if regex.HasTime() {
			timeStartIndex, timeEndIndex, matchedTimeFormat := regex.TimeIndices(match)
			localTime, err := time.Parse(matchedTimeFormat.FormatString(), text[timeStartIndex:timeEndIndex])
			if err != nil {
				continue
			}

			if b.Time.IsZero() {
				b.Time = localTime
			} else {
				b.Time = b.Time.Add(localTime.Sub(time.Time{}))
			}
		}

		b.HasTime = regex.HasTime()

		return true, text[:matchStartIndex], text[matchEndIndex:]
	}

	return false, "", ""
}
