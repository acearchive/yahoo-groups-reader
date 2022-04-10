package render

import (
	"fmt"
	"github.com/acearchive/yg-render/parse"
	"golang.org/x/text/language"
	textmessage "golang.org/x/text/message"
	"strings"
	"time"
)

const (
	MaxParentBodySize    = 500
	StringTruncateSymbol = "…"
)

type ParentArgs struct {
	ID            string
	User          string
	Body          string
	FormattedDate string
	FormattedTime string
}

type MessageArgs struct {
	ID                string
	Timestamp         string
	FormattedDatetime string
	Index             string
	TotalCount        string
	Parent            *ParentArgs
	User              string
	Flair             string
	Title             string
	Body              string
}

type TemplateArgs struct {
	Title    string
	Messages []MessageArgs
}

func formatTimestamp(input time.Time) string {
	return input.Format(time.RFC3339)
}

func formatDatetime(input time.Time) string {
	return input.Format("2 January 2006 15:04 MST")
}

func formatDate(input time.Time) string {
	return input.Format("_2 January 2006")
}

func formatTime(input time.Time) string {
	return input.Format("15:04 MST")
}

func formatID(index int) string {
	return fmt.Sprintf("message-%d", index)
}

func formatHumanReadableNumber(number int) string {
	localizedPrinter := textmessage.NewPrinter(language.English)
	return localizedPrinter.Sprintf("%d", number)
}

func formatBody(body string) string {
	trimmedBody := []rune(strings.TrimSpace(body))
	if len(trimmedBody) > MaxParentBodySize {
		return string(trimmedBody[:MaxParentBodySize]) + StringTruncateSymbol
	}

	return string(trimmedBody)
}

func MessageThreadToArgs(thread parse.MessageThread) []MessageArgs {
	argsList := make([]MessageArgs, len(thread))

	messagesByDate, messageIndices := thread.SortedByDate()

	for messageIndex, message := range messagesByDate {
		messageTitle := ""

		if message.Title != nil {
			messageTitle = *message.Title
		}

		var parentArgs *ParentArgs

		if message.Parent != nil {
			parentIndex, parentIndexExists := messageIndices[*message.Parent]
			parent, parentExists := thread[*message.Parent]

			if parentIndexExists && parentExists {
				parentArgs = &ParentArgs{
					ID:            formatID(parentIndex + 1),
					User:          parent.User,
					Body:          formatBody(parent.Body),
					FormattedDate: formatDate(parent.Date),
					FormattedTime: formatTime(parent.Date),
				}
			}
		}

		argsList[messageIndex] = MessageArgs{
			ID:                formatID(messageIndex + 1),
			Timestamp:         formatTimestamp(message.Date),
			FormattedDatetime: formatDatetime(message.Date),
			Index:             formatHumanReadableNumber(messageIndex + 1),
			TotalCount:        formatHumanReadableNumber(len(messagesByDate)),
			Parent:            parentArgs,
			User:              message.User,
			Flair:             message.Flair,
			Title:             messageTitle,
			Body:              formatBody(message.Body),
		}
	}

	return argsList
}
