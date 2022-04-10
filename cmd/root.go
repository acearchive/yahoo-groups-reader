package cmd

import (
	"github.com/spf13/cobra"
	"os"
)

func init() {
	rootCmd.Flags().StringP("output", "o", ".", "The directory to write the rendered output to")
	rootCmd.Flags().StringP("title", "t", "Yahoo Group", "The title of the group")
}

var rootCmd = &cobra.Command{
	Use:                   "yg-render [options] archive-path",
	Short:                 "Render an exported Yahoo Groups archive as HTML",
	Long:                  "Render an exported Yahoo Groups archive as HTML\n\nThis accepts the path of the directory containing the .eml files.",
	Args:                  cobra.ExactArgs(1),
	Version:               "0.1.0",
	DisableFlagsInUseLine: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
