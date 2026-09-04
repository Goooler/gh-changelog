# gh-changelog

A [GitHub CLI](https://cli.github.com/) extension to extract and parse release notes from changelog files (supporting [Keep a Changelog](https://keepachangelog.com/) and more) for GitHub Releases.

## Installation

```bash
gh extension install Goooler/gh-changelog
```

## Usage

### In GitHub Actions (CI)

Replace third-party extraction actions with `gh changelog extract`:

```yaml
- name: Extract release notes
  run: |
    gh extension install Goooler/gh-changelog
    gh changelog extract

- name: Create release
  run: gh release create ${{ github.ref_name }} --notes-file RELEASE_NOTES.md
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### CLI Commands

```bash
# Extract topmost release notes to RELEASE_NOTES.md (default)
gh changelog extract

# Extract specific version to RELEASE_NOTES.md
gh changelog extract v1.2.3

# Print release notes to stdout
gh changelog extract --stdout
gh changelog extract v1.2.3 --stdout

# Custom input changelog and output file
gh changelog extract v1.2.3 --file docs/CHANGELOG.md --output custom_notes.md
```

## Development

```bash
make help          # see all targets
make build         # build binary
make test          # run tests
make ci            # build + vet + test-race
make release-check # run the safe release preflight
make install-local # install extension from checkout
make relink-local  # reinstall after changes
```

`make release-check` checks formatting and module tidiness without retaining
changes, runs CI and vulnerability scanning, validates the GoReleaser config,
and creates a clean local snapshot in `dist/`. It does not tag, publish, or
create a GitHub release. Install a current GoReleaser v2 binary before running
it.

## Releasing

Update the changelog, commit all intended changes, and run the preflight before
tagging:

```bash
make release-check
git status --short
git tag -a v0.1.0 -m "Release v0.1.0"
git push origin v0.1.0
```

The tag push triggers GitHub Actions and GoReleaser. It creates standalone
binaries plus checksums for all six supported targets:

- macOS (`darwin/amd64`, `darwin/arm64`)
- Linux (`linux/amd64`, `linux/arm64`)
- Windows (`windows/amd64`, `windows/arm64`)

Once released, users install with:

```bash
gh extension install Goooler/gh-changelog
```

## What's included

| File | Purpose |
|------|---------|
| `Makefile` | Build, test, lint, install, and release-preflight targets |
| `.goreleaser.yml` | Cross-platform standalone binary releases |
| `.github/workflows/release.yml` | Validated automated releases on tag push |
| `.github/workflows/ci.yml` | CI on pushes and pull requests to `main` |
| `main.go` | Cobra CLI implementation with version and signal handling |
| `.gitignore` | Go, editor, and OS ignores |
