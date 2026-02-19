# TASKS.md

> Phased backlog. Each phase has tasks, size estimates, model recommendations, and acceptance criteria.
> Total MVP estimate: ~21-28 hours across 7 phases.

---

## Phase 0: Project Scaffolding
**Size:** S (1-2 hours) | **Model:** Sonnet

Foundation: running Go backend + React frontend + PostgreSQL, all wired together.

### Tasks
- [ ] Initialize Go module with project structure (see CLAUDE.md file structure)
- [ ] Set up PostgreSQL with podman-compose.yml
- [ ] Create React app with TypeScript, Vite, Tailwind CSS
- [ ] Makefile with `dev`, `infra`, `backend`, `frontend`, `migrate`, `test`, `build`, `down` targets
- [ ] Environment configuration (.env.example, config package)
- [ ] CORS middleware configured
- [ ] Health endpoint (`GET /health`)
- [ ] sqlc configured and code generation working
- [ ] Hot-reload for backend (air) and frontend (Vite HMR)

### Acceptance Criteria
- `make dev` starts backend + frontend + database via podman-compose
- Backend responds on `GET /health` with 200
- Frontend displays "Sikta" heading and can call backend
- Database connection verified
- sqlc generates code from a test query

---

## Phase 1: Document Ingestion & Chunking
**Size:** M (3-4 hours) | **Model:** Opus for chunking strategy, Sonnet for implementation

Upload documents, extract text, split into chunks with position metadata.

### Tasks
- [ ] Database migrations: `documents`, `chunks`
- [ ] sqlc queries for documents and chunks CRUD
- [ ] TXT file parser: split by chapters/sections, track line numbers
- [ ] PDF parser: pdftotext integration with page number tracking
- [ ] Chapter detection heuristic (regex + patterns for "Chapter X", "CHAPTER X", roman numerals, etc.)
- [ ] Chunking service: document → ordered chunks with narrative_position, page_start, page_end, chapter_title
- [ ] Document upload API endpoint (`POST /api/documents`)
- [ ] Document status tracking (uploaded → processing → ready → error)
- [ ] Processing progress API (`GET /api/documents/:id/status`)
- [ ] Download and store Pride and Prejudice (Project Gutenberg TXT) in `demo/`

### Acceptance Criteria
- Upload a TXT file → document stored, text split into chapter-level chunks
- Each chunk has: content, chapter_title, chapter_number, narrative_position, page references
- Pride and Prejudice splits into ~61 chapters correctly
- Upload a PDF → text extracted with page numbers preserved
- Status endpoint shows processing progress

**The test:** Upload Pride and Prejudice. Verify chapter 1 is chunk 1, chapter 61 is the last chunk, and each chunk contains the correct text.

---

## Phase 2: LLM Extraction Pipeline ✅ COMPLETE
**Size:** L (4-5 hours) | **Model:** Opus for prompt design, Sonnet for pipeline implementation

The intelligence layer. Send chunks to Claude API, get structured claims/entities/relationships back.

### Tasks
- [x] Database migrations: `events`, `entities`, `relationships`, `event_entities`, `source_references`
- [x] sqlc queries for all extraction tables
- [x] Design extraction prompt: given a chunk, extract events, entities, relationships as structured JSON
- [x] Claude API client (Go): call Sonnet for extraction
- [x] Extraction service: iterate chunks, call LLM, parse responses, store results
- [x] Source reference creation: link every extracted item to its chunk + excerpt + char positions
- [x] Narrative position tracking: record which chapter each event comes from
- [x] Chronological position estimation: LLM-assisted ordering of events on the actual timeline
- [x] Extraction CLI tool (`make extract doc=path/to/file.txt`)
- [x] Progress tracking: events emitted as chunks are processed

### Results
- Pride and Prejudice: 61 chapters → 178 events, 54 entities, 58 relationships extracted
- Timeline Hero View working with D3 dual-lane visualization
- Inconsistency detection implemented (Phase 3 also complete)

**The test:** Pick 10 random extracted events. Click through to source text. Verify each reference points to the correct passage and the extraction is accurate.

---

## Phase 2.5: Data Model Migration
**Size:** M (3-4 hours) | **Model:** Sonnet for implementation

Rename `documents` → `sources` and `events` → `claims` throughout the codebase. Add two-level confidence model (source trust + assertion confidence) and `claim_type` discriminator.

### Tasks
- [ ] SQL migration: rename `documents` table → `sources`, add `source_trust` and `trust_reason` columns
- [ ] SQL migration: rename `events` table → `claims`, add `claim_type` column (event/attribute/relation)
- [ ] SQL migration: rename all foreign keys `document_id` → `source_id`, `event_id` → `claim_id`
- [ ] SQL migration: rename `event_entities` → `claim_entities`
- [ ] Update all sqlc queries (`.sql` files) with new table/column names
- [ ] Run `sqlc generate` to regenerate Go database layer
- [ ] Update Go handlers: `documents.go`, `extraction.go`, `timeline.go`, `inconsistencies.go`
- [ ] Update Go extraction service: `service.go`, `deduplicator.go`, `chronology.go`, `inconsistency.go`, `types.go`
- [ ] Update Go CLI tools: `cmd/extract/main.go`, `cmd/chunk/main.go`
- [ ] Update repository.go with renamed methods
- [ ] Update frontend TypeScript types (`types/index.ts`)
- [ ] Update frontend API client (`api/timeline.ts`)
- [ ] Re-upload Pride and Prejudice and re-run extraction with new schema
- [ ] Verify timeline UI works with renamed fields

### Acceptance Criteria
- All `document_id` references in code → `source_id`
- All `event_id` references in code → `claim_id` (where referring to the claims table)
- `sources` table has `source_trust` and `trust_reason` columns
- `claims` table has `claim_type` discriminator column
- Extraction pipeline produces claims with `claim_type = 'event'`
- Timeline renders correctly with renamed data
- HTTP API routes remain `/api/documents/...` (external API unchanged)

**The test:** Full pipeline: upload P&P → chunking → extraction → timeline renders. All 61 chapters, ~178 claims, ~54 entities, ~58 relationships.

---

## Phase 3: Inconsistency Detection
**Size:** M (2-3 hours) | **Model:** Opus for detection logic design, Sonnet for implementation

Detect and store contradictions, temporal impossibilities, and narrative vs. chronological mismatches.

### Tasks
- [ ] Database migrations: `inconsistencies`, `inconsistency_items` (already created, may need migration for renamed FKs)
- [ ] sqlc queries for inconsistencies CRUD
- [ ] Narrative vs. chronological ordering: compare narrative_position (chapter order) with chronological_position (timeline order) for all claims
- [ ] Contradiction detection prompt: given overlapping claims/facts, identify conflicts
- [ ] Temporal impossibility detection: flag claims where timing is logically impossible
- [ ] Cross-reference detection: identify when the same event is described multiple times
- [ ] Inconsistency service: run detection passes after extraction, store results
- [ ] Link inconsistencies to involved claims/entities via inconsistency_items
- [ ] Severity classification: info (narrative order difference), warning (ambiguity), conflict (contradiction)
- [ ] API endpoints for inconsistencies (`GET /api/documents/:id/inconsistencies`)

### Acceptance Criteria
- Narrative vs. chronological mismatches detected (if any exist in Pride and Prejudice)
- Cross-references identified (same claim mentioned in different chapters)
- Each inconsistency links to the specific claims/entities and source references involved
- Severity levels assigned appropriately
- API returns inconsistencies with full context

**The test:** Review all detected inconsistencies. Each should be genuinely interesting — either a real narrative device, an actual ambiguity, or a cross-reference worth noting. No false positives from bad parsing.

---

## Phase 4: Timeline Hero View 🎯 THE MONEY SHOT
**Size:** L (4-5 hours) | **Model:** Sonnet for implementation, Opus for D3 architecture

The horizontal dual-lane timeline. This is what people see first and what makes them want to use Sikta.

### Tasks
- [ ] API endpoints: `GET /api/documents/:id/timeline` (claims with entities, sources, inconsistencies)
- [ ] API endpoints: `GET /api/documents/:id/entities`, `GET /api/documents/:id/relationships`
- [ ] D3 horizontal timeline component: render events on a scrollable horizontal axis
- [ ] Dual-lane layout: chronological lane (top) + narrative lane (bottom)
- [ ] Connector lines between lanes linking the same event — crossed lines reveal non-linear storytelling
- [ ] Event cards: title, date/period, entity chips (colored by type), confidence marker, page reference
- [ ] Click event → source text slide-out panel (right side)
- [ ] Source text viewer: excerpt highlighted, document + page/chapter reference, cross-reference count
- [ ] Filter bar: by entity, confidence level, event type, conflict status
- [ ] Smooth zoom and pan on the timeline
- [ ] Color coding: event types get distinct colors
- [ ] Inline inconsistency markers: ⚡ icon on conflicting events, dashed connector lines between them
- [ ] Responsive: works well on large screens (primary) and degrades gracefully on smaller
- [ ] Loading state with progress animation during extraction

### Acceptance Criteria
- Timeline renders all events from Pride and Prejudice horizontally
- Dual lanes visible: chronological and narrative order
- Connectors between lanes — any crossed lines immediately visible
- Click any event → source text appears in slide-out with correct passage highlighted
- Filters work: select "Elizabeth Bennet" → only her events visible
- Confidence markers visible on every event card
- Inconsistency markers (⚡) visible on conflicting events
- Smooth zoom/pan, no jank at 50-80 events
- Looks polished enough to screenshot and share

**The test:** Show the timeline to someone unfamiliar with the project. Within 10 seconds they should understand: this is a timeline of events from a book, the two lanes show different orderings, and they can click things to see the source text.

---

## Phase 5: Entity Panel & Relationship Graph
**Size:** M (3-4 hours) | **Model:** Sonnet for implementation

Entity sidebar and D3 force-directed relationship network.

### Tasks
- [ ] Entity sidebar panel: list all entities, grouped by type (people, places, organizations)
- [ ] Entity cards: name, type icon, first/last appearance, event count, relationship count
- [ ] Click entity → timeline filters to show only their events
- [ ] Entity search/filter within sidebar
- [ ] D3 force-directed graph component: entities as nodes, relationships as edges
- [ ] Node sizing: proportional to event involvement count
- [ ] Edge labels: relationship type (married, friends, enemies, siblings, etc.)
- [ ] Edge styling: line style indicates relationship nature (solid for family, dashed for social, etc.)
- [ ] Click node → highlights connected nodes, filters timeline
- [ ] Click edge → shows establishing events in a tooltip or panel
- [ ] Graph layout controls: zoom, pan, optional force adjustment
- [ ] Toggle between timeline view and graph view (tabs or toggle)

### Acceptance Criteria
- Entity sidebar shows all extracted characters from Pride and Prejudice
- Clicking "Mr. Darcy" filters timeline to his events only
- Relationship graph shows the Bennet family connections, Darcy-Elizabeth arc, Bingley-Jane arc
- Graph is readable — not a hairball. Key characters are central, minor ones peripheral.
- Clicking a relationship edge shows the events that define it

**The test:** Someone who hasn't read Pride and Prejudice can look at the relationship graph and understand the key character dynamics — who's related to whom, who's connected through marriage, who's in conflict.

---

## Phase 6: Review Workflow & Inconsistency Panel
**Size:** M (3-4 hours) | **Model:** Sonnet for implementation

Fast human review and dedicated inconsistency exploration.

### Tasks
- [ ] Review status management: API endpoints to update review_status (approve, reject, edit)
- [ ] Batch update endpoint: approve/reject multiple items at once
- [ ] Review mode UI: dedicated view or toggle that highlights pending items
- [ ] Keyboard navigation: J/K to navigate items, A to approve, R to reject, E to edit
- [ ] Edit modal: modify event title, description, date, type, confidence
- [ ] Review progress bar: "47 of 156 items reviewed" with completion percentage
- [ ] Queue sorting: lowest confidence first (surface uncertain items)
- [ ] Filter: show only pending, only approved, only rejected
- [ ] Inconsistency panel: dedicated tab/view listing all detected inconsistencies
- [ ] Inconsistency cards: type icon, title, description, both sides with source references
- [ ] Resolution actions per inconsistency: resolve (pick correct side), note (add explanation), dismiss
- [ ] Resolution note text field
- [ ] Inconsistency status tracking: unresolved → resolved / noted / dismissed
- [ ] Visual connection: clicking an inconsistency highlights the involved events on the timeline

### Acceptance Criteria
- Can approve/reject/edit any extracted event, entity, or relationship
- Keyboard shortcuts work fluidly in review mode
- Progress bar updates as items are reviewed
- Inconsistency panel shows all detected inconsistencies with full context
- Can resolve an inconsistency by picking a side or adding a note
- Clicking an inconsistency card navigates to and highlights the relevant timeline events

**The test:** Review 20 items using only keyboard shortcuts. It should feel fast — comparable to triaging email. No mouse required for basic approve/reject flow.

---

## Phase 7: Demo Polish & Landing
**Size:** M (2-3 hours) | **Model:** Sonnet for implementation

Make it demo-ready. Pre-loaded data, landing experience, visual polish.

### Tasks
- [ ] Pre-extract Pride and Prejudice and store as seed SQL (`demo/seed.sql`)
- [ ] Seed command: `make seed-demo` loads pre-extracted data
- [ ] Landing page: hero section explaining what Sikta does, "Explore Demo" button → timeline
- [ ] Upload section on landing: drag-and-drop to try your own document
- [ ] Processing animation: show progress when extracting a new document
- [ ] Dark mode: toggle in header, persist preference
- [ ] Visual polish pass: consistent spacing, typography, animations
- [ ] Timeline zoom animation on initial load (zoom from overview to detail)
- [ ] Screenshot-friendly: timeline and graph look good in static screenshots
- [ ] Error states: graceful handling for failed uploads, extraction errors
- [ ] Mobile: basic responsiveness (timeline scrolls horizontally, panels stack)
- [ ] Favicon and page title

### Acceptance Criteria
- First visit → landing page → click "Explore Demo" → immediately see Pride and Prejudice timeline
- Upload a new TXT file → processing animation → results appear
- Dark mode works and looks good
- Screenshots of timeline and graph are impressive enough to share on social media
- No broken states or ugly error screens

**The test:** Share the URL with 3 people who don't know what Sikta is. They should understand the product within 30 seconds and find it interesting enough to explore for at least 2 minutes.

---

## Future Phases (Post-MVP)

> **North Star Vision:** Multi-document, multi-type projects with unified timeline and cross-document anomaly detection.
>
> **Phasing:**
> - **Now (MVP):** Single document (novel) → prove extraction + timeline + review works brilliantly
> - **Phase 8:** Multiple documents of same type (e.g., 5 years of board protocols)
> - **Phase 9:** Mixed document types (protocols + invoices + emails) → unified timeline with cross-document anomaly detection
> - **Phase 10:** Multiple LLM provider support (OpenAI, Azure, local models)
>
> The data model already supports this trajectory: sources table is type-agnostic, source_references link to specific chunks, claims use claim_type discriminator for extensibility, and inconsistencies can span sources via inconsistency_items.

### Phase 8: Multi-Document Projects (Same Document Type)
- **Projects table:** Group multiple documents together (e.g., 5 years of board protocols)
- **Cross-document entity resolution:** "John Smith" in protocol 2020 and protocol 2024 → same entity
- **Cross-document event deduplication:** Same event mentioned in multiple sources → one entry with multiple source references
- **Unified timeline across project:** All events from all documents on one timeline
- **Project-based access control:** All documents in a project share permissions

### Phase 9: Mixed Document Types + Cross-Document Anomaly Detection
- **Document-type-specific extraction:**
  - Board protocols → decisions, votes, action items
  - Invoices → line items, totals, due dates
  - Emails → senders, recipients, attachments
  - Interviews → questions, answers, evidence
  - Legal docs → clauses, obligations, deadlines
  - CSV/Excel → data rows, calculations, references
- **Cross-document anomaly detection:**
  - Temporal contradictions: "Event A says March 15, Event B says March 16"
  - Entity conflicts: Same name, different roles/contexts across documents
  - Missing references: Invoice mentions PO that doesn't exist in any document
  - Duplicate events: Similar events across different documents
  - Data inconsistencies: Excel total ≠ sum of rows
- **Unified event model:** All document types map to common event/entity/relationship structures
- **Anomaly resolution workflow:** Flag conflicts, show both sources, human picks correct side

### Phase 10: Multiple LLM Provider Support
- **Provider abstraction layer:** Support Anthropic, OpenAI, Azure OpenAI, local models
- **Configuration:** LLM_PROVIDER=anthropic|openai|azure|local
- **Model selection per task:** Configure which model to use for extraction, classification, chronology
- **Cost optimization:** Use cheaper models (Haiku/GPT-4o-mini) for classification, premium models (Opus/GPT-4) for complex extraction
- **Fallback and redundancy:** Failover between providers, rate limiting across multiple API keys

### Phase 11: Board Protocol Mode (Swedish Market Focus)
- Specialized extraction prompts for Swedish board protocols (styrelsebeslut, protokoll)
- Decision trail view: follow one topic across multiple meetings
- Budget tracking across protocols
- Swedish UI option

### Phase 12: Export & Sharing

### Phase 12: Export & Sharing
- Export timeline as PDF/PNG
- Export data as JSON/CSV
- Shareable link (read-only view)
- Embeddable timeline widget

### Phase 13: Authentication & Multi-Tenant
- User accounts
- Project management (multiple document collections)
- Sharing and collaboration
- Billing (Stripe)

### Phase 14: Legal / Due Diligence Mode
- Contract obligation extraction
- Deadline tracking
- Entity relationship mapping across corporate documents
- Privileged document handling

---

## Icebox

Ideas that might be interesting but are not validated:
- Real-time collaborative review (multiple reviewers)
- AI-assisted resolution suggestions for inconsistencies
- Timeline comparison (two documents side by side)
- API for programmatic access to extraction results
- Plugin system for custom extraction types
- Integration with Sundla (association documents → timeline)
