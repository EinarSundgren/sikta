# CLAUDE.md

> **Entry point for Claude Code agents. Read this first.**

---

## Agent Instructions

### Before Any Task

1. **Read only what you need.** Start here, then navigate to specific docs.
2. **Assess the model requirement.** Before starting work, state which model (Opus/Sonnet/Haiku) is appropriate and why.
3. **Check `docs/STATE.md`** for current progress and blockers.
4. **Check `docs/TASKS.md`** for the current backlog and priorities.

### During Work

- Follow the architectural qualities in priority order: Security > Clarity > Extendability > Maintainability
- Write code that is obvious, not clever
- When making decisions, document them in the relevant doc file
- English-first for all code and comments

### After Completing a Task

1. **Stop.** Do not continue to the next task automatically.
2. **Update `docs/STATE.md`** with what was completed.
3. **Summarize** what was done in 2-3 sentences.
4. **Ask** if the user wants to continue to the next task.

### Model Selection

| Task Type | Model | Examples |
|-----------|-------|---------|
| Architecture, schema design, extraction prompt design | Opus | Data model, LLM pipeline design, inconsistency detection logic |
| Feature implementation, API handlers, frontend components | Sonnet | Timeline view, entity panel, review workflow |
| Bug fixes, string changes, simple CRUD | Haiku | Fix typo, add field, update copy |

**Default to Sonnet.** Use Opus only for decisions that compound. Use Haiku for mechanical work with clear context.

### Key Documentation

| Document | Purpose | When to Read |
|----------|---------|-------------|
| `docs/STATE.md` | Current progress, blockers | Before starting any task |
| `docs/TASKS.md` | Backlog with phase plan and estimates | When picking next task |
| `docs/features/` | Detailed feature specifications | Before implementing a feature |

---

## Product Overview

**Sikta** — Document Timeline Intelligence

A tool that takes unstructured documents and extracts structured, navigable timelines with source references. The AI does the heavy lifting; humans verify. Every item is traceable to its source. Inconsistencies are surfaced, not hidden.

**MVP scope:** Novel demo using public domain books (starting with Pride and Prejudice). Proves the core capability in a shareable, privacy-safe context. The same engine later applies to board protocols, legal documents, case files, and research.

### Core Value Proposition

"Make unstructured text auditable and navigable."

### What Makes Sikta Defensible

1. **Source references that actually work** — every extracted claim links to source + page/section
2. **Inconsistency detection** — contradictions, temporal impossibilities, narrative vs. chronological mismatches surfaced automatically
3. **Human review workflow** — fast batch approve/reject, not painful
4. **Cross-source intelligence** — same claim in multiple sources → one entry with multiple references (post-MVP)
5. **Two-level confidence** — source trust and assertion confidence are explicit, not hidden

### Portfolio Context

Sikta is part of a product portfolio:
- **Substrata** — SBOM monitoring for NIS2 compliance
- **Sundla** — Association administration platform
- **Sikta** — Document timeline intelligence (this project)

All share a philosophy: reveal structure hidden in complexity.

---

## Architecture

### Three-Layer Design

```
┌─────────────────────────────────────────────┐
│  PRESENTATION LAYER                         │
│  Timeline View · Entity Panel · Graph View  │
│  Inconsistency Panel · Review Workflow      │
│  Source Text Viewer                         │
└──────────────────┬──────────────────────────┘
                   │
┌──────────────────┴──────────────────────────┐
│  INTELLIGENCE LAYER                         │
│  LLM Extraction · Confidence Scoring        │
│  Inconsistency Detection · Entity Resolution│
│  Narrative vs. Chronological Ordering       │
└──────────────────┬──────────────────────────┘
                   │
┌──────────────────┴──────────────────────────┐
│  DATA LAYER                                 │
│  Sources · Chunks · Claims · Entities       │
│  Relationships · Inconsistencies · Reviews  │
│  Source References                          │
└─────────────────────────────────────────────┘
```

### Extraction Pipeline

```
Document Upload
    │
    ▼
Text Extraction (PDF/TXT → raw text with page/section positions)
    │
    ▼
Chunking (split by chapter/section, preserve location metadata)
    │
    ▼
LLM Extraction (Sonnet via API)
    → Claims: events, attributes, relations extracted as structured assertions
    → Entities: people, places, organizations, amounts
    → Relationships: connections between entities
    → Each with: source reference + confidence score + narrative position
    │
    ▼
Post-Processing
    → Inconsistency detection (contradictions, temporal impossibilities)
    → Narrative vs. chronological ordering
    → Entity deduplication (same character, different mentions)
    │
    ▼
Human Review Queue
    → Pending items sorted by confidence (lowest first)
    → Batch approve/reject/edit
    │
    ▼
Timeline + Entity Graph (the product)
```

### Confidence Markers

| Type | Example | Marker | Detection |
|------|---------|--------|-----------|
| Date — precise | "15 March 1805" | ✔ | LLM extraction |
| Date — approximate | "that spring" | ⚠️ | LLM classification |
| Date — inferred | "before the ball" (relative) | ❓ | LLM reasoning |
| Entity — named | "Elizabeth Bennet" | ✔ | NER confidence |
| Entity — referenced | "her sister" | ⚠️ | Coreference resolution |
| Fact — explicit | "She had five daughters" | ✔ | Direct extraction |
| Fact — contradicted | Doc says 5 in ch.1, 4 in ch.12 | ⚡ | Cross-reference |

Human sees: ✔ (high confidence), ⚠️ (review suggested), ❓ (uncertain), ⚡ (conflict detected)

### Inconsistency Types

| Type | Description | Display |
|------|-------------|---------|
| Narrative vs. Chronological | Story order ≠ timeline order (flashbacks, foreshadowing) | Dual-lane timeline with connectors |
| Contradicting accounts | Character A says X, character B says Y | Conflict card with both sources |
| Contradicting data | Amount/date/fact differs between passages | Inline ⚡ marker + conflict panel entry |
| Temporal impossibility | Character in two places at once, impossible travel | Flagged on timeline with explanation |

---

## Tech Stack

| Layer | Technology | Notes |
|-------|-----------|-------|
| Backend | Go | API server, orchestration, extraction pipeline |
| Frontend | React + TypeScript | Vite, Tailwind CSS |
| Database | PostgreSQL | JSONB for flexible metadata |
| Container | Podman Compose | PostgreSQL for local dev |
| Document parsing | pdftotext (poppler), plain text | PDF and TXT for MVP |
| LLM | Claude API | Sonnet for extraction, Haiku for classification |
| Timeline | D3.js | Custom horizontal dual-lane timeline |
| Graph | D3 force-directed | Entity relationship network |

### Why These Choices

- **Go + React + PostgreSQL** — Boring technology. Same stack as Sundla and Substrata. No context switching.
- **D3.js** — The dual-lane timeline with connectors and the force-directed graph are custom enough that a charting library won't cut it. D3 gives full control.
- **Podman** — Rootless containers, open-source preference, consistent with portfolio.
- **Claude API** — Structured JSON output, good at extraction tasks, tiered models for cost optimization.

---

## Data Model

> **Migrated from initial draft.** Key renames: `documents` → `sources`, `events` → `claims`.
> See Decision Log for rationale.

### Design Principles

1. **Two-level confidence:** Source trust (Level 1: how reliable is this source?) vs Assertion confidence (Level 2: how confident is the extraction?)
2. **Three atomic primitives:** Claims (event/attribute/relation), Entities, Relations
3. **Extensibility:** New document types = new parser + new prompt template, zero schema changes
4. **`claim_type` discriminator:** A single `claims` table holds events, attributes, and relational claims — distinguished by `claim_type`

### Core Tables

```sql
-- Sources (formerly "documents") — anything Sikta ingests
sources (
    id UUID PRIMARY KEY,
    title TEXT NOT NULL,
    filename TEXT NOT NULL,
    file_type TEXT NOT NULL,        -- 'pdf', 'txt', 'docx', 'csv' (extensible)
    total_pages INTEGER,
    upload_status TEXT NOT NULL,     -- 'uploaded', 'processing', 'ready', 'error'
    is_demo BOOLEAN DEFAULT FALSE,
    source_trust REAL,              -- Level 1: 0.0-1.0, how reliable is this source?
    trust_reason TEXT,              -- why this trust score (e.g., "official document", "novel")
    metadata JSONB,                 -- author, publication year, source type, etc.
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ
)

-- Text chunks with position info
chunks (
    id UUID PRIMARY KEY,
    source_id UUID REFERENCES sources,
    chunk_index INTEGER NOT NULL,
    content TEXT NOT NULL,
    chapter_title TEXT,
    chapter_number INTEGER,
    page_start INTEGER,
    page_end INTEGER,
    narrative_position INTEGER,     -- order in the story (chapter sequence)
    created_at TIMESTAMPTZ
)

-- Claims (formerly "events") — any assertion extracted from a source
-- claim_type discriminates: 'event', 'attribute', 'relation'
claims (
    id UUID PRIMARY KEY,
    source_id UUID REFERENCES sources,
    claim_type TEXT NOT NULL,        -- 'event', 'attribute', 'relation'
    title TEXT NOT NULL,
    description TEXT,
    event_type TEXT,                 -- subtype when claim_type='event': 'action', 'decision', 'encounter', etc.
    date_text TEXT,                  -- original text: "that spring", "15 March"
    date_start DATE,
    date_end DATE,
    date_precision TEXT,            -- 'exact', 'month', 'season', 'year', 'approximate', 'unknown'
    chronological_position INTEGER,
    narrative_position INTEGER,
    confidence REAL NOT NULL,       -- Level 2: 0.0-1.0 assertion confidence
    confidence_reason TEXT,
    review_status TEXT NOT NULL DEFAULT 'pending',
    metadata JSONB,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ
)

-- Source references (link extractions to text)
source_references (
    id UUID PRIMARY KEY,
    chunk_id UUID REFERENCES chunks,
    claim_id UUID REFERENCES claims,
    entity_id UUID REFERENCES entities,
    relationship_id UUID REFERENCES relationships,
    excerpt TEXT NOT NULL,
    char_start INTEGER,
    char_end INTEGER,
    created_at TIMESTAMPTZ
)

-- Extracted entities
entities (
    id UUID PRIMARY KEY,
    source_id UUID REFERENCES sources,
    name TEXT NOT NULL,
    entity_type TEXT NOT NULL,      -- 'person', 'place', 'organization', 'object', 'amount'
    aliases TEXT[],
    description TEXT,
    first_appearance_chunk INTEGER,
    confidence REAL NOT NULL,
    review_status TEXT NOT NULL DEFAULT 'pending',
    metadata JSONB,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ
)

-- Relationships between entities
relationships (
    id UUID PRIMARY KEY,
    source_id UUID REFERENCES sources,
    entity_a_id UUID REFERENCES entities,
    entity_b_id UUID REFERENCES entities,
    relationship_type TEXT NOT NULL,
    description TEXT,
    start_claim_id UUID REFERENCES claims,
    end_claim_id UUID REFERENCES claims,
    confidence REAL NOT NULL,
    review_status TEXT NOT NULL DEFAULT 'pending',
    metadata JSONB,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ
)

-- Entity involvement in claims
claim_entities (
    claim_id UUID REFERENCES claims,
    entity_id UUID REFERENCES entities,
    role TEXT,                      -- 'actor', 'subject', 'witness', 'mentioned', 'location'
    PRIMARY KEY (claim_id, entity_id)
)

-- Detected inconsistencies
inconsistencies (
    id UUID PRIMARY KEY,
    source_id UUID REFERENCES sources,
    inconsistency_type TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    severity TEXT NOT NULL,
    resolution_status TEXT NOT NULL DEFAULT 'unresolved',
    resolution_note TEXT,
    metadata JSONB,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ
)

-- Items involved in an inconsistency
inconsistency_items (
    inconsistency_id UUID REFERENCES inconsistencies,
    claim_id UUID REFERENCES claims,
    entity_id UUID REFERENCES entities,
    source_reference_id UUID REFERENCES source_references,
    side TEXT,                      -- 'a', 'b' for contradictions
    PRIMARY KEY (inconsistency_id, COALESCE(claim_id, entity_id, source_reference_id))
)
```

### Two-Level Confidence Model

| Level | What | Score | Example |
|-------|------|-------|---------|
| Level 1: Source Trust | How reliable is the source itself? | `sources.source_trust` | Novel = 0.9 (consistent author), minutes = 0.95 (official record) |
| Level 2: Assertion Confidence | How confident is this specific extraction? | `claims.confidence` | "15 March 1805" = 0.95, "that spring" = 0.5 |

**Effective confidence** = source_trust * assertion_confidence (displayed to user)

---

## UX Concept

### Hero: Horizontal Dual-Lane Timeline

The primary view is a horizontal timeline with two lanes:

- **Top lane:** Events in chronological order (the actual timeline)
- **Bottom lane:** Events in narrative order (chapter sequence)
- **Connectors:** Lines between lanes link the same event. Crossed lines reveal non-linear storytelling.

Event cards show: title, date/period, entity chips, confidence marker, page reference.
Click → slide-out panel with source text.

### Filter Bar

- By entity (show only events involving a specific character)
- By confidence level (hide high-confidence, show only review items)
- By event type (battles, marriages, deaths, etc.)
- By conflict status (show only events with inconsistencies)

### Entity Panel (Sidebar)

- List of all extracted characters/places/organizations
- Each: first appearance, last appearance, event count, relationship count
- Click entity → timeline filters to their events
- Search/filter entities

### Relationship Graph

- D3 force-directed network showing entity connections
- Edge labels: married, friends, enemies, siblings, etc.
- Click relationship → shows establishing events
- Click entity node → highlights connections, filters timeline

### Inconsistency Panel

- Dedicated view listing all detected inconsistencies
- Conflict cards: type icon, both sides with sources, resolution actions
- Inline on timeline: ⚡ markers on conflicting events, dashed connector lines

### Source Text Viewer

- Slide-out panel on click of any event, entity, or reference
- Original text with relevant section highlighted
- Document name, page/chapter reference
- Cross-reference count: "Referenced in N passages"

### Review Workflow

- Status per item: `pending` → `approved` / `rejected` / `edited`
- Visible markers: ✔ ⚠️ ❓ ⚡
- Batch review: keyboard-driven (J/K navigate, A approve, R reject, E edit)
- Progress bar: "47 of 156 items reviewed"
- Queue sorted by confidence (lowest first)

### Design Direction

Polished, shareable demo quality. Clean, modern, dark-mode-friendly. Between Notion and a data visualization tool. Generous whitespace, subtle animations, color-coded event types, smooth filter transitions.

---

## Demo Experience

### Pre-loaded: Pride and Prejudice

Demo ships with Pride and Prejudice fully extracted:
- ~50-80 events on the timeline
- ~20-30 entities (characters, places)
- ~15-25 relationships
- Narrative vs. chronological mismatches identified
- Example inconsistencies seeded

User lands on the timeline immediately. No upload required to experience Sikta.

### Upload Your Own

Upload button for users to try their own books (PDF/TXT). Processing shows real-time progress.

### Why Pride and Prejudice

- Public domain (no copyright)
- Rich social network (many characters, complex relationships)
- Manageable length (~120,000 words)
- Well-known — users can verify extraction quality
- Interesting timeline: parallel storylines, time skips, letters

---

## File Structure

```
sikta/
├── CLAUDE.md                    # This file — agent entry point
├── Makefile
├── podman-compose.yml           # PostgreSQL only (local dev)
├── .env.example
├── api/                         # Go backend
│   ├── Containerfile
│   ├── go.mod
│   ├── cmd/
│   │   ├── server/              # API server entry point
│   │   └── extract/             # CLI tool for running extraction
│   ├── internal/
│   │   ├── config/              # Environment + app config
│   │   ├── database/            # sqlc generated code
│   │   ├── handlers/            # HTTP handlers (thin layer)
│   │   ├── services/            # Business logic
│   │   ├── models/              # Domain types / DTOs
│   │   ├── middleware/          # HTTP middleware (CORS, logging)
│   │   ├── extraction/          # LLM extraction pipeline
│   │   │   ├── pipeline.go      # Orchestrates the full flow
│   │   │   ├── chunker.go       # Document → chunks with positions
│   │   │   ├── extractor.go     # LLM calls for extraction
│   │   │   ├── confidence.go    # Confidence scoring
│   │   │   ├── inconsistency.go # Inconsistency detection
│   │   │   ├── resolver.go      # Entity deduplication/merging
│   │   │   └── prompts/         # LLM prompt templates
│   │   └── document/            # Document parsing (PDF, TXT)
│   └── sql/
│       ├── queries/             # sqlc query definitions
│       └── schema/              # Database migrations
├── web/                         # React frontend
│   ├── Containerfile
│   ├── package.json
│   ├── vite.config.ts
│   ├── tailwind.config.ts
│   └── src/
│       ├── components/
│       │   ├── timeline/        # Horizontal dual-lane timeline (D3)
│       │   ├── entities/        # Entity panel and cards
│       │   ├── graph/           # D3 force-directed relationship graph
│       │   ├── review/          # Review workflow components
│       │   ├── inconsistencies/ # Inconsistency panel and cards
│       │   ├── source/          # Source text viewer slide-out
│       │   └── upload/          # Document upload with progress
│       ├── pages/
│       ├── hooks/
│       ├── stores/              # Zustand stores
│       ├── api/                 # React Query API client
│       └── types/
├── demo/                        # Pre-loaded demo data
│   ├── pride-and-prejudice.txt  # Source text
│   └── seed.sql                 # Pre-extracted data
└── docs/
    ├── STATE.md
    ├── TASKS.md
    └── features/
```

---

## Development Commands

```bash
# Start everything (podman-compose + backend + frontend)
make dev

# Start infrastructure only (postgres)
make infra

# Run backend only (requires infra)
make backend

# Run frontend only
make frontend

# Run migrations
make migrate

# Create new migration
make migration name=add_something

# Run extraction on a document
make extract doc=path/to/file.txt

# Seed demo data (Pride and Prejudice)
make seed-demo

# Run tests
make test

# Build containers
make build

# Stop everything
make down

# View logs
make logs
```

---

## Container Notes

Using Podman instead of Docker:
- `podman-compose.yml` is compatible with docker-compose syntax
- Use `Containerfile` instead of `Dockerfile` (Podman convention)
- Rootless by default — no sudo needed
- `podman-compose up -d` to start, `podman-compose down` to stop

---

## Testing Strategy

### Unit Tests (Sonnet/Haiku to write)
- Chunking logic (correct page/section boundaries)
- Confidence scoring calculations
- Inconsistency detection logic
- Entity deduplication/merging
- Date normalization

### Integration Tests (Sonnet to write)
- Full extraction pipeline (text → claims + entities)
- API endpoint responses
- Database operations

### Manual Testing (Einar)
- Extraction quality against known novel content
- Timeline visualization accuracy
- Review workflow usability
- Inconsistency detection accuracy

---

## Context for AI Assistants

This is a side project for Einar, who is:
- Technical Lead managing ~100 engineers (day job)
- Building multiple SaaS ideas in parallel (Substrata, Sundla, Sikta)
- Experienced with Go, React, PostgreSQL
- Values: boring tech, extensibility, validation before scale

The MVP is a polished demo using novels. If impressive, the engine applies to business documents. The novel is a proving ground, not the final market.

Keep suggestions practical. Every hour matters.
