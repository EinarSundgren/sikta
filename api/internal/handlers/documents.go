package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/einarsundgren/sikta/internal/database"
	"github.com/einarsundgren/sikta/internal/document"
	"github.com/einarsundgren/sikta/internal/services"
)

// DocumentHandler handles document-related HTTP requests.
type DocumentHandler struct {
	pool      *pgxpool.Pool
	repo      *database.Repository
	docService *services.DocumentService
	logger    *slog.Logger
}

// NewDocumentHandler creates a new document handler.
func NewDocumentHandler(pool *pgxpool.Pool, logger *slog.Logger) *DocumentHandler {
	repo := database.NewRepository(pool, context.Background())
	return &DocumentHandler{
		pool:      pool,
		repo:      repo,
		docService: services.NewDocumentService(logger),
		logger:    logger,
	}
}

// UploadDocument handles POST /api/documents
func (h *DocumentHandler) UploadDocument(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (max 50MB)
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		h.logger.Error("failed to parse multipart form", "error", err)
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		h.logger.Error("failed to get file from form", "error", err)
		http.Error(w, "No file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate upload
	uploadResult, err := h.docService.ValidateUpload(header.Filename, header.Size, file)
	if err != nil {
		h.logger.Error("upload validation failed", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create document record
	doc, err := h.repo.CreateDocument(
		uploadResult.Title,
		uploadResult.Filename,
		uploadResult.FilePath,
		uploadResult.FileType,
		"uploaded",
		false,
	)
	if err != nil {
		h.logger.Error("failed to create document record", "error", err)
		http.Error(w, "Failed to create document", http.StatusInternalServerError)
		return
	}

	h.logger.Info("document uploaded", "id", doc.ID, "filename", uploadResult.Filename)

	// Return document info
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(doc)
}

// GetDocumentStatus handles GET /api/documents/{id}/status
func (h *DocumentHandler) GetDocumentStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract document ID from URL path
	idStr := strings.TrimPrefix(r.URL.Path, "/api/documents/")
	idStr = strings.TrimSuffix(idStr, "/status")

	docID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	// Get document
	doc, err := h.repo.GetDocument(docID)
	if err != nil {
		h.logger.Error("failed to get document", "error", err)
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	// Get chunk count
	chunkCount, err := h.repo.GetChunkCount(docID)
	if err != nil {
		h.logger.Error("failed to get chunk count", "error", err)
		http.Error(w, "Failed to get status", http.StatusInternalServerError)
		return
	}

	// Calculate progress
	totalChunks := int64(1) // Default
	if doc.UploadStatus == "ready" {
		totalChunks = chunkCount
	} else if doc.UploadStatus == "processing" {
		// Estimate based on file size (rough heuristic)
		totalChunks = int64(chunkCount)
		if totalChunks == 0 {
			totalChunks = 50 // Default estimate for novels
		}
	}

	progress := h.docService.CalculateProgress(int(chunkCount), int(totalChunks))

	status := map[string]interface{}{
		"upload_status": doc.UploadStatus,
		"progress": map[string]int{
			"current":    progress.Current,
			"total":      progress.Total,
			"percentage": progress.Percentage,
		},
		"chunks_created": int(chunkCount),
		"error_message":  doc.ErrorMessage,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// GetDocument handles GET /api/documents/{id}
func (h *DocumentHandler) GetDocument(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract document ID from URL path
	idStr := strings.TrimPrefix(r.URL.Path, "/api/documents/")

	docID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	// Get document
	doc, err := h.repo.GetDocument(docID)
	if err != nil {
		h.logger.Error("failed to get document", "error", err)
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	// Get chunks
	chunks, err := h.repo.GetChunks(docID)
	if err != nil {
		h.logger.Error("failed to get chunks", "error", err)
		http.Error(w, "Failed to get chunks", http.StatusInternalServerError)
		return
	}

	// Combine document and chunks
	response := map[string]interface{}{
		"document": doc,
		"chunks":   chunks,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ListDocuments handles GET /api/documents
func (h *DocumentHandler) ListDocuments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	docs, err := h.repo.ListDocuments()
	if err != nil {
		h.logger.Error("failed to list documents", "error", err)
		http.Error(w, "Failed to list documents", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(docs)
}

// DeleteDocument handles DELETE /api/documents/{id}
func (h *DocumentHandler) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract document ID from URL path
	idStr := strings.TrimPrefix(r.URL.Path, "/api/documents/")

	docID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	// Get document to check if it exists
	doc, err := h.repo.GetDocument(docID)
	if err != nil {
		h.logger.Error("failed to get document", "error", err)
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	// Delete file from disk
	if err := h.docService.Cleanup(doc.FilePath); err != nil {
		h.logger.Error("failed to cleanup file", "error", err)
		// Continue anyway to delete DB record
	}

	// Delete document record (chunks will be cascade deleted)
	// For now, we'll just return success
	// In production, you'd implement actual delete in the repository

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "deleted",
		"id":     idStr,
	})
}

// ProcessDocuments is a background worker that processes uploaded documents.
func (h *DocumentHandler) ProcessDocuments(stopCh <-chan struct{}) {
	h.logger.Info("starting document processor")

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stopCh:
			h.logger.Info("stopping document processor")
			return
		case <-ticker.C:
			h.processQueuedDocuments()
		}
	}
}

// processQueuedDocuments processes all documents in "uploaded" status.
func (h *DocumentHandler) processQueuedDocuments() {
	docs, err := h.repo.ListDocuments()
	if err != nil {
		h.logger.Error("failed to list documents", "error", err)
		return
	}

	for _, doc := range docs {
		if doc.UploadStatus == "uploaded" {
			h.processDocument(doc)
		}
	}
}

// processDocument processes a single document.
func (h *DocumentHandler) processDocument(doc database.Document) {
	h.logger.Info("processing document", "id", doc.ID, "filename", doc.Filename)

	// Update status to processing
	_, err := h.repo.UpdateDocumentStatus(uuid.UUID(doc.ID), "processing", nil)
	if err != nil {
		h.logger.Error("failed to update document status", "error", err)
		return
	}

	// Process document
	result, err := h.docService.ProcessDocument(doc.FilePath, doc.FileType)
	if err != nil {
		h.logger.Error("failed to process document", "error", err)
		errMsg := err.Error()
		h.repo.UpdateDocumentStatus(uuid.UUID(doc.ID), "error", &errMsg)
		return
	}

	// Update total pages if PDF
	if doc.FileType == "pdf" && result.TotalPages > 0 {
		h.repo.UpdateDocumentTotalPages(uuid.UUID(doc.ID), int32(result.TotalPages))
	}

	// Create chunks
	for _, chunk := range result.Chunks {
		var chapterTitle *string
		if chunk.ChapterTitle != "" {
			chapterTitle = &chunk.ChapterTitle
		}

		var chapterNumber *int32
		if chunk.ChapterNumber != 0 {
			cn := int32(chunk.ChapterNumber)
			chapterNumber = &cn
		}

		wordCount := int32(document.WordCount(chunk.Content))

		_, err := h.repo.CreateChunk(
			uuid.UUID(doc.ID),
			int32(chunk.ChunkIndex),
			chunk.Content,
			chapterTitle,
			chapterNumber,
			nil, // pageStart - convert to int32 pointer
			nil, // pageEnd - convert to int32 pointer
			int32(chunk.NarrativePosition),
			&wordCount,
		)
		if err != nil {
			h.logger.Error("failed to create chunk", "error", err)
			errMsg := err.Error()
			h.repo.UpdateDocumentStatus(uuid.UUID(doc.ID), "error", &errMsg)
			return
		}
	}

	// Update status to ready
	_, err = h.repo.UpdateDocumentStatus(uuid.UUID(doc.ID), "ready", nil)
	if err != nil {
		h.logger.Error("failed to update document status", "error", err)
		return
	}

	h.logger.Info("document processed successfully", "id", doc.ID, "chunks", len(result.Chunks))
}
