package document

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode/utf8"
)

// ChapterPattern represents a detected chapter heading with its position.
type ChapterPattern struct {
	LineNumber  int    // 0-based line number in the file
	ChapterNum  int    // Parsed chapter number (0 if not parseable)
	Title       string // Full title including "Chapter X"
	PatternName string // Which pattern matched
}

// Parser state machine for chapter detection.
type parserState int

const (
	stateSearching    parserState = iota
	stateHeadingDetected
	stateInContent
)

// Chapter patterns ordered by priority (highest confidence first).
var chapterPatterns = []struct {
	name    string
	pattern *regexp.Regexp
}{
	{
		name:    "roman_chapter",
		pattern: regexp.MustCompile(`^Chapter\s+[IVXLCDM]+\b`),
	},
	{
		name:    "numeric_chapter",
		pattern: regexp.MustCompile(`^Chapter\s+\d+\b`),
	},
	{
		name:    "uppercase_roman",
		pattern: regexp.MustCompile(`^CHAPTER\s+[IVXLCDM]+\b`),
	},
	{
		name:    "uppercase_numeric",
		pattern: regexp.MustCompile(`^CHAPTER\s+\d+\b`),
	},
	{
		name:    "standalone_roman",
		pattern: regexp.MustCompile(`^[IVXLCDM]+\.\s+`),
	},
	{
		name:    "standalone_numeric",
		pattern: regexp.MustCompile(`^\d+\.\s+`),
	},
}

// Common false positives to ignore.
var antiPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^Page\s+\d+`),
	regexp.MustCompile(`^\d+\s+of\s+\d+`),
	regexp.MustCompile(`^©`),
	regexp.MustCompile(`^http`),
}

// romanToInt converts a Roman numeral to an integer.
func romanToInt(s string) int {
	romanValues := map[rune]int{'I': 1, 'V': 5, 'X': 10, 'L': 50, 'C': 100, 'D': 500, 'M': 1000}
	total := 0
	prev := 0

	for _, ch := range s {
		value := romanValues[ch]
		if value > prev {
			total += value - 2*prev
		} else {
			total += value
		}
		prev = value
	}
	return total
}

// extractChapterNumber attempts to parse a chapter number from the heading.
func extractChapterNumber(heading string) int {
	// Try "Chapter X" or "CHAPTER X" patterns
	re := regexp.MustCompile(`[Cc]hapter\s+([IVXLCDM\d]+)`)
	matches := re.FindStringSubmatch(heading)
	if len(matches) > 1 {
		numStr := matches[1]
		// Try Roman numeral first
		if regexp.MustCompile(`^[IVXLCDM]+$`).MatchString(numStr) {
			return romanToInt(numStr)
		}
		// Try numeric
		var num int
		if _, err := fmt.Sscanf(numStr, "%d", &num); err == nil {
			return num
		}
	}

	// Try standalone Roman numeral
	re = regexp.MustCompile(`^([IVXLCDM]+)\.\s+`)
	matches = re.FindStringSubmatch(heading)
	if len(matches) > 1 {
		return romanToInt(matches[1])
	}

	// Try standalone numeric
	re = regexp.MustCompile(`^(\d+)\.\s+`)
	matches = re.FindStringSubmatch(heading)
	if len(matches) > 1 {
		var num int
		if _, err := fmt.Sscanf(matches[1], "%d", &num); err == nil {
			return num
		}
	}

	return 0
}

// isLineEmpty returns true if the line is empty or contains only whitespace.
func isLineEmpty(line string) bool {
	return strings.TrimSpace(line) == ""
}

// isAntiPattern checks if a line matches known false positive patterns.
func isAntiPattern(line string) bool {
	for _, pattern := range antiPatterns {
		if pattern.MatchString(line) {
			return true
		}
	}
	return false
}

// detectChapters scans the text for chapter headings.
func detectChapters(content string) []ChapterPattern {
	var patterns []ChapterPattern
	scanner := bufio.NewScanner(strings.NewReader(content))

	lineNum := 0
	prevWasEmpty := false

	for scanner.Scan() {
		line := scanner.Text()
		lineTrimmed := strings.TrimSpace(line)

		// Check if this could be a chapter heading
		// Must follow an empty line or be at the start
		if prevWasEmpty || lineNum == 0 {
			for _, cp := range chapterPatterns {
				if cp.pattern.MatchString(lineTrimmed) {
					// Check it's not a false positive
					if !isAntiPattern(lineTrimmed) {
						chapterNum := extractChapterNumber(lineTrimmed)
						patterns = append(patterns, ChapterPattern{
							LineNumber:  lineNum,
							ChapterNum:  chapterNum,
							Title:       lineTrimmed,
							PatternName: cp.name,
						})
						break // Don't match multiple patterns for same line
					}
				}
			}
		}

		prevWasEmpty = isLineEmpty(line)
		lineNum++
	}

	// Filter and cluster: remove duplicates (consecutive matches are the same chapter)
	var filtered []ChapterPattern
	for i, p := range patterns {
		if i == 0 {
			filtered = append(filtered, p)
			continue
		}
		prev := patterns[i-1]
		// If same line or next line, it's a duplicate (e.g., "Chapter 1\nChapter 1")
		if p.LineNumber-prev.LineNumber > 1 {
			filtered = append(filtered, p)
		}
	}

	return filtered
}

// ParseTXT parses a TXT file and splits it into chapter-based chunks.
func ParseTXT(content string) ([]Chunk, error) {
	// Validate UTF-8
	if !utf8.ValidString(content) {
		return nil, fmt.Errorf("invalid UTF-8 encoding")
	}

	// Detect chapters
	chapterPatterns := detectChapters(content)

	// If no chapters detected, treat entire document as one chunk
	if len(chapterPatterns) == 0 {
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

	// Split content into chunks based on detected chapters
	lines := strings.Split(content, "\n")
	var chunks []Chunk

	for i, pattern := range chapterPatterns {
		// Determine content range: from this chapter to next chapter (or EOF)
		startLine := pattern.LineNumber
		var endLine int
		if i < len(chapterPatterns)-1 {
			endLine = chapterPatterns[i+1].LineNumber
		} else {
			endLine = len(lines)
		}

		// Extract chapter content
		chapterLines := lines[startLine+1 : endLine] // Skip heading itself

		// Trim leading/trailing empty lines
		firstNonEmpty := 0
		for firstNonEmpty < len(chapterLines) && isLineEmpty(chapterLines[firstNonEmpty]) {
			firstNonEmpty++
		}
		lastNonEmpty := len(chapterLines) - 1
		for lastNonEmpty >= 0 && isLineEmpty(chapterLines[lastNonEmpty]) {
			lastNonEmpty--
		}

		var contentLines []string
		if firstNonEmpty <= lastNonEmpty {
			contentLines = chapterLines[firstNonEmpty : lastNonEmpty+1]
		}

		// Collapse excessive blank lines (more than 2 consecutive -> 2)
		var cleanedLines []string
		blankCount := 0
		for _, line := range contentLines {
			if isLineEmpty(line) {
				blankCount++
				if blankCount <= 2 {
					cleanedLines = append(cleanedLines, line)
				}
			} else {
				blankCount = 0
				cleanedLines = append(cleanedLines, line)
			}
		}

		chunkContent := strings.Join(cleanedLines, "\n")

		// Validate minimum content threshold
		if WordCount(chunkContent) < 100 {
			// Too short, skip this pattern (might be false positive)
			continue
		}

		// Determine chapter number (use pattern if parsed, otherwise sequential)
		chapterNum := pattern.ChapterNum
		if chapterNum == 0 {
			chapterNum = i + 1
		}

		chunks = append(chunks, Chunk{
			ChunkIndex:        i,
			Content:           chunkContent,
			ChapterTitle:      pattern.Title,
			ChapterNumber:     chapterNum,
			NarrativePosition: i,
			PageStart:         nil,
			PageEnd:           nil,
		})
	}

	// Handle edge case: all patterns were false positives
	if len(chunks) == 0 {
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

	return chunks, nil
}

// ReadTXT reads a TXT file from a reader and returns the content as a string.
func ReadTXT(r io.Reader) (string, error) {
	// Read entire content
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		return "", fmt.Errorf("failed to read TXT file: %w", err)
	}

	content := buf.String()

	// Remove UTF-8 BOM if present
	if strings.HasPrefix(content, "\ufeff") {
		content = content[3:]
	}

	return content, nil
}
