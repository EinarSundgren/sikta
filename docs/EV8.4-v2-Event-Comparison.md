╔════════════════════════════════════════════════════════════════════════════════╗
║                    EVENT COMPARISON: MANIFEST vs EXTRACTION                   ║
╚════════════════════════════════════════════════════════════════════════════════╝

Extraction Result: results/brf-v2.json
Manifest:          corpora/brf/manifest.json
Corpus:             brf
Prompt Version:     v2.txt

════════════════════════════════════════════════════════════════════════════════
SUMMARY STATISTICS
════════════════════════════════════════════════════════════════════════════════

Manifest Events (Expected):     14
Extracted Events (Found):        45
Matched Events (Recall):         7 / 14 (50.0%)

Match Methods:
  exact_label: 6
  levenshtein_label: 1

Hallucinated Events:            45 (extracted but not in manifest)

════════════════════════════════════════════════════════════════════════════════

════════════════════════════════════════════════════════════════════════════════
MANIFEST EVENTS (Ground Truth)
════════════════════════════════════════════════════════════════════════════════

┌─ Manifest Event: V5
│  Label:          Offert från NorrBygg Fasad AB
│  Type:           document
│  Source Doc:     A2
│  Time:           2023-04-22
│  Entities:       E9, E8, E1
│  ✅ MATCHED
│     Extracted Label:  Offert från NorrBygg Fasad AB
│     Match Method:     exact_label
│     Match Score:      1.00
│     Entities Matched: 3/3
└───────────────────────────────────────────────────────────────────────────────

┌─ Manifest Event: V7
│  Label:          Ny budgetram fastställd till 600 000 kr
│  Type:           decision
│  Source Doc:     A3
│  Time:           2023-05-10
│  Entities:       E1, E8
│  ✅ MATCHED
│     Extracted Label:  Budgetram fastställd till 600 000 kr
│     Match Method:     levenshtein_label
│     Match Score:      0.70
│     Entities Matched: 2/2
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

┌─ Manifest Event: V1
│  Label:          Fuktsinspektion utförd
│  Type:           inspection
│  Source Doc:     A1
│  Time:           12 januari 2023
│  Entities:       E12
│  ✅ MATCHED
│     Extracted Label:  Fuktsinspektion utförd
│     Match Method:     exact_label
│     Match Score:      1.00
│     Entities Matched: 1/1
└───────────────────────────────────────────────────────────────────────────────

┌─ Manifest Event: V6
│  Label:          Val av entreprenör — NorrBygg Fasad AB antagen
│  Type:           decision
│  Source Doc:     A3
│  Time:           2023-05-10
│  Entities:       E1, E9, E8
│  ❌ NOT MATCHED
│     Best Candidate:   
│     No close match found
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
│  ❌ NOT MATCHED
│     Best Candidate:   
│     No close match found
└───────────────────────────────────────────────────────────────────────────────

┌─ Manifest Event: V13
│  Label:          Tilläggsarbeten godkända i efterhand av styrelsen
│  Type:           decision
│  Source Doc:     A5
│  Time:           2023-12-07
│  Entities:       E1, E5, E8
│  ❌ NOT MATCHED
│     Best Candidate:   
│     No close match found
└───────────────────────────────────────────────────────────────────────────────

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
│  Label:           Mötets öppnande kl. 18:30
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Val av justerare — Erik Johansson vald
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Protokoll 2023-01-20 godkänt
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
│  Label:           Fuktsinspektion utförd
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Fuktskador konstaterade vid fasad
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
│  Label:           Kostnad: 4 200 kr
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Fråga om laddstolpar för elbilar lyft
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Beslut om utredning av laddstolpar
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Mötets avslutande kl. 19:45
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Offert från NorrBygg Fasad AB
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Beräknad byggstart v.32 (7 augusti 2023)
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
│  Label:           Mötets öppnande kl. 18:15
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Val av justerare — Jonas Åkesson
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Godkännande av protokoll 2023-03-15
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
│  Label:           Offert från NorrBygg Fasad AB — 525 000 kr
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Offert från Sundsvalls Bygg AB — 610 000 kr
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Offert från Fasadgruppen Nord — 495 000 kr
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Beslut om val av entreprenör — NorrBygg Fasad AB antagen
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Bemyndigande att teckna avtal med NorrBygg Fasad AB
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Beslut om finansiering från underhållsfond
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Budgetram fastställd till 600 000 kr
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Redovisning av utredning om laddstolpar
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Beslut om hänvisning till föreningsstämma 2024
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Mötets avslutande kl. 19:30
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
│  Label:           Fasadrenovering — slutfaktura
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
│  Label:           Styrelsemöte 2023-12-07
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Mötets öppnande kl. 18:00
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Val av justerare — Maria Bergström vald
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
│  Label:           Fasadrenovering inklusive puts, fogning och fönsterbänkar
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
│  Label:           Beslut om godkännande av tilläggsarbetena i efterhand
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Kassören påpekar kraftig sänkning av likviditet
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Rekommendation om avgiftshöjning 10% från 2024-01-01
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Planering av årsstämma 2024-04-25 kl. 19:00
│  Excerpt:         Datum: 2023-03-15
│  Time:            2023-03-15
│  Confidence:      1.00
│  ⚠️  NO MATCH IN MANIFEST
└───────────────────────────────────────────────────────────────────────────────

┌─ Extracted Event: 
│  Label:           Mötets avslutande kl. 20:15
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
  [] Beslut om upphandling av fasadrenovering
       Time: 2023-03-15
  [] Beslut om utredning av laddstolpar
       Time: 2023-03-15
  [] Ekonomisk rapport presenterad
       Time: 2023-03-15
  [] Fråga om laddstolpar för elbilar lyft
       Time: 2023-03-15
  [] Fuktsinspektion utförd
       Time: 12 januari 2023
  [] Fuktskador konstaterade vid fasad
       Time: 12 januari 2023
  [] Jourutryckning Sundsvalls Rör AB
       Time: 2023-03-15
  [] Kostnad: 4 200 kr
       Time: 2023-03-15
  [] Mötets avslutande kl. 19:45
       Time: kl. 19:45
  [] Mötets öppnande kl. 18:30
       Time: kl. 18:30
  [] Protokoll 2023-01-20 godkänt
       Time: 2023-03-15
  [] Styrelsemöte 2023-03-15
       Time: 2023-03-15
  [] Val av justerare — Erik Johansson vald
       Time: 2023-03-15
  [] Vattenläcka lägenhet 7
       Time: 2023-03-15

Document: A2
────────────────────────────────────────────────────────────────────────────────
  [] Beräknad byggstart v.32 (7 augusti 2023)
       Time: v.32 (7 augusti 2023)
  [] Offert från NorrBygg Fasad AB
       Time: 2023-04-22

Document: A3
────────────────────────────────────────────────────────────────────────────────
  [] Bemyndigande att teckna avtal med NorrBygg Fasad AB
       Time: 2023-05-10
  [] Beslut om finansiering från underhållsfond
       Time: 2023-05-10
  [] Beslut om hänvisning till föreningsstämma 2024
       Time: 2023-05-10
  [] Beslut om val av entreprenör — NorrBygg Fasad AB antagen
       Time: 2023-05-10
  [] Budgetram fastställd till 600 000 kr
       Time: 2023-05-10
  [] Ekonomisk rapport presenterad
       Time: 2023-05-10
  [] Föreningsstämma 2024
       Time: 2024
       Confidence: 0.90
  [] Godkännande av protokoll 2023-03-15
       Time: 2023-05-10
  [] Mötets avslutande kl. 19:30
       Time: kl. 19:30
  [] Mötets öppnande kl. 18:15
       Time: kl. 18:15
  [] Offert från Fasadgruppen Nord — 495 000 kr
       Time: 2023-05-10
  [] Offert från NorrBygg Fasad AB — 525 000 kr
       Time: 2023-05-10
  [] Offert från Sundsvalls Bygg AB — 610 000 kr
       Time: 2023-05-10
  [] Redovisning av utredning om laddstolpar
       Time: 2023-05-10
  [] Styrelsemöte 2023-05-10
       Time: 2023-05-10
  [] Val av justerare — Jonas Åkesson
       Time: 2023-05-10

Document: A4
────────────────────────────────────────────────────────────────────────────────
  [] Fasadrenovering — slutfaktura
  [] Per Sandberg beställer tilläggsarbeten utan styrelsebeslut
       Time: 2023-09-14

Document: A5
────────────────────────────────────────────────────────────────────────────────
  [] Beslut om godkännande av tilläggsarbetena i efterhand
       Time: 2023-12-07
  [] Fasadrenovering inklusive puts, fogning och fönsterbänkar
       Time: innan 2023-11-22
       Confidence: 0.90
  [] Kassören påpekar kraftig sänkning av likviditet
       Time: 2023-12-07
  [] Mötets avslutande kl. 20:15
       Time: 2023-12-07 kl. 20:15
  [] Mötets öppnande kl. 18:00
       Time: 2023-12-07 kl. 18:00
  [] Per Sandberg beställer tilläggsarbeten utan styrelsebeslut
       Time: innan 2023-12-07
  [] Planering av årsstämma 2024-04-25 kl. 19:00
       Time: 2024-04-25 kl. 19:00
  [] Rekommendation om avgiftshöjning 10% från 2024-01-01
       Time: 2023-12-07
  [] Slutbesiktning fasadrenovering
       Time: 22 november 2023
  [] Styrelsemöte 2023-12-07
       Time: 2023-12-07
  [] Val av justerare — Maria Bergström vald
       Time: 2023-12-07

