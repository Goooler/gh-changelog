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
	Stdout        bool

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
		Use:     "extract [<version>]",
		Aliases: []string{"view", "get", "parse"},
		Short:   "Extract release notes from a changelog file",
		Long: `Extract release notes for a specified version (or the topmost released version by default)
from a changelog file and save to RELEASE_NOTES.md (or stdout with --stdout).

By default, it searches ./CHANGELOG.md and writes to ./RELEASE_NOTES.md.
Use '-' as the changelog file path to read from standard input.`,
		Example: `  # Extract topmost release notes to RELEASE_NOTES.md
  $ gh changelog

  # Extract specific version to RELEASE_NOTES.md
  $ gh changelog v1.2.3

  # Extract to stdout instead of file
  $ gh changelog --stdout
  $ gh changelog v1.2.3 --stdout

  # Extract from a custom changelog file to a custom output file
  $ gh changelog v1.2.3 --file docs/CHANGELOG.md --output notes.md

  # Read from stdin and print to stdout
  $ cat CHANGELOG.md | gh changelog -f - --stdout`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.Version = args[0]
			}
			return RunExtract(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.ChangelogFile, "file", "f", "CHANGELOG.md", "Path to changelog file (use \"-\" to read from standard input)")
	cmd.Flags().StringVarP(&opts.OutputFile, "output", "o", "RELEASE_NOTES.md", "Path to output file (use \"-\" to print to standard output)")
	cmd.Flags().BoolVar(&opts.Stdout, "stdout", false, "Print extracted release notes to standard output")
	cmd.Flags().BoolVarP(&opts.IncludeHeader, "header", "H", false, "Include version heading in output")

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

	if opts.Stdout || opts.OutputFile == "-" || opts.OutputFile == "" {
		fmt.Fprintln(opts.Out, notes)
		return nil
	}

	if err := os.WriteFile(opts.OutputFile, []byte(notes+"\n"), 0644); err != nil {
		return fmt.Errorf("failed to write output file %s: %w", opts.OutputFile, err)
	}

	return nil
}
