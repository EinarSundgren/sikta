# GLM-5 Token Limit Analysis

**Date:** 2026-02-22
**Model:** GLM-5 (via api.z.ai proxy)
**Issue:** Output truncated at 4096 tokens

---

## Summary of Document Extractions

| Document | Input Tokens | Output Tokens | Stop Reason | Error |
|----------|--------------|---------------|-------------|-------|
| A1 (protocol) | 2279 | 4096 | **max_tokens** | JSON cut off mid-stream |
| A2 (offert) | 2144 | 2324 | end_turn | Date format parsing error |
| A3 (protocol) | 639 | 4096 | **max_tokens** | JSON cut off mid-stream |
| A4 (faktura) | 534 | 3101 | end_turn | Date format parsing error |
| A5 (protocol) | 671 | 4096 | **max_tokens** | JSON cut off mid-stream |

**Total extraction time:** 5m 49s for 5 documents

---

## Two Distinct Issues

### Issue 1: Token Limit Truncation (A1, A3, A5)

**Symptoms:**
- `stop_reason=max_tokens`
- `output_tokens=4096` (exact limit)
- Error: "unexpected end of JSON input"
- JSON is valid but incomplete

**Root Cause:**
The `SendSystemPrompt` function hardcodes `MaxTokens: 4096`. GLM-5 hits this limit and stops generating mid-JSON.

**Location:** `api/internal/extraction/claude/client.go:152`
```go
func (c *Client) SendSystemPrompt(ctx context.Context, systemPrompt, userMessage string, model string) (*Response, error) {
    req := Request{
        Model:     model,
        MaxTokens: 4096,  // ← Too small for full document extraction
        System:    systemPrompt,
        Messages: []Message{
            {
                Role:    "user",
                Content: userMessage,
            },
        },
    }
    return c.SendMessage(ctx, req)
}
```

**Solution:**
Increase `MaxTokens` to 8192 or 16384. Alternatively, make it configurable via environment variable or CLI flag.

---

### Issue 2: Date Format Parsing (A2, A4)

**Symptoms:**
- `stop_reason=end_turn` (model finished naturally)
- Error: `parsing time "2023-04-22" as "2006-01-02T15:04:05Z07:00": cannot parse "" as "T"`

**Root Cause:**
GLM-5 generates dates in `YYYY-MM-DD` format, but the Go struct expects RFC3339 format (`YYYY-MM-DDTHH:MM:SSZ`).

**Location:** `api/internal/extraction/graph/types.go`
```go
type ExtractedNode struct {
    // ...
    ClaimedTimeStart  string `json:"claimed_time_start"`
    ClaimedTimeEnd    string `json:"claimed_time_end"`
    // ...
}
```

The JSON unmarshaling expects RFC3339 but GLM-5 provides simple dates.

**Example from GLM-5 response:**
```json
{
  "node_type": "event",
  "label": "Offert utfärdad",
  "claimed_time_text": "2023-04-22",
  "claimed_time_start": "2023-04-22",  // ← Not RFC3339 format
  "claimed_time_end": "2023-04-22"
}
```

**Solution:**
Change the struct tags to accept flexible date formats, or use a custom time type that handles both RFC3339 and simple dates.

---

## GLM-5 Quality Assessment

**The Good News:**
GLM-5 generates **excellent structured JSON** when it doesn't hit token limits:

✅ Correct node types (organization, person, event, value, obligation, document, place)
✅ Proper schema adherence
✅ Good entity extraction with aliases and roles
✅ Temporal extraction with claimed_time fields
✅ Appropriate confidence scores (mostly 1.0 for explicit statements)
✅ Correct modality classification (asserted, obligatory)
✅ Relevant excerpts from source text

**Example from A1 (before truncation):**
```json
{
  "node_type": "organization",
  "label": "Brf Stenbacken 3",
  "properties": {
    "type": "organization",
    "org_nr": "769612-4455",
    "aliases": ["föreningen", "styrelsen"]
  },
  "excerpt": "Brf Stenbacken 3, org.nr 769612-4455",
  "confidence": 1.0,
  "modality": "asserted"
}
```

This is high-quality extraction that matches the ground truth expectations.

---

## Recommendations

### Immediate Fix (Required for Baseline)

1. **Increase MaxTokens to 8192** in `client.go`
   - This will resolve truncation issues for A1, A3, A5
   - Cost: Minimal (we pay for tokens anyway)

2. **Fix date parsing** to accept `YYYY-MM-DD` format
   - Option A: Use custom time.Time unmarshaler
   - Option B: Change struct tags to string (already done)
   - Option C: Post-process dates to convert to RFC3339

### Long-term Improvements

1. **Make MaxTokens configurable** via environment variable
2. **Implement adaptive token limits** based on document size
3. **Add JSON schema validation** with clear error messages
4. **Consider chunking strategy** for very long documents

---

## Next Steps

1. Fix the token limit issue (1 line change)
2. Fix the date parsing issue
3. Re-run baseline extraction
4. Score results against ground truth
5. Begin prompt iteration (EV8.2)

---

## Evidence

Full logs available at:
- `results/brf-a1-full-log.txt` - Complete request/response logging
- Task outputs: `bf25208.output`, `bc141fd.output`
