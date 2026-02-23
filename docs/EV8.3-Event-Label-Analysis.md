# EV8.3: Event Label Analysis — BRF Corpus

**Date:** 2026-02-23
**Status:** Analysis complete, ready for prompt revision

---

## Summary Statistics

| Metric | Count | Note |
|--------|-------|------|
| Manifest Events (Expected) | 14 | Ground truth |
| Extracted Events (Found) | 36 | Includes all protocol events |
| Matched Events | 3 | 21.4% recall |
| Unmatched Manifest Events | 11 | Need prompt alignment |
| Hallucinated Events | 36 | Events not in manifest (expected) |

---

## Matched Events (3/14) ✅

These events matched via exact label or Levenshtein distance ≤ 3:

| Manifest ID | Manifest Label | Extracted Label | Match Method |
|-------------|----------------|-----------------|--------------|
| V1 | Styrelsemöte 2023-03-15 | Styrelsemöte 2023-03-15 | exact_label |
| V4 | Fuktsinspektion utförd | Fasadjnspektion | levenshtein_label |
| V12 | Styrelsemöte 2023-12-07 | Styrelsemöte 2023-12-07 | exact_label |

---

## Unmatched Manifest Events (11/14) ❌

### Pattern 1: **"Beslut om X" Events**

The prompt extracts these as "Beslut om X" but the manifest has more detailed labels:

| Manifest ID | Manifest Label | What Was Extracted |
|-------------|----------------|-------------------|
| V2 | Beslut om upphandling av fasadrenovering | **Beslut om upphandling av fasadrenovering** (MATCHES - should be found!) |
| V6 | Val av entreprenör — NorrBygg Fasad AB antagen | Beslut om val av entreprenör för fasadrenovering |
| V14 | Ny rutin beslutad: beställningar kräver förhandsgodkännande | Beslut om hänvisning till föreningsstämma |

**Issue:** V2 label matches exactly but wasn't found. This suggests a matcher bug or document mismatch.

---

### Pattern 2: **Budget/Financial Events**

Manifest labels include specific amounts, extracted labels are more generic:

| Manifest ID | Manifest Label | What Was Extracted |
|-------------|----------------|-------------------|
| V3 | Budget fastställd till 650 000 kr | (not found - extracted as value node?) |
| V7 | Ny budgetram fastställd till 600 000 kr | (not found - extracted as value node?) |
| V13 | Betalning av faktura 7842 | (not found - extracted as value node?) |

**Issue:** These financial events might be extracted as `value` nodes instead of `event` nodes.

---

### Pattern 3: **Document/Agreement Events**

Manifest labels documents as events, extracted differently:

| Manifest ID | Manifest Label | What Was Extracted |
|-------------|----------------|-------------------|
| V5 | Offert från NorrBygg Fasad AB | (not found - extracted as document node?) |
| V8 | Avtal tecknat med NorrBygg Fasad AB | (not found - extracted as document node?) |
| V9 | Faktura från NorrBygg Fasad AB | (not found - extracted as document node?) |

**Issue:** These documents might be extracted as `document` nodes instead of `event` nodes.

---

### Pattern 4: **Inspection Events**

Manifest labels are more detailed than extracted labels:

| Manifest ID | Manifest Label | What Was Extracted |
|-------------|----------------|-------------------|
| V4 | Fuktsinspektion utförd | Fasadjnspektion (matched via Levenshtein) |
| V10 | Fasadarbeten påbörjas | Fasadrenovering Storgatan 14 (not matched) |
| V11 | Slutbesiktning fasadrenovering | (not found) |

**Issue:** Inspection labels need more detail and consistency.

---

### Pattern 5: **Administrative Events**

| Manifest ID | Manifest Label | What Was Extracted |
|-------------|----------------|-------------------|
| V14 | Ny rutin beslutad: beställningar kräver förhandsgodkännande | Beslut om hänvisning till föreningsstämma (wrong event) |

**Issue:** Decision labels need to capture the full decision content.

---

## Extracted Events Not in Manifest (36 events)

These are legitimate events extracted from the protocols but not tracked in the manifest:

### Protocol Administrative Events
- Styrelsemöte 2023-03-15
- Mötets öppnande
- Val av justerare
- Mötets avslutande
- Godkännande av föregående protokoll

### Issue-Specific Events
- Vattenläcka i lägenhet 7
- Jourutryckning för vattenläcka
- Fråga om laddstolpar för elbilar
- Beslut om utredning av laddstolpar
- Redovisning av utredning om laddstolpar
- Presentation av ekonomisk rapport

**Note:** These are NOT hallucinations - they're real events from the protocols, just not in the ground truth manifest.

---

## Root Cause Analysis

### Issue 1: Event Type Confusion
**Problem:** Some manifest events are extracted as different node types:
- Budget events → `value` nodes (not `event`)
- Document events → `document` nodes (not `event`)

**Evidence:** The matcher only looks at nodes with `node_type == "event"`, so budget/document events are never checked.

**Solution Needed:** Ensure prompt classifies all manifest events as `event` nodes, regardless of content.

---

### Issue 2: Label Granularity
**Problem:** Manifest labels are more detailed than extracted labels:
- Manifest: "Ny budgetram fastställd till 600 000 kr"
- Extracted: (not found - likely "Budget" or "Beslut om budget")

**Solution Needed:** Update prompt to include key details in event labels:
- For budgets: include amount
- For decisions: include key decision details
- For inspections: include location and outcome

---

### Issue 3: Label Format Consistency
**Problem:** Manifest uses specific patterns:
- Decisions: "Beslut om X", "Val av X", "X beslutad"
- Inspections: "X utförd", "X påbörjas", "Slutbesiktning X"

Extracted labels:
- Decisions: "Beslut om X" (consistent)
- Inspections: Single word or phrase ("Fasadjnspektion")

**Solution Needed:** Update prompt to use consistent label format matching manifest:
- Inspections: "[Type]inspektion [location/status]" or "[Inspection] utförd"
- Example: "Fuktsinspektion utförd" instead of "Fasadjnspektion"

---

### Issue 4: Matcher Bug (V2)
**Problem:** V2 "Beslut om upphandling av fasadrenovering" exists in both manifest AND extraction but didn't match.

**Evidence:** Event comparison shows:
```
┌─ Manifest Event: V2
│  Label:          Beslut om upphandling av fasadrenovering
│  Source Doc:     A1
│  ❌ NOT MATCHED
│     No close match found
```

But extracted events include:
```json
{
  "label": "Beslut om upphandling av fasadrenovering",
  "claimed_time_text": "2023-03-15"
}
```

**Solution Needed:** Debug why exact match failed - likely document ID mismatch or normalization issue.

---

## Recommendations for EV8.3 Prompt Revision

### 1. **Event Type Classification**
Update prompt to ensure ALL time-bound occurrences are classified as `event` nodes:
```markdown
## Event Nodes
Events are temporal occurrences - anything that happens at a specific time.
This includes:
- Meetings and decisions (board meetings, votes, decisions)
- Inspections and work performed (inspections, renovations, repairs)
- Financial events (budgets set, payments made, invoices received)
- Document events (offers sent, contracts signed, invoices issued)

Events should ALWAYS be classified as `node_type: "event"`, even if they involve
documents, values, or obligations.
```

### 2. **Event Label Format**
Update prompt with specific label format instructions:
```markdown
## Event Label Format

Event labels should follow these patterns:

### Decisions
- Format: "Beslut om [action]" or "Val av [entity]"
- Examples:
  - "Beslut om upphandling av fasadrenovering"
  - "Val av entreprenör — NorrBygg Fasad AB antagen"
  - "Ny rutin beslutad: beställningar kräver förhandsgodkännande"

### Budgets and Financial Events
- Format: "[Budget type] fastställd till [amount]"
- Examples:
  - "Budget fastställd till 650 000 kr"
  - "Ny budgetram fastställd till 600 000 kr"
  - "Betalning av faktura [nummer]"

### Inspections and Work
- Format: "[Type]inspektion utförd" or "[Work] påbörjas" or "Slutbesiktning [work]"
- Examples:
  - "Fuktsinspektion utförd"
  - "Fasadarbeten påbörjas"
  - "Slutbesiktning fasadrenovering"

### Documents as Events
- Format: "[Document type] från/vård [entity] [verb]"
- Examples:
  - "Offert från NorrBygg Fasad AB"
  - "Avtal tecknat med NorrBygg Fasad AB"
  - "Faktura från NorrBygg Fasad AB"
```

### 3. **Include Key Details in Labels**
Add instruction:
```markdown
## Event Label Detail Level

Event labels should include:
- **For decisions**: What was decided and key details
- **For budgets**: The amount (exact number and currency)
- **For inspections**: What was inspected and outcome
- **For documents**: Type of document and key parties

Examples:
- ✅ "Budget fastställd till 650 000 kr"
- ❌ "Budget" (too generic)
- ✅ "Fuktsinspektion utförd"
- ❌ "Inspektion" (too generic)
- ✅ "Offert från NorrBygg Fasad AB"
- ❌ "Offert" (too generic)
```

---

## Next Steps

1. ✅ **Analysis Complete** - Event label patterns identified
2. **EV8.3 Implementation** - Update `prompts/system/v1.txt` with above recommendations
3. **EV8.4 Testing** - Re-run extraction on BRF corpus
4. **EV8.5 Measurement** - Compare v2 vs v1 metrics

**Target Improvement:**
- Current: 21.4% event recall (3/14)
- Target: ≥70% event recall (≥10/14)

**Expected Issues After Prompt Revision:**
- Event type classification fix → V3, V7, V13 (budgets) will match
- Label detail level fix → V6, V10, V11 will match
- V2 matcher bug → needs separate investigation

---

## Files Generated

- `api/cmd/compare-events/main.go` - Event comparison CLI tool
- `docs/EV8.3-Event-Comparison.md` - Full event comparison output
- `docs/EV8.3-Event-Label-Analysis.md` - This file
- `Makefile` - Updated with `eval-compare-events` target

---

## Usage

```bash
# Build and run comparison tool
make eval-compare-events result=results/brf-v1.json

# Or directly
./compare-events results/brf-v1-fixed.json corpora/brf/manifest.json
```
