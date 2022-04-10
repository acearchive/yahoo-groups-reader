package parse

import (
	"errors"
	"fmt"
	"github.com/acearchive/yg-render/logger"
	"io"
	"net/mail"
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

var ErrMalformedEmail = errors.New("malformed email")

func nameFromAddress(email *mail.Message) string {
	rawAddress := email.Header.Get(MailHeaderFrom)
	if rawAddress == "" {
		logger.Verbose.Printf("%v: missing `%s`", ErrMalformedEmail, MailHeaderFrom)

		return ""
	}

	fromAddress, err := mail.ParseAddress(rawAddress)
	if err != nil || fromAddress == nil {
		return rawAddress
	}

	if fromAddress.Name != "" {
		return fromAddress.Name
	}

	return fromAddress.Address
}

func userFromEmail(email *mail.Message) string {
	if profileName := email.Header.Get(MailHeaderProfile); profileName != "" {
		return profileName
	}

	if aliasName := email.Header.Get(MailHeaderAlias); aliasName != "" {
		return aliasName
	}

	return nameFromAddress(email)
}

func flairFromEmail(email *mail.Message) string {
	if profData := email.Header.Get(MailHeaderProfData); profData != "" {
		return profData
	}

	return nameFromAddress(email)
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
