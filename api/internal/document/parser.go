package document

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

const (
	targetWords = 3000
	maxWords    = 4500
	minWords    = 1500
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

// isLineEmpty returns true if the line is empty or contains only whitespace.
func isLineEmpty(line string) bool {
	return strings.TrimSpace(line) == ""
}

// splitIntoParagraphs splits text into paragraphs separated by blank lines.
func splitIntoParagraphs(content string) []string {
	lines := strings.Split(content, "\n")
	var paragraphs []string
	var current []string

	for _, line := range lines {
		if isLineEmpty(line) {
			if len(current) > 0 {
				paragraphs = append(paragraphs, strings.Join(current, "\n"))
				current = nil
			}
		} else {
			current = append(current, line)
		}
	}
	if len(current) > 0 {
		paragraphs = append(paragraphs, strings.Join(current, "\n"))
	}

	return paragraphs
}

// ParseTXT parses a TXT file into word-budget chunks split on paragraph boundaries.
// Works on any plain text regardless of chapter formatting conventions.
func ParseTXT(content string) ([]Chunk, error) {
	if !utf8.ValidString(content) {
		return nil, fmt.Errorf("invalid UTF-8 encoding")
	}

	content = StripGutenbergBoilerplate(content)

	paragraphs := splitIntoParagraphs(content)
	if len(paragraphs) == 0 {
		return []Chunk{{
			ChunkIndex:        0,
			Content:           content,
			ChapterTitle:      "",
			ChapterNumber:     0,
			NarrativePosition: 0,
			PageStart:         nil,
			PageEnd:           nil,
		}}, nil
	}

	var chunks []Chunk
	var current []string
	currentWords := 0

	flush := func() {
		if len(current) == 0 {
			return
		}
		chunkContent := strings.Join(current, "\n\n")
		chunks = append(chunks, Chunk{
			ChunkIndex:        len(chunks),
			Content:           chunkContent,
			ChapterTitle:      "",
			ChapterNumber:     0,
			NarrativePosition: len(chunks),
			PageStart:         nil,
			PageEnd:           nil,
		})
		current = nil
		currentWords = 0
	}

	for _, para := range paragraphs {
		paraWords := WordCount(para)
		// If adding this paragraph would exceed maxWords, flush first.
		if currentWords > 0 && currentWords+paraWords > maxWords {
			flush()
		}
		current = append(current, para)
		currentWords += paraWords
		// If we've reached the target, flush.
		if currentWords >= targetWords {
			flush()
		}
	}

	flush()

	// If the trailing chunk is too short, merge it into the previous.
	if len(chunks) >= 2 {
		last := chunks[len(chunks)-1]
		if WordCount(last.Content) < minWords {
			prev := chunks[len(chunks)-2]
			chunks[len(chunks)-2] = Chunk{
				ChunkIndex:        prev.ChunkIndex,
				Content:           prev.Content + "\n\n" + last.Content,
				ChapterTitle:      prev.ChapterTitle,
				ChapterNumber:     prev.ChapterNumber,
				NarrativePosition: prev.NarrativePosition,
				PageStart:         nil,
				PageEnd:           nil,
			}
			chunks = chunks[:len(chunks)-1]
		}
	}

	// Re-index after potential merge.
	for i := range chunks {
		chunks[i].ChunkIndex = i
		chunks[i].NarrativePosition = i
	}

	return chunks, nil
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
