package flowed_test

import (
	"github.com/acearchive/yg-render/render/flowed"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ParseLine", func() {
})

var _ = Describe("Tokenize", func() {
	Specify("fixed lines break paragraphs", func() {
		lines := []flowed.Line{
			{Kind: flowed.LineTypeFlowed, Content: "flowed line 1"},
			{Kind: flowed.LineTypeFlowed, Content: "flowed line 2"},
			{Kind: flowed.LineTypeFixed, Content: "fixed line 1"},
			{Kind: flowed.LineTypeFlowed, Content: "flowed line 3"},
			{Kind: flowed.LineTypeFixed, Content: "fixed line 2"},
		}

		Expect(flowed.Tokenize(lines)).To(Equal([]flowed.Token{
			flowed.TokenStartParagraph,
			flowed.NewTextToken("flowed line 1"),
			flowed.NewTextToken("flowed line 2"),
			flowed.NewTextToken("fixed line 1"),
			flowed.TokenEndParagraph,
			flowed.TokenStartParagraph,
			flowed.NewTextToken("flowed line 3"),
			flowed.NewTextToken("fixed line 2"),
			flowed.TokenEndParagraph,
		}))
	})

	Specify("EOF ends the last paragraph", func() {
		lines := []flowed.Line{
			{Kind: flowed.LineTypeFlowed, Content: "flowed line 1"},
			{Kind: flowed.LineTypeFlowed, Content: "flowed line 2"},
		}

		Expect(flowed.Tokenize(lines)).To(Equal([]flowed.Token{
			flowed.TokenStartParagraph,
			flowed.NewTextToken("flowed line 1"),
			flowed.NewTextToken("flowed line 2"),
			flowed.TokenEndParagraph,
		}))
	})

	Specify("adjacent fixed lines are separate paragraphs", func() {
		lines := []flowed.Line{
			{Kind: flowed.LineTypeFlowed, Content: "flowed line 1"},
			{Kind: flowed.LineTypeFixed, Content: "fixed line 1"},
			{Kind: flowed.LineTypeFixed, Content: "fixed line 2"},
			{Kind: flowed.LineTypeFixed, Content: "fixed line 3"},
		}

		Expect(flowed.Tokenize(lines)).To(Equal([]flowed.Token{
			flowed.TokenStartParagraph,
			flowed.NewTextToken("flowed line 1"),
			flowed.NewTextToken("fixed line 1"),
			flowed.TokenEndParagraph,
			flowed.TokenStartParagraph,
			flowed.NewTextToken("fixed line 2"),
			flowed.TokenEndParagraph,
			flowed.TokenStartParagraph,
			flowed.NewTextToken("fixed line 3"),
			flowed.TokenEndParagraph,
		}))
	})

	Specify("signature lines do not break paragraphs", func() {
		lines := []flowed.Line{
			{Kind: flowed.LineTypeFlowed, Content: "flowed line 1"},
			{Kind: flowed.LineTypeSignature, Content: "signature line"},
			{Kind: flowed.LineTypeFlowed, Content: "flowed line 2"},
			{Kind: flowed.LineTypeFixed, Content: "fixed line 1"},
		}
		Expect(flowed.Tokenize(lines)).To(Equal([]flowed.Token{
			flowed.TokenStartParagraph,
			flowed.NewTextToken("flowed line 1"),
			flowed.TokenSignatureLine,
			flowed.NewTextToken("flowed line 2"),
			flowed.NewTextToken("fixed line 1"),
			flowed.TokenEndParagraph,
		}))
	})
})
