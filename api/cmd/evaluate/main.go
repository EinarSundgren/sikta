package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"github.com/einarsundgren/sikta/internal/config"
	"github.com/einarsundgren/sikta/internal/extraction/claude"
	extraction "github.com/einarsundgren/sikta/internal/extraction/graph"
	"github.com/einarsundgren/sikta/internal/evaluation"
)

func main() {
	// Load .env file
	_ = godotenv.Load()

	// Setup logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "extract":
		runExtract(logger)
	case "score":
		runScore(logger)
	case "compare":
		runCompare(logger)
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("sikta-eval - Extraction Validation CLI")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  sikta-eval extract [options]")
	fmt.Println("  sikta-eval score [options]")
	fmt.Println("  sikta-eval compare [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  extract  Run extraction on a corpus")
	fmt.Println("  score    Score extraction results against ground truth")
	fmt.Println("  compare  Compare two extraction results")
	fmt.Println()
	fmt.Println("Environment:")
	fmt.Println("  ANTHROPIC_API_KEY  Required for extract command")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  sikta-eval extract --corpus corpora/brf --prompt prompts/system/v1.txt --fewshot prompts/fewshot/brf.txt --output results/brf-v1.json")
	fmt.Println("  sikta-eval score --result results/brf-v1.json --manifest corpora/brf/manifest.json")
	fmt.Println("  sikta-eval compare --a results/brf-v1.json --b results/brf-v2.json --manifest corpora/brf/manifest.json")
}

func runExtract(logger *slog.Logger) {
	flags := flag.NewFlagSet("extract", flag.ExitOnError)
	corpusDir := flags.String("corpus", "", "Path to corpus directory (e.g., corpora/brf)")
	systemPrompt := flags.String("prompt", "prompts/system/v1.txt", "Path to system prompt file")
	fewshotPrompt := flags.String("fewshot", "", "Path to few-shot example file (required)")
	model := flags.String("model", "claude-sonnet-4-20250514", "Claude model to use")
	outputPath := flags.String("output", "", "Output JSON file path (required)")

	if err := flags.Parse(os.Args[2:]); err != nil {
		logger.Error("failed to parse flags", "error", err)
		os.Exit(1)
	}

	// Validate required flags
	if *corpusDir == "" {
		fmt.Println("Error: --corpus is required")
		os.Exit(1)
	}
	if *fewshotPrompt == "" {
		fmt.Println("Error: --fewshot is required")
		os.Exit(1)
	}
	if *outputPath == "" {
		fmt.Println("Error: --output is required")
		os.Exit(1)
	}

	// Verify corpus directory exists
	docsDir := filepath.Join(*corpusDir, "docs")
	if _, err := os.Stat(docsDir); os.IsNotExist(err) {
		logger.Error("corpus docs directory not found", "path", docsDir)
		os.Exit(1)
	}

	// Verify prompt files exist
	if _, err := os.Stat(*systemPrompt); os.IsNotExist(err) {
		logger.Error("system prompt file not found", "path", *systemPrompt)
		os.Exit(1)
	}
	if _, err := os.Stat(*fewshotPrompt); os.IsNotExist(err) {
		logger.Error("few-shot prompt file not found", "path", *fewshotPrompt)
		os.Exit(1)
	}

	// Load config (minimal for extract command - no database needed)
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	apiURL := os.Getenv("ANTHROPIC_API_URL")

	if apiKey == "" {
		logger.Error("ANTHROPIC_API_KEY not configured")
		os.Exit(1)
	}

	// Create minimal config for client
	cfg := &config.Config{
		AnthropicAPIKey: apiKey,
		AnthropicAPIURL: apiURL,
	}

	// Create Claude client
	client := claude.NewClient(cfg, logger)

	// Create runner
	runner := extraction.NewRunner(client, logger, *model)

	// Load documents
	logger.Info("loading documents", "corpus", *corpusDir)
	docs, err := extraction.LoadDocumentsFromDirectory(docsDir)
	if err != nil {
		logger.Error("failed to load documents", "error", err)
		os.Exit(1)
	}
	logger.Info("loaded documents", "count", len(docs))

	// Extract corpus name from directory path
	corpusName := filepath.Base(filepath.Dir(*corpusDir))
	if corpusName == "corpora" {
		// Fallback: use parent directory name
		corpusName = filepath.Base(*corpusDir)
	}

	// Configure prompts
	prompt := extraction.PromptConfig{
		SystemPath:  *systemPrompt,
		FewshotPath: *fewshotPrompt,
	}

	// Run extraction
	logger.Info("starting extraction", "corpus", corpusName, "model", *model)
	startTime := time.Now()

	result, err := runner.RunExtraction(context.Background(), docs, prompt, corpusName)
	if err != nil {
		logger.Error("extraction failed", "error", err)
		os.Exit(1)
	}

	duration := time.Since(startTime)
	logger.Info("extraction complete", "duration", duration, "nodes", result.Metadata.TotalNodes, "edges", result.Metadata.TotalEdges)

	// Add timestamp to result
	result.Metadata.Timestamp = startTime.Format(time.RFC3339)

	// Serialize to JSON
	jsonOutput, err := result.ToJSON()
	if err != nil {
		logger.Error("failed to serialize results", "error", err)
		os.Exit(1)
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(*outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		logger.Error("failed to create output directory", "error", err)
		os.Exit(1)
	}

	// Write output
	if err := os.WriteFile(*outputPath, jsonOutput, 0644); err != nil {
		logger.Error("failed to write results", "error", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("✓ Extraction complete!\n")
	fmt.Printf("  Output: %s\n", *outputPath)
	fmt.Printf("  Nodes: %d, Edges: %d\n", result.Metadata.TotalNodes, result.Metadata.TotalEdges)
	if result.Metadata.FailedDocs > 0 {
		fmt.Printf("  ⚠ Failed documents: %d\n", result.Metadata.FailedDocs)
	}
}

func runScore(logger *slog.Logger) {
	flags := flag.NewFlagSet("score", flag.ExitOnError)
	resultPath := flags.String("result", "", "Extraction result JSON file (required)")
	manifestPath := flags.String("manifest", "", "Ground truth manifest JSON file (required)")
	_ = flags.Bool("full", false, "Enable LLM-as-judge for inconsistency scoring (requires API calls) -- reserved for EV6")

	if err := flags.Parse(os.Args[2:]); err != nil {
		logger.Error("failed to parse flags", "error", err)
		os.Exit(1)
	}

	// Validate required flags
	if *resultPath == "" {
		fmt.Println("Error: --result is required")
		os.Exit(1)
	}
	if *manifestPath == "" {
		fmt.Println("Error: --manifest is required")
		os.Exit(1)
	}

	// Verify files exist
	if _, err := os.Stat(*resultPath); os.IsNotExist(err) {
		logger.Error("result file not found", "path", *resultPath)
		os.Exit(1)
	}
	if _, err := os.Stat(*manifestPath); os.IsNotExist(err) {
		logger.Error("manifest file not found", "path", *manifestPath)
		os.Exit(1)
	}

	// Load extraction result
	resultData, err := os.ReadFile(*resultPath)
	if err != nil {
		logger.Error("failed to read result file", "error", err)
		os.Exit(1)
	}

	var extraction evaluation.Extraction
	if err := json.Unmarshal(resultData, &extraction); err != nil {
		logger.Error("failed to parse result JSON", "error", err)
		os.Exit(1)
	}

	// Load manifest
	manifestData, err := os.ReadFile(*manifestPath)
	if err != nil {
		logger.Error("failed to read manifest file", "error", err)
		os.Exit(1)
	}

	var manifest evaluation.Manifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		logger.Error("failed to parse manifest JSON", "error", err)
		os.Exit(1)
	}

	// Verify corpus matches
	if extraction.Corpus != manifest.Corpus {
		logger.Error("corpus mismatch", "result", extraction.Corpus, "manifest", manifest.Corpus)
		os.Exit(1)
	}

	// Run scorer
	scorer := evaluation.NewScorer(&manifest, &extraction)
	score := scorer.Score()

	// Print formatted results
	fmt.Println(evaluation.FormatTerminal(score))
}

func runCompare(logger *slog.Logger) {
	flags := flag.NewFlagSet("compare", flag.ExitOnError)
	resultAPath := flags.String("a", "", "First extraction result JSON file (required)")
	resultBPath := flags.String("b", "", "Second extraction result JSON file (required)")
	manifestPath := flags.String("manifest", "", "Ground truth manifest JSON file (required)")

	if err := flags.Parse(os.Args[2:]); err != nil {
		logger.Error("failed to parse flags", "error", err)
		os.Exit(1)
	}

	// Validate required flags
	if *resultAPath == "" {
		fmt.Println("Error: --a is required")
		os.Exit(1)
	}
	if *resultBPath == "" {
		fmt.Println("Error: --b is required")
		os.Exit(1)
	}
	if *manifestPath == "" {
		fmt.Println("Error: --manifest is required")
		os.Exit(1)
	}

	// Verify files exist
	if _, err := os.Stat(*resultAPath); os.IsNotExist(err) {
		logger.Error("result file A not found", "path", *resultAPath)
		os.Exit(1)
	}
	if _, err := os.Stat(*resultBPath); os.IsNotExist(err) {
		logger.Error("result file B not found", "path", *resultBPath)
		os.Exit(1)
	}
	if _, err := os.Stat(*manifestPath); os.IsNotExist(err) {
		logger.Error("manifest file not found", "path", *manifestPath)
		os.Exit(1)
	}

	// Load results
	resultAData, err := os.ReadFile(*resultAPath)
	if err != nil {
		logger.Error("failed to read result file A", "error", err)
		os.Exit(1)
	}
	var resultA evaluation.ScoreResult
	if err := json.Unmarshal(resultAData, &resultA); err != nil {
		logger.Error("failed to parse result JSON A", "error", err)
		os.Exit(1)
	}

	resultBData, err := os.ReadFile(*resultBPath)
	if err != nil {
		logger.Error("failed to read result file B", "error", err)
		os.Exit(1)
	}
	var resultB evaluation.ScoreResult
	if err := json.Unmarshal(resultBData, &resultB); err != nil {
		logger.Error("failed to parse result JSON B", "error", err)
		os.Exit(1)
	}

	// Load manifest
	manifestData, err := os.ReadFile(*manifestPath)
	if err != nil {
		logger.Error("failed to read manifest file", "error", err)
		os.Exit(1)
	}
	var manifest evaluation.Manifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		logger.Error("failed to parse manifest JSON", "error", err)
		os.Exit(1)
	}

	// Verify corpus matches
	if resultA.Corpus != resultB.Corpus || resultA.Corpus != manifest.Corpus {
		logger.Error("corpus mismatch", "resultA", resultA.Corpus, "resultB", resultB.Corpus, "manifest", manifest.Corpus)
		os.Exit(1)
	}

	// Compare results
	diff := evaluation.Compare(&resultA, &resultB)
	if diff == nil {
		logger.Error("failed to compare results")
		os.Exit(1)
	}

	// Print formatted diff
	fmt.Println(evaluation.FormatDiffTerminal(diff))
}
