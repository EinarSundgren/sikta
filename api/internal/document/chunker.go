// Package document handles document parsing and chunking.
// A chunk is one chapter (or the whole document if no chapters are detected).
// Position metadata on chunks enables source references in the extraction layer.
package document

import (
	"regexp"
	"strings"
)

// Chunk is the output of the parsing pipeline: one addressable unit of text
// with position metadata for source references.
type Chunk struct {
	ChunkIndex        int
	Content           string
	ChapterTitle      string // empty string if unknown
	ChapterNumber     int    // 0 if not a numbered chapter (prologue, epilogue, etc.)
	NarrativePosition int    // reading order (same as chunk_index for linear documents)

	// Section metadata for structured documents (protocols, reports)
	SectionID   string // e.g., "§5", "Section 3", "3.1"
	SectionName string // e.g., "Beslut om upphandling", "Financial Summary"

	// PDF only — nil for TXT files.
	PageStart *int
	PageEnd   *int
}

// ChunkStrategy defines how a document should be split into chunks
type ChunkStrategy interface {
	Chunk(content string) []Chunk
	Name() string
}

// wordCount returns the number of whitespace-separated words in s.
func WordCount(s string) int {
	return len(strings.Fields(s))
}

// DetectChunkStrategy analyzes content and returns the appropriate chunking strategy
func DetectChunkStrategy(content string) ChunkStrategy {
	wordCount := WordCount(content)

	// Short documents: no splitting needed (unless they have explicit structure)
	if wordCount < 3000 {
		// Still check for explicit section/chapter markers even in short docs
		if hasSectionMarkers(content) {
			return &SectionChunker{}
		}
		if hasChapterMarkers(content) {
			return &ChapterChunker{}
		}
		return &WholeDocChunker{}
	}

	// Long documents: check for structural markers
	if hasSectionMarkers(content) {
		return &SectionChunker{}
	}

	if hasChapterMarkers(content) {
		return &ChapterChunker{}
	}

	// Default: paragraph-based word budget
	return &FallbackChunker{}
}

// hasSectionMarkers checks for protocol/section-style formatting
func hasSectionMarkers(content string) bool {
	// Swedish section markers: §5, § 5
	if strings.Contains(content, "§") {
		return true
	}
	// English section markers: Section 5, SECTION 5
	sectionPattern := regexp.MustCompile(`(?i)^section\s+\d+`)
	if sectionPattern.MatchString(content) {
		return true
	}
	// Numbered headers: 1. Title, 2. Title, 3.1 Title
	numberedPattern := regexp.MustCompile(`(?m)^\d+\.\d*\s+[A-ZÄÖÅ]`)
	if numberedPattern.MatchString(content) {
		return true
	}
	return false
}

// hasChapterMarkers checks for novel-style chapter formatting
func hasChapterMarkers(content string) bool {
	// Chapter patterns: Chapter 1, CHAPTER I, Kapitel 1
	chapterPattern := regexp.MustCompile(`(?i)(^|\n)(chapter|kapitel)\s+(\d+|[ivxlcdm]+)`)
	return chapterPattern.MatchString(content)
}

// WholeDocChunker returns the entire document as a single chunk
type WholeDocChunker struct{}

func (c *WholeDocChunker) Name() string { return "whole_doc" }

func (c *WholeDocChunker) Chunk(content string) []Chunk {
	return []Chunk{{
		ChunkIndex:        0,
		Content:           strings.TrimSpace(content),
		ChapterTitle:      "",
		ChapterNumber:     0,
		NarrativePosition: 0,
		PageStart:         nil,
		PageEnd:           nil,
	}}
}

// SectionChunker splits on § markers and numbered headers
type SectionChunker struct{}

func (c *SectionChunker) Name() string { return "section" }

func (c *SectionChunker) Chunk(content string) []Chunk {
	// Pattern matches: §5, § 5, §5 Title, 5. Title, 5.1 Title
	sectionPattern := regexp.MustCompile(`(?m)^(\s*§\s*\d+|\d+\.\d*\s+[A-ZÄÖÅ])`)

	// Find all section start positions
	matches := sectionPattern.FindAllStringIndex(content, -1)
	if len(matches) == 0 {
		// No sections found, return as whole doc
		return (&WholeDocChunker{}).Chunk(content)
	}

	var chunks []Chunk

	for i, match := range matches {
		start := match[0]
		var end int
		if i+1 < len(matches) {
			end = matches[i+1][0]
		} else {
			end = len(content)
		}

		sectionContent := strings.TrimSpace(content[start:end])
		if len(sectionContent) < 50 {
			continue // Skip empty or tiny sections
		}

		// Extract section ID from the match
		firstLine := sectionContent
		if nl := strings.Index(sectionContent, "\n"); nl != -1 {
			firstLine = sectionContent[:nl]
		}

		sectionID := strings.TrimSpace(firstLine)
		var sectionName string
		if colon := strings.Index(firstLine, ":"); colon != -1 {
			sectionName = strings.TrimSpace(firstLine[colon+1:])
		} else if space := strings.Index(firstLine, " "); space != -1 && space < len(firstLine)-1 {
			// Extract name after the section number
			rest := firstLine[space+1:]
			// Skip leading dots or spaces
			rest = strings.TrimLeft(rest, ". ")
			if len(rest) > 0 {
				sectionName = rest
			}
		}

		chunks = append(chunks, Chunk{
			ChunkIndex:        len(chunks),
			Content:           sectionContent,
			ChapterTitle:      sectionName,
			ChapterNumber:     0,
			SectionID:         sectionID,
			SectionName:       sectionName,
			NarrativePosition: len(chunks),
			PageStart:         nil,
			PageEnd:           nil,
		})
	}

	// Handle content before first section (preamble)
	if len(matches) > 0 && matches[0][0] > 100 {
		preamble := strings.TrimSpace(content[:matches[0][0]])
		if len(preamble) >= 50 {
			// Insert preamble as first chunk
			preambleChunk := Chunk{
				ChunkIndex:        0,
				Content:           preamble,
				ChapterTitle:      "Preamble",
				ChapterNumber:     0,
				SectionID:         "0",
				SectionName:       "Preamble",
				NarrativePosition: 0,
				PageStart:         nil,
				PageEnd:           nil,
			}
			chunks = append([]Chunk{preambleChunk}, chunks...)
		}
	}

	// Re-index after potential preamble insertion
	for i := range chunks {
		chunks[i].ChunkIndex = i
		chunks[i].NarrativePosition = i
	}

	return chunks
}

// ChapterChunker splits on chapter markers (novels)
type ChapterChunker struct{}

func (c *ChapterChunker) Name() string { return "chapter" }

func (c *ChapterChunker) Chunk(content string) []Chunk {
	// Pattern matches: Chapter 1, CHAPTER I, Kapitel 1, etc.
	chapterPattern := regexp.MustCompile(`(?i)(^|\n)\s*(chapter|kapitel)\s+(\d+|[ivxlcdm]+)[^\n]*`)

	matches := chapterPattern.FindAllStringIndex(content, -1)
	if len(matches) == 0 {
		return (&FallbackChunker{}).Chunk(content)
	}

	var chunks []Chunk

	for i, match := range matches {
		start := match[0]
		var end int
		if i+1 < len(matches) {
			end = matches[i+1][0]
		} else {
			end = len(content)
		}

		chapterContent := strings.TrimSpace(content[start:end])

		// Extract chapter title
		firstLine := chapterContent
		if nl := strings.Index(chapterContent, "\n"); nl != -1 {
			firstLine = chapterContent[:nl]
		}
		chapterTitle := strings.TrimSpace(firstLine)

		// Extract chapter number
		chapterNum := 0
		numMatch := regexp.MustCompile(`\d+`).FindString(chapterTitle)
		if numMatch != "" {
			// Simple parse - could enhance for Roman numerals
			for _, c := range numMatch {
				chapterNum = chapterNum*10 + int(c-'0')
			}
		}

		chunks = append(chunks, Chunk{
			ChunkIndex:        len(chunks),
			Content:           chapterContent,
			ChapterTitle:      chapterTitle,
			ChapterNumber:     chapterNum,
			NarrativePosition: len(chunks),
			PageStart:         nil,
			PageEnd:           nil,
		})
	}

	return chunks
}

// FallbackChunker splits on paragraph boundaries with word budget
type FallbackChunker struct{}

func (c *FallbackChunker) Name() string { return "fallback" }

func (c *FallbackChunker) Chunk(content string) []Chunk {
	// This is essentially the existing ParseTXT logic
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
		}}
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
		if currentWords > 0 && currentWords+paraWords > maxWords {
			flush()
		}
		current = append(current, para)
		currentWords += paraWords
		if currentWords >= targetWords {
			flush()
		}
	}

	flush()

	// Merge trailing short chunk
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

	// Re-index
	for i := range chunks {
		chunks[i].ChunkIndex = i
		chunks[i].NarrativePosition = i
	}

	return chunks
}

// splitIntoParagraphs splits text into paragraphs separated by blank lines.
func splitIntoParagraphs(content string) []string {
	lines := strings.Split(content, "\n")
	var paragraphs []string
	var current []string

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
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
