package template

// MarkdownExpander processes Markdown content, tracking code fences and handling
// special comment markers of the form <!--BEGIN--KEY--> and <!--CLOSE--KEY-->.
// It supports single-line replacements (both markers on one line) and multi-line
// block replacements. The KEY in BEGIN/CLOSE tags must:
//   - Consist of uppercase letters, digits, or underscores (matching [A-Z0-9_]+)
//   - Be identical between BEGIN and CLOSE
// The type of marker (BEGIN or CLOSE) must consist of only uppercase letters (matching [A-Z]+).
// Markers placed inside fenced code blocks are ignored.

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// MarkdownExpander manages the parsing and expansion of Markdown content.
// It detects fenced code blocks and expands custom BEGIN/CLOSE comment markers
// with dynamically looked-up values using a user-defined Lookup function.
type MarkdownExpander struct {
	fenceChar    rune
	fenceLength  int
	lineNumber   int
	openMarkers  map[string]int
	closeMarkers map[string]int
	replaceKey   string

	fenceRegex  *regexp.Regexp
	markerRegex *regexp.Regexp

	Lookup func(string) (string, error)
}

// NewMarkdownExpander initializes and returns a MarkdownExpander with default regex patterns
// for detecting code fences and comment markers. It also sets a default Lookup function.
func NewMarkdownExpander() *MarkdownExpander {
	return &MarkdownExpander{
		openMarkers:  make(map[string]int),
		closeMarkers: make(map[string]int),
		fenceRegex:   regexp.MustCompile(`^(` + "`{3,}" + `|~{3,})(\w*)?`),
		markerRegex:  regexp.MustCompile(`<!--([A-Z]+)--([A-Z0-9_]+)-->`),
		Lookup: func(key string) (string, error) {
			return fmt.Sprintf("[Content for %s]", key), nil
		},
	}
}

// Expand reads and processes the given Markdown content. It scans for BEGIN/CLOSE markers
// and replaces the content between them using the Lookup function. Content inside code fences is ignored.
// Output is written to the provided io.Writer.
func (m *MarkdownExpander) Expand(content []byte, target io.Writer) error {
	scanner := bufio.NewScanner(bytes.NewReader(content))

	for scanner.Scan() {
		line := scanner.Text()
		m.lineNumber++
		trimmed := strings.TrimSpace(line)

		// Find all marker matches in the current line.
		matches := m.markerRegex.FindAllStringSubmatch(trimmed, -1)
		if !m.InFence() && len(matches) > 0 {
			var beginKey, closeKey string
			for _, match := range matches {
				typ := match[1]
				key := match[2]
				if typ == "BEGIN" {
					beginKey = key
				} else if typ == "CLOSE" {
					closeKey = key
				}
			}

			// Handle single-line replacement.
			if beginKey != "" && closeKey == beginKey {
				replacement, err := m.Lookup(beginKey)
				if err != nil {
					return fmt.Errorf("lookup failed for %s: %w", beginKey, err)
				}
				fmt.Fprintf(target, "<!--BEGIN--%s-->", beginKey)
				fmt.Fprintln(target, replacement)
				fmt.Fprintf(target, "<!--CLOSE--%s-->", closeKey)
				continue
			} else if beginKey != "" {
				// Start a multi-line replacement block.
				m.replaceKey = beginKey
				fmt.Fprintln(target, line)
				continue
			} else if closeKey != "" && m.replaceKey == closeKey {
				// End a multi-line replacement block.
				replacement, err := m.Lookup(closeKey)
				if err != nil {
					return fmt.Errorf("lookup failed for %s: %w", closeKey, err)
				}
				fmt.Fprintln(target, replacement)
				fmt.Fprintln(target, line)
				m.replaceKey = ""
				continue
			}
		}

		// If inside a replacement block, skip lines until the CLOSE marker.
		if m.InReplacement() {
			continue
		}

		// Process the line for code fences and output.
		m.processLine(line, target)
	}

	// Check for scanner errors.
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan error: %w", err)
	}

	return nil
}

// InReplacement returns true if the MarkdownExpander is currently inside a BEGIN block
// waiting for a corresponding CLOSE marker.
func (m *MarkdownExpander) InReplacement() bool {
	return m.replaceKey != ""
}

// processLine evaluates a line of Markdown, updating the fence state if needed,
// and writes the line to the output. Used when not inside a replacement block.
func (m *MarkdownExpander) processLine(line string, target io.Writer) {
	trimmed := strings.TrimSpace(line)

	// Check for code fence start/end.
	if m.fenceRegex.MatchString(trimmed) {
		m.handleFence(trimmed)
		fmt.Fprintln(target, line)
		return
	}

	// Output the line as-is.
	fmt.Fprintln(target, line)
}

// handleFence adjusts the internal fence state based on whether a new fence starts or
// an existing one ends. Fences are marked by triple backticks or tildes.
func (m *MarkdownExpander) handleFence(trimmed string) {
	char := rune(trimmed[0])
	count := m.countRunes(trimmed, char)

	if !m.InFence() {
		// Entering a new code fence.
		m.fenceChar = char
		m.fenceLength = count
	} else if char == m.fenceChar && count >= m.fenceLength {
		// Exiting the current code fence.
		m.fenceChar = 0
		m.fenceLength = 0
	}
}

// countRunes counts how many times a rune appears consecutively from the beginning of a string.
func (m *MarkdownExpander) countRunes(s string, r rune) int {
	count := 0
	for _, ch := range s {
		if ch == r {
			count++
		} else {
			break
		}
	}
	return count
}

// InFence returns true if currently inside a fenced code block.
func (m *MarkdownExpander) InFence() bool {
	return m.fenceLength > 0
}
