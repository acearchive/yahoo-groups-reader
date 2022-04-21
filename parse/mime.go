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

func decodeCharset(body io.Reader, contentTypeParams map[string]string) io.Reader {
	if charset, hasCharset := contentTypeParams[contentTypeParamCharset]; hasCharset {
		encoding, err := ianaindex.MIME.Encoding(charset)
		if err != nil {
			return body
		}

		return encoding.NewDecoder().Reader(body)
	}

	return body
}

func DecodeMessageBody(email *mail.Message) (io.Reader, error) {
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

			partMediaType, partContentTypeParams, err := mime.ParseMediaType(part.Header.Get(MimeHeaderContentType))
			if err != nil {
				return nil, err
			}

			if partMediaType == contentTypePlainText {
				return decodeCharset(part, partContentTypeParams), nil
			}
		}
	}

	if email.Header.Get(MimeHeaderContentTransferEncoding) == quotedPrintable {
		return decodeCharset(quotedprintable.NewReader(email.Body), contentTypeParams), nil
	}

	return decodeCharset(email.Body, contentTypeParams), nil
}
