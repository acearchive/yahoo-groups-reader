package parse

import (
	"errors"
	"golang.org/x/text/encoding/ianaindex"
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
	contentTypeParamCharset           = "charset"
	quotedPrintable                   = "quoted-printable"
)

func decodeMultipart(email *mail.Message) (io.Reader, error) {
	mediaType, contentTypeParams, err := mime.ParseMediaType(email.Header.Get(MimeHeaderContentType))

	if err == nil && strings.HasPrefix(mediaType, contentTypePrefixMultipart) {
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

func decodeCharset(body io.Reader, contentType string) (io.Reader, error) {
	_, contentTypeParams, err := mime.ParseMediaType(contentType)

	if charset, hasCharset := contentTypeParams[contentTypeParamCharset]; hasCharset && err == nil {
		encoding, err := ianaindex.MIME.Encoding(charset)
		if err != nil {
			return body, nil
		}

		return encoding.NewDecoder().Reader(body), nil
	}

	return body, nil
}

func DecodeMessageBody(email *mail.Message) (io.Reader, error) {
	multipartDecoded, err := decodeMultipart(email)
	if err != nil {
		return nil, err
	}

	return decodeCharset(multipartDecoded, email.Header.Get(MimeHeaderContentType))
}
