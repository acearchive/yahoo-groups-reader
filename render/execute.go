package render

import (
	"github.com/acearchive/yg-render/parse"
	"os"
	"path/filepath"
	"strconv"
)

const (
	outputFileMode = 0o644
	outputDirMode  = 0o755
)

func Execute(path string, config OutputConfig, thread parse.MessageThread) error {
	if err := os.Mkdir(path, outputDirMode); err != nil {
		return err
	}

	pages := BuildArgs(thread, config)

	for i, args := range pages {
		var outputFilePath string

		if i == 0 {
			outputFilePath = filepath.Join(path, "index.html")
		} else {
			if err := os.Mkdir(filepath.Join(path, strconv.Itoa(i+1)), outputDirMode); err != nil {
				return err
			}
			outputFilePath = filepath.Join(path, strconv.Itoa(i+1), "index.html")
		}

		file, err := os.OpenFile(outputFilePath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, outputFileMode)
		if err != nil {
			return err
		}

		if err := Template.Execute(file, args); err != nil {
			return err
		}

		if err := file.Close(); err != nil {
			return err
		}
	}

	return nil
}
