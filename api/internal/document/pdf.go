package document

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// PageMarker is inserted during PDF text extraction to mark page boundaries.
const PageMarker = "\n[[[PAGE %d]]]\n"

// ParsePDF extracts text from a PDF file using pdftotext.
func ParsePDF(filePath string) (string, map[int]int, error) {
	// Check if pdftotext is available
	if _, err := exec.LookPath("pdftotext"); err != nil {
		return "", nil, fmt.Errorf("pdftotext not found: %w (install poppler-utils)", err)
	}

	// Get total page count first
	pageCount, err := getPDFPageCount(filePath)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get PDF page count: %w", err)
	}

	// Extract text with page markers
	var textWithMarkers strings.Builder
	for page := 1; page <= pageCount; page++ {
		// Extract single page
		cmd := exec.Command("pdftotext", "-layout", "-f", strconv.Itoa(page), "-l", strconv.Itoa(page), filePath, "-")
		output, err := cmd.Output()
		if err != nil {
			return "", nil, fmt.Errorf("failed to extract page %d: %w", page, err)
		}

		// Add page marker before content
		fmt.Fprintf(&textWithMarkers, PageMarker, page)
		textWithMarkers.Write(output)
		textWithMarkers.WriteString("\n")
	}

	// Build character offset to page number lookup table
	offsetToPage := buildOffsetTable(textWithMarkers.String())

	// Remove page markers from final output
	cleanText := regexp.MustCompile(`\n\[\[\[PAGE \d+\]\]\]\n`).ReplaceAllString(textWithMarkers.String(), "\n")

	return cleanText, offsetToPage, nil
}

// getPDFPageCount returns the total number of pages in a PDF.
func getPDFPageCount(filePath string) (int, error) {
	cmd := exec.Command("pdftotext", "-layout", filePath, "-")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	// Count page markers in the output
	re := regexp.MustCompile(`\[\[\[PAGE \d+\]\]\]`)
	matches := re.FindAllString(string(output), -1)
	return len(matches), nil
}

// buildOffsetTable creates a lookup table from character offset to page number.
func buildOffsetTable(text string) map[int]int {
	table := make(map[int]int)

	re := regexp.MustCompile(`\[\[\[PAGE (\d+)\]\]\]`)

	// Initialize: before first page marker, assume page 1
	table[0] = 1

	// Find all page markers and record their positions
	matches := re.FindAllStringSubmatchIndex(text, -1)
	for _, match := range matches {
		// match[0] is start of full match, match[1] is end
		// match[2] is start of capture group, match[3] is end
		pageNumStr := text[match[2]:match[3]]
		pageNum, _ := strconv.Atoi(pageNumStr)

		// Record that from this offset onward, we're on this page
		table[match[0]] = pageNum
	}

	return table
}

// getPageRange returns the start and end page numbers for a chunk of text.
func getPageRange(text string, fullText string, offsetToPage map[int]int) (startPage, endPage int) {
	// Find where this chunk starts in the full text
	chunkOffset := strings.Index(fullText, text)
	if chunkOffset == -1 {
		return 1, 1 // Not found, default to page 1
	}

	// Find the page at this offset
	startPage = 1
	for offset, page := range offsetToPage {
		if offset <= chunkOffset && page > startPage {
			startPage = page
		}
	}

	// Find the page at the end of this chunk
	chunkEndOffset := chunkOffset + len(text)
	endPage = startPage
	for offset, page := range offsetToPage {
		if offset >= chunkEndOffset {
			break
		}
		if offset <= chunkEndOffset && page > endPage {
			endPage = page
		}
	}

	// Handle edge case where end page might be before start page
	if endPage < startPage {
		endPage = startPage
	}

	return startPage, endPage
}

// ParsePDFWithChunks extracts text from a PDF and splits it into chapter-based chunks with page info.
func ParsePDFWithChunks(filePath string) ([]Chunk, error) {
	// Extract text with page markers
	text, offsetToPage, err := ParsePDF(filePath)
	if err != nil {
		return nil, err
	}

	// Parse chapters using the same logic as TXT
	chunks, err := ParseTXT(text)
	if err != nil {
		return nil, err
	}

	// Add page information to each chunk
	for i := range chunks {
		startPage, endPage := getPageRange(chunks[i].Content, text, offsetToPage)
		chunks[i].PageStart = &startPage
		chunks[i].PageEnd = &endPage
	}

	return chunks, nil
}

// ReadPDF reads a PDF file and returns the raw text content.
func ReadPDF(filePath string) (string, error) {
	if _, err := exec.LookPath("pdftotext"); err != nil {
		return "", fmt.Errorf("pdftotext not found: %w (install poppler-utils)", err)
	}

	cmd := exec.Command("pdftotext", "-layout", filePath, "-")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to extract PDF text: %w", err)
	}

	return string(output), nil
}

// DetectFileType determines if a file is TXT or PDF based on its content.
func DetectFileType(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read first 512 bytes for magic number detection
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return "", fmt.Errorf("failed to read file header: %w", err)
	}

	// Check for PDF magic number (%PDF-)
	if bytes.HasPrefix(buffer, []byte("%PDF-")) {
		return "pdf", nil
	}

	// Check if it's valid UTF-8 text (likely TXT)
	if utf8Valid := isValidUTF8(buffer); utf8Valid {
		return "txt", nil
	}

	return "", fmt.Errorf("unable to detect file type (not PDF or valid UTF-8 text)")
}

// isValidUTF8 checks if a byte slice is valid UTF-8.
func isValidUTF8(b []byte) bool {
	// Check for common text indicators
	hasPrintable := false
	for _, c := range b {
		if c >= 32 && c <= 126 || c == '\t' || c == '\n' || c == '\r' {
			hasPrintable = true
		}
	}
	return hasPrintable
}
