package render

import (
	"encoding/json"
	"github.com/acearchive/yahoo-groups-reader/body"
	"github.com/acearchive/yahoo-groups-reader/parse"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const searchFileName = "search.json"

// The number of runes to truncate the message summaries in the search index
// to. This should roughly correspond to how much of the message we can
// feasibly display in the UI, but with some buffer room so that the actual
// truncation happens in CSS.
const messageSummaryTruncateLen = 400

type MessageSearchFields struct {
	Index      int    `json:"id"`
	PageNumber int    `json:"page"`
	Timestamp  string `json:"timestamp"`
	User       string `json:"user"`
	Flair      string `json:"flair"`
	Year       string `json:"year"`
	Title      string `json:"title"`
	Body       string `json:"body"`
	Summary    string `json:"summary"`
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

func truncateString(text string, length int) string {
	if length >= len(text) {
		return text
	} else {
		return string([]rune(text)[:length])
	}
}

// The reason for truncating the message summary at build time is to reduce the
// size of the search index we have to send to clients, which can be quite
// large for large data sets. We truncate it slightly larger than necessary so
// that the final truncation happens in CSS and we don't have to worry about
// splitting on word boundaries.
func truncateMessageSummary(text string) string {
	return truncateString(text, messageSummaryTruncateLen)
}

func buildSearchFields(thread parse.MessageThread, config OutputConfig) []MessageSearchFields {
	fields := make([]MessageSearchFields, 0, len(thread))

	sortedMessages, _ := thread.SortedByDate()

	for i, message := range sortedMessages {
		messageBody := tokensToSearchText(message.Body.Tokens)

		field := MessageSearchFields{
			Index:      i + 1,
			PageNumber: pageNumberOfMessage(i+1, config.PageSize),
			Timestamp:  message.Date.Format(time.RFC3339),
			User:       message.User,
			Flair:      message.Flair,
			Year:       strconv.Itoa(message.Date.Year()),
			Body:       messageBody,
			Summary:    truncateMessageSummary(messageBody),
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
