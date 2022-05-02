package cmd

import (
	"github.com/acearchive/yg-render/logger"
	"github.com/acearchive/yg-render/parse"
	"github.com/acearchive/yg-render/render"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
)

var (
	flagPageSize int
	flagTitle    string
	flagVerbose  bool
	flagNoSearch bool
)

const (
	DefaultPageSize = 25
	OutputPath      = "../output/"
)

func init() {
	rootCmd.Flags().StringVarP(&flagTitle, "title", "t", "Yahoo Group", "The title of the group")
	rootCmd.Flags().IntVar(&flagPageSize, "page-size", DefaultPageSize, "The maximum number of messages per page")
	rootCmd.Flags().BoolVar(&flagNoSearch, "no-search", false, "Disable the search functionality in the generated site")
	rootCmd.Flags().BoolVarP(&flagVerbose, "verbose", "v", false, "Print verbose output.")
}

var rootCmd = &cobra.Command{
	Use:                   "yg-render [options] archive-path",
	Short:                 "Render an exported Yahoo Groups archive as HTML",
	Long:                  "Render an exported Yahoo Groups archive as HTML\n\nThis accepts the path of the directory containing the .eml files.",
	Args:                  cobra.ExactArgs(1),
	Version:               "0.1.0",
	DisableFlagsInUseLine: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !flagVerbose {
			logger.Verbose.SetOutput(ioutil.Discard)
		}

		thread, err := parse.Directory(args[0])
		if err != nil {
			return err
		}

		config := render.OutputConfig{
			Title:         flagTitle,
			PageSize:      flagPageSize,
			IncludeSearch: !flagNoSearch,
		}

		if err := render.Execute(OutputPath, config, thread); err != nil {
			return err
		}

		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}