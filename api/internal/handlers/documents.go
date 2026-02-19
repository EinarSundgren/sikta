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
	pool       *pgxpool.Pool
	repo       *database.Repository
	docService *services.DocumentService
	logger     *slog.Logger
}

// NewDocumentHandler creates a new document handler.
func NewDocumentHandler(pool *pgxpool.Pool, logger *slog.Logger) *DocumentHandler {
	repo := database.NewRepository(pool, context.Background())
	return &DocumentHandler{
		pool:       pool,
		repo:       repo,
		docService: services.NewDocumentService(logger),
		logger:     logger,
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

	// Create source record
	src, err := h.repo.CreateSource(
		uploadResult.Title,
		uploadResult.Filename,
		uploadResult.FilePath,
		uploadResult.FileType,
		"uploaded",
		false,
	)
	if err != nil {
		h.logger.Error("failed to create source record", "error", err)
		http.Error(w, "Failed to create document", http.StatusInternalServerError)
		return
	}

	h.logger.Info("document uploaded", "id", src.ID, "filename", uploadResult.Filename)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(src)
}

// GetDocumentStatus handles GET /api/documents/{id}/status
func (h *DocumentHandler) GetDocumentStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/documents/")
	idStr = strings.TrimSuffix(idStr, "/status")

	srcID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	src, err := h.repo.GetSource(srcID)
	if err != nil {
		h.logger.Error("failed to get source", "error", err)
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	chunkCount, err := h.repo.GetChunkCount(srcID)
	if err != nil {
		h.logger.Error("failed to get chunk count", "error", err)
		http.Error(w, "Failed to get status", http.StatusInternalServerError)
		return
	}

	totalChunks := int64(1)
	if src.UploadStatus == "ready" {
		totalChunks = chunkCount
	} else if src.UploadStatus == "processing" {
		totalChunks = int64(chunkCount)
		if totalChunks == 0 {
			totalChunks = 50
		}
	}

	progress := h.docService.CalculateProgress(int(chunkCount), int(totalChunks))

	status := map[string]interface{}{
		"upload_status": src.UploadStatus,
		"progress": map[string]int{
			"current":    progress.Current,
			"total":      progress.Total,
			"percentage": progress.Percentage,
		},
		"chunks_created": int(chunkCount),
		"error_message":  database.TextPtr(src.ErrorMessage),
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

	idStr := strings.TrimPrefix(r.URL.Path, "/api/documents/")

	srcID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	src, err := h.repo.GetSource(srcID)
	if err != nil {
		h.logger.Error("failed to get source", "error", err)
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	chunks, err := h.repo.GetChunks(srcID)
	if err != nil {
		h.logger.Error("failed to get chunks", "error", err)
		http.Error(w, "Failed to get chunks", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"document": src,
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

	sources, err := h.repo.ListSources()
	if err != nil {
		h.logger.Error("failed to list sources", "error", err)
		http.Error(w, "Failed to list documents", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sources)
}

// DeleteDocument handles DELETE /api/documents/{id}
func (h *DocumentHandler) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/documents/")

	srcID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	src, err := h.repo.GetSource(srcID)
	if err != nil {
		h.logger.Error("failed to get source", "error", err)
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	if err := h.docService.Cleanup(src.FilePath); err != nil {
		h.logger.Error("failed to cleanup file", "error", err)
	}

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
			h.processQueuedSources()
		}
	}
}

// processQueuedSources processes all sources in "uploaded" status.
func (h *DocumentHandler) processQueuedSources() {
	sources, err := h.repo.ListSources()
	if err != nil {
		h.logger.Error("failed to list sources", "error", err)
		return
	}

	for _, src := range sources {
		if src.UploadStatus == "uploaded" {
			h.processSource(src)
		}
	}
}

// processSource processes a single source document.
func (h *DocumentHandler) processSource(src *database.Source) {
	h.logger.Info("processing source", "id", src.ID, "filename", src.Filename)

	srcID := uuid.UUID(src.ID.Bytes)

	_, err := h.repo.UpdateSourceStatus(srcID, "processing", nil)
	if err != nil {
		h.logger.Error("failed to update source status", "error", err)
		return
	}

	result, err := h.docService.ProcessDocument(src.FilePath, src.FileType)
	if err != nil {
		h.logger.Error("failed to process source", "error", err)
		errMsg := err.Error()
		h.repo.UpdateSourceStatus(srcID, "error", &errMsg)
		return
	}

	if src.FileType == "pdf" && result.TotalPages > 0 {
		h.repo.UpdateSourceTotalPages(srcID, int32(result.TotalPages))
	}

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
			srcID,
			int32(chunk.ChunkIndex),
			chunk.Content,
			chapterTitle,
			chapterNumber,
			nil,
			nil,
			int32(chunk.NarrativePosition),
			&wordCount,
		)
		if err != nil {
			h.logger.Error("failed to create chunk", "error", err)
			errMsg := err.Error()
			h.repo.UpdateSourceStatus(srcID, "error", &errMsg)
			return
		}
	}

	_, err = h.repo.UpdateSourceStatus(srcID, "ready", nil)
	if err != nil {
		h.logger.Error("failed to update source status", "error", err)
		return
	}

	h.logger.Info("source processed successfully", "id", src.ID, "chunks", len(result.Chunks))
}
