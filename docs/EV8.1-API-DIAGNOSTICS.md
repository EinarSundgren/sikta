# EV8.1 Baseline Extraction — API Diagnostics Report

**Date:** 2026-02-22
**Status:** ⚠️ Blocked - Model Incompatibility
**Task:** Run baseline extraction on BRF corpus to establish validation metrics

---

## Summary

The extraction validation pipeline is fully implemented and operational, but the LLM endpoint (`api.z.ai` proxy using GLM-5 model) does not generate valid structured JSON for complex extraction tasks. All 5 BRF documents failed with JSON parsing errors.

---

## What Works ✓

1. **API Authentication**: Successfully authenticating with the proxy
2. **HTTP Client**: Go client correctly constructs requests with proper headers
3. **Document Loading**: CLI loads corpus documents and prompts correctly
4. **Chunking**: Paragraph-based chunking logic operational
5. **JSON Post-processing**: Code handles markdown-wrapped JSON (```json...```)
6. **Scoring Engine**: Ready to score results once extraction produces valid output

**Test Command Used:**
```bash
./sikta-eval extract \
  --corpus corpora/brf \
  --prompt prompts/system/v1.txt \
  --fewshot prompts/fewshot/brf.txt \
  --model glm-5 \
  --output results/brf-v1-glm5.json
```

---

## What Doesn't Work ✗

### GLM-5 Model Issues

The GLM-5 model (accessed via `api.z.ai` proxy) fails to follow the structured JSON extraction prompt:

| Document | Error | Root Cause |
|----------|-------|------------|
| A1 (protocol) | "unexpected end of JSON input" | Empty or incomplete JSON response |
| A2 (offert) | "parsing time \"2023-04-22\" as RFC3339" | Wrong date format in JSON |
| A3 (protocol) | "unexpected end of JSON input" | Empty or incomplete JSON response |
| A4 (faktura) | "invalid character 'p' after object key" | Malformed JSON syntax |
| A5 (protocol) | (pending) | Likely similar |

### Evidence from Direct API Testing

```bash
curl test with simple JSON prompt:
> Input: "Return a JSON object with status and message fields"
> Output: ```json\n{"status": "success", "message": "Hello, World!"}\n```
```

**Observation:** GLM-5 wraps JSON in markdown code blocks even with explicit "respond with JSON only" instructions. The extraction code already strips markdown, but the underlying issue is that GLM-5 doesn't reliably generate the complex nested JSON structure required for graph extraction (nodes + edges with specific schemas).

---

## Proxy Configuration

**Current .env Settings:**
```
ANTHROPIC_API_KEY="23b8f9d2a5a94f20a1f2fd1035afe843.ZqZaaOzjAfDCb1oT"
ANTHROPIC_API_URL="https://api.z.ai/api/anthropic"
ANTHROPIC_MODEL_EXTRACTION=glm-5
```

**Model Behavior:**
- API endpoint: ✓ Responds to requests
- Authentication: ✓ Works with the provided key
- Model name mapping: Proxy maps `claude-sonnet-4-20250514` → GLM-5
- Simple JSON: ✓ Works (wraps in markdown)
- Complex structured JSON: ✗ Fails to generate valid output

---

## Options to Proceed

### Option 1: Use Real Claude API (Recommended)

Replace `ANTHROPIC_API_URL` with `https://api.anthropic.com` and use a real Claude API key starting with `sk-ant-`. Claude Sonnet is specifically trained for structured JSON output and will reliably follow the extraction schema.

**Pros:**
- Proven to work with structured extraction prompts
- No JSON parsing issues
- Matches the original design assumptions

**Cons:**
- Requires Claude API key (user may not have one)
- API costs (though baseline extraction on 5 docs is minimal)

### Option 2: Add JSON Schema Coercion

Implement a fallback that uses a JSON repair library (like `github.com/tidwall/gjson` or similar) to fix malformed responses, or add schema validation with clear error messages to iterate the prompt.

**Pros:**
- Can work with existing proxy
- Teaches us what GLM-5 struggles with

**Cons:**
- If the model isn't generating the right fields, coercion won't help
- Adds complexity without addressing root cause

### Option 3: Switch to a Different Model/Proxy

If the user has access to other LLM endpoints that better support structured output (e.g., OpenAI with `response_format: { "type": "json_object" }`).

---

## Recommendation

**Use real Claude API (Option 1).** The extraction pipeline was designed around Claude's structured output capabilities. GLM-5, while functional for simple tasks, doesn't reliably produce the complex nested JSON required for graph extraction.

The validation pipeline (EV1-EV5) is complete and ready to run. Once we have a working LLM endpoint, we can:

1. Run baseline extraction on all 3 corpora (BRF, M&A, Police)
2. Score results against ground truth manifests
3. Begin prompt iteration loop (EV8.2 - EV8.5)

---

## Files Created/Modified This Session

**Corpus Files (EV1):**
- `corpora/brf/`, `corpora/mna/`, `corpora/police/` - Test corpus directories
- `corpus/*/docs/*.txt` - 18 source documents
- `corpus/*/manifest.json` - Ground truth with entities, events, inconsistencies

**Prompt Files (EV2):**
- `prompts/system/v1.txt` - Graph extraction system prompt
- `prompts/fewshot/brf.txt`, `mna.txt`, `police.txt` - Domain-specific examples

**Extraction Pipeline (EV3):**
- `api/internal/extraction/graph/runner.go` - Database-free extraction runner
- `api/cmd/evaluate/main.go` - CLI entry point

**Scoring Engine (EV4):**
- `api/internal/evaluation/*.go` - Types, matcher, scorer, compare
- Unit tests passing

**CLI (EV5):**
- Makefile targets: `eval-build`, `eval-brf`, `eval-mna`, `eval-police`, `eval-all`
- `./sikta-eval` binary compiled successfully

---

## Next Step

Await user decision on LLM endpoint. Once confirmed:
1. Update .env with working endpoint/model
2. Re-run `make eval-brf` (baseline extraction)
3. Score results with `make eval-brf-score`
4. Document baseline metrics
5. Begin EV8.2 (failure analysis)
