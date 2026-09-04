package extract

import (
	"fmt"
	"io"
	"os"

	"github.com/Goooler/gh-changelog/pkg/changelog"
	"github.com/spf13/cobra"
)

// Options holds options for the extract command.
type Options struct {
	Version       string
	ChangelogFile string
	IncludeHeader bool
	OutputFile    string

	In  io.Reader
	Out io.Writer
	Err io.Writer
}

// NewCmdExtract creates the extract subcommand.
func NewCmdExtract() *cobra.Command {
	opts := &Options{
		In:  os.Stdin,
		Out: os.Stdout,
		Err: os.Stderr,
	}

	cmd := &cobra.Command{
		Use:     "extract <version>",
		Aliases: []string{"view", "get", "parse"},
		Short:   "Extract release notes for a version from a changelog file",
		Long: `Extract and print the release notes for a specified version from a changelog file.

By default, it searches ./CHANGELOG.md and prints the extracted markdown to stdout.
Use '-' as the file path to read the changelog from standard input.`,
		Example: `  # Extract notes for v1.2.3 from CHANGELOG.md
  $ gh changelog extract v1.2.3

  # Extract from a custom changelog file
  $ gh changelog extract 1.2.3 --file docs/CHANGELOG.md

  # Extract unreleased changes
  $ gh changelog extract unreleased

  # Read from stdin and save to a file
  $ cat CHANGELOG.md | gh changelog extract v1.2.3 -f - -o release_notes.md`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Version = args[0]
			return RunExtract(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.ChangelogFile, "file", "f", "CHANGELOG.md", "Path to changelog file (use \"-\" to read from standard input)")
	cmd.Flags().BoolVarP(&opts.IncludeHeader, "header", "H", false, "Include version heading in output")
	cmd.Flags().StringVarP(&opts.OutputFile, "output", "o", "", "Write extracted release notes to file instead of stdout")

	return cmd
}

// RunExtract executes the extraction logic.
func RunExtract(opts *Options) error {
	var source []byte
	var err error

	if opts.ChangelogFile == "-" {
		source, err = io.ReadAll(opts.In)
		if err != nil {
			return fmt.Errorf("failed to read changelog from stdin: %w", err)
		}
	} else {
		source, err = os.ReadFile(opts.ChangelogFile)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("changelog file not found: %s", opts.ChangelogFile)
			}
			return fmt.Errorf("failed to read changelog file %s: %w", opts.ChangelogFile, err)
		}
	}

	extractOpts := changelog.ExtractOptions{
		IncludeHeader:       opts.IncludeHeader,
		TrimLinkDefinitions: true,
	}

	notes, err := changelog.ExtractWithOptions(source, opts.Version, extractOpts)
	if err != nil {
		return err
	}

	if opts.OutputFile != "" {
		if err := os.WriteFile(opts.OutputFile, []byte(notes+"\n"), 0644); err != nil {
			return fmt.Errorf("failed to write output file %s: %w", opts.OutputFile, err)
		}
		return nil
	}

	fmt.Fprintln(opts.Out, notes)
	return nil
}
