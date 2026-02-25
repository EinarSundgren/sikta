package document

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

const (
	targetWords = 2000
	maxWords    = 3000
	minWords    = 1000
)

// StripGutenbergBoilerplate removes Project Gutenberg header and footer.
// Non-Gutenberg texts pass through unchanged.
func StripGutenbergBoilerplate(content string) string {
	const startMarker = "*** START OF THE PROJECT GUTENBERG EBOOK"
	const endMarker = "*** END OF THE PROJECT GUTENBERG EBOOK"

	if start := strings.Index(content, startMarker); start != -1 {
		rest := content[start+len(startMarker):]
		if nl := strings.Index(rest, "\n"); nl != -1 {
			content = rest[nl+1:]
		} else {
			content = rest
		}
	}

	if end := strings.Index(content, endMarker); end != -1 {
		content = content[:end]
	}

	return strings.TrimSpace(content)
}

// ParseTXT parses a TXT file using auto-detected chunking strategy.
// Works on any plain text regardless of formatting conventions.
func ParseTXT(content string) ([]Chunk, error) {
	if !utf8.ValidString(content) {
		return nil, fmt.Errorf("invalid UTF-8 encoding")
	}

	content = StripGutenbergBoilerplate(content)

	// Auto-detect and apply appropriate strategy
	strategy := DetectChunkStrategy(content)
	return strategy.Chunk(content), nil
}

// ParseTXTWithStrategy parses a TXT file using a specific chunking strategy.
func ParseTXTWithStrategy(content string, strategy ChunkStrategy) ([]Chunk, error) {
	if !utf8.ValidString(content) {
		return nil, fmt.Errorf("invalid UTF-8 encoding")
	}

	content = StripGutenbergBoilerplate(content)
	return strategy.Chunk(content), nil
}

// ReadTXT reads a TXT file from a reader and returns the content as a string.
func ReadTXT(r io.Reader) (string, error) {
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		return "", fmt.Errorf("failed to read TXT file: %w", err)
	}

	content := buf.String()

	// Remove UTF-8 BOM if present.
	if strings.HasPrefix(content, "\ufeff") {
		content = content[3:]
	}

	return content, nil
}
