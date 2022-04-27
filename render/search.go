package render

import (
	"encoding/json"
	"github.com/acearchive/yg-render/body"
	"github.com/acearchive/yg-render/parse"
	"os"
	"path/filepath"
	"strings"
)

const searchFileName = "search.json"

type MessageSearchFields struct {
	User  string `json:"user"`
	Flair string `json:"flair"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

func tokensToSearchText(tokens []body.Token) string {
	var builder strings.Builder

	quoteLevel := 0

	for _, token := range tokens {
		switch concreteToken := token.(type) {
		case body.StartQuoteToken:
			quoteLevel++
		case body.EndQuoteToken:
			quoteLevel--
		case body.TextToken:
			if quoteLevel > 0 {
				continue
			}

			builder.WriteString(string(concreteToken))
		case body.EndParagraphToken:
			builder.WriteString("\n")
		}
	}

	return builder.String()
}

func buildSearchFields(thread parse.MessageThread) []MessageSearchFields {
	fields := make([]MessageSearchFields, 0, len(thread))

	for _, message := range thread {
		field := MessageSearchFields{
			User:  message.User,
			Flair: message.Flair,
			Body:  tokensToSearchText(message.Body.Tokens),
		}

		if message.Title != nil {
			field.Title = *message.Title
		}

		fields = append(fields, field)
	}

	return fields
}

func writeSearchData(thread parse.MessageThread, path string) error {
	fields := buildSearchFields(thread)

	jsonFile, err := os.OpenFile(filepath.Join(path, searchFileName), os.O_CREATE|os.O_EXCL|os.O_WRONLY, outputFileMode)
	if err != nil {
		return err
	}

	if err := json.NewEncoder(jsonFile).Encode(fields); err != nil {
		return err
	}

	return jsonFile.Close()
}
