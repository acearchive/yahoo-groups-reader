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

	Specify("signature lines break paragraphs", func() {
		lines := []flowed.Line{
			{Kind: flowed.LineTypeFlowed, Content: "flowed line 1"},
			{Kind: flowed.LineTypeSignature},
			{Kind: flowed.LineTypeFlowed, Content: "flowed line 2"},
			{Kind: flowed.LineTypeFixed, Content: "fixed line 1"},
		}
		Expect(flowed.Tokenize(lines)).To(Equal([]flowed.Token{
			flowed.TokenStartParagraph,
			flowed.NewTextToken("flowed line 1"),
			flowed.TokenEndParagraph,
			flowed.TokenSignatureLine,
			flowed.TokenStartParagraph,
			flowed.NewTextToken("flowed line 2"),
			flowed.NewTextToken("fixed line 1"),
			flowed.TokenEndParagraph,
		}))
	})

	Specify("adjacent signature lines break paragraphs", func() {
		lines := []flowed.Line{
			{Kind: flowed.LineTypeFlowed, Content: "flowed line 1"},
			{Kind: flowed.LineTypeSignature},
			{Kind: flowed.LineTypeSignature},
			{Kind: flowed.LineTypeFlowed, Content: "flowed line 2"},
			{Kind: flowed.LineTypeFixed, Content: "fixed line 1"},
		}
		Expect(flowed.Tokenize(lines)).To(Equal([]flowed.Token{
			flowed.TokenStartParagraph,
			flowed.NewTextToken("flowed line 1"),
			flowed.TokenEndParagraph,
			flowed.TokenSignatureLine,
			flowed.TokenSignatureLine,
			flowed.TokenStartParagraph,
			flowed.NewTextToken("flowed line 2"),
			flowed.NewTextToken("fixed line 1"),
			flowed.TokenEndParagraph,
		}))
	})

	Specify("a signature line before EOF ends the paragraph", func() {
		// A paragraph ending with a signature line is forbidden by the spec,
		// but we want to handle it sanely anyway.
		lines := []flowed.Line{
			{Kind: flowed.LineTypeFlowed, Content: "flowed line"},
			{Kind: flowed.LineTypeSignature},
		}
		Expect(flowed.Tokenize(lines)).To(Equal([]flowed.Token{
			flowed.TokenStartParagraph,
			flowed.NewTextToken("flowed line"),
			flowed.TokenEndParagraph,
			flowed.TokenSignatureLine,
		}))
	})

	Specify("a quote block containing a single paragraph", func() {
		lines := []flowed.Line{
			{Kind: flowed.LineTypeFlowed, QuoteDepth: 1, Content: "quoted flowed line"},
			{Kind: flowed.LineTypeFixed, QuoteDepth: 1, Content: "quoted fixed line"},
		}

		Expect(flowed.Tokenize(lines)).To(Equal([]flowed.Token{
			flowed.TokenStartQuote,
			flowed.TokenStartParagraph,
			flowed.NewTextToken("quoted flowed line"),
			flowed.NewTextToken("quoted fixed line"),
			flowed.TokenEndParagraph,
			flowed.TokenEndQuote,
		}))
	})

	Specify("a quote block containing a multiple paragraphs", func() {
		lines := []flowed.Line{
			{Kind: flowed.LineTypeFixed, QuoteDepth: 1, Content: "quoted fixed line 1"},
			{Kind: flowed.LineTypeFlowed, QuoteDepth: 1, Content: "quoted flowed line"},
			{Kind: flowed.LineTypeFixed, QuoteDepth: 1, Content: "quoted fixed line 2"},
		}

		Expect(flowed.Tokenize(lines)).To(Equal([]flowed.Token{
			flowed.TokenStartQuote,
			flowed.TokenStartParagraph,
			flowed.NewTextToken("quoted fixed line 1"),
			flowed.TokenEndParagraph,
			flowed.TokenStartParagraph,
			flowed.NewTextToken("quoted flowed line"),
			flowed.NewTextToken("quoted fixed line 2"),
			flowed.TokenEndParagraph,
			flowed.TokenEndQuote,
		}))
	})

	Specify("paragraphs break on a quote depth change without a fixed line", func() {
		// This is considered an improperly formatted message by the spec, but
		// the RFC does specify how parsers should handle it.
		lines := []flowed.Line{
			{Kind: flowed.LineTypeFlowed, QuoteDepth: 1, Content: "quoted flowed line 1"},
			{Kind: flowed.LineTypeFlowed, QuoteDepth: 1, Content: "quoted flowed line 2"},
			{Kind: flowed.LineTypeFixed, Content: "fixed line"},
		}

		Expect(flowed.Tokenize(lines)).To(Equal([]flowed.Token{
			flowed.TokenStartQuote,
			flowed.TokenStartParagraph,
			flowed.NewTextToken("quoted flowed line 1"),
			flowed.NewTextToken("quoted flowed line 2"),
			flowed.TokenEndParagraph,
			flowed.TokenEndQuote,
			flowed.TokenStartParagraph,
			flowed.NewTextToken("fixed line"),
			flowed.TokenEndParagraph,
		}))
	})

	Specify("nested quote blocks with a depth change of 1", func() {
		lines := []flowed.Line{
			{Kind: flowed.LineTypeFixed, QuoteDepth: 1, Content: "quoted fixed line 1"},
			{Kind: flowed.LineTypeFixed, QuoteDepth: 2, Content: "quoted fixed line 2"},
			{Kind: flowed.LineTypeFixed, QuoteDepth: 1, Content: "quoted fixed line 3"},
		}

		Expect(flowed.Tokenize(lines)).To(Equal([]flowed.Token{
			flowed.TokenStartQuote,
			flowed.TokenStartParagraph,
			flowed.NewTextToken("quoted fixed line 1"),
			flowed.TokenEndParagraph,
			flowed.TokenStartQuote,
			flowed.TokenStartParagraph,
			flowed.NewTextToken("quoted fixed line 2"),
			flowed.TokenEndParagraph,
			flowed.TokenEndQuote,
			flowed.TokenStartParagraph,
			flowed.NewTextToken("quoted fixed line 3"),
			flowed.TokenEndParagraph,
			flowed.TokenEndQuote,
		}))
	})

	Specify("nested quote blocks abruptly ending", func() {
		lines := []flowed.Line{
			{Kind: flowed.LineTypeFixed, QuoteDepth: 1, Content: "quoted fixed line 1"},
			{Kind: flowed.LineTypeFixed, QuoteDepth: 2, Content: "quoted fixed line 2"},
			{Kind: flowed.LineTypeFixed, Content: "fixed line"},
		}

		Expect(flowed.Tokenize(lines)).To(Equal([]flowed.Token{
			flowed.TokenStartQuote,
			flowed.TokenStartParagraph,
			flowed.NewTextToken("quoted fixed line 1"),
			flowed.TokenEndParagraph,
			flowed.TokenStartQuote,
			flowed.TokenStartParagraph,
			flowed.NewTextToken("quoted fixed line 2"),
			flowed.TokenEndParagraph,
			flowed.TokenEndQuote,
			flowed.TokenEndQuote,
			flowed.TokenStartParagraph,
			flowed.NewTextToken("fixed line"),
			flowed.TokenEndParagraph,
		}))
	})

	Specify("nested quote blocks abruptly starting", func() {
		lines := []flowed.Line{
			{Kind: flowed.LineTypeFixed, QuoteDepth: 2, Content: "quoted fixed line 1"},
			{Kind: flowed.LineTypeFixed, QuoteDepth: 1, Content: "quoted fixed line 2"},
			{Kind: flowed.LineTypeFixed, Content: "fixed line"},
		}

		Expect(flowed.Tokenize(lines)).To(Equal([]flowed.Token{
			flowed.TokenStartQuote,
			flowed.TokenStartQuote,
			flowed.TokenStartParagraph,
			flowed.NewTextToken("quoted fixed line 1"),
			flowed.TokenEndParagraph,
			flowed.TokenEndQuote,
			flowed.TokenStartParagraph,
			flowed.NewTextToken("quoted fixed line 2"),
			flowed.TokenEndParagraph,
			flowed.TokenEndQuote,
			flowed.TokenStartParagraph,
			flowed.NewTextToken("fixed line"),
			flowed.TokenEndParagraph,
		}))
	})

	Specify("nested quote blocks abruptly starting and ending", func() {
		lines := []flowed.Line{
			{Kind: flowed.LineTypeFlowed, QuoteDepth: 2, Content: "quoted flowed line"},
			{Kind: flowed.LineTypeFixed, QuoteDepth: 2, Content: "quoted fixed line"},
		}

		Expect(flowed.Tokenize(lines)).To(Equal([]flowed.Token{
			flowed.TokenStartQuote,
			flowed.TokenStartQuote,
			flowed.TokenStartParagraph,
			flowed.NewTextToken("quoted flowed line"),
			flowed.NewTextToken("quoted fixed line"),
			flowed.TokenEndParagraph,
			flowed.TokenEndQuote,
			flowed.TokenEndQuote,
		}))
	})

	Specify("a signature line breaks paragraphs in a quote block", func() {
		lines := []flowed.Line{
			{Kind: flowed.LineTypeFlowed, QuoteDepth: 1, Content: "quoted flowed line"},
			{Kind: flowed.LineTypeSignature, QuoteDepth: 1},
			{Kind: flowed.LineTypeFixed, QuoteDepth: 1, Content: "quoted fixed line"},
		}

		Expect(flowed.Tokenize(lines)).To(Equal([]flowed.Token{
			flowed.TokenStartQuote,
			flowed.TokenStartParagraph,
			flowed.NewTextToken("quoted flowed line"),
			flowed.TokenEndParagraph,
			flowed.TokenSignatureLine,
			flowed.TokenStartParagraph,
			flowed.NewTextToken("quoted fixed line"),
			flowed.TokenEndParagraph,
			flowed.TokenEndQuote,
		}))
	})

	Specify("EOF ends the paragraph and quote block", func() {
		lines := []flowed.Line{
			{Kind: flowed.LineTypeFlowed, QuoteDepth: 1, Content: "quoted flowed line 1"},
			{Kind: flowed.LineTypeFlowed, QuoteDepth: 1, Content: "quoted flowed line 2"},
		}

		Expect(flowed.Tokenize(lines)).To(Equal([]flowed.Token{
			flowed.TokenStartQuote,
			flowed.TokenStartParagraph,
			flowed.NewTextToken("quoted flowed line 1"),
			flowed.NewTextToken("quoted flowed line 2"),
			flowed.TokenEndParagraph,
			flowed.TokenEndQuote,
		}))
	})
})
