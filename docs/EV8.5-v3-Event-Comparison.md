╔════════════════════════════════════════════════════════════════════════════════╗
║                    EVENT COMPARISON: MANIFEST vs EXTRACTION                   ║
╚════════════════════════════════════════════════════════════════════════════════╝

Extraction Result: results/brf-v3.json
Manifest:          corpora/brf/manifest.json
Corpus:             brf
Prompt Version:     v3.txt

════════════════════════════════════════════════════════════════════════════════
SUMMARY STATISTICS
════════════════════════════════════════════════════════════════════════════════

Manifest Events (Expected):     14
Extracted Events (Found):        36
Matched Events (Recall):         8 / 14 (57.1%)

Match Methods:
  exact_label: 8

Hallucinated Events:            36 (extracted but not in manifest)

════════════════════════════════════════════════════════════════════════════════

════════════════════════════════════════════════════════════════════════════════
MANIFEST EVENTS (Ground Truth)
════════════════════════════════════════════════════════════════════════════════

┌─ Manifest Event: V14
│  Label:          Ny rutin beslutad: beställningar kräver förhandsgodkännande
│  Type:           decision
│  Source Doc:     A5
│  Time:           2023-12-07
│  Entities:       E1, E8
│  ❌ NOT MATCHED
│     Best Candidate:   
│     No close match found
└───────────────────────────────────────────────────────────────────────────────

┌─ Manifest Event: V1
│  Label:          Fuktsinspektion utförd
│  Type:           inspection
│  Source Doc:     A1
│  Time:           12 januari 2023
│  Entities:       E12
│  ❌ NOT MATCHED
│     Best Candidate:   
│     No close match found
└───────────────────────────────────────────────────────────────────────────────

┌─ Manifest Event: V2
│  Label:          Beslut om upphandling av fasadrenovering
│  Type:           decision
│  Source Doc:     A1
│  Time:           2023-03-15
│  Entities:       E1, E8
│  ✅ MATCHED
│     Extracted Label:  Beslut om upphandling av fasadrenovering
│     Match Method:     exact_label
│     Match Score:      1.00
│     Entities Matched: 2/2
└───────────────────────────────────────────────────────────────────────────────

┌─ Manifest Event: V3
│  Label:          Budget fastställd till 650 000 kr
│  Type:           decision
│  Source Doc:     A1
│  Time:           2023-03-15
│  Entities:       E1, E8
│  ❌ NOT MATCHED
│     Best Candidate:   
│     No close match found
└───────────────────────────────────────────────────────────────────────────────

┌─ Manifest Event: V5
│  Label:          Offert från NorrBygg Fasad AB
│  Type:           document
│  Source Doc:     A2
│  Time:           2023-04-22
│  Entities:       E9, E8, E1
│  ❌ NOT MATCHED
│     Best Candidate:   
│     No close match found
└───────────────────────────────────────────────────────────────────────────────

┌─ Manifest Event: V7
│  Label:          Ny budgetram fastställd till 600 000 kr
│  Type:           decision
│  Source Doc:     A3
│  Time:           2023-05-10
│  Entities:       E1, E8
│  ✅ MATCHED
│     Extracted Label:  Ny budgetram fastställd till 600 000 kr
│     Match Method:     exact_label
│     Match Score:      1.00
│     Entities Matched: 2/2
└───────────────────────────────────────────────────────────────────────────────

┌─ Manifest Event: V8
│  Label:          Avtal tecknat med NorrBygg Fasad AB
│  Type:           agreement
│  Source Doc:     A4
│  Time:           2023-05-28
│  Entities:       E1, E9, E8
│  ❌ NOT MATCHED
│     Best Candidate:   
│     No close match found
└───────────────────────────────────────────────────────────────────────────────

┌─ Manifest Event: V11
│  Label:          Slutbesiktning fasadrenovering
│  Type:           inspection
│  Source Doc:     A5
│  Time:           22 november 2023
│  Entities:       E9, E8
│  ✅ MATCHED
│     Extracted Label:  Slutbesiktning fasadrenovering
│     Match Method:     exact_label
│     Match Score:      1.00
│     Entities Matched: 2/2
└───────────────────────────────────────────────────────────────────────────────

┌─ Manifest Event: V12
│  Label:          Slutfaktura NorrBygg Fasad AB
│  Type:           financial
│  Source Doc:     A4
│  Time:           2023-11-02
│  Entities:       E9, E8
│  ✅ MATCHED
│     Extracted Label:  Slutfaktura NorrBygg Fasad AB
│     Match Method:     exact_label
│     Match Score:      1.00
│     Entities Matched: 2/2
└───────────────────────────────────────────────────────────────────────────────

┌─ Manifest Event: V4
│  Label:          Vattenläcka lägenhet 7
│  Type:           incident
│  Source Doc:     A1
│  Time:           före 2023-03-15
│  Entities:       E13, E15
│  ✅ MATCHED
│     Extracted Label:  Vattenläcka lägenhet 7
│     Match Method:     exact_label
│     Match Score:      1.00
│     Entities Matched: 2/2
└───────────────────────────────────────────────────────────────────────────────

┌─ Manifest Event: V6
│  Label:          Val av entreprenör — NorrBygg Fasad AB antagen
│  Type:           decision
│  Source Doc:     A3
│  Time:           2023-05-10
│  Entities:       E1, E9, E8
│  ✅ MATCHED
│     Extracted Label:  Val av entreprenör — NorrBygg Fasad AB antagen
│     Match Method:     exact_label
│     Match Score:      1.00
│     Entities Matched: 3/3
└───────────────────────────────────────────────────────────────────────────────

┌─ Manifest Event: V9
│  Label:          Byggstart fasadrenovering
│  Type:           construction
│  Source Doc:     A2
│  Time:           v.32, 7 augusti 2023
│  Entities:       E9
│  ❌ NOT MATCHED
│     Best Candidate:   
│     No close match found
└───────────────────────────────────────────────────────────────────────────────

┌─ Manifest Event: V10
│  Label:          Per Sandberg beställer tilläggsarbeten utan styrelsebeslut
│  Type:           unauthorized_order
│  Source Doc:     A4
│  Time:           2023-09-14
│  Entities:       E5, E9
│  ✅ MATCHED
│     Extracted Label:  Per Sandberg beställer tilläggsarbeten utan styrelsebeslut
│     Match Method:     exact_label
│     Match Score:      1.00
│     Entities Matched: 2/2
└───────────────────────────────────────────────────────────────────────────────

┌─ Manifest Event: V13
│  Label:          Tilläggsarbeten godkända i efterhand av styrelsen
│  Type:           decision
│  Source Doc:     A5
│  Time:           2023-12-07
│  Entities:       E1, E5, E8
│  ✅ MATCHED
│     Extracted Label:  Tilläggsarbeten godkända i efterhand av styrelsen
│     Match Method:     exact_label
│     Match Score:      1.00
│     Entities Matched: 3/3
└───────────────────────────────────────────────────────────────────────────────

════════════════════════════════════════════════════════════════════════════════
HALLUCINATED EVENTS (Extracted but not in Manifest)
════════════════════════════════════════════════════════════════════════════════

┌─ Extracted Event: 
│  Label:           Styrelsemöte 2023-03-15
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Mötets öppnande
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Val av justerare
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Protokollet från styrelsemöte 2023-01-20 godkändes
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Ekonomisk rapport presenterad
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Fuktskador konstaterade
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Inspektion utförd av Byggkonsult Norrland AB
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Beslut om upphandling av fasadrenovering
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Vattenläcka lägenhet 7
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Jourutryckning Sundsvalls Rör AB
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Fråga om laddstolpar för elbilar i garaget
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Beslut att utreda frågan om laddstolpar
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Mötets avslutande
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Byggstart fasadrenovering v.32
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Styrelsemöte 2023-05-10
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Mötets öppnande
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Val av justerare
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Protokollet från 2023-03-15 godkändes
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Ekonomisk rapport presenterad
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Val av entreprenör — NorrBygg Fasad AB antagen
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Bemyndigande teckna avtal med NorrBygg Fasad AB
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Ny budgetram fastställd till 600 000 kr
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Fråga om laddstolpar hänvisad till föreningsstämman 2024
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Mötets avslutande
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Föreningsstämma 2024
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Slutfaktura NorrBygg Fasad AB
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Per Sandberg beställer tilläggsarbeten utan styrelsebeslut
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Mötets öppnande
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Val av justerare
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Slutbesiktning fasadrenovering
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Per Sandberg beställer tilläggsarbeten utan styrelsebeslut
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Tilläggsarbeten godkända i efterhand av styrelsen
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Kassören rekommenderar avgiftshöjning om 10%
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Årsstämma 2024 planeras
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Maria Bergström bokar lokal
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Mötets avslutande
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

════════════════════════════════════════════════════════════════════════════════
ALL EXTRACTED EVENTS (Reference)
════════════════════════════════════════════════════════════════════════════════

Document: A1
────────────────────────────────────────────────────────────────────────────────
  [] Beslut att utreda frågan om laddstolpar
       Time: 2023-03-15
  [] Beslut om upphandling av fasadrenovering
       Time: 2023-03-15
  [] Ekonomisk rapport presenterad
       Time: 2023-03-15
  [] Fråga om laddstolpar för elbilar i garaget
       Time: 2023-03-15
  [] Fuktskador konstaterade
       Time: 12 januari 2023
  [] Inspektion utförd av Byggkonsult Norrland AB
       Time: 12 januari 2023
  [] Jourutryckning Sundsvalls Rör AB
       Time: 2023-03-15
  [] Mötets avslutande
       Time: kl. 19:45
  [] Mötets öppnande
       Time: kl. 18:30
  [] Protokollet från styrelsemöte 2023-01-20 godkändes
       Time: 2023-03-15
  [] Styrelsemöte 2023-03-15
       Time: 2023-03-15
  [] Val av justerare
       Time: 2023-03-15
  [] Vattenläcka lägenhet 7
       Time: 2023-03-15

Document: A2
────────────────────────────────────────────────────────────────────────────────
  [] Byggstart fasadrenovering v.32
       Time: v.32 (7 augusti 2023)

Document: A3
────────────────────────────────────────────────────────────────────────────────
  [] Bemyndigande teckna avtal med NorrBygg Fasad AB
       Time: 2023-05-10
  [] Ekonomisk rapport presenterad
       Time: 2023-05-10
  [] Fråga om laddstolpar hänvisad till föreningsstämman 2024
       Time: 2023-05-10
  [] Föreningsstämma 2024
       Time: 2024
       Confidence: 0.95
  [] Mötets avslutande
       Time: kl. 19:30
  [] Mötets öppnande
       Time: kl. 18:15
  [] Ny budgetram fastställd till 600 000 kr
       Time: 2023-05-10
  [] Protokollet från 2023-03-15 godkändes
       Time: 2023-05-10
  [] Styrelsemöte 2023-05-10
       Time: 2023-05-10
  [] Val av entreprenör — NorrBygg Fasad AB antagen
       Time: 2023-05-10
  [] Val av justerare
       Time: 2023-05-10

Document: A4
────────────────────────────────────────────────────────────────────────────────
  [] Per Sandberg beställer tilläggsarbeten utan styrelsebeslut
       Time: 2023-09-14
  [] Slutfaktura NorrBygg Fasad AB
       Time: 2023-11-02

Document: A5
────────────────────────────────────────────────────────────────────────────────
  [] Kassören rekommenderar avgiftshöjning om 10%
       Time: 2023-12-07
  [] Maria Bergström bokar lokal
  [] Mötets avslutande
       Time: 2023-12-07 kl. 20:15
  [] Mötets öppnande
       Time: 2023-12-07 kl. 18:00
  [] Per Sandberg beställer tilläggsarbeten utan styrelsebeslut
       Time: före 2023-12-07
  [] Slutbesiktning fasadrenovering
       Time: 22 november 2023
  [] Tilläggsarbeten godkända i efterhand av styrelsen
       Time: 2023-12-07
  [] Val av justerare
       Time: 2023-12-07
  [] Årsstämma 2024 planeras
       Time: 2024-04-25 kl. 19:00

