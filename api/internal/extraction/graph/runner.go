package extraction

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/einarsundgren/sikta/internal/extraction/claude"
	"github.com/google/uuid"
)

// Document represents a document to be extracted
type Document struct {
	ID       string
	Filename string
	Content  string
}

// PromptConfig holds paths to prompt files
type PromptConfig struct {
	SystemPath  string // Path to system prompt file (e.g., prompts/system/v1.txt)
	FewshotPath string // Path to few-shot example file (e.g., prompts/fewshot/novel.txt)
}

// ExtractionResult is the complete extraction output for a corpus
type ExtractionResult struct {
	Corpus     string                  // Corpus identifier (e.g., "brf", "mna", "police")
	PromptVersion string               // Prompt version identifier (e.g., "v1")
	Documents  []DocumentExtraction    // Per-document extraction results
	Metadata   ExtractionMetadata     // Metadata about the extraction run
}

// DocumentExtraction holds extraction results for a single document
type DocumentExtraction struct {
	DocumentID string         // Document identifier (e.g., "A1", "B1")
	Filename   string         // Original filename
	Nodes      []ExtractedNode // Extracted nodes
	Edges      []ExtractedEdge // Extracted edges
	Error      string         // Error message if extraction failed
}

// ExtractionMetadata holds metadata about an extraction run
type ExtractionMetadata struct {
	Model        string    // Claude model used
	Timestamp    string    // ISO timestamp of extraction
	TotalDocs    int       // Total documents processed
	TotalNodes   int       // Total nodes extracted
	TotalEdges   int       // Total edges extracted
	FailedDocs   int       // Number of documents that failed
}

// Runner handles database-free extraction
type Runner struct {
	claude *claude.Client
	logger *slog.Logger
	model  string
}

// NewRunner creates a new extraction runner
func NewRunner(claude *claude.Client, logger *slog.Logger, model string) *Runner {
	return &Runner{
		claude: claude,
		logger: logger,
		model:  model,
	}
}

// RunExtraction processes a corpus without database, returning structured output
func (r *Runner) RunExtraction(ctx context.Context, docs []Document, prompt PromptConfig, corpus string) (*ExtractionResult, error) {
	r.logger.Info("starting database-free extraction", "corpus", corpus, "doc_count", len(docs))

	// Load prompts
	systemPrompt, err := os.ReadFile(prompt.SystemPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read system prompt from %s: %w", prompt.SystemPath, err)
	}

	fewshot, err := os.ReadFile(prompt.FewshotPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read few-shot prompt from %s: %w", prompt.FewshotPath, err)
	}

	result := &ExtractionResult{
		Corpus:        corpus,
		PromptVersion: filepath.Base(prompt.SystemPath),
		Documents:     make([]DocumentExtraction, 0, len(docs)),
		Metadata: ExtractionMetadata{
			Model:     r.model,
			Timestamp: "", // Will be set
		},
	}

	// Process each document
	for _, doc := range docs {
		docResult := r.extractFromDocument(ctx, doc, string(systemPrompt), string(fewshot))
		result.Documents = append(result.Documents, docResult)

		if docResult.Error != "" {
			result.Metadata.FailedDocs++
			r.logger.Error("document extraction failed", "doc_id", doc.ID, "error", docResult.Error)
		} else {
			result.Metadata.TotalNodes += len(docResult.Nodes)
			result.Metadata.TotalEdges += len(docResult.Edges)
		}
	}

	result.Metadata.TotalDocs = len(docs)
	r.logger.Info("extraction complete", "total_nodes", result.Metadata.TotalNodes, "total_edges", result.Metadata.TotalEdges, "failed", result.Metadata.FailedDocs)

	return result, nil
}

// extractFromDocument extracts nodes and edges from a single document
func (r *Runner) extractFromDocument(ctx context.Context, doc Document, systemPrompt, fewshot string) DocumentExtraction {
	r.logger.Info("extracting from document", "doc_id", doc.ID, "filename", doc.Filename)

	// Chunk the document (simple paragraph-based chunking)
	chunks := r.chunkDocument(doc.Content)

	docResult := DocumentExtraction{
		DocumentID: doc.ID,
		Filename:   doc.Filename,
		Nodes:      make([]ExtractedNode, 0),
		Edges:      make([]ExtractedEdge, 0),
	}

	// Process each chunk
	for i, chunk := range chunks {
		r.logger.Debug("processing chunk", "doc_id", doc.ID, "chunk", i, "length", len(chunk))

		nodes, edges, err := r.extractFromChunk(ctx, chunk, systemPrompt, fewshot)
		if err != nil {
			r.logger.Error("chunk extraction failed", "doc_id", doc.ID, "chunk", i, "error", err)
			// Continue with other chunks, don't fail entire document
			continue
		}

		// Add source document to node properties for tracking
		for j := range nodes {
			if nodes[j].Properties == nil {
				nodes[j].Properties = make(map[string]interface{})
			}
			nodes[j].Properties["source_doc"] = doc.ID
			nodes[j].Properties["source_file"] = doc.Filename
		}

		docResult.Nodes = append(docResult.Nodes, nodes...)
		docResult.Edges = append(docResult.Edges, edges...)
	}

	if len(docResult.Nodes) == 0 && len(docResult.Edges) == 0 {
		docResult.Error = "no nodes or edges extracted"
	}

	r.logger.Info("document extraction complete", "doc_id", doc.ID, "nodes", len(docResult.Nodes), "edges", len(docResult.Edges))

	return docResult
}

// extractFromChunk extracts nodes and edges from a single chunk using Claude
func (r *Runner) extractFromChunk(ctx context.Context, chunk, systemPrompt, fewshot string) ([]ExtractedNode, []ExtractedEdge, error) {
	// Build user message: few-shot example + chunk content
	userMessage := fmt.Sprintf("%s\n\n%s", fewshot, chunk)

	// Log request details
	chunkPreview := chunk
	if len(chunkPreview) > 200 {
		chunkPreview = chunkPreview[:200] + "..."
	}
	r.logger.Info("=== LLM REQUEST ===",
		"model", r.model,
		"system_prompt_length", len(systemPrompt),
		"user_message_length", len(userMessage),
		"fewshot_length", len(fewshot),
		"chunk_length", len(chunk),
		"chunk_preview", chunkPreview,
	)

	// Log full system prompt to stdout for inspection
	fmt.Fprintf(os.Stderr, "=== SYSTEM PROMPT ===\n%s\n=== END SYSTEM PROMPT ===\n\n", systemPrompt)

	// Log full user message to stdout for inspection
	fmt.Fprintf(os.Stderr, "=== USER MESSAGE ===\n%s\n=== END USER MESSAGE ===\n\n", userMessage)

	// Call Claude API
	apiResp, err := r.claude.SendSystemPrompt(ctx, systemPrompt, userMessage, r.model)
	if err != nil {
		r.logger.Error("LLM API call failed", "error", err)
		return nil, nil, fmt.Errorf("Claude API call failed: %w", err)
	}

	// Log response details
	if len(apiResp.Content) == 0 {
		r.logger.Error("Empty response from LLM")
		return nil, nil, fmt.Errorf("empty response from Claude")
	}

	rawResponse := apiResp.Content[0].Text
	r.logger.Info("=== LLM RESPONSE ===",
		"response_length", len(rawResponse),
		"input_tokens", apiResp.Usage.InputTokens,
		"output_tokens", apiResp.Usage.OutputTokens,
		"stop_reason", apiResp.StopReason,
	)

	// Log full raw response to stdout for inspection
	fmt.Fprintf(os.Stderr, "=== RAW RESPONSE ===\n%s\n=== END RAW RESPONSE ===\n\n", rawResponse)

	// Parse JSON response
	responseText := stripMarkdownCodeBlocks(rawResponse)

	// Fix trailing commas (common LLM output issue)
	responseText = fixTrailingCommas(responseText)

	// Log stripped response
	fmt.Fprintf(os.Stderr, "=== STRIPPED RESPONSE (markdown removed) ===\n%s\n=== END STRIPPED RESPONSE ===\n\n", responseText)

	var resp GraphExtractionResponse
	if err := json.Unmarshal([]byte(responseText), &resp); err != nil {
		// Fallback: try to extract JSON from markdown
		r.logger.Warn("JSON parse failed, attempting fallback extraction", "error", err)
		if jsonStr := extractJSONFromMarkdown(responseText); jsonStr != "" {
			responseText = jsonStr
			fmt.Fprintf(os.Stderr, "=== FALLBACK EXTRACTED JSON ===\n%s\n=== END FALLBACK JSON ===\n\n", responseText)
			if err := json.Unmarshal([]byte(responseText), &resp); err != nil {
				r.logger.Error("Fallback JSON parse also failed", "error", err)
				return nil, nil, fmt.Errorf("failed to parse JSON response: %w", err)
			}
		} else {
			r.logger.Error("JSON extraction failed", "parse_error", err, "response_preview", responseText[:min(200, len(responseText))])
			return nil, nil, fmt.Errorf("failed to parse JSON response: %w", err)
		}
	}

	r.logger.Info("=== EXTRACTION SUCCESS ===",
		"nodes_extracted", len(resp.Nodes),
		"edges_extracted", len(resp.Edges),
	)

	return resp.Nodes, resp.Edges, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// fixTrailingCommas removes trailing commas before } and ] in JSON.
// Common issue with LLM-generated JSON output.
func fixTrailingCommas(s string) string {
	// Remove trailing commas before closing braces/brackets
	// Handles: ,} ,] , } , ] and with whitespace/newlines between
	result := regexp.MustCompile(`,\s*([}\]])`).ReplaceAllString(s, "$1")
	return result
}

// chunkDocument splits a document into chunks based on paragraph boundaries
// Target: ~3000 words per chunk, max 4500, merge trailing chunks <1500
func (r *Runner) chunkDocument(content string) []string {
	// Split into paragraphs
	paragraphs := strings.Split(content, "\n\n")

	var chunks []string
	currentChunk := ""
	currentWords := 0
	const (
		targetWords = 3000
		maxWords     = 4500
		minWords     = 1500
	)

	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}

		// Estimate words (split by spaces)
		paraWords := len(strings.Fields(para))

		// If adding this paragraph would exceed max, flush current chunk
		if currentWords > 0 && currentWords+paraWords > maxWords {
			chunks = append(chunks, strings.TrimSpace(currentChunk))
			currentChunk = ""
			currentWords = 0
		}

		// Add paragraph to current chunk
		if currentChunk != "" {
			currentChunk += "\n\n"
		}
		currentChunk += para
		currentWords += paraWords
	}

	// Flush remaining chunk
	if currentChunk != "" {
		// Check if last chunk is too small and should be merged
		if len(chunks) > 0 && currentWords < minWords {
			// Merge with previous chunk
			prevChunk := chunks[len(chunks)-1]
			chunks[len(chunks)-1] = prevChunk + "\n\n" + currentChunk
		} else {
			chunks = append(chunks, strings.TrimSpace(currentChunk))
		}
	}

	return chunks
}

// ToJSON serializes the extraction result to JSON
func (r *ExtractionResult) ToJSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

// LoadDocumentsFromDirectory loads all .txt files from a directory
func LoadDocumentsFromDirectory(dirPath string) ([]Document, error) {
	var docs []Document

	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dirPath, err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".txt") {
			continue
		}

		filePath := filepath.Join(dirPath, file.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
		}

		// Extract document ID from filename (e.g., "A1-protocol-2023-03-15.txt" -> "A1")
		docID := strings.Split(file.Name(), "-")[0]

		docs = append(docs, Document{
			ID:       docID,
			Filename: file.Name(),
			Content:  string(content),
		})
	}

	return docs, nil
}

// GenerateUUID creates a new UUID for node/edge identification
func GenerateUUID() string {
	return uuid.New().String()
}
