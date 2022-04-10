package parse

import (
	"errors"
	"fmt"
	"github.com/acearchive/yg-render/logger"
	"os"
	"path/filepath"
)

const EmailExtension = ".eml"

func Directory(path string) (MessageThread, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	thread := make(MessageThread, len(entries))

	for _, entry := range entries {
		emailPath := filepath.Join(path, entry.Name())

		if entry.IsDir() {
			logger.Verbose.Printf("file is a directory: %s", emailPath)
			continue
		}

		if filepath.Ext(emailPath) != EmailExtension {
			logger.Verbose.Printf("file is not a `.eml` file: %s", emailPath)
			continue
		}

		file, err := os.Open(emailPath)
		if err != nil {
			return nil, err
		}

		message, err := Email(file)
		if errors.Is(err, ErrMalformedEmail) {
			logger.Verbose.Printf("%v: '%s'", err, emailPath)
		} else if err != nil {
			return nil, fmt.Errorf("%w: '%s'", err, emailPath)
		}

		if err := file.Close(); err != nil {
			return nil, err
		}

		thread[message.ID] = message
	}

	return thread, nil
}
