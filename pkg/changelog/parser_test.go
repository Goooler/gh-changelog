package changelog

import (
	"errors"
	"testing"
)

func TestExtract(t *testing.T) {
	tests := []struct {
		name          string
		changelog     string
		targetVersion string
		opts          *ExtractOptions
		want          string
		wantErr       error
	}{
		{
			name: "standard keep a changelog with v prefix",
			changelog: `# Changelog
## [Unreleased]
- Ongoing work

## [1.1.1] - 2024-05-01
### Added
- Feature A
- Feature B

### Fixed
- Bug X

## [1.1.0] - 2024-04-01
- Older feature

[Unreleased]: https://github.com/owner/repo/compare/v1.1.1...HEAD
[1.1.1]: https://github.com/owner/repo/compare/v1.1.0...v1.1.1
[1.1.0]: https://github.com/owner/repo/releases/tag/v1.1.0
`,
			targetVersion: "v1.1.1",
			want: `### Added
- Feature A
- Feature B

### Fixed
- Bug X`,
		},
		{
			name: "auto-detect topmost released version when targetVersion is empty",
			changelog: `# Changelog
## [Unreleased]
- Ongoing work

## [1.1.1] - 2024-05-01
### Added
- Feature A

## [1.1.0] - 2024-04-01
- Older feature
`,
			targetVersion: "",
			want: `### Added
- Feature A`,
		},
		{
			name: "auto-detect topmost version when only unreleased exists",
			changelog: `# Changelog
## [Unreleased]
- Ongoing work
`,
			targetVersion: "",
			want:          "- Ongoing work",
		},
		{
			name: "standard keep a changelog without v prefix",
			changelog: `# Changelog
## [1.1.1] - 2024-05-01
### Added
- Feature A
`,
			targetVersion: "1.1.1",
			want: `### Added
- Feature A`,
		},
		{
			name: "unreleased extraction",
			changelog: `# Changelog
## [Unreleased]
### Added
- Upcoming feature

## [1.0.0] - 2024-01-01
- First release
`,
			targetVersion: "unreleased",
			want: `### Added
- Upcoming feature`,
		},
		{
			name: "headings without brackets and with date in parentheses",
			changelog: `# Changes
## 2.0.0 (2024-06-10)
- Major rewrite

## 1.0.0 (2023-01-01)
- Initial
`,
			targetVersion: "2.0.0",
			want:          "- Major rewrite",
		},
		{
			name: "headings with Release prefix",
			changelog: `# Changelog
## Release 3.0.0
* New UI

## Release 2.9.0
* Old UI
`,
			targetVersion: "v3.0.0",
			want:          "* New UI",
		},
		{
			name: "code block containing fake heading should not terminate section",
			changelog: `# Changelog
## [1.0.0]
Here is an example:

` + "```markdown" + `
## Not a real heading
` + "```" + `

And more text.

## [0.9.0]
Old notes.
`,
			targetVersion: "1.0.0",
			want: `Here is an example:

` + "```markdown" + `
## Not a real heading
` + "```" + `

And more text.`,
		},
		{
			name: "include header option",
			changelog: `# Changelog
## [1.0.0] - 2024-01-01
- Feature 1
`,
			targetVersion: "1.0.0",
			opts: &ExtractOptions{
				IncludeHeader:       true,
				TrimLinkDefinitions: true,
			},
			want: `## [1.0.0] - 2024-01-01
- Feature 1`,
		},
		{
			name: "last version in file with trailing reference links stripped",
			changelog: `# Changelog
## [1.0.0] - 2024-01-01
- Initial release

[1.0.0]: https://github.com/owner/repo/releases/tag/v1.0.0
`,
			targetVersion: "1.0.0",
			want:          "- Initial release",
		},
		{
			name: "h1 heading style",
			changelog: `# 2.0.0
- Feature in H1

# 1.0.0
- Feature in older H1
`,
			targetVersion: "2.0.0",
			want:          "- Feature in H1",
		},
		{
			name: "setext heading style",
			changelog: `2.0.0
-----
- Setext feature

1.0.0
-----
- Old setext feature
`,
			targetVersion: "2.0.0",
			want:          "- Setext feature",
		},
		{
			name: "version not found",
			changelog: `# Changelog
## [1.0.0]
- Initial
`,
			targetVersion: "2.0.0",
			wantErr:       ErrVersionNotFound,
		},
		{
			name:          "empty changelog",
			changelog:     "",
			targetVersion: "1.0.0",
			wantErr:       ErrEmptyChangelog,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			var err error
			if tt.opts != nil {
				got, err = ExtractWithOptions([]byte(tt.changelog), tt.targetVersion, *tt.opts)
			} else {
				got, err = Extract([]byte(tt.changelog), tt.targetVersion)
			}

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error to be %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.want {
				t.Errorf("Extract() =\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}
