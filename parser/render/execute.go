package render

import (
	"embed"
	"github.com/acearchive/yg-render/parse"
	fileminify "github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	"github.com/yosssi/gohtml"
	"io"
	"mime"
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

var minifier = fileminify.New()

func init() {
	minifier.AddFunc("text/html", html.Minify)
	minifier.AddFunc("text/css", css.Minify)
	minifier.AddFunc("text/javascript", js.Minify)
}

func assetFileNames() []string {
	return []string{
		"search.js",
		"time.js",
		"variables.css",
		"global.css",
		"components.css",
		"thread.css",
		"search.css",
	}
}

func writeAssets(path string, minify bool) error {
	for _, fileName := range assetFileNames() {
		assetFile, err := assets.Open(filepath.Join(assetsDirectory, fileName))
		if err != nil {
			return err
		}

		outputFile, err := os.OpenFile(filepath.Join(path, fileName), os.O_CREATE|os.O_EXCL|os.O_WRONLY, outputFileMode)
		if err != nil {
			return err
		}

		if minify {
			mimeType := mime.TypeByExtension(filepath.Ext(fileName))
			if err := minifier.Minify(mimeType, outputFile, assetFile); err != nil {
				return err
			}
		} else {
			if _, err := io.Copy(outputFile, assetFile); err != nil {
				return err
			}
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

func formatHtml(input string, output io.Writer, minify bool) error {
	if minify {
		return minifier.Minify("text/html", output, strings.NewReader(input))
	} else {
		_, err := io.WriteString(output, gohtml.Format(input))
		return err
	}
}

func Execute(path string, config OutputConfig, thread parse.MessageThread) error {
	if err := os.Mkdir(path, outputDirMode); err != nil {
		return err
	}

	pages := BuildArgs(thread, config)

	for pageIndex, args := range pages {
		var outputFilePath string

		if pageIndex == 0 {
			outputFilePath = filepath.Join(path, "index.html")
		} else {
			if err := os.Mkdir(filepath.Join(path, strconv.Itoa(pageIndex+1)), outputDirMode); err != nil {
				return err
			}
			outputFilePath = filepath.Join(path, strconv.Itoa(pageIndex+1), "index.html")
		}

		file, err := os.OpenFile(outputFilePath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, outputFileMode)
		if err != nil {
			return err
		}

		var outputHtml strings.Builder

		if err := Template.Execute(&outputHtml, args); err != nil {
			return err
		}

		if err := formatHtml(outputHtml.String(), file, config.Minify); err != nil {
			return err
		}

		if err := file.Close(); err != nil {
			return err
		}
	}

	if err := writeAssets(path, config.Minify); err != nil {
		return err
	}

	if config.IncludeSearch {
		if err := writeSearchData(thread, config, path); err != nil {
			return err
		}
	}

	return nil
}
