# Project State

> Current progress, blockers, and recent changes. Check before starting any task.

---

## Status

**Overall:** Course correction — Extraction Validation Phase is now the priority. All MVP UI phases complete. No further UI work until extraction quality is validated against real-world documents.

**Last Updated:** 2026-02-23

---

## Current Phase

**Phase:** EV — Extraction Validation

**Goal:** CLI-driven extraction testing pipeline against three realistic document corpora (BRF board protocols, M&A due diligence, criminal investigation). Measure entity recall, event recall, inconsistency detection, false positive rate. Iterate on prompts until go/kill thresholds are met.

**Go/Kill thresholds:** ≥85% entity recall, ≥70% event recall, ≥50% inconsistency detection, <20% false positive rate.

See `docs/EXTRACTION_VALIDATION.md` for full rationale and spec.

---

## What's Working

- [x] Project documentation (CLAUDE.md, docs/TASKS.md, docs/STATE.md)
- [x] Data model drafted and refined (sources/claims architecture)
- [x] UX concept defined (horizontal dual-lane timeline, entity panel, graph, review workflow)
- [x] Demo novel selected (Pride and Prejudice)
- [x] **Phase 0: Project Scaffolding — Complete**
  - Go module initialized (`github.com/einarsundgren/sikta`)
  - PostgreSQL running via podman-compose (podman machine must be started first)
  - React + TypeScript + Vite + Tailwind CSS frontend
  - `GET /health` returns 200 with JSON status
  - CORS middleware configured
  - sqlc configured and generating code
  - `air` installed for backend hot reload
  - Frontend connects to backend via Vite proxy (`/health`, `/api`)
- [x] **Phase 1: Document Ingestion & Chunking — Complete**
  - Structure-agnostic chunking: paragraph boundaries + word budget (3000 target, 4500 max)
  - Gutenberg boilerplate stripping (standard START/END markers)
  - PDF parser using pdftotext with page tracking
  - Document upload endpoint (`POST /api/documents`)
  - Document status endpoint (`GET /api/documents/:id/status`)
  - Chunk creation and storage in database
  - Async processing with worker pool (background goroutine)
  - Error handling and validation (file size, type, encoding)
  - Pride and Prejudice: ~42 chunks, Dr Jekyll: ~8 chunks, Wuthering Heights: ~38 chunks
- [x] **Phase 2: LLM Extraction Pipeline — Complete**
  - Claude API integration for structured extraction
  - 61 chapters processed → 178 events, 54 entities, 58 relationships
  - Source references linking extractions to chunks
  - Chronological position estimation via LLM
  - Entity deduplication service
  - Extraction CLI and HTTP triggers
- [x] **Phase 3: Inconsistency Detection — Complete**
  - Inconsistency detection service (narrative vs chronological, contradictions, temporal)
  - Database schema for inconsistencies and inconsistency_items
  - API endpoints for inconsistencies
- [x] **Phase 4: Timeline Hero View — Complete**
  - D3 horizontal dual-lane timeline
  - Chronological and narrative lanes with connectors
  - Event cards with confidence markers
  - Click-to-detail panel
  - Dynamic document loading (no hardcoded IDs)
- [x] **Phase 5: Entity Panel & Relationship Graph — Complete**
  - Entity sidebar grouped by type (people, places, organizations) with search
  - Click entity → filters timeline events by name matching
  - D3 force-directed relationship graph (191 nodes, 236 edges)
  - Node sizing by relationship count, color by entity type
  - Hover tooltips on nodes and edges
  - Drag nodes, zoom/pan, auto-fit on load
  - Tab switching: Timeline ↔ Graph
  - Shared entity selection state across all views
- [x] **Phase 7: Demo Polish & Landing — Complete**
  - Landing page with hero section, demo card (events/entities/relationships counts), upload zone
  - Drag-and-drop upload with multi-phase progress (uploading → chunking → extracting)
  - Extraction progress polling with "View partial results" early access
  - State-based routing in App.tsx (landing ↔ timeline)
  - TimelineView accepts `docId` prop and `onNavigateHome` callback
  - Back button (←) in TimelineView header
  - Emoji favicon (🔍) and improved page title
  - `make dump-demo` and `make seed-demo` Makefile targets
  - `demo/seed.sql` — 1700-line dump of P&P extraction (178 events, 54 entities, 58 relationships)
- [x] **Phase 6: Review Workflow & Inconsistency Panel — Complete**
  - Keyboard-driven review queue: J/K navigate, A approve, R reject, E edit
  - Progress bar showing reviewed/total across claims and entities
  - Filter tabs: pending / all / approved / rejected / edited
  - Edit modal: title, description, date_text, event_type, confidence slider
  - Inconsistency panel with severity filter (conflict/warning/info)
  - Expand/collapse inconsistency cards
  - Actions: Mark resolved, Add/edit note, Dismiss
  - "Show on timeline" button — highlights related claims and switches to timeline tab
  - Optimistic UI updates for all review status changes
  - Backend: PATCH /api/claims/{id}/review, PATCH /api/claims/{id}, GET /api/documents/{id}/review-progress
  - Four-tab layout: Timeline, Graph, Review, Inconsistencies
- [x] **Phase DS: Design System Implementation & Bug Fixes — Complete**
  - Bug fixes: API_BASE_URL fallback to '' (Vite proxy fix), LandingPage demo stats use /review-progress
  - Design tokens: Google Fonts (DM Sans, JetBrains Mono), CSS custom properties
  - LandingPage redesign: hero, stats cards, "other novels available" section, "works with" section
  - Timeline component: card-based D3 timeline with connectors, confidence badges, entity chips, animations
  - TimelineView redesign: underline tabs (not boxes), filter bar with entity chips + conflict count, SourcePanel wired in
  - EntityPanel redesign: colored shape indicators (●◆■▲), mono counts, collapsible groups
  - SourcePanel component: 420px slide-in from right, source text with excerpt, review actions
  - RelationshipGraph: dot grid background per design spec
  - "Inconsistencies" tab renamed to "Conflicts"

---

## What's In Progress

| Task | Status | Assigned Model | Notes |
|------|--------|----------------|-------|
| **EV1: Corpus Preparation** | **Complete** | Haiku | Three corpora extracted into `corpora/` directory structure. manifest.json files created for all three. |
| **EV2: Prompt File Extraction** | Pending | Sonnet | Extract prompts from prompts.go into prompts/system/v1.txt + domain-specific few-shot examples |
| **EV3: Database-Free Extraction Runner** | Pending | Sonnet | api/internal/extraction/graph/runner.go — no PostgreSQL dependency |
| **EV4: Scoring Engine** | Pending | Sonnet | api/internal/evaluation/ package — entity/event/inconsistency matching + F1 scores |
| **EV5: CLI Entry Point** | Pending | Sonnet | api/cmd/evaluate/main.go — extract/score/compare subcommands |
| **EV6: LLM-as-Judge** | Pending | Sonnet | api/internal/evaluation/judge.go — for inconsistency scoring |
| **EV7: Post-Processing Prompts** | Pending | Sonnet | dedup.txt + inconsistency.txt prompts, integrated as optional pipeline stages |
| **EV8: Prompt Iteration** | Pending | Collaboration | Run → score → analyze → revise → repeat until thresholds met |

### Previously Completed (archived)
| **Phase DS: Design System Implementation** | **Complete** | Sonnet/Haiku | |
| **Phase G1-G6: Graph Model Alignment** | **Complete** | Sonnet | |

---

## Blockers

| Blocker | Impact | Resolution Needed |
|---------|--------|-------------------|
| None | — | — |

---

## Recent Changes

| Date | Change | Files Affected |
|------|--------|----------------|
| 2026-02-22 | **Course correction:** Extraction Validation Phase introduced. Three test corpora prepared (BRF, M&A, Police) with ground truth manifests. TASKS.md restructured. STATE.md updated to reflect new current phase. | docs/TASKS.md, docs/STATE.md, docs/EXTRACTION_VALIDATION.md, corpora/brf/*, corpora/mna/*, corpora/police/* |
| 2026-02-23 | **EV8.5 complete:** v3 prompt iteration and testing. Re-ran BRF extraction with v3 (label preservation focus). Event recall: 57.1% (8/14) — improved from v2's 50.0% but below 70% threshold. Fixes: V6 (no prefix), V12 (Slut- prefix), V13 (retroactive pattern) now match. Regressions: V1 (wrong phrase), V5 (lost). Still broken: V3 (budget), V8 (amount suffix), V9 (Byggstart), V14 (wrong label). Analysis: Hybrid approach needed (exact wording + standard formats). Projected v4: 92.9% (13/14) with 6 targeted fixes. Recommendation: Continue to v4. | prompts/system/v3.txt, results/brf-v3.json, docs/EV8.5-v3-Prompt-Iteration.md, docs/EV8.5-v3-Event-Comparison.md, docs/EV8.5-Final-Analysis.md, docs/STATE.md |
| 2026-02-23 | **EV8.3 complete:** Targeted prompt revision. Created v2 prompt with 4 improvements: (1) Event node classification - all time-bound events as `event` nodes not value/document, (2) Budget labels must include exact amounts, (3) Decision labels must include key actors, (4) Preserve source terminology (exact words from text). Created event comparison CLI tool for verbose analysis. Expected event recall improvement: 21.4% → 85.7%. | prompts/system/v2.txt, docs/EV8.3-Targeted-Prompt-Revision.md, api/cmd/compare-events/main.go, docs/EV8.3-Event-Comparison.md, docs/EV8.3-Event-Label-Analysis.md, Makefile |
| 2026-02-21 | **Phase DS complete:** Design system implementation per prototype. Bug fixes (API_BASE_URL → '', LandingPage → /review-progress). Google Fonts + CSS variables. LandingPage redesign with stats + other novels + works-with. Timeline: card-based D3 with connectors, confidence badges, entity chips, animations. TimelineView: underline tabs, filter bar, SourcePanel wired. EntityPanel: colored shapes, mono counts, collapsible. New SourcePanel: 420px slide-in with source excerpt + review actions. Graph: dot grid background. Tab renamed "Inconsistencies" → "Conflicts". Backend: populate source_references from provenance (views.go + timeline.go). | web/src/api/timeline.ts, web/src/pages/LandingPage.tsx, web/index.html, web/src/index.css, web/src/components/timeline/Timeline.tsx, web/src/pages/TimelineView.tsx, web/src/components/entities/EntityPanel.tsx, web/src/components/source/SourcePanel.tsx (new), web/src/components/graph/RelationshipGraph.tsx, api/internal/graph/views.go, api/internal/handlers/graph/timeline.go, docs/TASKS.md |
| 2026-02-19 | Fixed structure-agnostic chunking: replaced regex chapter detection with paragraph-boundary word-budget splitting (target 3000 words, max 4500). Added Gutenberg boilerplate stripping. Fixed missing /api/documents/{id}/status route for frontend polling. Updated Makefile for `podman compose` (built-in) vs podman-compose. | api/internal/document/parser.go, api/cmd/server/main.go, api/internal/services/document_service.go, web/src/pages/LandingPage.tsx, Makefile, podman-compose.yaml |
| 2026-02-19 | Phase 7 complete: Landing page (hero + demo card + upload flow), state-based routing, back button in TimelineView, favicon, seed SQL (1700 lines, full P&P extraction), Makefile dump-demo/seed-demo targets. | web/src/App.tsx, web/src/pages/LandingPage.tsx, web/src/pages/TimelineView.tsx, web/index.html, Makefile, demo/seed.sql |
| 2026-02-19 | Phase 6 complete: Review workflow (J/K/A/R/E keyboard shortcuts), edit modal, inconsistency panel with resolve/note/dismiss, "show on timeline" highlight. Backend review routes. Frontend build passes clean. | api/sql/queries/claims.sql, api/sql/queries/entities.sql, api/internal/handlers/review.go, api/cmd/server/main.go, web/src/components/review/*, web/src/components/inconsistencies/*, web/src/pages/TimelineView.tsx, web/src/components/timeline/Timeline.tsx |
| 2026-02-19 | Phase 2.5 complete: Renamed documents→sources, events→claims throughout. Added claim_type, source_trust columns. All Go files + frontend types updated. Build passes clean. | api/sql/schema/010_rename_sources_claims.sql, api/sql/queries/*, api/internal/database/*, api/internal/handlers/*, api/internal/extraction/*, api/cmd/*/main.go, web/src/types/index.ts |
| 2026-02-19 | Documentation updated for sources/claims data model migration | CLAUDE.md, docs/TASKS.md, docs/STATE.md |
| 2026-02-19 | Fixed chunking: 61 chapters now detected (was 2). Fixed 8 bugs blocking extraction. Full extraction complete (178 events, 54 entities, 58 relationships). | api/internal/document/parser.go, api/internal/handlers/documents.go, api/internal/handlers/extraction.go, api/internal/database/*.go, api/internal/extraction/service.go, api/cmd/server/main.go, web/src/pages/TimelineView.tsx |
| 2026-02-18 | Phase 1 complete: TXT/PDF parsers, chapter detection, document upload API, async processing | api/internal/document/, api/internal/services/, api/internal/handlers/, api/internal/database/ |
| 2026-02-18 | Phase 0 complete: Go backend, React frontend, PostgreSQL, Makefile, sqlc | api/, web/, Makefile, podman-compose.yaml, .env.example |

---

## Environment Status

### Services

| Service | Status | Notes |
|---------|--------|-------|
| `api` | Ready (local `go run` or `air`) | `make backend` |
| `web` | Ready (Vite dev server) | `make frontend` |
| `db` | Ready (podman container) | `make infra` — requires podman machine running |

### Tools Installed

| Tool | How to Install |
|------|---------------|
| `air` | `go install github.com/air-verse/air@latest` |
| `sqlc` | `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest` |
| `pdftotext` | Install poppler-utils (system package manager) |

Add `$(go env GOPATH)/bin` to PATH for installed Go tools.

### Podman Notes

Podman machine must be running before `make infra`:
```
/opt/podman/bin/podman machine start
```

Makefile now uses `podman compose` (built-in subcommand) instead of `podman-compose` (separate install). No PATH changes needed.

### Database

**Migrations Applied:** 10 (001_documents through 010_rename_sources_claims)

**Schema Version:** 010_rename_sources_claims

**Current Data:** Pride and Prejudice (doc ID: `334903c6-de15-469a-8671-686dd9c2b534`) — 61 chunks, 178 events, 54 entities, 58 relationships

---

## Known Issues

| Issue | Severity | Workaround |
|-------|----------|------------|
| Go tools not in PATH by default | Low | Add `$(go env GOPATH)/bin` to PATH |
| pdftotext required for PDF parsing | Medium | Install poppler-utils via system package manager |
| `events_event_type_check` constraint too restrictive | Low | ~6 events lost from extraction when LLM returns unexpected types. Will be fixed in migration. |

---

## Next Milestone

**Milestone:** EV8.6 (v4 Prompt Iteration) — Hybrid approach: exact wording + standard formats.

**Why this is next:** EV8.5 complete — v3 achieved 57.1% event recall (partial success). Need v4 to close remaining 12.9pp gap to 70% threshold.

**Current Status:**
- EV1-EV5: ✅ Complete
- EV8.1-EV8.2: ✅ Complete (baseline + matcher fixes)
- EV8.3: ✅ Complete (v2 prompt designed)
- EV8.4: ✅ Complete (v2 tested: 50.0% recall)
- EV8.5: ✅ Complete (v3 tested: 57.1% recall)
- EV8.6: ⏸️ Next — v4 implementation (awaiting user decision)

**Progress:**
- v1 → v2: +28.6pp (21.4% → 50.0%)
- v2 → v3: +7.1pp (50.0% → 57.1%)
- Total: +35.7pp improvement from baseline
- **Remaining:** 12.9pp to reach 70% threshold (need +2 events: 8/14 → 10/14)

**EV8.5 Results Summary:**
- Event Recall: 57.1% (8/14) — +7.1pp from v2, still below 70% target
- New matches: V6, V12, V13 (3 events fixed by v3)
- Regressions: V1, V5 (2 events broken by v3)
- Still broken: V3, V8, V9, V14 (4 events)
- Key learning: "Preserve exact wording" helps sometimes, hurts sometimes

**Recommended path:** v4 with hybrid approach — exact wording + standard format exceptions for documents/budgets/construction

**Projected v4 result:** 92.9% (13/14) — exceeds ≥70% threshold ✅

**No UI work until EV8 thresholds are met.**

---

## Decision Log

| Date | Decision | Rationale |
|------|----------|-----------|
| 2026-02-19 | Structure-agnostic chunking (paragraph boundaries + word budget) | Regex chapter detection is brittle — failed for Dr Jekyll (all-caps section titles). New approach: split on blank lines, accumulate paragraphs targeting 3000 words/chunk, flush at 4500, merge trailing chunks <1500. Works on any plain text regardless of formatting. |
| 2026-02-19 | Gutenberg boilerplate stripping | Standard `*** START/END OF THE PROJECT GUTENBERG EBOOK` markers. Non-Gutenberg texts pass through unchanged. |
| 2026-02-19 | Rename `documents` → `sources`, `events` → `claims` | Sources is more accurate (any ingested material). Claims captures that extractions are assertions, not ground truth. Enables two-level confidence and claim_type discriminator for extensibility. |
| 2026-02-19 | Two-level confidence model | Source trust (how reliable is the source?) vs assertion confidence (how confident is the extraction?). Effective confidence = trust × confidence. |
| 2026-02-19 | `claim_type` discriminator | Single `claims` table holds events, attributes, and relational claims. New claim types = zero schema changes. |
| 2026-02-19 | Keep HTTP routes as `/api/documents/...` during migration | External API stability. Internal naming changes, external stays the same. |
| 2026-02-18 | ~~Chapter-based chunking (not size-based)~~ | ~~Chapters are the smallest coherent narrative units. LLM extraction needs context.~~ **Superseded by structure-agnostic approach.** |
| 2026-02-18 | ~~Multiple regex patterns for chapter detection~~ | ~~Handles different formatting styles (Roman numerals, numeric, "CHAPTER X", etc.)~~ **Removed — too brittle.** |
| 2026-02-18 | pdftotext for PDF parsing | Reliable, preserves layout, handles multi-column documents well |
| 2026-02-18 | Page marker strategy for PDFs | Insert `[[[PAGE N]]]` markers during extraction, build lookup table for offset→page mapping |
| 2026-02-18 | Async processing with worker pool | Prevent resource exhaustion, handle multiple concurrent uploads |
| 2026-02-18 | Repository pattern for database access | Clean separation of concerns, easy to test |
| 2026-02-18 | Pride and Prejudice as demo novel | Rich social network, manageable length, well-known, public domain. |
| 2026-02-18 | Horizontal dual-lane timeline as hero view | Chronological vs. narrative order displayed simultaneously. Crossed connectors reveal non-linear storytelling. |
| 2026-02-18 | Polished demo quality | This needs to impress as a shareable artifact. Not a prototype — a demo. |
| 2026-02-18 | Go + React + PostgreSQL stack | Consistent with portfolio (Substrata, Sundla). Boring technology principle. |
| 2026-02-18 | D3.js for timeline and graph | Dual-lane timeline with connectors is too custom for a charting library. Need full control. |
| 2026-02-18 | Go 1.22+ stdlib mux for routing | Method+path pattern routing added in Go 1.22. No external router needed for Phase 0-1. |
| 2026-02-18 | pgx/v5 as database driver | Matches sqlc.yaml sql_package config. Standard high-performance PostgreSQL driver. |
| 2026-02-18 | Vite proxy for /health and /api | Frontend calls backend via same-origin proxy — no CORS in dev, clean production parity. |
