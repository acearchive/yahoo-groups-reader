package cmd

import (
	"errors"
	"fmt"
	"github.com/acearchive/yahoo-groups-reader/logger"
	"github.com/acearchive/yahoo-groups-reader/parse"
	"github.com/acearchive/yahoo-groups-reader/render"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"strings"
)

var ErrInvalidLinkInput = errors.New("malformed --link input")

var (
	flagPageSize    int
	flagTitle       string
	flagVerbose     bool
	flagNoSearch    bool
	flagOutput      string
	flagBase        string
	flagNoRepo      bool
	flagLinks       []string
	flagLocale      string
	flagDescription string
)

const (
	DefaultPageSize   = 25
	DefaultOutputPath = "../output"
	DefaultBasePath   = "/"
	DefaultGroupName  = "Yahoo Group"
)

func init() {
	rootCmd.Flags().StringVarP(&flagTitle, "title", "t", DefaultGroupName, "The title of the group")
	rootCmd.Flags().IntVar(&flagPageSize, "page-size", DefaultPageSize, "The maximum number of messages per page")
	rootCmd.Flags().BoolVar(&flagNoSearch, "no-search", false, "Disable the search functionality in the generated site")
	rootCmd.Flags().BoolVar(&flagNoRepo, "no-repo", false, "Don't add a link to the GitHub repo in the generated site")
	rootCmd.Flags().StringArrayVar(&flagLinks, "link", nil, "Add a link to the top of the page in the generated site")
	rootCmd.Flags().StringVar(&flagLocale, "locale", "en_US", "The locale of the generated site")
	rootCmd.Flags().StringVar(&flagDescription, "description", "", "Override the default site description for search results and social previews")
	rootCmd.Flags().StringVarP(&flagOutput, "output", "o", DefaultOutputPath, "The directory to write the generated HTML to")
	rootCmd.Flags().StringVarP(&flagBase, "base", "b", DefaultBasePath, "The base URL for the generated site")
	rootCmd.Flags().BoolVarP(&flagVerbose, "verbose", "v", false, "Print verbose output.")
}

func parseLinkInputs(inputs []string) ([]render.ExternalLinkConfig, error) {
	configs := make([]render.ExternalLinkConfig, len(inputs))

	for i, input := range inputs {
		components := strings.Split(input, ",")
		if len(components) != 3 {
			return nil, fmt.Errorf("%s: %s", ErrInvalidLinkInput, input)
		}

		configs[i] = render.ExternalLinkConfig{
			IconName: components[0],
			Label:    components[1],
			Url:      components[2],
		}
	}

	return configs, nil
}

var rootCmd = &cobra.Command{
	Use:                   "yahoo-groups-reader [options] archive-path",
	Short:                 "Render an exported Yahoo Groups archive as HTML",
	Long:                  "Render an exported Yahoo Groups archive as HTML.\n\nThis accepts the path of the directory containing the `.eml` files.\n\nYou can add external links at the top of the page with --link. It accepts the\nname of a Feather icon, a label, and a target URL in the form `icon,label,url`.\nYou can add more than one by passing --link multiple times.\n\nExample: `mail,Contact Us,https://example.com/contact`",
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

		linkConfigs, err := parseLinkInputs(flagLinks)
		if err != nil {
			return err
		}

		config := render.OutputConfig{
			Title:             flagTitle,
			CustomDescription: flagDescription,
			PageSize:          flagPageSize,
			IncludeSearch:     !flagNoSearch,
			BaseUrl:           flagBase,
			AddRepoLink:       !flagNoRepo,
			Links:             linkConfigs,
			Locale:            flagLocale,
		}

		if err := render.Execute(flagOutput, config, thread); err != nil {
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
