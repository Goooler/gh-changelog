package cmd

import (
	"github.com/Goooler/gh-changelog/pkg/cmd/extract"
	"github.com/Goooler/gh-changelog/pkg/cmd/release"
	"github.com/spf13/cobra"
)

// NewRootCmd creates the root cobra command for gh-changelog.
func NewRootCmd(version string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "gh-changelog",
		Short:   "Extract release notes from changelog files for GitHub Releases",
		Long: `gh-changelog is a GitHub CLI extension that parses release notes from changelog files
(supporting Keep a Changelog and other markdown formats) and seamlessly integrates with GitHub Releases.`,
		Version: version,
		Example: `  # Extract release notes for a version from CHANGELOG.md
  $ gh changelog extract v1.2.3

  # Create a GitHub release with notes extracted from CHANGELOG.md
  $ gh changelog release v1.2.3

  # Preview the release notes without publishing (dry run)
  $ gh changelog release v1.2.3 --dry-run`,
	}

	rootCmd.AddCommand(extract.NewCmdExtract())
	rootCmd.AddCommand(release.NewCmdRelease())

	return rootCmd
}
