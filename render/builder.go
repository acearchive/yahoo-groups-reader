package render

import (
	"github.com/acearchive/yg-render/parse"
	"golang.org/x/text/language"
	textmessage "golang.org/x/text/message"
	"html/template"
	"time"
)

type ParentArgs struct {
	Index         int
	User          string
	Body          template.HTML
	FormattedDate string
	FormattedTime string
}

type MessageArgs struct {
	Index             int
	Number            string
	TotalCount        string
	Timestamp         string
	FormattedDatetime string
	Parent            *ParentArgs
	User              string
	Flair             string
	Title             string
	Body              template.HTML
}

type PagePath string

type PageRef struct {
	Path      PagePath
	Number    string
	IsCurrent bool
}

type PaginationArgs struct {
	Pages []PageRef
	Next  *PagePath
	Prev  *PagePath
	First PagePath
	Last  PagePath
}

type TemplateArgs struct {
	Title      string
	Messages   []MessageArgs
	Pagination PaginationArgs
}

func formatTimestamp(input time.Time) string {
	return input.Format(time.RFC3339)
}

func formatDatetime(input time.Time) string {
	return input.Format("2 January 2006 15:04 MST")
}

func formatDate(input time.Time) string {
	return input.Format("2 January 2006")
}

func formatTime(input time.Time) string {
	return input.Format("15:04 MST")
}

func formatHumanReadableNumber(number int) string {
	localizedPrinter := textmessage.NewPrinter(language.English)
	return localizedPrinter.Sprintf("%d", number)
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
					Index:         parentIndex + 1,
					User:          parent.User,
					Body:          template.HTML(parent.Body),
					FormattedDate: formatDate(parent.Date),
					FormattedTime: formatTime(parent.Date),
				}
			}
		}

		argsList[messageIndex] = MessageArgs{
			Index:             messageIndex + 1,
			Number:            formatHumanReadableNumber(messageIndex + 1),
			TotalCount:        formatHumanReadableNumber(len(messagesByDate)),
			Timestamp:         formatTimestamp(message.Date),
			FormattedDatetime: formatDatetime(message.Date),
			Parent:            parentArgs,
			User:              message.User,
			Flair:             message.Flair,
			Title:             messageTitle,
			Body:              template.HTML(message.Body),
		}
	}

	return argsList
}
