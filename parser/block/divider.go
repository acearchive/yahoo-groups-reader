package block

import (
	"fmt"
	"regexp"
)

var dividerRegex = regexp.MustCompile(fmt.Sprintf(`(?m)^%[1]s(?:-{2,}|_{2,}|#{2,})%[1]s$`, nonNewlineWhitespaceRegexPart))

type DividerBlock struct{}

func (DividerBlock) FromText(text string) (ok bool, before, after string) {
	match := dividerRegex.FindStringIndex(text)
	if match == nil {
		return false, "", ""
	}

	matchStartIndex, matchEndIndex := match[0], match[1]

	return true, text[:matchStartIndex], text[matchEndIndex:]
}
