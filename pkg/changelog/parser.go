package changelog

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

var (
	// ErrVersionNotFound is returned when the specified version is not found in the changelog.
	ErrVersionNotFound = errors.New("version not found in changelog")
	// ErrEmptyChangelog is returned when the changelog content is empty.
	ErrEmptyChangelog = errors.New("changelog content is empty")
)

// ExtractOptions configures changelog extraction behavior.
type ExtractOptions struct {
	// IncludeHeader, if true, includes the version heading line in the extracted output.
	IncludeHeader bool
	// TrimLinkDefinitions, if true, strips trailing reference links (e.g., `[1.0.0]: https://...`)
	// from the end of the extracted notes. Defaults to true.
	TrimLinkDefinitions bool
}

// DefaultOptions returns the default extraction options.
func DefaultOptions() ExtractOptions {
	return ExtractOptions{
		IncludeHeader:       false,
		TrimLinkDefinitions: true,
	}
}

// Result contains the extracted release notes and metadata.
type Result struct {
	Version string
	Notes   string
}

// Extract parses the markdown changelog and extracts release notes for the target version.
// If targetVersion is empty, it automatically extracts the topmost released version.
func Extract(source []byte, targetVersion string) (string, error) {
	res, err := ExtractVersion(source, targetVersion, DefaultOptions())
	if err != nil {
		return "", err
	}
	return res.Notes, nil
}

// ExtractWithOptions parses the markdown changelog with customized options.
func ExtractWithOptions(source []byte, targetVersion string, opts ExtractOptions) (string, error) {
	res, err := ExtractVersion(source, targetVersion, opts)
	if err != nil {
		return "", err
	}
	return res.Notes, nil
}

// ExtractVersion parses the markdown changelog and returns the Result.
func ExtractVersion(source []byte, targetVersion string, opts ExtractOptions) (Result, error) {
	if len(bytes.TrimSpace(source)) == 0 {
		return Result{}, ErrEmptyChangelog
	}

	targetVersion = strings.TrimSpace(targetVersion)

	parser := goldmark.DefaultParser()
	doc := parser.Parse(text.NewReader(source))

	var targetHeading *ast.Heading
	var targetHeadingIndex int
	var children []ast.Node

	var unreleasedHeading *ast.Heading
	var unreleasedIndex int

	for child := doc.FirstChild(); child != nil; child = child.NextSibling() {
		children = append(children, child)
		if h, ok := child.(*ast.Heading); ok {
			text := headingText(h, source)
			if targetVersion != "" {
				if targetHeading == nil && matchesVersion(text, targetVersion) {
					targetHeading = h
					targetHeadingIndex = len(children) - 1
				}
			} else {
				// Automatic topmost version detection
				if isDocumentTitle(text, h.Level) {
					continue
				}
				if isUnreleased(text) {
					if unreleasedHeading == nil {
						unreleasedHeading = h
						unreleasedIndex = len(children) - 1
					}
					continue
				}
				// Found the topmost released version
				if targetHeading == nil {
					targetHeading = h
					targetHeadingIndex = len(children) - 1
				}
			}
		}
	}

	if targetHeading == nil {
		if targetVersion == "" && unreleasedHeading != nil {
			targetHeading = unreleasedHeading
			targetHeadingIndex = unreleasedIndex
		} else {
			if targetVersion == "" {
				return Result{}, errors.New("no version headings found in changelog")
			}
			return Result{}, fmt.Errorf("%w: %q", ErrVersionNotFound, targetVersion)
		}
	}

	// Determine start offset
	var startOffset int
	lines := targetHeading.Lines()
	if lines.Len() == 0 {
		return Result{}, fmt.Errorf("invalid heading structure in changelog")
	}

	if opts.IncludeHeader {
		startOffset = lines.At(0).Start
		for startOffset > 0 && source[startOffset-1] != '\n' {
			startOffset--
		}
	} else {
		// Start immediately after the heading line
		lastLine := lines.At(lines.Len() - 1)
		startOffset = lastLine.Stop
		for startOffset < len(source) && source[startOffset] != '\n' {
			startOffset++
		}
		if startOffset < len(source) && source[startOffset] == '\n' {
			startOffset++
		}
		// In case of Setext heading (e.g. `2.0.0\n---`), skip the underline line
		startOffset = skipSetextUnderline(source, startOffset)
	}

	// Determine end offset: find the next heading of same or higher level (level <= targetHeading.Level)
	endOffset := len(source)
	for i := targetHeadingIndex + 1; i < len(children); i++ {
		if h, ok := children[i].(*ast.Heading); ok {
			if h.Level <= targetHeading.Level {
				if h.Lines().Len() > 0 {
					nextStart := h.Lines().At(0).Start
					for nextStart > 0 && source[nextStart-1] != '\n' {
						nextStart--
					}
					endOffset = nextStart
				}
				break
			}
		}
	}

	if startOffset > endOffset {
		startOffset = endOffset
	}

	extracted := string(source[startOffset:endOffset])

	if opts.TrimLinkDefinitions {
		extracted = stripTrailingLinkDefinitions(extracted)
	}

	return Result{
		Version: headingText(targetHeading, source),
		Notes:   strings.TrimSpace(extracted),
	}, nil
}

func isDocumentTitle(title string, level int) bool {
	if level > 1 {
		return false
	}
	cleaned := strings.ToLower(strings.TrimSpace(title))
	return cleaned == "changelog" || cleaned == "change log" || cleaned == "release notes" || cleaned == "history" || cleaned == "releases"
}

func isUnreleased(title string) bool {
	cleaned := strings.TrimSpace(title)
	cleaned = strings.Trim(cleaned, "[]()")
	return strings.EqualFold(cleaned, "unreleased")
}

// headingText extracts all text contents inside a heading node.
func headingText(h *ast.Heading, source []byte) string {
	var sb strings.Builder
	_ = ast.Walk(h, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			switch node := n.(type) {
			case *ast.Text:
				sb.Write(node.Segment.Value(source))
			case *ast.String:
				sb.Write(node.Value)
			case *ast.CodeSpan:
				for i := 0; i < node.Lines().Len(); i++ {
					seg := node.Lines().At(i)
					sb.Write(seg.Value(source))
				}
			}
		}
		return ast.WalkContinue, nil
	})
	return sb.String()
}

// matchesVersion checks if the heading text represents the target version.
func matchesVersion(headerText, targetVersion string) bool {
	headerText = strings.TrimSpace(headerText)
	cleanTarget := strings.TrimPrefix(strings.TrimPrefix(targetVersion, "v"), "V")

	// 1. Direct case-insensitive match for keywords like "Unreleased"
	if strings.EqualFold(headerText, targetVersion) || strings.EqualFold(headerText, cleanTarget) {
		return true
	}

	// 2. Pattern matching for version tags in headers
	// Matches:
	// - `[1.0.0]` / `[v1.0.0]`
	// - `1.0.0` / `v1.0.0`
	// - `Release 1.0.0` / `Version 1.0.0`
	// - `1.0.0 - 2024-01-01`
	quotedClean := regexp.QuoteMeta(cleanTarget)
	pattern := fmt.Sprintf(`(?i)(?:^|[\s\[vV(])%s(?:$|[\s\])\-:,])`, quotedClean)
	re, err := regexp.Compile(pattern)
	if err == nil && re.MatchString(headerText) {
		return true
	}

	// Also check with original targetVersion (in case of non-semver names like "2024.1")
	if cleanTarget != targetVersion {
		quotedRaw := regexp.QuoteMeta(targetVersion)
		patternRaw := fmt.Sprintf(`(?i)(?:^|[\s\[(])%s(?:$|[\s\])\-:,])`, quotedRaw)
		if reRaw, err := regexp.Compile(patternRaw); err == nil && reRaw.MatchString(headerText) {
			return true
		}
	}

	return false
}

var linkDefinitionRegex = regexp.MustCompile(`(?m)^\s*\[[^\]]+\]:\s*https?://\S+\s*$`)

// stripTrailingLinkDefinitions removes trailing markdown reference link definitions like:
// [1.0.0]: https://github.com/owner/repo/releases/tag/v1.0.0
func stripTrailingLinkDefinitions(content string) string {
	lines := strings.Split(content, "\n")
	lastContentIndex := -1

	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		if linkDefinitionRegex.MatchString(line) {
			continue
		}
		lastContentIndex = i
		break
	}

	if lastContentIndex == -1 {
		return ""
	}

	return strings.Join(lines[:lastContentIndex+1], "\n")
}

// skipSetextUnderline skips a Setext heading underline line if present at offset.
func skipSetextUnderline(source []byte, offset int) int {
	if offset >= len(source) {
		return offset
	}
	lineEnd := offset
	for lineEnd < len(source) && source[lineEnd] != '\n' {
		lineEnd++
	}
	trimmed := strings.TrimSpace(string(source[offset:lineEnd]))
	if len(trimmed) > 0 && (strings.Trim(trimmed, "-") == "" || strings.Trim(trimmed, "=") == "") {
		if lineEnd < len(source) && source[lineEnd] == '\n' {
			lineEnd++
		}
		return lineEnd
	}
	return offset
}
