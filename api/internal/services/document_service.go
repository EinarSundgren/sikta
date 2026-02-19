package services

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/einarsundgren/sikta/internal/document"
)

const (
	maxFileSize      = 50 * 1024 * 1024 // 50MB
	uploadsDir       = "uploads"
	uploadsDemoDir   = "uploads/demo"
	statusUploaded   = "uploaded"
	statusProcessing = "processing"
	statusReady      = "ready"
	statusError      = "error"
)

// DocumentService handles document processing business logic.
type DocumentService struct {
	logger *slog.Logger
}

// NewDocumentService creates a new document service.
func NewDocumentService(logger *slog.Logger) *DocumentService {
	return &DocumentService{
		logger: logger,
	}
}

// UploadResult contains the result of a document upload.
type UploadResult struct {
	Filename string
	FilePath string
	FileType string
	Title    string
}

// ValidateUpload validates an uploaded file.
func (s *DocumentService) ValidateUpload(filename string, size int64, reader io.Reader) (*UploadResult, error) {
	// Check file size
	if size > maxFileSize {
		return nil, fmt.Errorf("file exceeds %dMB limit", maxFileSize/(1024*1024))
	}

	if size == 0 {
		return nil, fmt.Errorf("file is empty")
	}

	// Create uploads directory if it doesn't exist
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create uploads directory: %w", err)
	}

	// Generate unique filename
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return nil, fmt.Errorf("file has no extension")
	}

	// Validate file type by extension
	if ext != ".txt" && ext != ".pdf" {
		return nil, fmt.Errorf("only PDF and TXT files supported")
	}

	fileType := strings.TrimPrefix(ext, ".")

	// Generate unique filename
	uniqueFilename := fmt.Sprintf("%s_%s%s", uuid.New().String(), filename, ext)
	filePath := filepath.Join(uploadsDir, uniqueFilename)

	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy content
	copied, err := io.Copy(file, reader)
	if err != nil {
		os.Remove(filePath) // Clean up on error
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	if copied != size {
		os.Remove(filePath) // Clean up on error
		return nil, fmt.Errorf("file size mismatch")
	}

	// Detect actual file type from content
	detectedType, err := document.DetectFileType(filePath)
	if err != nil {
		os.Remove(filePath) // Clean up on error
		return nil, fmt.Errorf("failed to detect file type: %w", err)
	}

	// Verify extension matches detected type
	if detectedType != fileType {
		os.Remove(filePath) // Clean up on error
		return nil, fmt.Errorf("file extension (%s) doesn't match detected type (%s)", fileType, detectedType)
	}

	// Generate title from filename
	title := strings.TrimSuffix(filename, ext)
	title = strings.ReplaceAll(title, "_", " ")
	title = strings.ReplaceAll(title, "-", " ")

	return &UploadResult{
		Filename: uniqueFilename,
		FilePath: filePath,
		FileType: detectedType,
		Title:    title,
	}, nil
}

// ProcessResult contains the result of document processing.
type ProcessResult struct {
	Chunks    []document.Chunk
	TotalPages int
	Warnings  []string
}

// ProcessDocument processes a document file and extracts chunks.
func (s *DocumentService) ProcessDocument(filePath, fileType string) (*ProcessResult, error) {
	var chunks []document.Chunk
	var warnings []string
	var totalPages int

	switch fileType {
	case "txt":
		content, err := s.readTXTFile(filePath)
		if err != nil {
			return nil, err
		}

		chunks, err = document.ParseTXT(content)
		if err != nil {
			return nil, fmt.Errorf("failed to parse TXT: %w", err)
		}

	case "pdf":
		var err error
		chunks, err = document.ParsePDFWithChunks(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PDF: %w", err)
		}

		// Calculate total pages from chunks
		for _, chunk := range chunks {
			if chunk.PageEnd != nil && *chunk.PageEnd > totalPages {
				totalPages = *chunk.PageEnd
			}
		}

	default:
		return nil, fmt.Errorf("unsupported file type: %s", fileType)
	}

	// Check for warnings
	if len(chunks) == 1 && chunks[0].ChapterNumber == 0 {
		warnings = append(warnings, "Document produced a single chunk (short text)")
	}

	s.logger.Info("document processed",
		"chunks", len(chunks),
		"total_pages", totalPages,
		"warnings", len(warnings))

	return &ProcessResult{
		Chunks:     chunks,
		TotalPages: totalPages,
		Warnings:   warnings,
	}, nil
}

// readTXTFile reads a TXT file and returns its content.
func (s *DocumentService) readTXTFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	content, err := document.ReadTXT(file)
	if err != nil {
		return "", fmt.Errorf("failed to read TXT: %w", err)
	}

	return content, nil
}

// Cleanup removes a document file from disk.
func (s *DocumentService) Cleanup(filePath string) error {
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove file: %w", err)
	}
	return nil
}

// Progress represents processing progress.
type Progress struct {
	Current     int
	Total       int
	Percentage  int
}

// CalculateProgress calculates progress based on chunks created vs estimated total.
func (s *DocumentService) CalculateProgress(chunksCreated int, totalChunks int) *Progress {
	if totalChunks == 0 {
		totalChunks = 1 // Avoid division by zero
	}

	percentage := (chunksCreated * 100) / totalChunks

	return &Progress{
		Current:    chunksCreated,
		Total:      totalChunks,
		Percentage: percentage,
	}
}

// EstimateChunks estimates the number of chunks based on word count.
func (s *DocumentService) EstimateChunks(wordCount int) int {
	// Rough heuristic: ~3000 words per chunk
	estimated := wordCount / 3000
	if estimated < 1 {
		estimated = 1
	}
	return estimated
}

func timestamp() string {
	return time.Now().Format("20060102_150405")
}
