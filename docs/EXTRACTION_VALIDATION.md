# Sikta — Project Course Correction: Extraction Validation Phase

> Date: February 2026
> Status: Active — this supersedes the original Phase 0-7 implementation plan in TASKS.md

---

## What Changed and Why

The original plan was to build the full stack (Go backend → PostgreSQL → React frontend) across seven phases, starting with project scaffolding and ending with a polished demo. Three rounds of external review converged on the same conclusion:

**Extraction quality is the existential risk.** If the LLM can't reliably extract structured claims from documents, nothing else matters — not the timeline UI, not the relationship graph, not the review workflow. Building 21-28 hours of full-stack product before validating extraction is building on sand.

The project now enters an **Extraction Validation Phase** before any frontend work begins. The goal is a CLI-driven prompt development and testing pipeline that measures extraction accuracy against known ground truth.

### What this phase IS:

- A Go CLI tool that runs extraction prompts against test documents
- Machine-readable ground truth manifests for three test corpora
- Automated scoring: entity recall, event recall, inconsistency detection rate, false positive rate
- A prompt iteration workflow: change prompt → run extraction → measure improvement
- Collaboration between the developer and AI to refine prompts based on scored results

### What this phase is NOT:

- No React frontend. No UI of any kind.
- No database required for the validation pipeline (extraction output is JSON files)
- No deployment, no podman-compose, no infrastructure
- Not a rewrite — builds on the existing extraction package

---

## Existing Code Assessment

The current codebase has a working extraction pipeline under `api/internal/extraction/graph/`:

**`api/internal/extraction/graph/service.go`** — Orchestrates chunk-level extraction via Claude API, stores results as nodes/edges/provenance in PostgreSQL. Well-structured. The `extractFromChunk` method is the core: sends a system prompt + chunk content, parses JSON response into `ExtractedNode` and `ExtractedEdge` structs.

**`api/internal/extraction/graph/types.go`** — `ExtractedNode` and `ExtractedEdge` already align with the three-primitive model. Nodes carry `modality`, `confidence`, temporal claims (`claimed_time_*`), and spatial claims (`claimed_geo_*`). Edges carry `is_negated`, `modality`, `confidence`. This is the right structure.

**`api/internal/extraction/graph/prompts.go`** — Current prompt is a reasonable starting point. Has node types, edge types, modality taxonomy, confidence guidelines, temporal/spatial extraction instructions, and a few-shot example from Pride and Prejudice.

**`api/internal/extraction/claude/client.go`** — Clean Claude API client with retry logic, system prompts, and JSON parsing. Uses `SendSystemPrompt(ctx, systemPrompt, userMessage, model)`. Ready to use as-is.

**Note:** There is also a legacy extraction pipeline at `api/internal/extraction/service.go` and `api/internal/extraction/prompts.go` (non-graph). This is left untouched — the validation phase works exclusively with the `graph/` subpackage.

### What's missing for the validation phase:

1. **CLI entry point** — No `api/cmd/evaluate/` or equivalent. Extraction currently runs only through the web service handler.
2. **Ground truth manifests** — No machine-readable expected output to compare against.
3. **Scoring logic** — No way to measure extraction quality.
4. **Prompt versioning** — Prompts are hardcoded constants in `api/internal/extraction/graph/prompts.go`. Need to be loadable from files.
5. **Database-free execution** — Current extraction in `graph/service.go` requires PostgreSQL (stores via `database.Queries` and `graph.Service`). The validation pipeline should work without a database — extract to JSON, score against manifests.
6. **Multi-document awareness** — Current pipeline processes one document at a time. The test corpora (BRF, M&A, Police) are multi-document sets where cross-document inconsistencies are the key test.

---

## What Needs to Be Built

### 1. CLI Tool: `cmd/evaluate/main.go`

A standalone CLI with three subcommands:

```bash
# Extract: run a prompt against a corpus, output JSON
# (run from project root — binary built from api/cmd/evaluate/)
sikta-eval extract \
  --corpus corpora/brf \
  --prompt prompts/system/v1.txt \
  --fewshot prompts/fewshot/brf.txt \
  --model claude-sonnet-4-20250514 \
  --output results/brf-v1.json

# Score: compare extraction output against ground truth (fast, no LLM)
sikta-eval score \
  --result results/brf-v1.json \
  --manifest corpora/brf/manifest.json

# Full evaluation: score + LLM-as-judge for inconsistencies (slower)
sikta-eval score \
  --result results/brf-v1.json \
  --manifest corpora/brf/manifest.json \
  --full

# Compare: diff two prompt versions
sikta-eval compare \
  --a results/brf-v1.json \
  --b results/brf-v3.json \
  --manifest corpora/brf/manifest.json
```

Add a Makefile target for convenience:

```makefile
# In Makefile (project root)
eval-build:
	cd api && go build -o ../sikta-eval ./cmd/evaluate/

eval-brf:
	./sikta-eval extract --corpus corpora/brf --prompt prompts/system/v1.txt --output results/brf-v1.json
	./sikta-eval score --result results/brf-v1.json --manifest corpora/brf/manifest.json

eval-all:
	./sikta-eval extract --corpus corpora/brf --prompt prompts/system/v1.txt --output results/brf-v1.json
	./sikta-eval extract --corpus corpora/mna --prompt prompts/system/v1.txt --output results/mna-v1.json
	./sikta-eval extract --corpus corpora/police --prompt prompts/system/v1.txt --output results/police-v1.json
	./sikta-eval score --result results/brf-v1.json --manifest corpora/brf/manifest.json
	./sikta-eval score --result results/mna-v1.json --manifest corpora/mna/manifest.json
	./sikta-eval score --result results/police-v1.json --manifest corpora/police/manifest.json
```

The CLI reuses the existing `claude.Client` for API calls but does NOT touch PostgreSQL. Extraction output is pure JSON.

### 2. Extraction Runner (Database-Free)

A new function that wraps the existing extraction logic but outputs JSON instead of writing to the database:

```go
// api/internal/extraction/graph/runner.go

// RunExtraction processes a corpus without database, returns structured output
func RunExtraction(ctx context.Context, client *claude.Client, docs []Document, prompt PromptConfig, model string) (*ExtractionResult, error)
```

This reads documents from disk, chunks them, sends each chunk to Claude with the configured prompt, and collects all extracted nodes/edges into a single `ExtractionResult` struct. The existing `extractFromChunk` logic from `graph/service.go` is reused — just the storage layer is bypassed.

For multi-document corpora: each document is processed separately, but the output is merged into one result with source document tracking. Cross-document entity deduplication and inconsistency detection run as a post-processing pass over the merged result.

### 3. Ground Truth Manifests (JSON)

Machine-readable versions of the ground truth from the three test corpora. Structure:

```json
{
  "corpus": "brf",
  "description": "BRF Stenbacken 3 — Protocols and invoices",
  "documents": [
    { "id": "A1", "filename": "protocol-2023-03-15.txt", "type": "protocol", "trust": 0.95 },
    { "id": "A2", "filename": "offert-norrbygg.txt", "type": "offert", "trust": 0.90 }
  ],
  "entities": [
    {
      "id": "E1",
      "label": "Anna Lindqvist",
      "type": "person",
      "aliases": ["Lindqvist", "ordföranden", "Anna"],
      "properties": { "role": "ordförande" },
      "mentioned_in": ["A1", "A2", "A3", "A5"]
    }
  ],
  "events": [
    {
      "id": "V1",
      "label": "Fuktinspektion",
      "type": "inspection",
      "claimed_time_text": "12 januari 2023",
      "claimed_time_start": "2023-01-12",
      "entities": ["E12"],
      "source_doc": "A1",
      "source_section": "§5"
    }
  ],
  "inconsistencies": [
    {
      "id": "I1",
      "type": "amount",
      "severity": "high",
      "description": "Budget set to 650k in A1 §5 but lowered to 600k in A3 §5 without explanation",
      "documents": ["A1", "A3"],
      "entities_involved": ["E8"],
      "evidence": {
        "side_a": { "doc": "A1", "section": "§5", "claim": "inte överstiga 650 000 kr" },
        "side_b": { "doc": "A3", "section": "§5", "claim": "maximalt 600 000 kr" }
      }
    }
  ]
}
```

Entity matching uses the `aliases` array for fuzzy matching. Event matching uses entity involvement + temporal proximity + source document. Inconsistency matching uses LLM-as-judge.

### 4. Scoring Engine

```go
// api/internal/evaluation/scorer.go

type ScoreResult struct {
    // Entity scores
    EntityRecall     float64  // found / expected
    EntityPrecision  float64  // correct / found
    EntityF1         float64
    EntityDetails    []EntityMatch  // per-entity breakdown

    // Event scores
    EventRecall      float64
    EventPrecision   float64
    EventF1          float64
    EventDetails     []EventMatch

    // Inconsistency scores (only in --full mode)
    InconsistencyRecall    float64
    InconsistencyDetails   []InconsistencyMatch

    // Quality metrics
    FalsePositiveRate      float64
    AvgConfidenceAccuracy  float64  // how well confidence scores predict correctness
}
```

**Entity matching logic:**
1. Normalize both labels: lowercase, trim, remove titles (Mr./Dr./etc.)
2. Check exact match against label or any alias
3. If no exact match: Levenshtein distance ≤ 3 against label and all aliases
4. If still no match: check if extracted label is a substring of any alias or vice versa
5. A manifest entity is "found" if any extracted node matches it
6. An extracted node is a "false positive" if it matches no manifest entity

**Event matching logic:**
1. Event matches if it: references at least one correct entity AND comes from the correct source document AND temporal claims overlap (if both specify time)
2. Fuzzy: event label similarity is secondary — the structural match (entities + source + time) is primary

**Inconsistency matching (LLM-as-judge, --full mode only):**
Send to Claude (Haiku is fine for judging):
```
Here is a planted inconsistency from the ground truth:
[description from manifest]

Here are the inconsistencies detected by the extraction:
[list of detected inconsistencies]

Does any detected inconsistency correspond to this planted one?
Answer with JSON: {"match": true/false, "matched_id": "...", "reasoning": "..."}
```

### 5. Prompt Files (Versionable)

Move prompts from hardcoded constants to files:

```
prompts/
├── system/
│   ├── v1.txt              # Current prompt from prompts.go
│   ├── v2.txt              # First iteration
│   └── ...
├── fewshot/
│   ├── novel.txt           # Pride and Prejudice example (current)
│   ├── brf.txt             # BRF protocol example
│   ├── mna.txt             # M&A document example
│   └── police.txt          # Police report example
└── postprocess/
    ├── dedup.txt            # Entity deduplication prompt
    └── inconsistency.txt   # Cross-document inconsistency detection prompt
```

The `--prompt` flag to the CLI specifies the system prompt file. Few-shot examples are selected automatically based on the corpus type, or overridden with `--fewshot`.

### 6. Post-Processing Prompts

Two additional LLM passes after per-chunk extraction:

**Entity deduplication:** After all chunks are processed, collect all extracted entity nodes. Send to Claude: "Here are all extracted entities. Identify which ones likely refer to the same real-world entity. Output same_as edges with confidence scores."

**Cross-document inconsistency detection:** After all documents in a corpus are processed, send the merged entity/event/edge list to Claude: "Here are all claims extracted from multiple documents. Identify contradictions, amount discrepancies, temporal impossibilities, and obligation conflicts."

These are separate prompt files so they can be iterated independently from the per-chunk extraction prompt.

---

## File Structure (New / Modified)

```
sikta/
├── CLAUDE.md                        # Existing — project entry point for AI agents
├── Makefile                         # Existing — add eval targets
├── podman-compose.yaml              # Existing — untouched
├── .env.example                     # Existing — untouched
│
├── docs/
│   ├── STATE.md                     # Existing
│   ├── TASKS.md                     # Existing
│   ├── DESIGN_SYSTEM.md             # Existing
│   └── EXTRACTION_VALIDATION.md     # THIS DOCUMENT
│
├── api/                             # Existing Go backend
│   ├── go.mod
│   ├── go.sum
│   ├── sqlc.yaml
│   ├── cmd/
│   │   ├── server/                  # Existing HTTP server — untouched
│   │   └── evaluate/                # NEW — extraction validation CLI
│   │       └── main.go
│   ├── internal/
│   │   ├── config/                  # Existing — untouched
│   │   ├── database/               # Existing — untouched
│   │   │   ├── models.go
│   │   │   └── graph.go
│   │   ├── handlers/               # Existing — untouched
│   │   │   ├── documents.go
│   │   │   ├── graph/
│   │   │   │   ├── review.go
│   │   │   │   └── timeline.go
│   │   │   └── inconsistencies.go
│   │   ├── graph/                   # Existing graph service — untouched
│   │   │   ├── service.go
│   │   │   ├── migrator.go
│   │   │   ├── types.go
│   │   │   └── views.go
│   │   ├── extraction/
│   │   │   ├── service.go           # Existing legacy orchestrator — untouched
│   │   │   ├── prompts.go           # Existing legacy prompts — untouched
│   │   │   ├── claude/
│   │   │   │   └── client.go        # Existing — reused by validation CLI
│   │   │   └── graph/
│   │   │       ├── service.go       # Existing graph extraction — untouched
│   │   │       ├── prompts.go       # Existing graph prompts — kept as fallback
│   │   │       ├── types.go         # Existing — may extend
│   │   │       └── runner.go        # NEW — database-free extraction runner
│   │   ├── evaluation/              # NEW — scoring and comparison
│   │   │   ├── scorer.go
│   │   │   ├── matcher.go
│   │   │   ├── judge.go
│   │   │   ├── compare.go
│   │   │   └── types.go
│   │   └── services/                # Existing — untouched
│   └── sql/
│       ├── schema/                  # Existing — untouched
│       └── queries/                 # Existing — untouched
│
├── web/                             # Existing React frontend — UNTOUCHED this phase
│   ├── index.html
│   ├── package.json
│   ├── vite.config.ts
│   ├── tailwind.config.ts
│   └── src/
│       ├── main.tsx
│       ├── App.tsx
│       ├── types/
│       ├── api/
│       ├── pages/
│       └── components/
│           ├── timeline/
│           ├── entities/
│           ├── graph/
│           ├── review/
│           ├── inconsistencies/
│           └── source/
│
├── demo/                            # Existing — untouched
│   └── seed.sql
│
├── corpora/                         # NEW — test corpora with ground truth
│   ├── brf/
│   │   ├── docs/
│   │   │   ├── A1-protocol-2023-03-15.txt
│   │   │   ├── A2-offert-norrbygg.txt
│   │   │   ├── A3-protocol-2023-05-10.txt
│   │   │   ├── A4-faktura-norrbygg.txt
│   │   │   └── A5-protocol-2023-12-07.txt
│   │   └── manifest.json
│   ├── mna/
│   │   ├── docs/
│   │   │   ├── B1-management-presentation.txt
│   │   │   ├── B2-financial-statements.txt
│   │   │   ├── B3-employment-agreement.txt
│   │   │   ├── B4-ip-assignment.txt
│   │   │   ├── B5-board-minutes.txt
│   │   │   └── B6-customer-contract.txt
│   │   └── manifest.json
│   └── police/
│       ├── docs/
│       │   ├── C1-polisrapport.txt
│       │   ├── C2-rattsmedicinskt.txt
│       │   ├── C3-forhor-samira.txt
│       │   ├── C4-forhor-viktor.txt
│       │   ├── C5-forhor-nadia.txt
│       │   ├── C6-kameraovervakning.txt
│       │   └── C7-telefonlistor.txt
│       └── manifest.json
│
├── prompts/                         # NEW — versionable prompt files
│   ├── system/
│   │   └── v1.txt
│   ├── fewshot/
│   │   ├── novel.txt
│   │   ├── brf.txt
│   │   ├── mna.txt
│   │   └── police.txt
│   └── postprocess/
│       ├── dedup.txt
│       └── inconsistency.txt
│
└── results/                         # NEW — timestamped extraction outputs
    └── .gitkeep
```

### Notes on structure

- All new Go code lives under `api/` alongside the existing module (`api/go.mod`)
- The CLI at `api/cmd/evaluate/` is a sibling of `api/cmd/server/`, sharing the same module and internal packages
- `corpora/`, `prompts/`, and `results/` are at the project root (not inside `api/`) because they're data, not code — and may be used by future tooling outside Go
- `results/` is gitignored except for `.gitkeep` — extraction outputs are ephemeral
- Nothing in `web/` is touched during this phase

---

## Development Tasks

### Task 1: Corpus Preparation

Convert the three test corpora (already written as markdown with embedded documents) into the file structure above. Extract each document into its own `.txt` file. Convert the ground truth manifests into `manifest.json` files with the schema defined in section 3.

**Estimate:** 2-3 hours (mostly mechanical conversion + careful alias list definition)

### Task 2: Prompt File Extraction

Extract the current `GraphExtractionSystemPrompt` and `GraphFewShotExample` from `prompts.go` into `prompts/system/v1.txt` and `prompts/fewshot/novel.txt`. Write domain-specific few-shot examples for BRF, M&A, and police corpora.

**Estimate:** 2-3 hours (the BRF/M&A/police examples require careful construction)

### Task 3: Database-Free Runner

Create `api/internal/extraction/graph/runner.go` that:
- Reads documents from a directory
- Chunks them (reuse existing chunker or simple section-based splitting for non-novel documents)
- Calls Claude API via existing `claude.Client`
- Loads prompt from file instead of hardcoded constant
- Returns `ExtractionResult` as a Go struct (serializable to JSON)

**Estimate:** 3-4 hours

### Task 4: Scoring Engine

Create `api/internal/evaluation/` package with:
- Entity matcher (normalized + Levenshtein + alias lookup)
- Event matcher (entity involvement + source document + temporal overlap)
- Score calculator (precision, recall, F1, false positive rate)
- JSON output of detailed per-item match results

**Estimate:** 4-5 hours

### Task 5: CLI Entry Point

Create `api/cmd/evaluate/main.go` with subcommands:
- `extract` — run extraction on a corpus
- `score` — compare against ground truth
- `compare` — diff two prompt versions

Use standard library `flag` or simple arg parsing. No framework needed.

**Estimate:** 2-3 hours

### Task 6: LLM-as-Judge (for --full mode)

Create `api/internal/evaluation/judge.go` that:
- Takes detected inconsistencies + manifest inconsistencies
- Sends structured comparison prompts to Claude (Haiku)
- Parses match/no-match responses
- Integrates into the scoring pipeline as an optional step

**Estimate:** 2-3 hours

### Task 7: Post-Processing Prompts

Create the entity deduplication and cross-document inconsistency detection prompts. These run after per-chunk extraction and before scoring:
- `dedup.txt` — takes all extracted entities, outputs same_as edges
- `inconsistency.txt` — takes all claims across documents, outputs detected contradictions

Integrate into the extraction runner as optional pipeline stages.

**Estimate:** 3-4 hours (prompt design is the hard part, not the code)

### Task 8: Prompt Iteration (Ongoing)

With the infrastructure from Tasks 1-7 in place:
1. Run `sikta-eval extract` on all three corpora with v1 prompt
2. Run `sikta-eval score` to get baseline numbers
3. Analyze failures: which entities missed? which inconsistencies not found?
4. Revise prompt → v2 → re-run → re-score
5. Repeat until thresholds met

**Go/kill thresholds:**
- ≥85% entity recall across all three corpora
- ≥70% event recall across all three corpora
- ≥50% inconsistency detection rate
- <20% false positive rate
- If these cannot be reached after 10 prompt iterations → reassess approach

---

## Prompt Iteration Workflow

The developer and AI collaborate on prompt refinement:

```
1. Developer runs: sikta-eval extract --corpus corpora/brf --prompt prompts/system/v3.txt
2. Developer runs: sikta-eval score --result results/brf-v3.json --manifest corpora/brf/manifest.json
3. Developer shares the score output + a sample of missed/wrong extractions with AI
4. AI analyzes failures, proposes prompt changes, writes prompts/system/v4.txt
5. Developer runs v4 → scores → shares results
6. Repeat
```

The AI never runs the extraction — the developer does. The AI sees the scores and the failure cases, then proposes prompt refinements. This keeps the human in control of execution and cost while leveraging AI for prompt design.

When reviewing failures, focus on:
- **Missed entities:** Is the prompt too restrictive on entity types? Does the few-shot example not cover this pattern?
- **Missed events:** Are events buried in passive voice or indirect speech that the prompt doesn't instruct for?
- **Missed inconsistencies:** Does the cross-document prompt have enough context? Does it understand what constitutes a contradiction in this domain?
- **False positives:** Is the prompt hallucinating entities or relationships not in the text?
- **Confidence calibration:** Are high-confidence extractions actually correct more often than low-confidence ones?

---

## What Remains Unchanged

- All existing project documentation (CLAUDE.md, TASKS.md, STATE.md, UPDATES.md, DESIGN_SYSTEM.md, SIKTA_DATA_MODEL.md, STRATEGY.md) remains valid
- The three-primitive data model (Node/Edge/Provenance) is the extraction target
- The dual-frontend architecture (Explore/Evidence) is deferred, not cancelled
- The original Phase 0-7 plan in TASKS.md becomes the post-validation implementation plan
- The existing `service.go` pipeline continues to work for the web application path
- No UI work until extraction validation passes go/kill thresholds

---

## Success Criteria for This Phase

| Metric | Target | Measured By |
|--------|--------|-------------|
| Entity recall (avg across 3 corpora) | ≥85% | `sikta-eval score` |
| Event recall (avg across 3 corpora) | ≥70% | `sikta-eval score` |
| Inconsistency detection (avg) | ≥50% | `sikta-eval score --full` |
| False positive rate | <20% | `sikta-eval score` |
| Confidence calibration | High-conf items correct >90% of the time | `sikta-eval score` |
| Prompt iterations to reach thresholds | ≤10 | Count |

**Timeline estimate:** 2-3 weeks for infrastructure (Tasks 1-7), then 1-2 weeks of prompt iteration (Task 8). Total: 3-5 weeks part-time.

**If thresholds are not met:** Document what fails and why. The failures inform whether the approach needs fundamental redesign (different extraction strategy) or just more prompt engineering. Do not proceed to UI development.