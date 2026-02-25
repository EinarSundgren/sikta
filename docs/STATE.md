# Project State

> Current progress, blockers, and recent changes. Check before starting any task.

---

## Status

**Overall:** E2E (End-to-End Pipeline) phase started. Backend project support complete. Extraction validation passed all thresholds.

**Last Updated:** 2026-02-25

---

## Current Phase

**Phase:** E2E — End-to-End Pipeline

**Goal:** Deliver a working pipeline: Upload real municipality documents → extract to graph → see results in browser. Test corpus: `corpora/hsand2024/` (Swedish kommun meeting protocols).

**Completed:**
- [x] E2E.1: Connect validated prompts to backend (PromptLoader with fallback)
- [x] E2E.2: Document-type-aware chunking (ChunkStrategy interface, auto-detection)
- [x] E2E.3: Multi-document project support (projects table, API endpoints, routes)
- [x] E2E.4: Cross-document post-processing (deduplication, inconsistency detection)
- [x] E2E.5: Minimal UI results viewer (ProjectsPage, ProjectView)

**In Progress:**
- [ ] E2E.6: Integration testing with hsand2024 corpus
- [ ] E2E.5: Minimal UI results viewer
- [ ] E2E.6: Integration testing

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
| **Phase 9: Multi-File Projects** | **Next** | Sonnet | Projects table, multi-file upload, cross-document entity resolution |

### Completed in EV Phase
| **EV1-EV8: Extraction Validation** | **Complete** | All | Entity recall 100%, event recall 70-86% on all corpora |
| **Phase 8: Extraction Progress UX** | **Complete** | Sonnet | ETA calculation, retry extraction, SSE completion bug fix |

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
| 2026-02-25 | **E2E.5 complete:** Minimal UI results viewer. ProjectsPage for listing/creating projects. ProjectView with tabs for documents, entities, events, analysis results. Upload documents, add to project, run post-processing. Added "Projects" button to LandingPage. | web/src/api/projects.ts (new), web/src/pages/ProjectsPage.tsx (new), web/src/pages/ProjectView.tsx (new), web/src/App.tsx, web/src/pages/LandingPage.tsx |
| 2026-02-25 | **E2E.4 complete:** Cross-document post-processing. Added PostProcessor service for entity deduplication and inconsistency detection. Created dedup.txt prompt. API endpoints: POST /projects/{id}/postprocess, /deduplicate, /detect-inconsistencies. | api/internal/graph/postprocess.go (new), prompts/postprocess/dedup.txt (new), api/internal/handlers/projects.go, api/cmd/server/main.go |
| 2026-02-25 | **E2E.2 complete:** Document-type-aware chunking. Added ChunkStrategy interface with WholeDocChunker, SectionChunker, ChapterChunker, FallbackChunker. Auto-detection based on document structure (§ markers, chapters, word count). | api/internal/document/chunker.go, api/internal/document/parser.go, api/internal/document/chunker_test.go (new) |
| 2026-02-25 | **E2E.1 + E2E.3 complete:** Added PromptLoader for file-based prompts with fallback to hardcoded. Added projects table, project API handlers, and routes. Projects support: CRUD, add documents, get merged graph. | api/internal/config/config.go, api/internal/extraction/graph/prompt_loader.go (new), api/internal/extraction/graph/service.go, api/sql/schema/000015_projects.up.sql (new), api/internal/database/projects.sql.go (new), api/internal/handlers/projects.go (new), api/cmd/server/main.go |
| 2026-02-23 | **Phase 8 complete:** Extraction Progress UX improvements. Added ETA calculation based on chunk progress. Fixed SSE completion detection bug. Added retry extraction functionality. Improved error state UI with separate retry/upload buttons. | web/src/pages/LandingPage.tsx, docs/TASKS.md |
| 2026-02-23 | **EV8.6 complete:** v5 prompt iteration. Entity recall reaches 100% on all three corpora. Added entity extraction from events, expanded node types (address, vehicle, technology), fixed matcher to include new entity types. EV8.7 (FP reduction) deferred to icebox. | prompts/system/v5.txt, prompts/fewshot/mna-v5.txt, prompts/fewshot/police-v5.txt, api/internal/evaluation/matcher.go, docs/TASKS.md, docs/EV8-Prompt-Iteration-Summary.md |
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

**Milestone:** Phase 9 — Multi-File Projects

**Why this is next:** Extraction validation complete. All three corpora pass thresholds (100% entity recall, 70-86% event recall). Post-validation phases unblocked.

**Extraction Validation Results (v5):**
| Corpus | Entity Recall | Event Recall | Status |
|--------|---------------|--------------|--------|
| BRF | 100% | 71.4% | ✅ PASS |
| M&A | 100% | 70.0% | ✅ PASS |
| Police | 100% | 86.4% | ✅ PASS |

**Phase 9 Scope:**
- Projects table (group documents)
- Multi-file upload
- Cross-document entity resolution
- Unified timeline showing all project documents
- Source document badges on events

---

## Decision Log

| Date | Decision | Rationale |
|------|----------|-----------|
| 2026-02-23 | Defer false positive reduction (EV8.7) to icebox | Entity and event recall thresholds met. ~64% FP rate acceptable for MVP demo. Can iterate later if needed. |
| 2026-02-23 | v5 prompt: Entity extraction from events | Every event must create entity nodes for persons/orgs/places mentioned. Single-mention entities OK at 0.7-0.8 confidence. Expanded node types: address, vehicle, technology. |
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
