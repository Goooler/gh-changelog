package cmd

import (
	"os"

	"github.com/Goooler/gh-changelog/pkg/cmd/extract"
	"github.com/spf13/cobra"
)

// NewRootCmd creates the root cobra command for gh-changelog.
func NewRootCmd(version string) *cobra.Command {
	opts := &extract.Options{
		In:  os.Stdin,
		Out: os.Stdout,
		Err: os.Stderr,
	}

	rootCmd := &cobra.Command{
		Use:     "gh-changelog [<version>]",
		Short:   "Extract release notes from changelog files",
		Long: `gh-changelog is a GitHub CLI extension that parses release notes from changelog files
(supporting Keep a Changelog and other markdown formats) and outputs them to RELEASE_NOTES.md.`,
		Version: version,
		Example: `  # Extract topmost release notes to RELEASE_NOTES.md
  $ gh changelog extract

  # Extract specific version to RELEASE_NOTES.md
  $ gh changelog extract v1.2.3

  # Extract to stdout instead of file
  $ gh changelog extract --stdout
  $ gh changelog extract v1.2.3 --stdout

  # Extract from a custom changelog file to a custom output file
  $ gh changelog extract v1.2.3 --file docs/CHANGELOG.md --output notes.md`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.Version = args[0]
			}
			return extract.RunExtract(opts)
		},
	}

	rootCmd.Flags().StringVarP(&opts.ChangelogFile, "file", "f", "CHANGELOG.md", "Path to changelog file (use \"-\" to read from standard input)")
	rootCmd.Flags().StringVarP(&opts.OutputFile, "output", "o", "RELEASE_NOTES.md", "Path to output file (use \"-\" to print to standard output)")
	rootCmd.Flags().BoolVar(&opts.Stdout, "stdout", false, "Print extracted release notes to standard output")
	rootCmd.Flags().BoolVarP(&opts.IncludeHeader, "header", "H", false, "Include version heading in output")

	// Also add extract as explicit subcommand for users who prefer `gh changelog extract`
	rootCmd.AddCommand(extract.NewCmdExtract())

	return rootCmd
}
