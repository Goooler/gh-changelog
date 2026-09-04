package extract

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractCommand(t *testing.T) {
	tempDir := t.TempDir()
	changelogPath := filepath.Join(tempDir, "CHANGELOG.md")
	content := `# Changelog

## [1.2.3] - 2024-05-01
### Added
- Super feature

### Fixed
- Bug fix

## [1.2.2] - 2024-04-01
- Old fix
`
	if err := os.WriteFile(changelogPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test changelog: %v", err)
	}

	t.Run("extract from file to stdout", func(t *testing.T) {
		var stdout bytes.Buffer
		opts := &Options{
			Version:       "v1.2.3",
			ChangelogFile: changelogPath,
			Out:           &stdout,
		}

		err := RunExtract(opts)
		if err != nil {
			t.Fatalf("RunExtract() failed: %v", err)
		}

		want := "### Added\n- Super feature\n\n### Fixed\n- Bug fix"
		if got := strings.TrimSpace(stdout.String()); got != want {
			t.Errorf("RunExtract() = %q, want %q", got, want)
		}
	})

	t.Run("extract from stdin", func(t *testing.T) {
		var stdout bytes.Buffer
		opts := &Options{
			Version:       "1.2.2",
			ChangelogFile: "-",
			In:            strings.NewReader(content),
			Out:           &stdout,
		}

		err := RunExtract(opts)
		if err != nil {
			t.Fatalf("RunExtract() failed: %v", err)
		}

		want := "- Old fix"
		if got := strings.TrimSpace(stdout.String()); got != want {
			t.Errorf("RunExtract() = %q, want %q", got, want)
		}
	})

	t.Run("extract to output file", func(t *testing.T) {
		outputPath := filepath.Join(tempDir, "output.md")
		opts := &Options{
			Version:       "v1.2.3",
			ChangelogFile: changelogPath,
			OutputFile:    outputPath,
		}

		err := RunExtract(opts)
		if err != nil {
			t.Fatalf("RunExtract() failed: %v", err)
		}

		data, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("failed to read output file: %v", err)
		}

		want := "### Added\n- Super feature\n\n### Fixed\n- Bug fix"
		if got := strings.TrimSpace(string(data)); got != want {
			t.Errorf("OutputFile content = %q, want %q", got, want)
		}
	})

	t.Run("missing file returns error", func(t *testing.T) {
		opts := &Options{
			Version:       "v1.2.3",
			ChangelogFile: filepath.Join(tempDir, "non_existent.md"),
		}

		err := RunExtract(opts)
		if err == nil {
			t.Fatal("expected error for non existent changelog, got nil")
		}
	})
}
