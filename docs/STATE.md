# Project State

> Current progress, blockers, and recent changes. Check before starting any task.

---

## Status

**Overall:** Phase 5 Complete — Entity Panel & Relationship Graph done. Phase 6 (Review Workflow) is next.

**Last Updated:** 2026-02-19

---

## Current Phase

**Phase:** 2.5 — Data Model Migration

**Goal:** Rename `documents` → `sources`, `events` → `claims`. Add two-level confidence model and `claim_type` discriminator.

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
  - TXT parser with chapter detection (multiple regex patterns)
  - PDF parser using pdftotext with page tracking
  - Document upload endpoint (`POST /api/documents`)
  - Document status endpoint (`GET /api/documents/:id/status`)
  - Chunk creation and storage in database
  - Async processing with worker pool (background goroutine)
  - Error handling and validation (file size, type, encoding)
  - Pride and Prejudice: 61 chapters correctly chunked
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

---

## What's In Progress

| Task | Status | Assigned Model | Notes |
|------|--------|----------------|-------|
| Phase 2.5: Data Model Migration | **Complete** | Sonnet | Migration done. DB + all Go files + frontend types updated. Build passes. |
| Phase 5: Entity Panel & Relationship Graph | **Complete** | Sonnet | EntityPanel sidebar + D3 RelationshipGraph + tab switching. claim_entities table empty (extraction didn't link events↔entities), entity filter uses text matching fallback. |

---

## Blockers

| Blocker | Impact | Resolution Needed |
|---------|--------|-------------------|
| None | — | — |

---

## Recent Changes

| Date | Change | Files Affected |
|------|--------|----------------|
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
| `podman-compose` | `pip3 install podman-compose` |
| `pdftotext` | Install poppler-utils (system package manager) |

Add `$(go env GOPATH)/bin` to PATH for installed Go tools.

### Podman Notes

Podman machine must be running before `make infra`:
```
/opt/podman/bin/podman machine start
```

podman-compose binary is at `/Users/einar.sundgren/Library/Python/3.9/bin/podman-compose`.
Add to PATH: `export PATH="$PATH:/Users/einar.sundgren/Library/Python/3.9/bin:$(go env GOPATH)/bin"`

### Database

**Migrations Applied:** 10 (001_documents through 010_rename_sources_claims)

**Schema Version:** 010_rename_sources_claims

**Current Data:** Pride and Prejudice (doc ID: `334903c6-de15-469a-8671-686dd9c2b534`) — 61 chunks, 178 events, 54 entities, 58 relationships

---

## Known Issues

| Issue | Severity | Workaround |
|-------|----------|------------|
| podman-compose not in PATH by default | Low | Add `/Users/einar.sundgren/Library/Python/3.9/bin` to PATH |
| Go tools not in PATH by default | Low | Add `$(go env GOPATH)/bin` to PATH |
| pdftotext required for PDF parsing | Medium | Install poppler-utils via system package manager |
| `events_event_type_check` constraint too restrictive | Low | ~6 events lost from extraction when LLM returns unexpected types. Will be fixed in migration. |

---

## Next Milestone

**Milestone:** Phase 6 — Review Workflow & Inconsistency Panel

**Why this is next:** Entity panel and graph are done. Phase 6 adds the human review loop: approve/reject/edit extracted items, keyboard shortcuts, and the inconsistency panel.

**Recommended model:** Sonnet for implementation.

**Known issue:** `claim_entities` table is empty — the extraction pipeline stored events and entities separately but never linked them. Entity filtering on the timeline uses text-match fallback (works for Pride and Prejudice). To fix properly: update the extraction prompt to return participant entities per event, add `storeEvent` to call `CreateClaimEntity` for each participant.

**After Phase 6:** Phase 7 (Demo Polish & Landing).

---

## Decision Log

| Date | Decision | Rationale |
|------|----------|-----------|
| 2026-02-19 | Rename `documents` → `sources`, `events` → `claims` | Sources is more accurate (any ingested material). Claims captures that extractions are assertions, not ground truth. Enables two-level confidence and claim_type discriminator for extensibility. |
| 2026-02-19 | Two-level confidence model | Source trust (how reliable is the source?) vs assertion confidence (how confident is the extraction?). Effective confidence = trust × confidence. |
| 2026-02-19 | `claim_type` discriminator | Single `claims` table holds events, attributes, and relational claims. New claim types = zero schema changes. |
| 2026-02-19 | Keep HTTP routes as `/api/documents/...` during migration | External API stability. Internal naming changes, external stays the same. |
| 2026-02-18 | Chapter-based chunking (not size-based) | Chapters are the smallest coherent narrative units. LLM extraction needs context. |
| 2026-02-18 | Multiple regex patterns for chapter detection | Handles different formatting styles (Roman numerals, numeric, "CHAPTER X", etc.) |
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
