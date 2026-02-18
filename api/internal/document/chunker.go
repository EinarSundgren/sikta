// Package document handles document parsing and chunking.
// A chunk is one chapter (or the whole document if no chapters are detected).
// Position metadata on chunks enables source references in the extraction layer.
package document

import "strings"

// Chunk is the output of the parsing pipeline: one addressable unit of text
// with position metadata for source references.
type Chunk struct {
	ChunkIndex        int
	Content           string
	ChapterTitle      string // empty string if unknown
	ChapterNumber     int    // 0 if not a numbered chapter (prologue, epilogue, etc.)
	NarrativePosition int    // reading order (same as chunk_index for linear documents)

	// PDF only — nil for TXT files.
	PageStart *int
	PageEnd   *int
}

// wordCount returns the number of whitespace-separated words in s.
func WordCount(s string) int {
	return len(strings.Fields(s))
}
