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

func MultipartMessageBody(email *mail.Message) (io.Reader, error) {
	mediaType, contentTypeParams, err := mime.ParseMediaType(email.Header.Get(MimeHeaderContentType))

	if err != nil && strings.HasPrefix(mediaType, contentTypePrefixMultipart) {
		multipartReader := multipart.NewReader(email.Body, contentTypeParams[contentTypeParamBoundary])
		for {
			part, err := multipartReader.NextPart()
			if errors.Is(err, io.EOF) {
				break
			} else if err != nil {
				return nil, err
			}

			mediaType, _, err := mime.ParseMediaType(part.Header.Get(MimeHeaderContentType))
			if err != nil {
				return nil, err
			}

			if mediaType == contentTypePlainText {
				return part, nil
			}
		}
	}

	// The email is not a MIME multipart message or no plain text part was
	// found.
	return email.Body, nil
}
