package render

import (
	"errors"
	"fmt"
	"github.com/acearchive/yg-render/parse"
	"golang.org/x/text/language"
	textmessage "golang.org/x/text/message"
	"time"
)

var ErrInvalidParentRef = errors.New("invalid parent ID")

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
	return input.Format("_2 January 2006 15:04 MST")
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

func MessageThreadToArgs(thread parse.MessageThread) ([]MessageArgs, error) {
	argsList := make([]MessageArgs, len(thread))

	messagesByDate, messageIndices := thread.SortedByDate()

	for messageIndex, message := range messagesByDate {
		messageTitle := ""

		if message.Title != nil {
			messageTitle = *message.Title
		}

		var parentArgs *ParentArgs

		if message.Parent != nil {
			parentIndex, parentExists := messageIndices[*message.Parent]
			if !parentExists {
				return nil, fmt.Errorf("%w: '%s'", ErrInvalidParentRef, *message.Parent)
			}

			parent, parentExists := thread[*message.Parent]
			if !parentExists {
				return nil, fmt.Errorf("%w: '%s'", ErrInvalidParentRef, *message.Parent)
			}

			parentArgs = &ParentArgs{
				ID:            formatID(parentIndex),
				User:          parent.User,
				Body:          parent.Body,
				FormattedDate: formatDate(parent.Date),
				FormattedTime: formatTime(parent.Date),
			}
		}

		argsList[messageIndex] = MessageArgs{
			ID:                formatID(messageIndex),
			Timestamp:         formatTimestamp(message.Date),
			FormattedDatetime: formatDatetime(message.Date),
			Index:             formatHumanReadableNumber(messageIndex),
			TotalCount:        formatHumanReadableNumber(len(messagesByDate)),
			Parent:            parentArgs,
			User:              message.User,
			Flair:             message.Flair,
			Title:             messageTitle,
			Body:              message.Body,
		}
	}

	return argsList, nil
}
