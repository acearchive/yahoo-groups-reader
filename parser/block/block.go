package block

const nonNewlineWhitespaceRegexPart = `[\t ]*`

type Block interface {
	ToHtml() string
	FromText(text string) (ok bool, before, after string)
}

func AllBlocks() []Block {
	return []Block{
		&HardBreakBlock{},
		&DividerBlock{},
		&MessageHeaderBlock{},
		&AttributionBlock{},
	}
}
