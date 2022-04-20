package parse

import (
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"strings"
)

const (
	MimeHeaderContentType             = "Content-Type"
	MimeHeaderContentTransferEncoding = "Content-Transfer-Encoding"
	contentTypePrefixMultipart        = "multipart/"
	contentTypePlainText              = "text/plain"
	contentTypeParamBoundary          = "boundary"
	quotedPrintable                   = "quoted-printable"
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

	if email.Header.Get(MimeHeaderContentTransferEncoding) == quotedPrintable {
		return quotedprintable.NewReader(email.Body), nil
	}

	return email.Body, nil
}
