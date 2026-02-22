# Sikta — Data Model & Extraction Pipeline Brief

> For continuing work in a new conversation. Focuses on the core data model and the next critical challenge: the extraction pipeline.

---

## What Sikta Is (30-second version)

An evidence synthesis engine. Takes fragmented documents, extracts structured claims as a graph (nodes + edges + provenance), and detects contradictions across sources. Every claim is traceable to its source. The human decides what to trust — the system never resolves truth.

**MVP:** Novel demo (Jekyll & Hyde). **Target:** Corporate evidence synthesis (M&A, legal, audit, board governance).

**Stack:** Go backend, React frontend(s), PostgreSQL, Claude API for extraction, D3.js for visualization.

---

## The Data Model (Reviewed, Approved)

Three primitives. Everything else is a query pattern over these three.

### Node — anything that exists

```
id            UUID
node_type     text        -- open: 'entity', 'event', 'value', 'document',
                          --        'claim', 'obligation' (never an enum)
label         text        -- "Henry Jekyll", "Board Meeting 2019-03"
properties    jsonb       -- type-specific extensible data
```

No time, no location. These are claims — they live in Provenance.

### Edge — any directed connection

```
id            UUID
edge_type     text        -- open: 'involved_in', 'same_as', 'has_value',
                          --        'asserts', 'contradicts', 'precedes',
                          --        'caused_by', 'bound_by'
source_node   UUID → Node
target_node   UUID → Node
properties    jsonb       -- role, weight, amounts as literals
is_negated    boolean     -- explicit denial only (source says "NOT")
                          -- absence = no provenance record, not is_negated
```

Values (amounts, dates) are literals in Edge `properties` by default. Promoted to nodes only when contested.

### Provenance — where anything came from

```
id                  UUID
target_type         text        -- 'node' or 'edge'
target_id           UUID
source_id           UUID → Node -- source document (itself a node)

excerpt             text        -- relevant text passage
location            jsonb       -- { page, section, char_start, char_end, chapter }

confidence          real        -- 0.0–1.0 extraction confidence
trust               real        -- 0.0–1.0 source reliability
status              text        -- 'pending', 'approved', 'rejected', 'edited'

modality            text        -- 'asserted', 'hypothetical', 'denied',
                                --  'conditional', 'inferred',
                                --  'obligatory', 'permitted'

claimed_time_start  timestamptz
claimed_time_end    timestamptz
claimed_time_text   text        -- raw: "last spring", "around 10pm"
claimed_geo_region  text
claimed_geo_text    text        -- raw: "at the corner of the street"
```

### Why This Model

1. **No resolved values anywhere.** Even human review is a provenance record. Timeline applies view strategies at query time (trust-weighted, majority, human-decided, conflict mode).
2. **Time and space are claims.** Two sources can disagree on when/where. Both coexist in Provenance.
3. **Identity is a claim.** "John Doe" and "J. Doe" are two nodes linked by Edge(same_as) with its own confidence.
4. **Seven modalities** cover assertion, denial, hypothesis, condition, inference, obligation, and permission.
5. **Adding a new document type = new parser + new extraction prompt. Zero schema changes.**

### Stress-tested against:

- "In a Grove" (seven contradictory murder accounts)
- "An Occurrence at Owl Creek Bridge" (hallucinated vs. real event sequences)
- "The Lady, or the Tiger?" (unresolved hypotheticals)
- Police investigation (CCTV, phone records, unreliable witnesses)
- BRF board protocols + invoices (three different amounts for one project)
- M&A due diligence (non-compete waiver vs. surviving IP assignment)

All scenarios model cleanly with three primitives.

---

## The Extraction Pipeline — The Next Critical Challenge

The data model is sound. Whether the LLM can *populate it accurately* from messy text is the unproven bet. This is the Opus-level work.

### Pipeline Architecture

```
Document Upload
    │
    ▼
Text Extraction (PDF → pdftotext, TXT → direct read)
    │
    ▼
Chunking (split by chapter/section, preserve page + position metadata)
    │
    ▼
LLM Extraction (Claude Sonnet API)
    Input: text chunk + context (document title, chapter, position)
    Output: structured JSON → nodes, edges, provenance records
    │
    ▼
Post-Processing
    → Entity deduplication (Edge: same_as with confidence)
    → Temporal ordering (compare claimed_time across provenance)
    → Inconsistency detection (contradictions, temporal impossibilities)
    │
    ▼
Human Review Queue (sorted by confidence, lowest first)
    │
    ▼
Evidence Graph → Timeline, Entity Panel, Relationship Graph
```

### The Hard Problems

#### 1. Prompt Design — Extraction

The extraction prompt must take a text chunk and produce structured JSON matching the three-primitive model. This is the single highest-leverage design decision remaining.

**What the prompt must extract per chunk:**

- **Nodes:** entities (people, places, orgs, objects), events, values — each with node_type, label, properties
- **Edges:** relationships between extracted nodes — each with edge_type, directionality, properties, is_negated
- **Provenance:** for every node and edge — excerpt, location in chunk, confidence score, modality, temporal claims (claimed_time_start/end/text), spatial claims (claimed_geo_region/text)

**What makes this hard:**

- The LLM must assign confidence honestly, not uniformly high
- Temporal expressions are wildly varied: "that spring", "15 March 1805", "before the ball", "three days after the murder", "the following year"
- Entity coreference: "Elizabeth", "Lizzy", "Miss Bennet", "she" may all be the same person within a chunk — but across chunks the LLM doesn't have context
- Modality detection: distinguishing "X happened" from "X might have happened" from "X was supposed to happen" requires nuanced reading
- The prompt must produce output that maps cleanly to the database schema — node_type and edge_type strings must be consistent across chunks

**Approach considerations:**

- Single mega-prompt per chunk vs. multi-pass (extract entities first, then events, then relationships)?
- How much context to include? Just the chunk? Chapter title + chunk? Previous chunk summary?
- Output format: flat JSON? Nested? How to handle entity references across chunks?

#### 2. Chunking Strategy

How to split a document into chunks that are:
- Small enough for the LLM context window
- Large enough to preserve meaning (don't split mid-scene)
- Trackable to their source position (page, section, chapter)

For novels: chapter boundaries are natural. For board protocols: agenda items / sections. For contracts: clauses. For emails: individual messages in a thread.

The chunker needs to be document-type-aware but the chunk output format must be universal (chunk_index, content, chapter_title, page_start, page_end, narrative_position).

#### 3. Entity Deduplication Across Chunks

The LLM processes one chunk at a time. It might extract "Henry Jekyll" from Chapter 1 and "Dr. Jekyll" from Chapter 3 as separate entities. The system needs to:

1. Detect potential duplicates (string similarity, alias patterns, LLM-assisted matching)
2. Create Edge(same_as) between them with confidence
3. NOT merge them — identity resolution is a claim, not a fact

Options:
- Post-processing pass: collect all entities, run a dedup prompt
- Running context: pass accumulated entity list to each chunk extraction as context
- Both: extract naively, then deduplicate with a focused prompt

#### 4. Temporal Normalization

"That spring" in a novel set in 1811 → `claimed_time_start: 1811-03-01, claimed_time_end: 1811-05-31, claimed_time_text: "that spring"`.

This requires the LLM to:
- Identify temporal expressions
- Resolve them against the story's internal chronology
- Assign appropriate precision markers
- Handle relative references ("three days later" = three days after what?)

For corporate documents this is easier (dates are usually explicit) but for novels it's the hardest extraction task.

#### 5. Inconsistency Detection

After extraction, detect:
- **Same event, different claims:** Two provenance records for the same node with different claimed_time or claimed_geo → flag
- **Temporal impossibility:** Entity at two locations with overlapping claimed_time → flag
- **Contradicting values:** Same entity with two different has_value edges → flag
- **Narrative displacement:** Compare provenance.location.chapter (narrative position) with claimed_time ordering (evidence position) → reveals non-linear storytelling

Some detection is query-based (SQL). Some requires LLM reasoning (is this genuinely contradictory, or are they talking about different things?).

#### 6. Confidence Calibration

The LLM tends to over-assign confidence. The extraction prompt must:
- Define explicit criteria for each confidence tier
- Use examples showing what 0.9 vs. 0.5 vs. 0.2 looks like
- Penalize uniform high confidence
- Separately score: entity identification confidence, temporal claim confidence, relationship confidence

### Model Selection for Extraction Work

| Task | Model | Rationale |
|------|-------|-----------|
| Extraction prompt design | Opus | Highest-leverage decision. Quality compounds across every document. |
| Extraction prompt iteration | Opus | Prompt refinement based on test results |
| Chunking logic | Sonnet | Implementation task with clear spec |
| Entity dedup prompt | Opus | Nuanced matching, high impact on quality |
| Confidence calibration | Opus | Requires careful reasoning about uncertainty |
| Temporal normalization | Sonnet | Structured task, well-defined input/output |
| API client / pipeline orchestration | Sonnet | Standard Go implementation |
| Inconsistency detection queries | Sonnet | SQL + straightforward logic |

### Extraction Output Format (Draft)

The LLM should return JSON matching this structure per chunk:

```json
{
  "nodes": [
    {
      "temp_id": "n1",
      "node_type": "entity",
      "label": "Henry Jekyll",
      "properties": { "entity_type": "person", "description": "London doctor" }
    },
    {
      "temp_id": "n2",
      "node_type": "event",
      "label": "Jekyll begins duality research",
      "properties": { "event_type": "discovery" }
    }
  ],
  "edges": [
    {
      "edge_type": "involved_in",
      "source": "n1",
      "target": "n2",
      "properties": { "role": "actor" },
      "is_negated": false
    }
  ],
  "provenance": [
    {
      "target": "n2",
      "excerpt": "I was born in the year 18— to a large fortune...",
      "char_start": 0,
      "char_end": 142,
      "confidence": 0.7,
      "modality": "asserted",
      "claimed_time_text": "~15 years before main events",
      "claimed_time_start": null,
      "claimed_time_end": null,
      "claimed_geo_region": "London",
      "claimed_geo_text": null
    }
  ]
}
```

`temp_id` fields are chunk-local references resolved to UUIDs during ingestion. `source_id` and `trust` are set by the pipeline (inherited from the document node), not the LLM.

---

## Project Status

- ✅ Data model: reviewed, approved (v0.2)
- ✅ UX/UI: design system complete, interactive prototype exists
- ✅ Architecture: dual frontend, shared API client, phase plan
- ✅ All planning documents written (CLAUDE.md, TASKS.md, STATE.md, UPDATES.md, DESIGN_SYSTEM.md, SIKTA_DATA_MODEL.md)
- ⬜ Phase 0: Project scaffolding (not started)
- ⬜ Phase 1: Document ingestion & chunking
- ⬜ **Phase 2: Extraction pipeline — THIS IS NEXT AND CRITICAL**

---

## Reference Documents

| Document | Content |
|----------|---------|
| CLAUDE.md | Full agent instructions, architecture, file structure, dev commands |
| docs/TASKS.md | 7-phase backlog with acceptance criteria |
| docs/STATE.md | Current progress tracker |
| docs/UPDATES.md | All post-planning decisions (evidence synthesis reframing, two-level architecture, three primitives, dual frontend, reviewer feedback) |
| docs/SIKTA_DATA_MODEL.md | Canonical data model with stress tests and peer review |
| docs/DESIGN_SYSTEM.md | Complete UX/UI spec (both editions) |
| docs/HANDOFF.md | General project handoff (broader scope than this document) |
