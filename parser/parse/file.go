package parse

import (
	"errors"
	"fmt"
	"github.com/acearchive/yahoo-groups-reader/logger"
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

		message, parseErr := Email(file)

		if err := file.Close(); err != nil {
			return nil, err
		}

		if errors.Is(parseErr, ErrMalformedEmail) {
			logger.Verbose.Printf("%v: '%s'", parseErr, emailPath)
			continue
		} else if parseErr != nil {
			return nil, fmt.Errorf("%w: '%s'", parseErr, emailPath)
		}

		thread[message.ID] = message
	}

	return thread, nil
}
