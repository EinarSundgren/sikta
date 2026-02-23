package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
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
	case "view":
		runView(logger)
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
	fmt.Println("  sikta-eval view [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  extract  Run extraction on a corpus")
	fmt.Println("  score    Score extraction results against ground truth")
	fmt.Println("  compare  Compare two extraction results")
	fmt.Println("  view     View detailed score results (entities, events, false positives)")
	fmt.Println()
	fmt.Println("Extract Options:")
	fmt.Println("  --corpus PATH           Path to corpus directory (required)")
	fmt.Println("  --prompt PATH           Path to system prompt file (default: prompts/system/v1.txt)")
	fmt.Println("  --fewshot PATH          Path to few-shot example file (required)")
	fmt.Println("  --model MODEL           Claude model to use (default: claude-sonnet-4-20250514)")
	fmt.Println("  --output PATH           Output JSON file path (required)")
	fmt.Println("  --detect-inconsistencies  Run cross-document inconsistency detection")
	fmt.Println()
	fmt.Println("Environment:")
	fmt.Println("  ANTHROPIC_API_KEY  Required for extract and score --full commands")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  sikta-eval extract --corpus corpora/brf --prompt prompts/system/v5.txt --fewshot prompts/fewshot/brf-v4.txt --output results/brf-v5.json")
	fmt.Println("  sikta-eval extract --corpus corpora/brf --detect-inconsistencies --output results/brf-v5-inc.json")
	fmt.Println("  sikta-eval score --result results/brf-v5.json --manifest corpora/brf/manifest.json --full")
	fmt.Println("  sikta-eval view --score results/brf-v5-score.json")
	fmt.Println("  sikta-eval compare --a results/brf-v1.json --b results/brf-v2.json --manifest corpora/brf/manifest.json")
}

func runExtract(logger *slog.Logger) {
	flags := flag.NewFlagSet("extract", flag.ExitOnError)
	corpusDir := flags.String("corpus", "", "Path to corpus directory (e.g., corpora/brf)")
	systemPrompt := flags.String("prompt", "prompts/system/v1.txt", "Path to system prompt file")
	fewshotPrompt := flags.String("fewshot", "", "Path to few-shot example file (required)")
	model := flags.String("model", "claude-sonnet-4-20250514", "Claude model to use")
	outputPath := flags.String("output", "", "Output JSON file path (required)")
	detectInconsistencies := flags.Bool("detect-inconsistencies", false, "Run cross-document inconsistency detection after extraction")

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
	logger.Info("starting extraction", "corpus", corpusName, "model", *model, "detect_inconsistencies", *detectInconsistencies)
	startTime := time.Now()

	result, err := runner.RunExtractionWithOptions(context.Background(), docs, prompt, corpusName, *detectInconsistencies)
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
	if len(result.Inconsistencies) > 0 {
		fmt.Printf("  Inconsistencies detected: %d\n", len(result.Inconsistencies))
	}
	if result.Metadata.FailedDocs > 0 {
		fmt.Printf("  ⚠ Failed documents: %d\n", result.Metadata.FailedDocs)
	}
}

func runScore(logger *slog.Logger) {
	flags := flag.NewFlagSet("score", flag.ExitOnError)
	resultPath := flags.String("result", "", "Extraction result JSON file (required)")
	manifestPath := flags.String("manifest", "", "Ground truth manifest JSON file (required)")
	fullMode := flags.Bool("full", false, "Enable LLM-as-judge for unmatched event matching (requires API calls)")
	judgeModel := flags.String("model", "claude-haiku-3-5-20241022", "Model to use for LLM judge when --full is set")
	outputPath := flags.String("output", "", "Output file for detailed score results (JSON)")

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

	// Default output path to results directory
	if *outputPath == "" {
		base := filepath.Base(*resultPath)
		base = strings.TrimSuffix(base, ".json")
		*outputPath = filepath.Join("results", base+"-score.json")
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

	// Try to load as ExtractionResult (with Documents array)
	var resultExt evaluation.ExtractionResult
	if err := json.Unmarshal(resultData, &resultExt); err != nil {
		logger.Error("failed to parse result JSON", "error", err)
		os.Exit(1)
	}

	// Flatten to Extraction for scoring
	extraction := resultExt.Flatten()

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

	// Create scorer with optional judge
	var scorer *evaluation.Scorer
	if *fullMode {
		// Load API config for judge
		apiKey := os.Getenv("ANTHROPIC_API_KEY")
		apiURL := os.Getenv("ANTHROPIC_API_URL")

		if apiKey == "" {
			logger.Error("ANTHROPIC_API_KEY not configured (required for --full mode)")
			os.Exit(1)
		}

		cfg := &config.Config{
			AnthropicAPIKey: apiKey,
			AnthropicAPIURL: apiURL,
		}

		client := claude.NewClient(cfg, logger)
		judge := evaluation.NewEventJudge(client, logger, *judgeModel)
		scorer = evaluation.NewScorerWithJudge(&manifest, extraction, judge, logger)

		logger.Info("LLM judge enabled", "model", *judgeModel)
	} else {
		scorer = evaluation.NewScorer(&manifest, extraction)
	}

	// Run scorer
	score := scorer.ScoreWithContext(context.Background())

	// Save detailed JSON output
	jsonOutput, err := json.MarshalIndent(score, "", "  ")
	if err != nil {
		logger.Error("failed to serialize score results", "error", err)
		os.Exit(1)
	}

	if err := os.WriteFile(*outputPath, jsonOutput, 0644); err != nil {
		logger.Error("failed to write score results", "error", err)
		os.Exit(1)
	}

	logger.Info("score results saved", "path", *outputPath)

	// Print summary to terminal
	fmt.Println(evaluation.FormatTerminal(score))
	fmt.Printf("\nDetailed results saved to: %s\n", *outputPath)
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

func runView(logger *slog.Logger) {
	flags := flag.NewFlagSet("view", flag.ExitOnError)
	scorePath := flags.String("score", "", "Score result JSON file (required)")
	showMatches := flags.Bool("matches", true, "Show matched items")
	showUnmatched := flags.Bool("unmatched", true, "Show unmatched items")
	showFalsePositives := flags.Bool("fp", true, "Show false positives (hallucinations)")

	if err := flags.Parse(os.Args[2:]); err != nil {
		logger.Error("failed to parse flags", "error", err)
		os.Exit(1)
	}

	if *scorePath == "" {
		fmt.Println("Error: --score is required")
		os.Exit(1)
	}

	// Load score result
	scoreData, err := os.ReadFile(*scorePath)
	if err != nil {
		logger.Error("failed to read score file", "error", err)
		os.Exit(1)
	}

	var score evaluation.ScoreResult
	if err := json.Unmarshal(scoreData, &score); err != nil {
		logger.Error("failed to parse score JSON", "error", err)
		os.Exit(1)
	}

	// Print header
	fmt.Printf("=== Score Results: %s (prompt %s) ===\n\n", score.Corpus, score.PromptVersion)
	fmt.Printf("Timestamp: %s\n\n", score.Timestamp.Format("2006-01-02 15:04:05"))

	// Summary metrics
	fmt.Println("SUMMARY")
	fmt.Println("-------")
	fmt.Printf("Entity Recall:    %.1f%%\n", score.EntityRecall*100)
	fmt.Printf("Event Recall:     %.1f%%\n", score.EventRecall*100)
	fmt.Printf("False Positive:   %.1f%%\n", score.FalsePositiveRate*100)
	fmt.Println()

	if *showMatches {
		fmt.Println("=== ENTITY MATCHES ===")
		for _, m := range score.EntityDetails {
			if m.ManifestID != "" {
				if m.IsCorrect {
					fmt.Printf("✓ %s: \"%s\" → \"%s\" [%s]\n",
						m.ManifestID, m.ManifestLabel, m.MatchedNodeLabel, m.MatchMethod)
				}
			}
		}
		fmt.Println()

		fmt.Println("=== EVENT MATCHES ===")
		for _, m := range score.EventDetails {
			if m.ManifestID != "" {
				if m.IsCorrect {
					fmt.Printf("✓ %s: \"%s\" → \"%s\" [%s]\n",
						m.ManifestID, m.ManifestLabel, m.MatchedNodeLabel, m.MatchMethod)
				}
			}
		}
		fmt.Println()
	}

	if *showUnmatched {
		fmt.Println("=== UNMATCHED (NOT FOUND) ===")
		fmt.Println("Entities:")
		hasUnmatched := false
		for _, m := range score.EntityDetails {
			if m.ManifestID != "" && !m.IsCorrect {
				hasUnmatched = true
				fmt.Printf("  ✗ %s: \"%s\"\n", m.ManifestID, m.ManifestLabel)
			}
		}
		if !hasUnmatched {
			fmt.Println("  (all matched)")
		}

		fmt.Println("\nEvents:")
		hasUnmatched = false
		for _, m := range score.EventDetails {
			if m.ManifestID != "" && !m.IsCorrect {
				hasUnmatched = true
				fmt.Printf("  ✗ %s: \"%s\"\n", m.ManifestID, m.ManifestLabel)
			}
		}
		if !hasUnmatched {
			fmt.Println("  (all matched)")
		}
		fmt.Println()
	}

	if *showFalsePositives {
		fmt.Println("=== FALSE POSITIVES (HALLUCINATIONS) ===")
		fmt.Println("Entities:")
		entityFP := 0
		for _, m := range score.EntityDetails {
			if m.IsHallucination {
				entityFP++
				fmt.Printf("  - \"%s\"\n", m.MatchedNodeLabel)
			}
		}
		if entityFP == 0 {
			fmt.Println("  (none)")
		}

		fmt.Println("\nEvents:")
		eventFP := 0
		for _, m := range score.EventDetails {
			if m.IsHallucination {
				eventFP++
				fmt.Printf("  - \"%s\"\n", m.MatchedNodeLabel)
			}
		}
		if eventFP == 0 {
			fmt.Println("  (none)")
		}
	}
}
