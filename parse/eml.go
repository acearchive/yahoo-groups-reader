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

	if profileName := rawMessage.Header.Get(MailHeaderProfile); profileName != "" {
		message.User = profileName
	} else if aliasName := rawMessage.Header.Get(MailHeaderAlias); aliasName != "" {
		message.User = aliasName
	} else {
		return Message{}, fmt.Errorf("%w: missing `%s` or `%s`", ErrMalformedEmail, MailHeaderProfData, MailHeaderAlias)
	}

	if profData := rawMessage.Header.Get(MailHeaderProfData); profData != "" {
		message.Flair = profData
	} else {
		fromAddress, err := mail.ParseAddress(rawMessage.Header.Get(MailHeaderFrom))
		if err != nil {
			logger.Verbose.Printf("%v: %v", ErrMalformedEmail, err)
		}

		if fromAddress.Name != "" {
			message.Flair = fromAddress.Name
		} else {
			message.Flair = fromAddress.Address
		}
	}

	if message.Date, err = rawMessage.Header.Date(); err != nil {
		return Message{}, fmt.Errorf("%w, %v", ErrMalformedEmail, err)
	}

	if message.Title = rawMessage.Header.Get(MailHeaderSubject); message.Title == "" {
		return Message{}, fmt.Errorf("%w, missing `%s`", ErrMalformedEmail, MailHeaderSubject)
	}

	bodyBuffer := new(strings.Builder)

	if _, err := io.Copy(bodyBuffer, rawMessage.Body); err != nil {
		return Message{}, fmt.Errorf("%w: %v", ErrMalformedEmail, err)
	}

	message.Body = bodyBuffer.String()

	return message, nil
}
