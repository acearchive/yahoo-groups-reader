package render

import (
	"github.com/Masterminds/sprig/v3"
	"github.com/acearchive/yg-render/parse"
	"os"
)

const outputFileMode = 0o644

func Execute(title, path string, thread parse.MessageThread) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, outputFileMode)
	if err != nil {
		return err
	}

	defer func() {
		if e := file.Close(); err == nil {
			err = e
		}
	}()

	args := TemplateArgs{
		Title:    title,
		Messages: MessageThreadToArgs(thread),
	}

	err = Template.Funcs(sprig.FuncMap()).Execute(file, args)

	return err
}
