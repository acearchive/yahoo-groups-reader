package parse

import (
	"errors"
	"fmt"
	"github.com/acearchive/yg-render/logger"
	"github.com/acearchive/yg-render/parse/flowed"
	"io"
	"net/mail"
	"regexp"
)

type MimeHeader string

const (
	MimeHeaderFrom      = "From"
	MimeHeaderSubject   = "Subject"
	MimeHeaderMessageID = "Message-ID"
	MimeHeaderInReplyTo = "In-Reply-To"
	MimeHeaderProfile   = "X-Yahoo-Profile"
	MimeHeaderAlias     = "X-Yahoo-Alias"
	MimeHeaderProfData  = "X-Yahoo-ProfData"
)

var (
	ErrMalformedEmail = errors.New("malformed email")
	// The "correct" way to do this would be to use `ParseAddress` in
	// `net/mail`, however in the email address in the `From` field, the domain
	// name is sometimes redacted and replaced with '...', which is not a valid
	// email address and will cause the function to return an error. Instead,
	// we need to take a dumber approach that's more tolerant of invalid data.
	addressRegex = regexp.MustCompile(`^"?([\w\s]*\w)"? <[^<>]+>$`)
)

const addressRegexNameIndex = 1

func userFromEmail(email *mail.Message) string {
	if profileName := email.Header.Get(MimeHeaderProfile); profileName != "" {
		return profileName
	}

	if aliasName := email.Header.Get(MimeHeaderAlias); aliasName != "" {
		return aliasName
	}

	rawAddress := email.Header.Get(MimeHeaderFrom)
	if rawAddress == "" {
		logger.Verbose.Printf("%v: missing `%s`", ErrMalformedEmail, MimeHeaderFrom)
		return ""
	}

	matches := addressRegex.FindStringSubmatch(rawAddress)
	if matches != nil {
		return matches[addressRegexNameIndex]
	}

	return rawAddress
}

func flairFromEmail(email *mail.Message) string {
	if profData := email.Header.Get(MimeHeaderProfData); profData != "" {
		return profData
	}

	rawAddress := email.Header.Get(MimeHeaderFrom)
	if rawAddress == "" {
		logger.Verbose.Printf("%v: missing `%s`", ErrMalformedEmail, MimeHeaderFrom)
		return ""
	}

	matches := addressRegex.FindStringSubmatch(rawAddress)
	if matches != nil {
		return matches[addressRegexNameIndex]
	}

	return ""
}

func bodyFromEmail(email *mail.Message) (string, error) {
	rawTextBody, err := MultipartMessageBody(email)
	if err != nil {
		return "", err
	}

	return flowed.ToHtml(rawTextBody)
}

func Email(contents io.Reader) (Message, error) {
	rawMessage, err := mail.ReadMessage(contents)
	if err != nil {
		return Message{}, fmt.Errorf("%w: %v", ErrMalformedEmail, err)
	}

	message := Message{}

	if message.ID = MessageID(rawMessage.Header.Get(MimeHeaderMessageID)); message.ID == "" {
		return Message{}, fmt.Errorf("%w: missing `%s`", ErrMalformedEmail, MimeHeaderMessageID)
	}

	if parentID := MessageID(rawMessage.Header.Get(MimeHeaderInReplyTo)); parentID != "" {
		message.Parent = &parentID
	}

	message.User = userFromEmail(rawMessage)
	message.Flair = flairFromEmail(rawMessage)

	if message.Flair == message.User {
		message.Flair = ""
	}

	if message.Date, err = rawMessage.Header.Date(); err != nil {
		return Message{}, fmt.Errorf("%w: %v", ErrMalformedEmail, err)
	}

	if messageTitle := rawMessage.Header.Get(MimeHeaderSubject); messageTitle != "" {
		message.Title = &messageTitle
	}

	message.Body, err = bodyFromEmail(rawMessage)
	if err != nil {
		return Message{}, fmt.Errorf("%w: %v", ErrMalformedEmail, err)
	}

	return message, nil
}
