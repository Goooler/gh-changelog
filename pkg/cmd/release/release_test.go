package release

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReleaseCommand(t *testing.T) {
	tempDir := t.TempDir()
	changelogPath := filepath.Join(tempDir, "CHANGELOG.md")
	content := `# Changelog

## [1.2.3] - 2024-05-01
### Added
- Feature A

## [1.2.2]
- Old
`
	if err := os.WriteFile(changelogPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test changelog: %v", err)
	}

	t.Run("dry run outputs planned command and notes", func(t *testing.T) {
		var stdout bytes.Buffer
		opts := &Options{
			Version:       "v1.2.3",
			ChangelogFile: changelogPath,
			Title:         "Custom Release Title",
			Draft:         true,
			Prerelease:    true,
			DryRun:        true,
			Assets:        []string{"dist/app.zip"},
			Out:           &stdout,
		}

		err := RunRelease(opts)
		if err != nil {
			t.Fatalf("RunRelease() error = %v", err)
		}

		out := stdout.String()
		if !strings.Contains(out, "gh release create v1.2.3 --notes-file - --title Custom Release Title --draft --prerelease dist/app.zip") {
			t.Errorf("expected planned command in dry-run output, got:\n%s", out)
		}
		if !strings.Contains(out, "### Added\n- Feature A") {
			t.Errorf("expected extracted notes in dry-run output, got:\n%s", out)
		}
	})

	t.Run("executes gh command with extracted notes in stdin", func(t *testing.T) {
		var capturedArgs []string
		var capturedStdin string

		mockExecutor := func(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
			capturedArgs = args
			buf, _ := io.ReadAll(stdin)
			capturedStdin = string(buf)
			return nil
		}

		opts := &Options{
			Version:       "v1.2.3",
			ChangelogFile: changelogPath,
			Target:        "main",
			Executor:      mockExecutor,
		}

		err := RunRelease(opts)
		if err != nil {
			t.Fatalf("RunRelease() error = %v", err)
		}

		wantArgs := []string{"release", "create", "v1.2.3", "--notes-file", "-", "--target", "main"}
		if strings.Join(capturedArgs, " ") != strings.Join(wantArgs, " ") {
			t.Errorf("capturedArgs = %v, want %v", capturedArgs, wantArgs)
		}

		wantStdin := "### Added\n- Feature A"
		if strings.TrimSpace(capturedStdin) != wantStdin {
			t.Errorf("capturedStdin = %q, want %q", capturedStdin, wantStdin)
		}
	})
}
