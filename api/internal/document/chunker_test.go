package document

import (
	"strings"
	"testing"
)

func TestDetectChunkStrategy(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "short document without markers uses whole doc",
			content:  strings.Repeat("word ", 1000), // ~1000 words
			expected: "whole_doc",
		},
		{
			name: "short document with section markers uses section chunker",
			content: `Protokoll fört vid sammanträde 2024-02-13

§1 Beslut om upphandling
Styrelsen beslutade att upphandla fasadrenovering.

§2 Ekonomisk rapport
Ekonomisk redovisning presenterades.

§3 Övriga frågor
Inga övriga frågor.`,
			expected: "section",
		},
		{
			name: "short document with chapter markers uses chapter chunker",
			content: `Chapter 1
It is a truth universally acknowledged...

Chapter 2
Mr. Bingley had soon made himself acquainted...

Chapter 3
The ladies of Longbourn...`,
			expected: "chapter",
		},
		{
			name:     "long plain text uses fallback",
			content:  strings.Repeat("This is plain text without any special markers. ", 400), // ~4000 words
			expected: "fallback",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy := DetectChunkStrategy(tt.content)
			if strategy.Name() != tt.expected {
				t.Errorf("DetectChunkStrategy() = %v, want %v", strategy.Name(), tt.expected)
			}
		})
	}
}

func TestWholeDocChunker(t *testing.T) {
	content := "This is a test document.\n\nIt has multiple paragraphs.\n\nBut should be one chunk."
	chunker := &WholeDocChunker{}

	chunks := chunker.Chunk(content)

	if len(chunks) != 1 {
		t.Errorf("WholeDocChunker returned %d chunks, want 1", len(chunks))
	}
	if chunks[0].Content != strings.TrimSpace(content) {
		t.Errorf("Content mismatch")
	}
}

func TestSectionChunker(t *testing.T) {
	content := `Protokoll 2024-02-13

§1 Beslut om upphandling
Styrelsen beslutade att upphandla fasadrenovering till en budget om 650 000 kr.
Arbetet ska vara slutfört senast oktober 2024.

§2 Ekonomisk rapport
Ekonomisk redovisning för Q1 presenterades av kassören.
Balansen i föreningen är positiv.

§3 Övriga frågor
Inga övriga frågor togs upp under mötet.`

	chunker := &SectionChunker{}
	chunks := chunker.Chunk(content)

	if len(chunks) < 2 {
		t.Errorf("SectionChunker returned %d chunks, expected at least 2", len(chunks))
	}

	// Check that sections have IDs
	for _, chunk := range chunks {
		if chunk.SectionID == "" && chunk.ChunkIndex > 0 {
			t.Errorf("Chunk %d missing SectionID", chunk.ChunkIndex)
		}
	}
}

func TestChapterChunker(t *testing.T) {
	content := `Chapter 1
It is a truth universally acknowledged, that a single man in possession
of a good fortune, must be in want of a wife.

Chapter 2
Mr. Bingley was good-looking and gentlemanlike; he had a pleasant countenance.

Chapter 3
The ladies of Longbourn soon waited on those of Netherfield.`

	chunker := &ChapterChunker{}
	chunks := chunker.Chunk(content)

	if len(chunks) != 3 {
		t.Errorf("ChapterChunker returned %d chunks, want 3", len(chunks))
	}

	// Check chapter numbers
	for i, chunk := range chunks {
		expectedNum := i + 1
		if chunk.ChapterNumber != expectedNum {
			t.Errorf("Chunk %d has ChapterNumber %d, want %d", i, chunk.ChapterNumber, expectedNum)
		}
	}
}

func TestFallbackChunker(t *testing.T) {
	// Generate content with paragraph breaks that's long enough to trigger multiple chunks
	var content string
	for i := 0; i < 200; i++ {
		content += "This is paragraph number " + string(rune('0'+i%10)) + " with enough words to fill the buffer properly and trigger splitting logic.\n\n"
	}

	chunker := &FallbackChunker{}
	chunks := chunker.Chunk(content)

	if len(chunks) < 2 {
		t.Errorf("FallbackChunker returned %d chunks, expected at least 2 for long content", len(chunks))
	}

	// Verify chunks are indexed correctly
	for i, chunk := range chunks {
		if chunk.ChunkIndex != i {
			t.Errorf("Chunk %d has ChunkIndex %d", i, chunk.ChunkIndex)
		}
	}
}
