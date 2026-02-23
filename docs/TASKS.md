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

### EV6: LLM-as-Judge for Inconsistency Detection
**Model: Sonnet** | **Size:** S (2-3 hours)

`api/internal/evaluation/judge.go` — uses Claude Haiku to judge whether detected inconsistencies match planted ones.

**Tasks:**
- [ ] Implement `JudgeInconsistencies(ctx, client, detected, manifest) ([]InconsistencyMatch, error)`
- [ ] Prompt: "Here is a planted inconsistency. Here are the detected inconsistencies. Does any detected one match? JSON: {match, matched_id, reasoning}"
- [ ] Integrate as optional step in scoring pipeline (only when `--full` flag is set)

**Acceptance:**
- Judge correctly matches a clearly-worded detected inconsistency against the manifest
- Returns structured match result with reasoning
- Does not run without `--full` flag (Claude API calls cost money)

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

### EV8: Prompt Iteration (Ongoing)
**Model: Collaboration (AI proposes, developer runs)** | **Size:** Open-ended

With EV1-EV7 in place, iterate on prompts until go/kill thresholds are met.

**Workflow:**
1. Developer runs `sikta-eval extract` on all three corpora
2. Developer runs `sikta-eval score` to get baseline numbers
3. Developer shares score output + failure cases with AI
4. AI analyzes failures, proposes prompt changes, writes next version
5. Developer re-runs → re-scores → shares results
6. Repeat until thresholds met or 10 iterations exhausted

**Focus areas for prompt improvement:**
- Missed entities: too-restrictive type list? few-shot doesn't cover this pattern?
- Missed events: passive voice, indirect speech not captured?
- Missed inconsistencies: cross-document context insufficient?
- False positives: hallucinating entities or relationships?
- Confidence calibration: do high-confidence extractions actually prove out?

**Go/Kill decision:**
- If thresholds met within 10 iterations: proceed to post-validation UI and product work
- If thresholds not met: document failures, reassess extraction approach before building more product

---

## Post-Validation Phases (deferred until EV8 thresholds met)

> These phases remain valid but are blocked until extraction validation passes.

### Phase 8: Extraction Progress UX
**Size:** S (1-2 hours) | **Model:** Sonnet

- [ ] Frontend: Estimated time remaining
- [ ] Frontend: Error state per chunk with retry
- [ ] **BUG:** Fix completion detection — SSE `status: "complete"` doesn't always navigate to timeline

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

- Real-time collaborative review
- AI-assisted resolution suggestions for inconsistencies
- Timeline comparison (two documents side by side)
- API for programmatic access
- Plugin system for custom extraction types
- Integration with Sundla (association documents)
