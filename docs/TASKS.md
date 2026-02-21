# TASKS.md

> Phased backlog. Each phase has tasks, model recommendations, and acceptance criteria.

---

## Completed MVP Phases (0-7)

All MVP phases are complete. See `docs/STATE.md` for details.

- **Phase 0:** Project Scaffolding — Go + React + PostgreSQL + Makefile
- **Phase 1:** Document Ingestion & Chunking — Structure-agnostic paragraph chunking
- **Phase 2:** LLM Extraction Pipeline — 178 events, 54 entities, 58 relationships from P&P
- **Phase 2.5:** Data Model Migration — documents->sources, events->claims, two-level confidence
- **Phase 3:** Inconsistency Detection — Narrative vs chronological, contradictions, temporal
- **Phase 4:** Timeline Hero View — D3 dual-lane timeline with connectors
- **Phase 5:** Entity Panel & Relationship Graph — D3 force-directed graph
- **Phase 6:** Review Workflow & Inconsistency Panel — Keyboard-driven J/K/A/R/E
- **Phase 7:** Demo Polish & Landing — Landing page, upload flow, seed SQL

---

## Current: Graph Model Alignment

**Context:** The data model spec defines three primitives (Node, Edge, Provenance) with strict rules. The database schema is correct. The Go service layer has 6 critical violations against the spec. This is fixable — not a restart.

**Verdict: Fix, don't restart.** Schema is sound. sqlc layer is sound. Violations are in ~5 Go files.

### What's Correct (Keep As-Is)

- `api/sql/schema/011_graph_primitives.sql` — schema matches spec (except modality CHECK)
- `api/sql/schema/012_graph_indexes.sql` — good indexes
- `api/sql/queries/nodes.sql`, `edges.sql`, `provenance.sql` — correct
- `api/internal/database/models.go` — generated, correct
- `api/internal/database/graph.go` — helper methods fine; only type declarations need changing
- `api/internal/extraction/graph/types.go` — already uses `string` for node_type/edge_type
- `api/internal/extraction/graph/service.go` — correctly stores temporal claims in provenance
- Frontend (`web/`) — no changes needed

---

### Phase G1: Open the Type System ✅ COMPLETE
**Model: Sonnet** | **Size:** S (30 min)

NodeType, EdgeType, Modality become untyped string constants. All function signatures accept `string`.

#### Tasks
- [ ] `api/internal/database/graph.go`: Remove `type NodeType string`, `type EdgeType string`, `type Modality string`. Keep consts as plain strings
- [ ] `api/internal/graph/types.go`: Change CreateNodeParams.NodeType, CreateEdgeParams.EdgeType, CreateProvenanceParams.Modality/Status to `string`
- [ ] `api/internal/graph/service.go`: Change ListNodesByType, ListEdgesByType, UpdateProvenanceStatus params to `string`. Remove `string()` casts
- [ ] `api/internal/graph/migrator.go`: Remove all `database.NodeType(...)`, `database.EdgeType(...)` casts
- [ ] `api/internal/extraction/graph/service.go`: Remove `database.NodeType()`, `database.Modality()`, `database.EdgeType()` casts
- [ ] `api/internal/graph/views.go`: Remove `string()` casts on type comparisons

#### Acceptance
- `go build ./...` passes. Zero behavior change.
- Any string value can be passed as node_type or edge_type without compile error.

---

### Phase G2: Drop Modality CHECK Constraint ✅ COMPLETE
**Model: Haiku** | **Size:** XS (10 min)

Allow arbitrary modality values in provenance.

#### Tasks
- [ ] New migration `api/sql/schema/000014_open_modality.up.sql`: `ALTER TABLE provenance DROP CONSTRAINT IF EXISTS provenance_modality_check`
- [ ] Corresponding `.down.sql` to restore the constraint

#### Acceptance
- `make migrate` succeeds.
- Can insert provenance with `modality = 'narrative_ordering'`.

---

### Phase G3: Move Ordering Data to Provenance ✅ COMPLETE
**Model: Sonnet** | **Size:** M (1.5 hr)

`narrative_position` and `chronological_position` stop being node properties. They become provenance records with ordering in `location` JSONB.

#### Design Decision

Extend `Location` struct with `PositionType` ("narrative"/"chronological") and `Position` (int). Narrative position = "where in the source text." Chronological position = "where in the inferred timeline." Both are claims with confidence/trust.

#### Tasks
- [x] `api/internal/database/graph.go`: Add `PositionType string` and `Position int` to Location struct
- [x] `api/internal/graph/migrator.go` — `MigrateClaimToNode`: Remove position from properties, create two additional provenance records (narrative + chronological ordering)
- [x] `api/internal/graph/migrator.go` — `MigrateChunkToNode`: Remove `narrative_position` from properties
- [x] `api/internal/graph/migrator.go` — `MigrateEntityToNode`: Remove `first_appearance_chunk` and `last_appearance_chunk` from properties
- [x] `api/internal/graph/views.go` — `GetEventsForTimeline`: Replace property-based position reading with provenance-based. New helper: `extractOrderingFromProvenance()`

#### Acceptance
- `go build ./...` passes
- After `MigrateDocument`: `SELECT count(*) FROM provenance WHERE location->>'position_type' = 'narrative'` returns ~178 rows
- Timeline endpoint returns correct positions

---

### Phase G4: Fix Runtime Bugs in Views ✅ COMPLETE
**Model: Sonnet** | **Size:** M (1.5 hr)

Eliminate nil pointer panics, type assertion panics. Implement relationship retrieval.

#### Tasks
- [x] `api/internal/graph/views.go`: Add nil guard after `selectProvenance` (line 75 and line 184)
- [x] `api/internal/graph/views.go`: Fix aliases type assertion — `[]interface{}` not `[]string` from JSON unmarshal
- [x] `api/internal/graph/views.go`: Extract `findDocumentNode` helper (deduplicate 3 copies of lookup pattern)
- [x] `api/sql/queries/nodes.sql`: Fix `GetDocumentNodeByLegacySourceID` param type (`$1::text`)
- [x] `api/sql/queries/edges.sql`: Add `ListEdgesBySourceDocument` query (edges with provenance from a document)
- [x] `api/internal/graph/views.go`: Implement `GetRelationshipsForGraph` using new query
- [x] Run `sqlc generate`

#### Acceptance
- No panics when nodes have empty provenance
- `GET /api/documents/{id}/relationships` returns relationship data
- `sqlc generate` succeeds

---

### Phase G5: Graph-Based Review Handlers ✅ COMPLETE
**Model: Sonnet** | **Size:** M (1.5 hr)

Replace legacy review handlers. Review status lives on provenance records.

#### Design

When a user approves a claim, they approve the **provenance record** that supports it. For frontend compatibility, the handler receives a node/edge ID, finds its provenance records, and updates their status.

#### Tasks
- [x] `api/sql/queries/provenance.sql`: Add `CountClaimProvenanceByStatusForSource`, `CountEntityProvenanceByStatusForSource`, `UpdateProvenanceStatusByTarget` queries
- [x] New file `api/internal/handlers/graph/review.go`: ReviewHandler with UpdateNodeReview, UpdateEdgeReview, UpdateNodeData, GetReviewProgress
- [x] `api/cmd/server/main.go`: Wire graph review handlers inside UseGraphModel block; legacy review handlers moved inside else block
- [x] Run `sqlc generate`

#### Acceptance
- Frontend review workflow works: approve/reject/edit claims
- Review progress bar shows correct counts
- `go build ./...` passes

---

### Phase G6: Remove Feature Flag, Clean Up Legacy ✅ COMPLETE
**Model: Sonnet** | **Size:** S (30 min)

Graph model becomes the only path.

#### Tasks
- [x] `api/cmd/server/main.go`: Remove `if cfg.UseGraphModel` conditional, keep only graph handlers
- [x] `api/internal/config/config.go`: Remove `UseGraphModel` field
- [x] Delete `api/internal/handlers/timeline.go` (replaced by graph handler)
- [x] Delete `api/internal/handlers/review.go` (replaced by graph handler)
- [x] Clean up `api/internal/handlers/extraction.go` — removed `GetEvents`, `GetEntities`, `GetRelationships`, `GetExtractionStatus`

#### Keep
- `api/internal/handlers/documents.go` — upload staging
- `api/internal/handlers/inconsistencies.go` — no graph equivalent yet
- Legacy SQL queries for sources/chunks — staging tables

#### Acceptance
- `go build ./...` passes with no feature flag
- All four frontend tabs work: Timeline, Graph, Review, Inconsistencies
- `USE_GRAPH_MODEL` not referenced anywhere

---

### Phase G7 (Future): Full Legacy Table Removal

Not in scope. Requires:
- Graph-based inconsistency detection
- Graph-based extraction replaces legacy extraction
- Migration of demo seed data to graph format
- Drop legacy tables (claims, entities, relationships, inconsistencies, etc.)

---

## Files Modified (Graph Alignment Summary)

| File | Phase | Change |
|------|-------|--------|
| `api/internal/database/graph.go` | G1, G3 | Remove typed enums, extend Location struct |
| `api/internal/graph/types.go` | G1 | Open signatures to string |
| `api/internal/graph/service.go` | G1 | Open signatures to string |
| `api/internal/graph/migrator.go` | G1, G3 | Remove casts, move ordering to provenance |
| `api/internal/graph/views.go` | G1, G3, G4 | Fix nil bugs, read ordering from provenance, implement relationships |
| `api/internal/extraction/graph/service.go` | G1 | Remove type casts |
| `api/sql/schema/000014_open_modality.up.sql` | G2 | New migration |
| `api/sql/queries/nodes.sql` | G4 | Fix parameter type |
| `api/sql/queries/edges.sql` | G4 | Add ListEdgesBySourceDocument |
| `api/sql/queries/provenance.sql` | G5 | Add review queries |
| `api/internal/handlers/graph/review.go` | G5 | New file |
| `api/cmd/server/main.go` | G5, G6 | Wire review handlers, remove flag |
| `api/internal/config/config.go` | G6 | Remove UseGraphModel |
| `api/internal/handlers/timeline.go` | G6 | Delete |
| `api/internal/handlers/review.go` | G6 | Delete |

---

## Future Phases (Post-MVP)

> **North Star Vision:** Multi-document, multi-type projects with unified timeline and cross-document anomaly detection.

### Phase 8: Extraction Progress UX
**Size:** S (1-2 hours) | **Model:** Sonnet

- [x] Backend: SSE streaming chunk-by-chunk progress
- [x] Frontend: Real-time progress bar and live counters
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
