package block

import "regexp"

var hardBreakRegex = regexp.MustCompile(`(?:<br>\s*)+`)

type HardBreakBlock struct{}

func (b *HardBreakBlock) FromText(text string) (ok bool, before, after string) {
	if match := hardBreakRegex.FindStringIndex(text); match != nil {
		return true, text[:match[0]], text[match[1]:]
	}

	return false, "", ""
}
