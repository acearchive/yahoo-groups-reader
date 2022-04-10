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

func userFromEmail(email *mail.Message) (string, error) {
	if profileName := email.Header.Get(MailHeaderProfile); profileName != "" {
		return profileName, nil
	}

	if aliasName := email.Header.Get(MailHeaderAlias); aliasName != "" {
		return aliasName, nil
	}

	return "", fmt.Errorf("%w: missing `%s` or `%s`", ErrMalformedEmail, MailHeaderProfData, MailHeaderAlias)
}

func flairFromEmail(email *mail.Message) (string, error) {
	if profData := email.Header.Get(MailHeaderProfData); profData != "" {
		return profData, nil
	}

	fromAddress, err := mail.ParseAddress(email.Header.Get(MailHeaderFrom))
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrMalformedEmail, err)
	} else if fromAddress == nil {
		return "", fmt.Errorf("%w: missing `%s`", ErrMalformedEmail, MailHeaderFrom)
	}

	if fromAddress.Name != "" {
		return fromAddress.Name, nil
	} else {
		return fromAddress.Address, nil
	}
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

	message.User, err = userFromEmail(rawMessage)
	if err != nil {
		return Message{}, err
	}

	message.Flair, err = flairFromEmail(rawMessage)
	if err != nil {
		logger.Verbose.Println(err)
		message.Flair = ""
	}

	if message.Date, err = rawMessage.Header.Date(); err != nil {
		return Message{}, fmt.Errorf("%w: %v", ErrMalformedEmail, err)
	}

	if message.Title = rawMessage.Header.Get(MailHeaderSubject); message.Title == "" {
		return Message{}, fmt.Errorf("%w: missing `%s`", ErrMalformedEmail, MailHeaderSubject)
	}

	bodyBuffer := new(strings.Builder)

	if _, err := io.Copy(bodyBuffer, rawMessage.Body); err != nil {
		return Message{}, fmt.Errorf("%w: %v", ErrMalformedEmail, err)
	}

	message.Body = bodyBuffer.String()

	return message, nil
}
