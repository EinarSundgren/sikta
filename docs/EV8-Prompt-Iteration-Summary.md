# EV8: Prompt Iteration Analysis — v4 Baseline

**Date**: 2026-02-23
**Prompt Version**: v4
**Status**: Baseline established — entity recall below threshold on 2/3 corpora

---

## Quick Summary

| Corpus | Entity Recall | Event Recall | Status |
|--------|---------------|--------------|--------|
| BRF | 86.7% (13/15) | 92.9% (13/14) | ✓ PASS (with LLM judge) |
| M&A | 64.3% (9/14) | 70.0% (7/10) | ✗ Entity FAIL |
| Police | 64.7% (11/17) | 86.4% (19/22) | ✗ Entity FAIL |

**Go/Kill Thresholds**: Entity ≥85%, Event ≥70%

**Current State**: Only BRF passes. M&A and Police miss ~20% of entities.

---

## How to Run Tests

```bash
# Build
cd /Users/einar.sundgren/projects/sikta/api
go build -o ../sikta-eval ./cmd/evaluate/

# Extract (run from project root)
./sikta-eval extract \
  --corpus corpora/brf \
  --prompt prompts/system/v4.txt \
  --fewshot prompts/fewshot/brf-v4.txt \
  --output results/brf-v4.json

# Score with LLM judge
./sikta-eval score \
  --result results/brf-v4.json \
  --manifest corpora/brf/manifest.json \
  --full \
  --model glm-5 \
  --output results/brf-v4-score.json

# View detailed results
./sikta-eval view --score results/brf-v4-score.json
```

---

## How to Review Results

1. **View summary**:
   ```bash
   ./sikta-eval view --score results/mna-v4-score.json
   ```

2. **Inspect raw JSON** for unmatched entities/events:
   ```bash
   cat results/mna-v4-score.json | jq '.EntityDetails[] | select(.ManifestID != "" and .IsCorrect == false)'
   ```

3. **Compare extracted vs manifest** to find patterns in what's being missed

---

## BRF Corpus (Passing — 86.7% Entity, 92.9% Event)

### Unmatched Entities (2/15)
- **E14**: "Storgatan 14" — Not extracted (place/address)
- **E15**: "Karlsson" — Not extracted (person mentioned as witness)

### Unmatched Events (1/14)
- **V3**: "Budget fastställd till 650 000 kr" — Not extracted

### Analysis
- BRF is the strongest corpus — few-shot examples were built from BRF source text
- Missing entities are edge cases: addresses, minor persons
- Missing event (V3) is a budget decision not explicitly stated in protocol text

**Prompt improvement for BRF**: Likely minimal gains possible. Already near ceiling.

---

## M&A Corpus (Failing — 64.3% Entity, 70.0% Event)

### Unmatched Entities (5/14)
| ID | Label | Issue |
|----|-------|-------|
| E4 | Johan Nylund | Extracted as event "Johan Nylund joined as CFO" |
| E5 | Anders Blom | Extracted as event "Anders Blom joined as VP Sales" |
| E6 | Lars Wennerström | Not extracted |
| E7 | Sofia Eriksson | Not extracted |
| E14 | SynCore | Not extracted (organization) |

### Unmatched Events (3/10)
| ID | Label | Issue |
|----|-------|-------|
| V5 | Alexei Petrov employed at Lithuania office | Not extracted |
| V6 | Alexei Petrov leaves company | Not extracted |
| V8 | Customer Alpha contract signed | Not extracted |

### Key Issues

1. **People extracted as events instead of entities**
   - "Johan Nylund joined as CFO" is an **event**, but Johan Nylund is a **person entity**
   - The prompt is extracting employment events but not creating separate person nodes for the employees
   - This is a data modeling issue: we need both the person entity AND the employment event

2. **Missing events related to Alexei Petrov**
   - V5 (employment in Lithuania), V6 (leaves company) are not in the extracted timeline
   - These may be described subtly in contracts/emails rather than explicitly stated

3. **Minor characters not extracted**
   - Lars Wennerström, Sofia Eriksson are mentioned briefly
   - May be below extraction threshold for "significant" entities

---

## Police Corpus (Failing — 64.7% Entity, 86.4% Event)

### Unmatched Entities (6/17)
| ID | Label | Issue |
|----|-------|-------|
| E7 | Daniel Falk | Not extracted |
| E11 | Marcus Öhman | Not extracted |
| E12 | Sofia Bergh | Not extracted |
| E15 | Björlandavägen 34 | Not extracted (address) |
| E16 | Kvillegatan 18 | Not extracted (address) |
| E17 | Svart Audi A4, ABC 123 | Not extracted (vehicle) |

### Unmatched Events (3/22)
| ID | Label | Issue |
|----|-------|-------|
| V14 | Person 1 returns and descends ramp again | Not extracted |
| V16 | Estimated time of death | Not extracted |
| V19 | Viktor's phone reappears on Hisingen base station | Not extracted |

### Key Issues

1. **Addresses not extracted**
   - Kvillegatan 18, Björlandavägen 34 — addresses as entities
   - These are places but not marked as significant in current prompt

2. **Vehicle as entity**
   - "Svart Audi A4, ABC 123" — a vehicle that's a key piece of evidence
   - Should be an object/entity but not extracted

3. **Minor characters**
   - Daniel Falk, Marcus Öhman, Sofia Bergh — brief mentions in witness statements

4. **Subtle timeline events**
   - V14, V16, V19 are timeline reconstruction events that may require inference

---

## Is Prompt Improvement Feasible?

### Assessment: **Limited gains expected from prompt changes alone**

### Why?

1. **Entity vs Event confusion is structural** — The LLM is extracting "Johan Nylund joined as CFO" as an event, which is correct. But the manifest expects Johan Nylund as a person entity. This requires either:
   - Two-node extraction (event + person) for every employment event
   - A post-processing pass to extract entities from events

2. **Low-mention entities are genuinely hard** — Lars Wennerström, Sofia Eriksson, Daniel Falk are mentioned 1-2 times in 50+ pages. Even humans scanning for these would need careful reading.

3. **Addresses as entities** — The current prompt focuses on "person, place, organization". An address like "Kvillegatan 18" is a place, but not extracted because it's not "significant" — it's just a location reference, not a named entity with agency.

### What Could Help

1. **Post-processing deduplication** (EV7) — Would reduce false positive rate, not recall

2. **Two-pass extraction** — First pass extracts entities/events, second pass looks for "mentions of already-extracted entities" to catch low-mention cases

3. **Entity expansion from events** — For every event like "X joined as Y", also create a person entity for X

4. **More permissive entity threshold** — Extract entities even if mentioned only once (current behavior may filter these)

5. **Entity type expansion** — Add "address" and "vehicle" as explicit node types with examples

---

## v5 Results (2026-02-23) — Entity Recall Fixed

**Changes:**
1. Added "Entity Extraction from Events" section to system prompt — every event must create entity nodes for persons/orgs/places mentioned
2. Updated confidence guidance — single-mention entities OK at 0.7-0.8 confidence
3. Expanded node types: `address`, `vehicle`, `technology`
4. Created v5 few-shot examples for M&A and Police showing these new types
5. **Fixed matcher bug** — added `place`, `address`, `vehicle`, `technology` to entity types for matching

**Results:**
| Corpus | Entity Recall | Event Recall | False Positive |
|--------|---------------|--------------|----------------|
| BRF | **100.0%** (15/15) | 71.4% (10/14) | 63.3% |
| M&A | **100.0%** (14/14) | 70.0% (7/10) | 63.6% |
| Police | **100.0%** (17/17) | 86.4% (19/22) | 64.5% |

**Key insight:** The matcher was only indexing `person` and `organization` types. Adding the new types (`place`, `address`, `vehicle`, `technology`) to the matcher immediately fixed entity recall.

**Remaining issue:** False positive rate is ~64%, needs to be <20%. This is the focus of EV8.7.

---

## Recommended Next Steps

1. ~~**Accept that 65-70% entity recall may be the ceiling**~~ — RESOLVED: v5 achieves 100% entity recall
2. **Focus on false positive reduction** — EV8.7: add hallucination guards, deduplication instructions
3. **Implement EV7** (post-processing deduplication) to improve precision
4. ~~**Consider two-pass entity extraction**~~ — Not needed, single-pass v5 works
5. ~~**Re-evaluate go/kill thresholds**~~ — Thresholds are achievable

---

## Files to Reference

- Prompts: `prompts/system/v4.txt`, `prompts/system/v5.txt`, `prompts/fewshot/brf-v4.txt`, `prompts/fewshot/mna-v5.txt`, `prompts/fewshot/police-v5.txt`
- Score results: `results/*-v5-score.json`
- Manifests: `corpora/*/manifest.json`
- CLI: `./sikta-eval` (binary)
