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
	Index      int    `json:"id"`
	PageNumber int    `json:"page"`
	User       string `json:"user"`
	Flair      string `json:"flair"`
	Title      string `json:"title"`
	Body       string `json:"body"`
}

func pageNumberOfMessage(index, pageSize int) int {
	return calculateTotalPages(index, pageSize)
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

func buildSearchFields(thread parse.MessageThread, config OutputConfig) []MessageSearchFields {
	fields := make([]MessageSearchFields, 0, len(thread))

	sortedMessages, _ := thread.SortedByDate()

	for i, message := range sortedMessages {
		field := MessageSearchFields{
			Index:      i + 1,
			PageNumber: pageNumberOfMessage(i+1, config.PageSize),
			User:       message.User,
			Flair:      message.Flair,
			Body:       tokensToSearchText(message.Body.Tokens),
		}

		if message.Title != nil {
			field.Title = *message.Title
		}

		fields = append(fields, field)
	}

	return fields
}

func writeSearchData(thread parse.MessageThread, config OutputConfig, path string) error {
	fields := buildSearchFields(thread, config)

	jsonFile, err := os.OpenFile(filepath.Join(path, searchFileName), os.O_CREATE|os.O_EXCL|os.O_WRONLY, outputFileMode)
	if err != nil {
		return err
	}

	if err := json.NewEncoder(jsonFile).Encode(fields); err != nil {
		return err
	}

	return jsonFile.Close()
}
