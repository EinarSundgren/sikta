# Sikta — Graph Model Migration Plan

## Context

Sikta is evolving from a domain-specific relational model (sources, claims, entities, relationships) to a **universal graph-based three-primitive model** (Node, Edge, Provenance). This change enables:

1. **Universal evidence modeling** — same system for novels, police investigations, board protocols, M&A due diligence
2. **First-class contradiction handling** — conflicting claims coexist, humans decide
3. **Complete provenance traceability** — every assertion links to source with confidence
4. **Zero-schema extensibility** — new evidence types need only parser + prompt changes

**Current state:** Working MVP with Pride and Prejudice demo, timeline, entity graph, inconsistency panel.

**Estimated effort:** 30-40 hours across 8 phases

---

## The Three Primitives

### 1. Node
Anything in the evidence graph — person, place, event, document, value, obligation.

```
id            UUID
node_type     TEXT        -- open: 'entity', 'event', 'document', 'value', etc.
label         TEXT        -- human-readable
properties    JSONB       -- type-specific data
```

### 2. Edge
Directed connection between nodes — relationships, involvement, attribution, contradiction.

```
id            UUID
edge_type     TEXT        -- open: 'involved_in', 'same_as', 'asserts', etc.
source_node   UUID → Node
target_node   UUID → Node
properties    JSONB
is_negated    BOOLEAN     -- explicit denial
```

### 3. Provenance
Where any node/edge came from — source, location, confidence, modality, temporal/spatial claims.

```
target_type         TEXT        -- 'node' or 'edge'
target_id           UUID
source_id           UUID → Node
excerpt             TEXT
location            JSONB       -- page, chapter, char positions
confidence          REAL        -- 0-1, extraction confidence
trust               REAL        -- 0-1, source reliability
status              TEXT        -- pending, approved, rejected, edited
modality            TEXT        -- asserted, hypothetical, denied, conditional, inferred, obligatory, permitted
claimed_time_start  TIMESTAMPTZ
claimed_time_end    TIMESTAMPTZ
claimed_time_text   TEXT
claimed_geo_region  TEXT
claimed_geo_text    TEXT
claimed_by          UUID → Node -- who made this claim (for human decisions)
```

---

## Mapping Current Tables to New Primitives

| Current Table | New Primitive | Notes |
|---------------|---------------|-------|
| `sources` | Node (`type='document'`) | Direct mapping, file metadata in `properties` |
| `chunks` | Node (`type='chunk'`) | Text fragments become nodes |
| `claims` | Node (`type='event'/'attribute'/'relation'`) | Time/space move to Provenance |
| `entities` | Node (`type='person'/'place'/'org'/'object'`) | Direct mapping |
| `relationships` | Edge | Already edges, new `source_node`/`target_node` |
| `claim_entities` | Edge (`type='participates_in'`) | Entity participation |
| `source_references` | Provenance | All source references |
| `inconsistencies` | Node + Edges | Inconsistency node + `contradicts` edges |

---

## Implementation Phases

### Phase 1: Schema and Database Layer
**Size:** L (4-5 hours) | **Model:** Haiku (SQL generation is straightforward)

Create new tables alongside existing ones (parallel existence).

**Files to create:**
- `api/sql/schema/011_graph_primitives.sql` — nodes, edges, provenance tables
- `api/sql/schema/012_graph_indexes.sql` — performance indexes
- `api/sql/queries/nodes.sql` — CreateNode, GetNode, ListNodesByType
- `api/sql/queries/edges.sql` — CreateEdge, GetEdge, ListEdgesBySource/Target
- `api/sql/queries/provenance.sql` — CreateProvenance, GetProvenance, ListProvenanceByTarget
- `api/internal/database/graph.go` — wrapper types with convenience methods

**Tasks:**
1. Create migration files for nodes, edges, provenance with proper constraints
2. Create sqlc query files for CRUD operations
3. Run `make generate` to regenerate Go models
4. Add helper functions for graph operations

**Acceptance:** Migration applies cleanly, can insert/query via generated code, old tables still work.

---

### Phase 2: Graph Storage Service
**Size:** XL (5-7 hours) | **Model:** Sonnet (service layer design, moderate complexity)

Create service layer operating on graph model. No data migration yet.

**Files to create:**
- `api/internal/graph/service.go` — Core graph operations (CreateNode, CreateEdge, CreateProvenance, GetNodeWithProvenance)
- `api/internal/graph/migrator.go` — MigrateSourceToNode, MigrateClaimToNode, MigrateEntityToNode, MigrateRelationshipToEdge
- `api/internal/graph/views.go` — GetEventsForTimeline, GetEntitiesForGraph, ResolveIdentity with view strategies
- `api/internal/graph/types.go` — Domain types for graph operations

**Key functions:**
```go
CreateNode(node_type, label, properties) → UUID
CreateEdge(source_id, target_id, edge_type, properties, is_negated) → UUID
CreateProvenance(target_type, target_id, source_id, excerpt, confidence, modality, ...) → UUID
GetNodeWithProvenance(id) → Node + []Provenance
GetEventsForTimeline(source_id, view_strategy) → TimelineEvent[]
```

**Acceptance:** Can create nodes/edges/provenance, migrate single document, view strategies return compatible data.

---

### Phase 3: Updated Extraction Pipeline
**Size:** XL (6-8 hours) | **Model:** Opus for prompts, Sonnet for code

Create new extraction service writing directly to graph. Old service for comparison.

**Files to create:**
- `api/internal/extraction/graph/prompts.go` — Updated prompts for node/edge extraction with modality
- `api/internal/extraction/graph/service.go` — ExtractDocumentToGraph, storeExtractedNode, storeExtractedEdge
- `api/internal/extraction/graph/types.go` — GraphExtractionResponse, ExtractedNode, ExtractedEdge
- `api/cmd/migrate/main.go` — CLI to migrate existing sources to graph

**Files to modify:**
- `Makefile` — add `make migrate-to-graph` target

**Key changes:**
- Prompts extract `node_type` explicitly, return edge relationships, include `modality`
- Storage creates nodes/edges/provenance directly instead of domain tables
- Identity resolution creates `same_as` edges instead of merging

**Acceptance:** New extraction creates graph data directly, fresh document extraction works, migrate command handles P&P.

---

### Phase 4: API Adapter Layer
**Size:** L (4-5 hours) | **Model:** Sonnet (mechanical mapping, existing patterns)

Present graph data in existing API contract. No frontend changes yet.

**Files to create:**
- `api/internal/handlers/graph/timeline.go` — GetTimelineGraph() returns same TimelineEvent struct
- `api/internal/handlers/graph/entities.go` — GetEntitiesGraph() returns same Entity struct
- `api/internal/handlers/graph/relationships.go` — GetRelationshipsGraph() returns same Relationship struct

**Files to modify:**
- `api/cmd/server/main.go` — add `USE_GRAPH_MODEL` feature flag
- `api/internal/config/config.go` — read feature flag
- `web/src/types/index.ts` — add graph types, keep old types for compatibility
- `.env.example` — add `USE_GRAPH_MODEL=true/false`

**Approach:** Feature flag routes to graph handlers or legacy handlers. Same JSON response format.

**Acceptance:** Toggle via env var, frontend works identically with both backends, performance within 2x.

---

### Phase 5: Frontend Graph Components
**Size:** XL (5-7 hours) | **Model:** Sonnet (React components, established patterns)

Add new UI capabilities enabled by graph model.

**Files to create:**
- `web/src/components/graph/IdentityPanel.tsx` — Shows entity identity claims (same_as edges), confirm/reject
- `web/src/components/graph/ProvenancePanel.tsx` — All provenance for selected node/edge, view strategy selector
- `web/src/components/graph/ProvenanceViewer.tsx` — Source text with related claims
- `web/src/hooks/useGraphData.ts` — Hook for graph queries

**Files to modify:**
- `web/src/components/graph/RelationshipGraph.tsx` — Use direct graph data, add edge type filter, show same_as edges
- `web/src/api/timeline.ts` — Add `useGraph` flag to API calls
- `web/src/pages/TimelineView.tsx` — Integrate new panels

**Features:**
- Identity panel with duplicate entity candidates
- Provenance panel with modality/confidence/status filters
- View strategy selector (trust-weighted, majority, human-decided)
- Relationship graph with edge type filtering

**Acceptance:** Identity panel shows duplicates, provenance panel shows all sources, view strategies work.

---

### Phase 6: Human Review as Claims
**Size:** M (3-4 hours) | **Model:** Sonnet (CRUD operations, established patterns)

Human review actions create graph nodes/edges with provenance.

**Files to create:**
- `api/internal/handlers/graph/review.go` — PostReviewDecision, GetReviewHistory
- `web/src/components/review/ReviewHistory.tsx` — Show review history per item

**Files to modify:**
- `web/src/components/review/ReviewPanel.tsx` — Show review history
- `api/internal/graph/views.go` — Update view strategies for human decisions

**Approach:**
- Approve/reject creates 'review_action' node
- Links via 'performed_by', 'approves', 'rejects' edges
- 'human-decided' view strategy prioritizes human-reviewed provenance

**Acceptance:** Each approve/reject creates audit trail, can see who decided what, view strategies respect decisions.

---

### Phase 7: Data Migration and Cutover
**Size:** M (3-4 hours) | **Model:** Sonnet (data transformation, well-defined mapping)

Migrate all existing data, cutover production.

**Files to create:**
- `api/internal/graph/migrator_full.go` — Complete migration script

**Files to modify:**
- `Makefile` — add `make backup-db`, `make rollback-graph`
- `demo/seed.sql` — regenerate with graph data
- `.env` — set `USE_GRAPH_MODEL=true`

**Migration order:**
1. sources → document nodes
2. chunks → chunk nodes + provenance
3. claims → event/attribute/relation nodes + provenance
4. entities → entity nodes + provenance
5. relationships → edges + provenance
6. claim_entities → participates_in edges
7. inconsistencies → inconsistency nodes + contradicts edges

**Rollback plan:** Keep legacy tables, can revert by changing `USE_GRAPH_MODEL=false`.

**Acceptance:** All P&P data migrated, timeline/graph/entity panel work, rollback tested.

---

### Phase 8: Legacy Cleanup (Post-Cutover)
**Size:** S (1-2 hours) | **Model:** Haiku (deletion and removal)

Remove feature flag, delete legacy code, cleanup migration.

**Files to create:**
- `api/sql/schema/013_cleanup_legacy.sql` — Drop legacy tables (after verification period)

**Files to modify:**
- All handler files — remove legacy code paths
- Documentation — update to reflect graph model

---

## Critical Implementation Details

### View Strategies (Query-Time Logic)
- **Single source:** Use one provenance record (novel analysis)
- **Trust-weighted:** Highest `trust * confidence` wins (default multi-source)
- **Majority:** Most sources agree (investigation consensus)
- **Human-decided:** `status='approved'` wins (reviewed evidence base)
- **Conflict:** Show all claims, highlight disagreement (audit, due diligence)

### Values: Literals vs Nodes
- **Default:** Values are literals in Edge `properties` (e.g., `{amount: 450000, currency: "SEK"}`)
- **Exception:** Value earns node only when contested or independently referenced
- **Rationale:** Prevents supernodes (10k invoices pointing to same "100.00 SEK" node)

### Identity Resolution
- **Never merge:** Different sources referring to "J. Doe" vs "John Doe" create separate entity nodes
- **Instead:** Create `same_as` edge with provenance and confidence
- **Human decides:** Identity panel shows candidates, user confirms/rejects

### Time and Space are Claims
- Not stored on Node — stored in Provenance records
- Multiple sources can disagree on when/where — all coexist
- Timeline applies view strategy at query time

---

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Data loss | Backup before migration, keep legacy tables until verified |
| Performance | Indexes, query optimization, caching layer for view strategies |
| Frontend breakage | API adapter maintains same contract, gradual rollout |
| Extraction regression | Run old and new in parallel, compare results |

---

## User Decisions (Resolved)

1. **View strategy persistence:** Per-user preference (saved default, with per-query override option)
2. **Identity display:** Primary entity shown by default, with expandable IdentityPanel for duplicate candidates and same_as edges
3. **Performance target:** <500ms for timeline, <1s for entity graph

## Model Recommendations per Phase

| Phase | Model | Rationale |
|-------|-------|-----------|
| 1 | Haiku | SQL schema generation is straightforward, well-defined patterns |
| 2 | Sonnet | Graph service logic requires moderate complexity, design patterns |
| 3 | Opus (prompts), Sonnet (code) | Prompt design needs LLM expertise; code implementation is straightforward |
| 4 | Sonnet | API adapter is mechanical mapping work |
| 5 | Sonnet | Frontend component implementation, React patterns |
| 6 | Sonnet | Review workflow is straightforward CRUD |
| 7 | Sonnet | Migration script is data transformation, well-defined |
| 8 | Haiku | Cleanup is deletion and removal |

---

## Files to Create (Summary)

**Schema/Database:**
- `api/sql/schema/011_graph_primitives.sql`
- `api/sql/schema/012_graph_indexes.sql`
- `api/sql/queries/nodes.sql`
- `api/sql/queries/edges.sql`
- `api/sql/queries/provenance.sql`

**Backend:**
- `api/internal/database/graph.go`
- `api/internal/graph/service.go`
- `api/internal/graph/migrator.go`
- `api/internal/graph/migrator_full.go`
- `api/internal/graph/views.go`
- `api/internal/graph/types.go`
- `api/internal/extraction/graph/prompts.go`
- `api/internal/extraction/graph/service.go`
- `api/internal/extraction/graph/types.go`
- `api/internal/handlers/graph/timeline.go`
- `api/internal/handlers/graph/entities.go`
- `api/internal/handlers/graph/relationships.go`
- `api/internal/handlers/graph/review.go`
- `api/cmd/migrate/main.go`

**Frontend:**
- `web/src/components/graph/IdentityPanel.tsx`
- `web/src/components/graph/ProvenancePanel.tsx`
- `web/src/components/graph/ProvenanceViewer.tsx`
- `web/src/components/review/ReviewHistory.tsx`
- `web/src/hooks/useGraphData.ts`

---

## Files to Modify (Summary)

**Backend:**
- `Makefile` — add migrate-to-graph, backup-db, rollback-graph targets
- `api/cmd/server/main.go` — add USE_GRAPH_MODEL feature flag
- `api/internal/config/config.go` — read feature flag

**Frontend:**
- `web/src/types/index.ts` — add graph types
- `web/src/components/graph/RelationshipGraph.tsx` — use direct graph data
- `web/src/api/timeline.ts` — add useGraph flag
- `web/src/pages/TimelineView.tsx` — integrate new panels
- `web/src/components/review/ReviewPanel.tsx` — show review history

**Config:**
- `.env.example` — add USE_GRAPH_MODEL

---

## Verification

After each phase:
1. Run existing tests: `make test`
2. Check timeline loads with Pride and Prejudice
3. Verify extraction works with fresh document
4. Performance check: timeline <500ms, entity graph <1s

After Phase 7 (full migration):
1. All P&P events appear in timeline
2. Entity graph shows same connections
3. Inconsistency panel works
4. Can approve/reject items
5. Rollback procedure tested

---

## Effort Summary

| Phase | Description | Model | Size | Hours |
|-------|-------------|-------|------|-------|
| 1 | Schema and Database Layer | Haiku | L | 4-5 |
| 2 | Graph Storage Service | Sonnet | XL | 5-7 |
| 3 | Updated Extraction Pipeline | Opus/Sonnet | XL | 6-8 |
| 4 | API Adapter Layer | Sonnet | L | 4-5 |
| 5 | Frontend Graph Components | Sonnet | XL | 5-7 |
| 6 | Human Review as Claims | Sonnet | M | 3-4 |
| 7 | Data Migration and Cutover | Sonnet | M | 3-4 |
| 8 | Legacy Cleanup | Haiku | S | 1-2 |
| **Total** | | | | **30-40 hours** |

**Starting Options:**
- **Phase 1 only** (4-5h) — Validate database approach, minimal commitment
- **Phases 1-3** (15-20h) — Backend graph extraction ready, can compare with legacy
- **Phases 1-4** (19-25h) — Full backend with feature flag, frontend can toggle models
