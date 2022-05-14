package render

import (
	"fmt"
	"github.com/acearchive/yg-render/body"
	"github.com/acearchive/yg-render/parse"
	"golang.org/x/text/language"
	textmessage "golang.org/x/text/message"
	"html/template"
	"strings"
	"time"
)

const (
	pagesToDisplayInNavigation = 9
	pagesToDisplayOnEitherSide = pagesToDisplayInNavigation / 2
	firstPageNumber            = 1
	repoUrl                    = "https://github.com/acearchive/yg-render"
)

type ParentArgs struct {
	Index             int
	PagePath          PagePath
	User              string
	Body              template.HTML
	Timestamp         string
	FormattedDatetime string
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
	Number    int
	IsCurrent bool
}

type PaginationArgs struct {
	Pages            []PageRef
	PageNumber       int
	Current          PagePath
	CurrentCanonical PagePath
	Next             *PagePath
	Prev             *PagePath
	First            PagePath
	Last             PagePath
}

type ExternalLinkArgs struct {
	IconName string
	Label    string
	Url      string
}

type TemplateArgs struct {
	Title         string
	BaseUrl       string
	IncludeSearch bool
	Links         []ExternalLinkArgs
	Messages      []MessageArgs
	Pagination    PaginationArgs
}

func formatTimestamp(input time.Time) string {
	return input.Format(time.RFC3339)
}

func formatDatetime(input time.Time) string {
	return input.Format("2 Jan 2006, 15:04 -07:00")
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
					Index:             parentIndex + 1,
					User:              parent.User,
					Body:              template.HTML(strings.TrimSpace(body.IndentMultilineString(parent.Body.Html, messageParentBodyIndent))),
					Timestamp:         formatTimestamp(parent.Date),
					FormattedDatetime: formatDatetime(parent.Date),
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
			Body:              template.HTML(strings.TrimSpace(body.IndentMultilineString(message.Body.Html, messageBodyIndent))),
		}
	}

	return argsList
}

type OutputConfig struct {
	PageSize      int
	Title         string
	IncludeSearch bool
	BaseUrl       string
	AddRepoLink   bool
}

func pagePath(pageNumber int) PagePath {
	if pageNumber == firstPageNumber {
		return "/"
	} else {
		return PagePath(fmt.Sprintf("/%d/", pageNumber))
	}
}

func canonicalPagePath(baseUrl string, pageNumber int) PagePath {
	return PagePath(strings.TrimSuffix(baseUrl, "/")) + pagePath(pageNumber)
}

func navPagesRange(pageNumber, totalPages int) (first, last int) {
	switch {
	case pageNumber < firstPageNumber+pagesToDisplayOnEitherSide:
		first = firstPageNumber
		last = firstPageNumber + pagesToDisplayInNavigation - 1
	case pageNumber > totalPages-pagesToDisplayOnEitherSide:
		first = totalPages - pagesToDisplayInNavigation + 1
		last = totalPages
	default:
		first = pageNumber - pagesToDisplayOnEitherSide
		last = pageNumber + pagesToDisplayOnEitherSide
	}

	if last > totalPages {
		last = totalPages
	}

	return first, last
}

func calculateTotalPages(messages, pageSize int) int {
	if messages%pageSize > 0 {
		return (messages / pageSize) + 1
	}

	return messages / pageSize
}

func BuildArgs(thread parse.MessageThread, config OutputConfig) []TemplateArgs {
	messages := messageThreadToArgs(thread)

	totalPages := calculateTotalPages(len(messages), config.PageSize)

	var args []TemplateArgs

	for pageNumber := firstPageNumber; pageNumber <= totalPages; pageNumber++ {
		firstPageInNav, lastPageInNav := navPagesRange(pageNumber, totalPages)

		var pageRefs []PageRef

		for pageInNavNumber := firstPageInNav; pageInNavNumber <= lastPageInNav; pageInNavNumber++ {
			pageRefs = append(pageRefs, PageRef{
				Path:      pagePath(pageInNavNumber),
				Number:    pageInNavNumber,
				IsCurrent: pageInNavNumber == pageNumber,
			})
		}

		paginationArgs := PaginationArgs{
			Pages:            pageRefs,
			PageNumber:       pageNumber,
			Current:          pagePath(pageNumber),
			CurrentCanonical: canonicalPagePath(config.BaseUrl, pageNumber),
			First:            pagePath(firstPageNumber),
			Last:             pagePath(totalPages),
		}

		if pageNumber > firstPageNumber {
			prevPath := pagePath(pageNumber - 1)
			paginationArgs.Prev = &prevPath
		}

		if pageNumber < totalPages {
			nextPath := pagePath(pageNumber + 1)
			paginationArgs.Next = &nextPath
		}

		messageStartIndex := (pageNumber - 1) * config.PageSize
		messageEndIndex := messageStartIndex + config.PageSize
		if messageEndIndex > len(messages) {
			messageEndIndex = len(messages)
		}

		messagesInPage := messages[messageStartIndex:messageEndIndex]

		for _, message := range messagesInPage {
			if message.Parent != nil {
				parentPageNumber := (message.Parent.Index / config.PageSize) + 1
				message.Parent.PagePath = pagePath(parentPageNumber)
			}
		}

		var linkArgs []ExternalLinkArgs

		if config.AddRepoLink {
			linkArgs = append(linkArgs, ExternalLinkArgs{
				IconName: "github",
				Label:    "Source Code",
				Url:      repoUrl,
			})
		}

		args = append(args, TemplateArgs{
			Title:         config.Title,
			BaseUrl:       config.BaseUrl,
			IncludeSearch: config.IncludeSearch,
			Links:         linkArgs,
			Messages:      messagesInPage,
			Pagination:    paginationArgs,
		})
	}

	return args
}
