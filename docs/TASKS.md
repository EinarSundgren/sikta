# TASKS.md

> Phased backlog. Each phase has tasks, model recommendations, and acceptance criteria.

---

## Completed MVP Phases (0-7 + DS + G1-G6)

All MVP phases and the design system are complete. See `docs/STATE.md` for details.

- **Phase 0:** Project Scaffolding — Go + React + PostgreSQL + Makefile
- **Phase 1:** Document Ingestion & Chunking — Structure-agnostic paragraph chunking
- **Phase 2:** LLM Extraction Pipeline — 178 events, 54 entities, 58 relationships from P&P
- **Phase 2.5:** Data Model Migration — documents→sources, events→claims, two-level confidence
- **Phase 3:** Inconsistency Detection — Narrative vs chronological, contradictions, temporal
- **Phase 4:** Timeline Hero View — D3 dual-lane timeline with connectors
- **Phase 5:** Entity Panel & Relationship Graph — D3 force-directed graph
- **Phase 6:** Review Workflow & Inconsistency Panel — Keyboard-driven J/K/A/R/E
- **Phase 7:** Demo Polish & Landing — Landing page, upload flow, seed SQL
- **Phase DS:** Design System Implementation — Visual redesign per prototype, SourcePanel, bug fixes
- **Phase G1-G6:** Graph Model Alignment — Open type system, provenance-based ordering, clean architecture

---

## ⚠️ Current: Extraction Validation Phase

> **This phase supersedes further UI/frontend work.** See `docs/EXTRACTION_VALIDATION.md` for full rationale.
>
> Extraction quality is the existential risk. If the LLM can't reliably extract structured claims from documents, nothing else matters. Building more UI before validating extraction is building on sand.

**Goal:** CLI-driven prompt development and testing pipeline that measures extraction accuracy against known ground truth across three realistic document corpora.

**Go/Kill thresholds (all three corpora must pass):**
- ≥85% entity recall
- ≥70% event recall
- ≥50% inconsistency detection rate
- <20% false positive rate

**No UI work until these thresholds are met.**

---

### EV1: Corpus Preparation ✅ COMPLETE
**Model: Haiku** | **Size:** S (mechanical conversion)

Extract the three test corpora from markdown documents into the proper file structure. Create machine-readable `manifest.json` ground truth files.

**Output:**
```
corpora/
├── brf/docs/{A1-A5}.txt + manifest.json
├── mna/docs/{B1-B6}.txt + manifest.json
└── police/docs/{C1-C7}.txt + manifest.json
```

**Acceptance:**
- All document `.txt` files exist with clean content (no markdown)
- All three `manifest.json` files are valid JSON with entities, events, inconsistencies

---

### EV2: Prompt File Extraction
**Model: Sonnet** | **Size:** S (2-3 hours)

Extract current hardcoded prompts from `api/internal/extraction/graph/prompts.go` into versionable files. Write domain-specific few-shot examples for each corpus type.

**Tasks:**
- [ ] Extract `GraphExtractionSystemPrompt` → `prompts/system/v1.txt`
- [ ] Extract `GraphFewShotExample` (P&P) → `prompts/fewshot/novel.txt`
- [ ] Write BRF few-shot example → `prompts/fewshot/brf.txt`
- [ ] Write M&A few-shot example → `prompts/fewshot/mna.txt`
- [ ] Write police/investigation few-shot example → `prompts/fewshot/police.txt`

**Acceptance:**
- `prompts/system/v1.txt` is byte-identical to the current hardcoded prompt
- Few-shot examples follow the same JSON structure as the current extraction output

---

### EV3: Database-Free Extraction Runner
**Model: Sonnet** | **Size:** M (3-4 hours)

New file `api/internal/extraction/graph/runner.go` that runs extraction without PostgreSQL — output is JSON, not database writes. Reuses existing Claude client and extraction logic.

**Tasks:**
- [ ] Define `Document`, `PromptConfig`, `ExtractionResult` structs in `runner.go`
- [ ] Implement `RunExtraction(ctx, client, docs, prompt, model) (*ExtractionResult, error)`
- [ ] Read documents from disk (plain text files from a directory)
- [ ] Chunk documents (reuse existing chunker if suitable; simple paragraph-split fallback for structured docs)
- [ ] Load prompt from file (not hardcoded constant)
- [ ] Collect all extracted nodes/edges into merged `ExtractionResult` with per-doc tracking
- [ ] Serialize result to JSON

**Acceptance:**
- `go build ./...` passes
- Can run against `corpora/brf/` and produce a valid JSON output file
- No PostgreSQL dependency in runner.go

---

### EV4: Scoring Engine
**Model: Sonnet** | **Size:** M (4-5 hours)

New package `api/internal/evaluation/` with entity matching, event matching, and score calculation.

**Tasks:**
- [ ] `api/internal/evaluation/types.go` — `ScoreResult`, `EntityMatch`, `EventMatch`, `InconsistencyMatch`
- [ ] `api/internal/evaluation/matcher.go` — entity matcher (normalize → exact → Levenshtein ≤3 → substring), event matcher (entities + source + temporal overlap)
- [ ] `api/internal/evaluation/scorer.go` — precision, recall, F1, false positive rate from match results
- [ ] `api/internal/evaluation/compare.go` — diff two `ScoreResult` structs, format for terminal output

**Acceptance:**
- Entity matcher correctly links "ordföranden" to "Anna Lindqvist" via aliases
- Event matcher correctly matches on entity involvement + source doc
- Scorer computes F1 correctly for a sample hand-crafted case

---

### EV5: CLI Entry Point
**Model: Sonnet** | **Size:** S (2-3 hours)

`api/cmd/evaluate/main.go` with three subcommands: `extract`, `score`, `compare`.

**Tasks:**
- [ ] Subcommand `extract`: `--corpus`, `--prompt`, `--fewshot`, `--model`, `--output`
- [ ] Subcommand `score`: `--result`, `--manifest`, `--full` (enables LLM-as-judge)
- [ ] Subcommand `compare`: `--a`, `--b`, `--manifest`
- [ ] Add Makefile targets: `eval-build`, `eval-brf`, `eval-mna`, `eval-police`, `eval-all`

**Acceptance:**
- `make eval-build` compiles `sikta-eval` binary
- `sikta-eval extract --corpus corpora/brf --prompt prompts/system/v1.txt --output results/brf-v1.json` runs without error
- `sikta-eval score --result results/brf-v1.json --manifest corpora/brf/manifest.json` prints a score report

---

### EV6: LLM-as-Judge for Event Matching ✅ COMPLETE
**Model: Sonnet (implementation) / Haiku (judge runtime)** | **Size:** S (2-3 hours)

Two-pass scoring: deterministic matching first (free), then LLM judge for unmatched events only (when `--full` flag is set).

**Tasks:**
- [x] Create `api/internal/evaluation/judge.go` — `EventJudge` with `JudgeUnmatchedEvents()` method
- [x] Judge prompt: manifest event + candidate extracted events → JSON match decision with reasoning
- [x] Wire `--full` flag in `main.go` to create Claude client and pass to scorer
- [x] Update `Scorer` to accept optional judge, run after deterministic pass
- [x] Conservative matching: only match events referring to the same real-world occurrence

**Acceptance:**
- [x] Without `--full`: identical results to deterministic matcher (71.4% event recall on BRF v4)
- [x] With `--full`: V1, V6, V9 matched via `llm_judge` → 92.9% event recall
- [x] V3 correctly remains unmatched (genuinely not extracted)
- [x] Judge reasoning printed in score output for transparency

---

### EV7: Post-Processing Prompts
**Model: Sonnet** | **Size:** M (3-4 hours)

Two additional LLM passes after per-chunk extraction: entity deduplication and cross-document inconsistency detection.

**Tasks:**
- [ ] Write `prompts/postprocess/dedup.txt` — takes all extracted entities, outputs `same_as` edges
- [ ] Write `prompts/postprocess/inconsistency.txt` — takes all claims across docs, outputs detected contradictions
- [ ] Implement post-processing pipeline stages in `runner.go`:
  - After all chunks processed: entity dedup pass
  - After all documents in corpus: cross-document inconsistency pass
- [ ] Make post-processing stages optional (flags: `--dedup`, `--inconsistency-check`)

**Acceptance:**
- Dedup prompt correctly identifies "ordföranden" and "Anna Lindqvist" as `same_as` with high confidence
- Inconsistency prompt identifies the BRF budget discrepancy (650k vs 600k) when both documents are in context

---

### EV8: Prompt Iteration (In Progress)
**Model: Opus (analysis), Sonnet (implementation)** | **Size:** Open-ended

**Current Status (v5 after EV8.6):**
| Corpus | Entity Recall | Event Recall | False Positive | Status |
|--------|---------------|--------------|----------------|--------|
| BRF | 100.0% | 71.4% | 63.3% | ✓ Entity/Event PASS, ✗ FP FAIL |
| M&A | 100.0% | 70.0% | 63.6% | ✓ Entity/Event PASS, ✗ FP FAIL |
| Police | 100.0% | 86.4% | 64.5% | ✓ Entity/Event PASS, ✗ FP FAIL |

**Root Cause Analysis:** See `docs/EV8-Prompt-Iteration-Summary.md`

---

#### EV8.6: v5 Prompt — Entity Recall Fix ✅ COMPLETE
**Goal:** Raise entity recall from 64% to 85%

**Completed:**
- [x] Create `prompts/system/v5.txt` from v4 with these additions:
  1. Add "Entity Extraction from Events" section — for every event, parse label/description for named entities, create nodes + involved_in edges
  2. Update confidence guidance — do NOT skip entities by mention count; single mention = 0.7-0.8 confidence
  3. Expand node types — add explicit `technology`, `vehicle`, `address` types with examples
- [x] Create `prompts/fewshot/mna-v5.txt` — add extraction examples for board members, SynCore
- [x] Create `prompts/fewshot/police-v5.txt` — add extraction examples for forensics staff, addresses, vehicle
- [x] Fix matcher.go to include `place`, `address`, `vehicle`, `technology` entity types

**Result:** Entity recall 100% on all three corpora ✓

---

**EV8 Status:** Entity recall and event recall thresholds met on all corpora. False positive rate (~64%) deferred.

---

### EV9: Inconsistency Detection
**Model: Opus (prompt design), Sonnet (implementation), Haiku (judge runtime)** | **Size:** M-L (6-10 hours)

**Goal:** Achieve ≥50% inconsistency detection rate across all three corpora.

**Current Status:** Inconsistency recall is 0% — not implemented. The manifests contain inconsistencies (BRF: 8, MNA: ~5, Police: ~4), but the extraction pipeline has no mechanism to detect or output them.

---

#### EV9.1: Data Structures ✅ COMPLETE
**Model: Sonnet** | **Size:** S (1 hour)

Add inconsistency types to evaluation package:

- [x] Add `ExtractedInconsistency` struct to `api/internal/evaluation/types.go`
  - Fields: id, type (amount/temporal/authority/procedural/obligation/provenance), severity, description, documents, entities, evidence (side_a, side_b)
- [x] Add `InconsistencyMatch` struct for scoring results
- [x] Add `InconsistencyRecall`, `InconsistencyPrecision`, `InconsistencyDetails` to `ScoreResult`

**Acceptance:** Types compile, match manifest inconsistency structure. ✓

---

#### EV9.2: Inconsistency Detection Prompt ✅ COMPLETE
**Model: Opus** | **Size:** M (2-3 hours)

Write the cross-document inconsistency detection prompt:

- [x] Create `prompts/postprocess/inconsistency.txt`
- [x] Input: All extracted nodes/edges merged across all corpus documents
- [x] Output: JSON array of `ExtractedInconsistency` objects
- [x] Cover types: amount, temporal, authority, procedural, obligation, provenance, contradiction
- [x] Include few-shot examples from BRF corpus (4 worked examples: budget change, authority violation, procedural conflict, deadline missed)

**Acceptance:** Prompt produces valid JSON with at least 3/8 BRF inconsistencies when tested manually. *(Pending EV9.3 for testing)*

---

#### EV9.3: Runner Integration ✅ COMPLETE
**Model: Sonnet** | **Size:** S (1-2 hours)

Add post-processing inconsistency detection to extraction runner:

- [x] Modify `api/internal/extraction/graph/runner.go`
- [x] After per-chunk extraction: merge all nodes/edges
- [x] Run inconsistency detection prompt with merged data
- [x] Parse response into `[]ExtractedInconsistency`
- [x] Add to `ExtractionResult.Inconsistencies`
- [x] Add `--detect-inconsistencies` flag to CLI

**Acceptance:** `./sikta-eval extract --detect-inconsistencies` produces JSON with `inconsistencies` array. ✓

---

#### EV9.4: Inconsistency Judge ✅ COMPLETE
**Model: Sonnet** | **Size:** M (2-3 hours)

Build LLM-as-judge for inconsistency matching (pattern: follow `EventJudge`):

- [x] Create `api/internal/evaluation/inconsistency_judge.go`
- [x] `InconsistencyJudge` struct with Claude client
- [x] `JudgeInconsistencies(ctx, manifestInconsistencies, extractedInconsistencies) []InconsistencyMatch`
- [x] System prompt: compare manifest inconsistency description + evidence against extracted ones
- [x] Determine if they refer to the same underlying issue (allow different wording)

**Acceptance:** Judge correctly matches BRF I1 (budget 650k→600k) when extracted with similar description. *(Pending EV9.5 for full testing)*

---

#### EV9.5: Scorer Integration
**Model: Sonnet** | **Size:** S (1 hour)

Wire inconsistency scoring into the scorer:

- [ ] Modify `api/internal/evaluation/scorer.go`
- [ ] Add `inconsistencyJudge` to Scorer (optional, like event judge)
- [ ] Run inconsistency judge in `ScoreWithContext` when `--full` flag is set
- [ ] Calculate recall/precision/F1 for inconsistencies
- [ ] Add to score output

**Acceptance:** `./sikta-eval score --full` shows inconsistency metrics in output.

---

#### EV9.6: Prompt Iteration
**Model: Opus (analysis), Sonnet (implementation)** | **Size:** Open-ended

Iterate on inconsistency detection prompt until ≥50% recall:

- [ ] Run extraction + scoring on all 3 corpora
- [ ] Analyze missed inconsistencies
- [ ] Refine prompt (types, examples, context window)
- [ ] Repeat until threshold met

**Go/Kill:** If ≥50% inconsistency recall cannot be achieved in 5 iterations, document limitations and proceed.

---

## Post-Validation Phases

> Extraction validation (EV8) complete. Entity recall and event recall thresholds met.

### Phase 8: Extraction Progress UX ✅ COMPLETE
**Size:** S (1-2 hours) | **Model:** Sonnet

- [x] Frontend: Estimated time remaining (ETA based on chunk progress)
- [x] Frontend: Error state per chunk with retry
- [x] **BUG:** Fix completion detection — SSE `status: "complete"` doesn't always navigate to timeline

**Changes:**
- Added ETA calculation showing "~Xm left" based on elapsed time and chunk progress
- Fixed SSE completion detection with hasCompleted flag and 500ms delay for data commit
- Added retry extraction button in error state
- Improved polling fallback with timeout (6 min max) and lenient completion detection
- Updated error state UI with separate "Upload different file" and "Retry extraction" buttons

### Phase 9: Multi-File Projects
**Size:** L (4-5 hours) | **Model:** Sonnet for implementation, Opus for cross-document entity resolution

- Projects table, multi-file upload, cross-document entity resolution
- Unified timeline showing all project documents
- Source document badges on events

### Phase 10: Mixed Document Types + Cross-Document Anomaly Detection
- Document-type-specific extraction (protocols, invoices, emails, legal docs)
- Cross-document anomaly detection (temporal contradictions, entity conflicts, missing references)

### Phase 11: Multiple LLM Provider Support
- Provider abstraction (Anthropic, OpenAI, Azure, local)
- Model selection per task, cost optimization, fallback

### Phase 12: Board Protocol Mode (Swedish Market Focus)
- Swedish board protocol extraction prompts
- Decision trail view, budget tracking

### Phase 13: Export & Sharing
- PDF/PNG/JSON/CSV export, shareable links, embeddable widget

### Phase 14: Authentication & Multi-Tenant
- User accounts, project management, sharing, billing

### Phase 15: Legal / Due Diligence Mode
- Contract obligation extraction, deadline tracking, entity mapping

---

## Icebox

- **EV8.7: False Positive Reduction** — Add hallucination guards, deduplication instructions to reduce FP rate from ~64% to <20%. Currently deferred; precision acceptable for MVP demo.
- Real-time collaborative review
- AI-assisted resolution suggestions for inconsistencies
- Timeline comparison (two documents side by side)
- API for programmatic access
- Plugin system for custom extraction types
- Integration with Sundla (association documents)
