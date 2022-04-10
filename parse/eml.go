package parse

import (
	"errors"
	"fmt"
	"github.com/acearchive/yg-render/logger"
	"io"
	"net/mail"
	"regexp"
	"strings"
)

type MailHeader string

const (
	MailHeaderFrom      = "From"
	MailHeaderSubject   = "Subject"
	MailHeaderMessageID = "Message-ID"
	MailHeaderInReplyTo = "In-Reply-To"
	MailHeaderProfile   = "X-Yahoo-Profile"
	MailHeaderAlias     = "X-Yahoo-Alias"
	MailHeaderProfData  = "X-Yahoo-ProfData"
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
	if profileName := email.Header.Get(MailHeaderProfile); profileName != "" {
		return profileName
	}

	if aliasName := email.Header.Get(MailHeaderAlias); aliasName != "" {
		return aliasName
	}

	rawAddress := email.Header.Get(MailHeaderFrom)
	if rawAddress == "" {
		logger.Verbose.Printf("%v: missing `%s`", ErrMalformedEmail, MailHeaderFrom)
		return ""
	}

	matches := addressRegex.FindStringSubmatch(rawAddress)
	if matches != nil {
		return matches[addressRegexNameIndex]
	}

	return rawAddress
}

func flairFromEmail(email *mail.Message) string {
	if profData := email.Header.Get(MailHeaderProfData); profData != "" {
		return profData
	}

	rawAddress := email.Header.Get(MailHeaderFrom)
	if rawAddress == "" {
		logger.Verbose.Printf("%v: missing `%s`", ErrMalformedEmail, MailHeaderFrom)
		return ""
	}

	matches := addressRegex.FindStringSubmatch(rawAddress)
	if matches != nil {
		return matches[addressRegexNameIndex]
	}

	return ""
}

func Email(contents io.Reader) (Message, error) {
	rawMessage, err := mail.ReadMessage(contents)
	if err != nil {
		return Message{}, fmt.Errorf("%w: %v", ErrMalformedEmail, err)
	}

	message := Message{}

	if message.ID = MessageID(rawMessage.Header.Get(MailHeaderMessageID)); message.ID == "" {
		return Message{}, fmt.Errorf("%w: missing `%s`", ErrMalformedEmail, MailHeaderMessageID)
	}

	if parentID := MessageID(rawMessage.Header.Get(MailHeaderInReplyTo)); parentID != "" {
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

	if messageTitle := rawMessage.Header.Get(MailHeaderSubject); messageTitle != "" {
		message.Title = &messageTitle
	}

	bodyBuffer := new(strings.Builder)

	if _, err := io.Copy(bodyBuffer, rawMessage.Body); err != nil {
		return Message{}, fmt.Errorf("%w: %v", ErrMalformedEmail, err)
	}

	message.Body = bodyBuffer.String()

	return message, nil
}
