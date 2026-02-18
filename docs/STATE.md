# Project State

> Current progress, blockers, and recent changes. Check before starting any task.

---

## Status

**Overall:** Phase 1 Complete — Ready for Phase 2

**Last Updated:** 2026-02-18

---

## Current Phase

**Phase:** 2 — LLM Extraction Pipeline (Next)

**Goal:** Send chunks to Claude API, get structured events/entities/relationships back.

---

## What's Working

- [x] Project documentation (CLAUDE.md, docs/TASKS.md, docs/STATE.md)
- [x] Data model drafted (see CLAUDE.md)
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
  - Pride and Prejudice downloaded to `api/uploads/demo/`
  - Full API implementation with repository pattern

---

## What's In Progress

| Task | Status | Assigned Model | Notes |
|------|--------|----------------|-------|
| Phase 2: LLM Extraction Pipeline | Not started | Opus (prompts) + Sonnet (pipeline) | Depends on Phase 1 ✅ |

---

## Blockers

| Blocker | Impact | Resolution Needed |
|---------|--------|-------------------|
| None | — | — |

---

## Recent Changes

| Date | Change | Files Affected |
|------|--------|----------------|
| 2026-02-18 | Phase 1 complete: TXT/PDF parsers, chapter detection, document upload API, async processing | api/internal/document/, api/internal/services/, api/internal/handlers/, api/internal/database/ |
| 2026-02-18 | Pride and Prejudice downloaded from Project Gutenberg | api/uploads/demo/pride-and-prejudice.txt |
| 2026-02-18 | Project initialized with full documentation | CLAUDE.md, docs/TASKS.md, docs/STATE.md |
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

**Migrations Applied:** 2 (001_documents.sql, 002_chunks.sql)

**Schema Version:** 002_chunks

---

## Known Issues

| Issue | Severity | Workaround |
|-------|----------|------------|
| podman-compose not in PATH by default | Low | Add `/Users/einar.sundgren/Library/Python/3.9/bin` to PATH |
| Go tools not in PATH by default | Low | Add `$(go env GOPATH)/bin` to PATH |
| pdftotext required for PDF parsing | Medium | Install poppler-utils via system package manager |

---

## Next Milestone

**Milestone:** Phase 2 — LLM Extraction Pipeline

**Why this is next:** We now have structured chunks with position metadata. Next step is to extract events, entities, and relationships using Claude API.

**Recommended model:** Opus for prompt design (critical decisions that compound), Sonnet for pipeline implementation.

**After Phase 2:** Phase 3 (Inconsistency Detection) — detection logic design is an Opus-level task.

---

## Decision Log

| Date | Decision | Rationale |
|------|----------|-----------|
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

