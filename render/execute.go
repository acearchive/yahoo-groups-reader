package render

import (
	"embed"
	"github.com/acearchive/yg-render/parse"
	"github.com/yosssi/gohtml"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	outputFileMode  = 0o644
	outputDirMode   = 0o755
	assetsDirectory = "assets"
)

//go:embed assets/*
var assets embed.FS

func assetFileNames() []string {
	return []string{
		"thread.css",
	}
}

func writeAssets(path string) error {
	for _, fileName := range assetFileNames() {
		assetFile, err := assets.Open(filepath.Join(assetsDirectory, fileName))
		if err != nil {
			return err
		}

		outputFile, err := os.OpenFile(filepath.Join(path, fileName), os.O_CREATE|os.O_EXCL|os.O_WRONLY, outputFileMode)
		if err != nil {
			return err
		}

		if _, err := io.Copy(outputFile, assetFile); err != nil {
			return err
		}

		if err := assetFile.Close(); err != nil {
			return err
		}

		if err := outputFile.Close(); err != nil {
			return err
		}
	}

	return nil
}

func formatHtml(input string, output io.Writer) error {
	_, err := io.WriteString(output, gohtml.Format(input))
	return err
}

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

		var outputHtml strings.Builder

		if err := Template.Execute(&outputHtml, args); err != nil {
			return err
		}

		if err := formatHtml(outputHtml.String(), file); err != nil {
			return err
		}

		if err := file.Close(); err != nil {
			return err
		}
	}

	if err := writeAssets(path); err != nil {
		return err
	}

	return writeSearchData(thread, path)
}
