package release

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/Goooler/gh-changelog/pkg/changelog"
	"github.com/cli/safeexec"
	"github.com/spf13/cobra"
)

// GhExecutor executes a gh CLI command.
type GhExecutor func(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error

func defaultGhExecutor(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	ghPath, err := safeexec.LookPath("gh")
	if err != nil {
		return fmt.Errorf("could not find 'gh' executable in PATH: %w", err)
	}

	cmd := exec.Command(ghPath, args...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}

// Options holds options for the release command.
type Options struct {
	Version            string
	ChangelogFile      string
	Title              string
	Draft              bool
	Prerelease         bool
	Latest             string
	Target             string
	DiscussionCategory string
	VerifyTag          bool
	DryRun             bool
	Assets             []string

	Executor GhExecutor
	In       io.Reader
	Out      io.Writer
	Err      io.Writer
}

// NewCmdRelease creates the release subcommand.
func NewCmdRelease() *cobra.Command {
	opts := &Options{
		In:       os.Stdin,
		Out:      os.Stdout,
		Err:      os.Stderr,
		Executor: defaultGhExecutor,
	}

	cmd := &cobra.Command{
		Use:   "release <version> [flags] [<assets>...]",
		Short: "Create a GitHub Release using release notes extracted from a changelog",
		Long: `Extract release notes for a specified version from a changelog and create
a GitHub Release via the GitHub CLI.

All release notes are extracted from the changelog file and piped directly to
'gh release create <version> --notes-file -'. Any additional flags (such as asset files)
are passed through.`,
		Example: `  # Create a release for v1.2.3 from CHANGELOG.md
  $ gh changelog release v1.2.3

  # Create a draft release with a custom title
  $ gh changelog release v1.2.3 --title "v1.2.3 - Major Update" --draft

  # Preview release notes without publishing (dry run)
  $ gh changelog release v1.2.3 --dry-run

  # Create release from a custom changelog file and attach release assets
  $ gh changelog release v1.2.3 -f docs/CHANGELOG.md ./dist/*.zip`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Version = args[0]
			if len(args) > 1 {
				opts.Assets = args[1:]
			}
			return RunRelease(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.ChangelogFile, "file", "f", "CHANGELOG.md", "Path to changelog file (use \"-\" to read from standard input)")
	cmd.Flags().StringVarP(&opts.Title, "title", "t", "", "Release title (defaults to version tag)")
	cmd.Flags().BoolVarP(&opts.Draft, "draft", "d", false, "Save the release as a draft")
	cmd.Flags().BoolVarP(&opts.Prerelease, "prerelease", "p", false, "Mark the release as a prerelease")
	cmd.Flags().StringVar(&opts.Latest, "latest", "", "Mark the release as latest (true/false/auto)")
	cmd.Flags().StringVar(&opts.Target, "target", "", "Target branch or full commit SHA")
	cmd.Flags().StringVar(&opts.DiscussionCategory, "discussion-category", "", "Start a discussion in the specified category")
	cmd.Flags().BoolVar(&opts.VerifyTag, "verify-tag", false, "Abort release creation if tag does not exist")
	cmd.Flags().BoolVarP(&opts.DryRun, "dry-run", "n", false, "Print extracted notes and planned gh command without creating release")

	return cmd
}

// RunRelease executes the release creation logic.
func RunRelease(opts *Options) error {
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

	notes, err := changelog.Extract(source, opts.Version)
	if err != nil {
		return err
	}

	// Construct gh release create arguments
	ghArgs := []string{"release", "create", opts.Version, "--notes-file", "-"}

	if opts.Title != "" {
		ghArgs = append(ghArgs, "--title", opts.Title)
	}
	if opts.Draft {
		ghArgs = append(ghArgs, "--draft")
	}
	if opts.Prerelease {
		ghArgs = append(ghArgs, "--prerelease")
	}
	if opts.Latest != "" {
		ghArgs = append(ghArgs, "--latest="+opts.Latest)
	}
	if opts.Target != "" {
		ghArgs = append(ghArgs, "--target", opts.Target)
	}
	if opts.DiscussionCategory != "" {
		ghArgs = append(ghArgs, "--discussion-category", opts.DiscussionCategory)
	}
	if opts.VerifyTag {
		ghArgs = append(ghArgs, "--verify-tag")
	}
	if len(opts.Assets) > 0 {
		ghArgs = append(ghArgs, opts.Assets...)
	}

	if opts.DryRun {
		fmt.Fprintf(opts.Out, "==> Planned command:\ngh %s\n\n", strings.Join(ghArgs, " "))
		fmt.Fprintf(opts.Out, "==> Extracted Release Notes for %s:\n%s\n", opts.Version, notes)
		return nil
	}

	executor := opts.Executor
	if executor == nil {
		executor = defaultGhExecutor
	}

	notesReader := bytes.NewReader([]byte(notes))
	return executor(ghArgs, notesReader, opts.Out, opts.Err)
}
