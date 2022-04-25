package render

import (
	"fmt"
	"github.com/acearchive/yg-render/parse"
	"golang.org/x/text/language"
	textmessage "golang.org/x/text/message"
	"html/template"
	"strconv"
	"time"
)

const (
	pagesToDisplayInNavigation = 9
	pagesToDisplayOnEitherSide = pagesToDisplayInNavigation / 2
	firstPageNumber            = 1
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
	Pages       []PageRef
	IsFirstPage bool
	Next        *PagePath
	Prev        *PagePath
	First       PagePath
	Last        PagePath
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

func messageThreadToArgs(thread parse.MessageThread) []MessageArgs {
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

type OutputConfig struct {
	PageSize int
	Title    string
}

func pagePath(pageNumber, currentPageNumber int) PagePath {
	switch {
	case pageNumber == currentPageNumber:
		return "."
	case currentPageNumber == firstPageNumber:
		return PagePath(fmt.Sprintf("./%d", pageNumber))
	case pageNumber == firstPageNumber:
		return "../"
	default:
		return PagePath(fmt.Sprintf("../%d", pageNumber))
	}
}

func navPagesRange(pageNumber, totalPages int) (first, last int) {
	switch {
	case pageNumber < firstPageNumber+pagesToDisplayOnEitherSide:
		return firstPageNumber, firstPageNumber + pagesToDisplayInNavigation - 1
	case pageNumber > totalPages-pagesToDisplayOnEitherSide:
		return totalPages - pagesToDisplayInNavigation + 1, totalPages
	default:
		return pageNumber - pagesToDisplayOnEitherSide, pageNumber + pagesToDisplayOnEitherSide
	}
}

func BuildArgs(thread parse.MessageThread, config OutputConfig) []TemplateArgs {
	messages := messageThreadToArgs(thread)

	totalPages := len(messages) / config.PageSize
	if len(messages)%config.PageSize > 0 {
		totalPages++
	}

	var args []TemplateArgs

	for pageNumber := firstPageNumber; pageNumber <= totalPages; pageNumber++ {
		firstPageInNav, lastPageInNav := navPagesRange(pageNumber, totalPages)

		var pageRefs []PageRef

		for pageInNavNumber := firstPageInNav; pageInNavNumber <= lastPageInNav; pageInNavNumber++ {
			pageRefs = append(pageRefs, PageRef{
				Path:      pagePath(pageInNavNumber, pageNumber),
				Number:    strconv.Itoa(pageInNavNumber),
				IsCurrent: pageInNavNumber == pageNumber,
			})
		}

		paginationArgs := PaginationArgs{
			Pages:       pageRefs,
			First:       pagePath(firstPageNumber, pageNumber),
			Last:        pagePath(totalPages, pageNumber),
			IsFirstPage: pageNumber == firstPageNumber,
		}

		if pageNumber > firstPageNumber {
			prevPath := pagePath(pageNumber-1, pageNumber)
			paginationArgs.Prev = &prevPath
		}

		if pageNumber < totalPages {
			nextPath := pagePath(pageNumber+1, pageNumber)
			paginationArgs.Next = &nextPath
		}

		messageStartIndex := (pageNumber - 1) * config.PageSize
		messageEndIndex := messageStartIndex + config.PageSize
		if messageEndIndex > len(messages) {
			messageEndIndex = len(messages)
		}

		args = append(args, TemplateArgs{
			Title:      config.Title,
			Messages:   messages[messageStartIndex:messageEndIndex],
			Pagination: paginationArgs,
		})
	}

	return args
}
