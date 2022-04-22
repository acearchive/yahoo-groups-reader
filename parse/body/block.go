package body

import (
	"regexp"
	"time"
)

var (
	attributionNameRegex = regexp.MustCompile(`^--- In [^\s@]+@(?:yahoogroups\.com|y?\.{3}), ([^<>]+)(?: <[^<>\s]+>)? wrote:$`)
	dividerRegex         = regexp.MustCompile(`^[-_]{2,}$`)
	fieldRegex           = regexp.MustCompile(`^(From|Reply-To|To|Subject|Date|Sent|Message): +(\S.*)$`)
	messageHeaderRegex   = regexp.MustCompile(`^-+ ?Original Message ?-+$`)
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

func (b *MessageHeaderBlock) FromText(text string) (ok bool, before, after string) {
	// TODO: Implement
	return false, "", ""
}

type DividerBlock struct{}

func (DividerBlock) FromText(text string) (ok bool, before, after string) {
	// TODO: Implement
	return false, "", ""
}

type AttributionBlock struct {
	Name        string
	Time        *time.Time
	IncludeTime bool
}

func (b *AttributionBlock) FromText(text string) (ok bool, before, after string) {
	// TODO: Implement
	return false, "", ""
}
