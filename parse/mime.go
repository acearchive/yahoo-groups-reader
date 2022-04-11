package parse

import (
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"strings"
)

const (
	MimeHeaderContentType      = "Content-Type"
	contentTypePrefixMultipart = "multipart/"
	contentTypePlainText       = "text/plain"
	contentTypeParamBoundary   = "boundary"
)

func bodyFromEmail(email *mail.Message) (string, error) {
	mediaType, contentTypeParams, err := mime.ParseMediaType(email.Header.Get(MimeHeaderContentType))
	if err != nil {
		return "", err
	}

	var bodyReader io.Reader

	if strings.HasPrefix(mediaType, contentTypePrefixMultipart) {
		multipartReader := multipart.NewReader(email.Body, contentTypeParams[contentTypeParamBoundary])
		for {
			part, err := multipartReader.NextPart()
			if errors.Is(err, io.EOF) {
				break
			} else if err != nil {
				return "", err
			}

			mediaType, _, err := mime.ParseMediaType(part.Header.Get(MimeHeaderContentType))
			if err != nil {
				return "", err
			}

			if mediaType == contentTypePlainText {
				bodyReader = part
				break
			}
		}
	}

	if bodyReader == nil {
		// The email is not a MIME multipart message or no plain text part was
		// found.
		bodyReader = email.Body
	}

	bodyBuffer := new(strings.Builder)

	if _, err := io.Copy(bodyBuffer, bodyReader); err != nil {
		return "", err
	}

	return bodyBuffer.String(), nil
}
