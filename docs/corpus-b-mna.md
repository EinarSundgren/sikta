# Test Corpus B: M&A Due Diligence — Acquisition of DataSyn Nordic AB

## About This Corpus

A private equity firm is evaluating the acquisition of DataSyn Nordic AB, a 45-person SaaS company in Gothenburg selling data integration tools to mid-market companies. Six documents from the data room. Contains 9 planted inconsistencies.

**Language:** English (standard for Nordic M&A)
**Extraction challenges:** Contradicting financials across documents, obligation conflicts, undisclosed liabilities, temporal gaps, entity resolution across legal and trade names.

---

## Document B1: Company Overview (Management Presentation)

```
CONFIDENTIAL — PROJECT NORDIC

DataSyn Nordic AB — Management Presentation
Prepared for: Potential Acquirer
Date: November 2024

COMPANY OVERVIEW

DataSyn Nordic AB ("DataSyn" or "the Company") is a Gothenburg-based SaaS 
company providing data integration and transformation tools to mid-market 
companies across the Nordics.

Founded: 2017 by CEO Marcus Lindgren and CTO Helena Friberg
Employees: 45 (38 in Sweden, 7 in Lithuania)
Headquarters: Gothenburg, Sweden
Legal entity: DataSyn Nordic AB, org.nr 559187-2234

KEY FINANCIALS (MSEK)
                    2021    2022    2023    2024E
Revenue             12.4    18.7    26.1    34.5
ARR (Dec)           14.1    21.3    28.9    36.0
Gross Margin        82%     83%     84%     85%
EBITDA              -1.2    1.8     4.1     7.2
Headcount (avg)     28      34      40      45

CUSTOMER BASE
- 127 active customers (Dec 2023)
- Average contract value: 228 kSEK/year
- Net Revenue Retention: 118%
- Logo churn: 8% annually
- Top 10 customers: 34% of ARR

TECHNOLOGY
Platform built on Kubernetes, deployed on AWS (eu-north-1, Stockholm).
Core IP: proprietary transformation engine ("SynCore") written in Rust.
14 granted patents related to data mapping algorithms.

TEAM
Marcus Lindgren, CEO — Co-founder, previously VP Sales at Qlik
Helena Friberg, CTO — Co-founder, previously Principal Engineer at Spotify  
Johan Nylund, CFO — Joined 2021, previously controller at PE-backed SaaS
Anders Blom, VP Sales — Joined 2022, manages 8-person sales team

The Company has received SEK 15M in venture funding:
  Seed round (2018): 5 MSEK from NorthCap Ventures
  Series A (2020): 10 MSEK from NorthCap Ventures + Baltic Innovation Fund
```

---

## Document B2: Audited Financial Statements 2023

```
DATASYN NORDIC AB
ANNUAL REPORT 2023
Org.nr: 559187-2234

INCOME STATEMENT (KSEK)
                              2023        2022
Net Revenue                  26,847      18,732
Cost of Revenue              -4,563      -3,184
───────────────────────────────────────────────
Gross Profit                 22,284      15,548
Gross Margin                   83.0%       83.0%

Sales & Marketing            -8,412      -6,230
Research & Development       -7,893      -5,811
General & Administration     -2,340      -1,892
───────────────────────────────────────────────
Operating Profit (EBIT)       3,639       1,615
EBIT Margin                   13.6%        8.6%

Financial items                -312        -187
───────────────────────────────────────────────
Profit Before Tax             3,327       1,428
Tax                            -685        -294
───────────────────────────────────────────────
Net Profit                    2,642       1,134

BALANCE SHEET (KSEK)
                              2023        2022
ASSETS
Intangible assets             4,230       3,150
  Capitalized development     3,890       2,810
  Patents                       340         340
Property & equipment            567         423
Right-of-use assets           2,100       2,100
Financial assets                 45          45
Total non-current assets      6,942       5,718

Accounts receivable           3,456       2,187
Prepaid expenses                890         654
Cash and cash equivalents     8,234       6,412
Total current assets         12,580       9,253
───────────────────────────────────────────────
TOTAL ASSETS                 19,522      14,971

EQUITY AND LIABILITIES
Share capital                   500         500
Reserves                     11,687       9,045
Total equity                 12,187       9,545

Non-current liabilities
  Lease liability              1,575       2,100
  Deferred tax liability         234         187

Current liabilities
  Accounts payable             1,234         876
  Accrued expenses             2,456       1,543
  Deferred revenue             1,456         456
  Current lease liability        380         264
Total current liabilities      5,526       3,139
───────────────────────────────────────────────
TOTAL EQUITY & LIABILITIES   19,522      14,971

NOTES

Note 4 — Revenue
Revenue is recognized over the contract period. Annual contracts 
constitute 89% of revenue, monthly contracts 11%.

Note 7 — Capitalized Development
The Company capitalizes development costs for SynCore platform
enhancements meeting IAS 38 criteria. Amortization: 3 years straight-line.
Total capitalized in 2023: 1,420 kSEK.

Note 11 — Contingent Liabilities
The Company is party to a dispute with former employee regarding IP 
assignment. The Company's legal counsel assesses the risk as low. 
No provision has been made.

Note 12 — Related Party Transactions
The Company leases office premises from Lindgren Fastigheter AB, 
a company owned by CEO Marcus Lindgren. Annual lease: 840 kSEK.
Terms assessed as market-rate by the board.

Note 14 — Employees
Average number of employees: 42 (2022: 35).
  Sweden: 36 (2022: 32)
  Lithuania: 6 (2022: 3)
```

---

## Document B3: Employment Agreement — CTO (Helena Friberg)

```
EMPLOYMENT AGREEMENT

Between:  DataSyn Nordic AB, org.nr 559187-2234 ("the Company")
And:      Helena Friberg, 830415-XXXX ("the Employee")

Date: 2017-06-01
Amended: 2022-03-15

1. POSITION
The Employee serves as Chief Technology Officer (CTO), reporting to the 
CEO. The Employee is a member of the executive management team.

2. COMPENSATION
Monthly salary: 95,000 SEK (as of 2022-03-15 amendment)
Pension: ITP1 according to collective agreement
Variable: Annual bonus of up to 3 monthly salaries based on company 
performance targets set by the Board.

3. EQUITY
The Employee holds 22% of the shares in DataSyn Nordic AB through 
direct ownership. Subject to the Shareholders' Agreement dated 2020-09-01.

4. INTELLECTUAL PROPERTY
All inventions, software, designs, and other intellectual property created 
by the Employee during the course of employment, or using Company resources, 
shall be the exclusive property of the Company.

This includes work performed outside regular working hours if related to the 
Company's business area (data integration and transformation technology).

5. NON-COMPETE
Upon termination of employment, the Employee agrees not to:
  a) Engage in competing business for a period of 18 months
  b) Solicit Company customers for a period of 24 months
  c) Recruit Company employees for a period of 12 months

Competing business is defined as: development, sale, or distribution of 
data integration, data transformation, or ETL software products.

Compensation during non-compete period: 60% of base salary.

6. CONFIDENTIALITY
The Employee shall not disclose confidential information during or after 
employment. This obligation survives termination without time limitation.

7. NOTICE PERIOD
Company: 6 months
Employee: 6 months

8. GOVERNING LAW
Swedish law. Disputes settled by Stockholm Chamber of Commerce arbitration.

Signed:
Marcus Lindgren, CEO          Helena Friberg
DataSyn Nordic AB
```

---

## Document B4: IP Assignment Agreement

```
INTELLECTUAL PROPERTY ASSIGNMENT AGREEMENT

Between:  DataSyn Nordic AB, org.nr 559187-2234 ("the Company")
And:      Helena Friberg, 830415-XXXX ("the Assignor")

Date: 2020-09-01

BACKGROUND
The Assignor co-founded the Company in 2017 and has developed core 
technology forming the basis of the Company's SynCore platform.

Certain components of the SynCore platform were developed by the 
Assignor prior to the formal incorporation of the Company.

1. ASSIGNMENT
The Assignor hereby assigns to the Company all right, title, and interest 
in and to:
  a) The SynCore transformation engine (all versions)
  b) The data mapping algorithms (patent applications SE1800234-1 through 
     SE1800248-1)
  c) Any derivative works created since 2016-01-01

2. CONSIDERATION
The assignment is made in consideration of:
  a) The Assignor's 25% equity stake in the Company
  b) A one-time payment of 150,000 SEK

3. REPRESENTATIONS
The Assignor represents that:
  a) She is the sole creator of the assigned IP
  b) No third party has any claim to the assigned IP
  c) The IP has not been previously assigned or licensed

4. FURTHER ASSURANCE
The Assignor agrees to execute any further documents necessary to 
perfect the Company's ownership of the assigned IP.

Signed:
Marcus Lindgren, CEO          Helena Friberg
DataSyn Nordic AB
```

---

## Document B5: Board Minutes — Extraordinary Meeting

```
BOARD MINUTES
DataSyn Nordic AB, org.nr 559187-2234

Extraordinary Board Meeting
Date: 2024-08-15
Location: Company offices, Lindholmen, Gothenburg

Present:
  Marcus Lindgren, Chairman & CEO
  Helena Friberg, Board member & CTO
  Lars Wennerström, Board member (NorthCap Ventures)
  Sofia Eriksson, Board member (Baltic Innovation Fund)

1. Opening
Marcus Lindgren opened the meeting.

2. Purpose
The purpose of this extraordinary meeting is to discuss and approve 
the engagement of Setterwalls Advokatbyrå to conduct a legal review 
in preparation for a potential transaction process.

3. Transaction Readiness
Marcus Lindgren reported that the Company has received inbound interest 
from multiple parties regarding a potential acquisition. The Board 
discussed timing and process.

Lars Wennerström noted that NorthCap's fund has a 2025 exit horizon 
and recommended proceeding with preparation.

4. Known Issues for Disclosure
The Board discussed items requiring disclosure:

  a) IP Dispute: Former developer Alexei Petrov (employed 2019-2021, 
     Lithuania office) has claimed co-authorship of the SynCore 
     transformation engine's parallel processing module. His lawyers 
     sent a formal letter in June 2024. Setterwalls assesses the 
     claim as "not without merit" but manageable.

  b) Revenue Recognition: CFO Johan Nylund flagged that 3 multi-year 
     contracts signed in Q4 2023 include implementation services 
     bundled with SaaS licenses. These have been recognized as SaaS 
     revenue. Auditors were not specifically consulted on the 
     bundling treatment.

  c) Lithuanian Employment: The Lithuania team operates as direct 
     employees of DataSyn Nordic AB. No local entity exists. Tax 
     counsel has advised this structure carries risk of permanent 
     establishment classification.

5. Decisions
The Board resolved:
  - To engage Setterwalls for legal review (budget: 400 kSEK)
  - To instruct management to prepare a vendor due diligence report
  - To maintain strict confidentiality regarding the process

6. Closing
Meeting closed at 16:30.

Minutes: Marcus Lindgren
```

---

## Document B6: Customer Contract (Top Customer — Redacted)

```
SAAS SUBSCRIPTION AGREEMENT

Between: DataSyn Nordic AB ("Provider")
And:     [REDACTED — "Customer Alpha"] ("Customer")

Agreement Date: 2023-10-01
Agreement Number: SA-2023-0891

1. SERVICES
Provider shall deliver:
  - DataSyn Platform, Enterprise tier
  - Up to 50 user licenses
  - Data transformation volume: unlimited
  - Implementation services (one-time): data migration and 
    configuration, estimated 200 hours

2. TERM
Initial term: 36 months from Go-Live Date
Renewal: Automatic 12-month renewal unless 90 days written notice
Go-Live Date: To be confirmed upon completion of implementation, 
estimated January 2024

3. FEES
Annual SaaS license fee: 1,250,000 SEK
Implementation fee (one-time): 450,000 SEK
Total Year 1: 1,700,000 SEK
Year 2 and 3: 1,250,000 SEK per year

Price escalation: CPI + 2% annually

4. PAYMENT TERMS
SaaS fee: invoiced annually in advance
Implementation fee: 50% at signing, 50% at Go-Live

5. SERVICE LEVELS
Uptime: 99.5% measured monthly (excluding planned maintenance)
Support response: Critical issues within 4 hours, standard within 
24 hours

Penalty for SLA breach: 5% credit of monthly fee per percentage 
point below 99.5%, capped at 30% of monthly fee.

6. DATA PROCESSING
Customer data processed in AWS eu-north-1 (Stockholm).
Provider acts as data processor under GDPR Article 28.
Data Processing Agreement attached as Appendix B.

7. TERMINATION
Either party may terminate for material breach with 30 days notice 
to cure.
Customer may terminate for convenience with 6 months notice, 
subject to payment of remaining fees in the initial term.

8. LIMITATION OF LIABILITY
Provider's total liability capped at 12 months of fees paid.

9. GOVERNING LAW
Swedish law. Stockholm District Court.

Signed:
Johan Nylund, CFO              [REDACTED]
DataSyn Nordic AB               Customer Alpha
2023-10-01                       2023-10-15
```

---

## GROUND TRUTH MANIFEST

### Entities to Extract

| ID | Label | Type | Properties | Notes |
|----|-------|------|-----------|-------|
| E1 | DataSyn Nordic AB | organization | org_nr: 559187-2234, founded: 2017 | The target company |
| E2 | Marcus Lindgren | person | role: CEO + Chairman, founder | Also owns Lindgren Fastigheter AB |
| E3 | Helena Friberg | person | role: CTO, founder, board member | 22% or 25% ownership (inconsistency) |
| E4 | Johan Nylund | person | role: CFO, joined 2021 | Signs customer contract |
| E5 | Anders Blom | person | role: VP Sales, joined 2022 | |
| E6 | Lars Wennerström | person | role: board member, org: NorthCap | |
| E7 | Sofia Eriksson | person | role: board member, org: Baltic Innovation Fund | |
| E8 | Alexei Petrov | person | role: former developer 2019-2021 | IP claimant |
| E9 | NorthCap Ventures | organization | investor | 2025 exit horizon |
| E10 | Baltic Innovation Fund | organization | investor | |
| E11 | Lindgren Fastigheter AB | organization | owner: Marcus Lindgren | Related party |
| E12 | Setterwalls Advokatbyrå | organization | role: legal counsel | |
| E13 | Customer Alpha | organization | redacted | Top customer |
| E14 | SynCore | technology | type: transformation engine, language: Rust | Core IP |

### Planted Inconsistencies

| # | Type | Severity | Description | Documents |
|---|------|----------|-------------|-----------|
| **I1** | FINANCIAL | HIGH | Management presentation (B1) states 2023 revenue as 26.1 MSEK. Audited financials (B2) show 26,847 kSEK (26.8 MSEK). Difference of ~750 kSEK. Presentation rounds down to make growth trajectory look smoother. | B1 vs B2 |
| **I2** | FINANCIAL | HIGH | Management presentation (B1) states gross margin 84% for 2023. Audited financials (B2) show gross profit 22,284 / revenue 26,847 = 83.0%. Presentation inflates margin by 1 percentage point. | B1 vs B2 |
| **I3** | EQUITY | CRITICAL | Employment agreement (B3) states Helena Friberg holds 22% of shares. IP Assignment (B4) states consideration includes "the Assignor's 25% equity stake." 3% discrepancy — was equity diluted between 2020 (IP assignment) and 2022 (employment amendment), or is one document wrong? | B3 vs B4 |
| **I4** | HEADCOUNT | MEDIUM | Management presentation (B1) states 45 employees (38 Sweden, 7 Lithuania). Audited financials (B2) Note 14 states average 42 employees (36 Sweden, 6 Lithuania). Different metrics (year-end vs. average) but the Lithuania count differs: 7 vs 6. When was the 7th person hired? | B1 vs B2 |
| **I5** | IP_RISK | CRITICAL | IP Assignment (B4) — Helena represents she is "sole creator" of assigned IP. Board minutes (B5) — Alexei Petrov claims co-authorship of the parallel processing module. Setterwalls assesses claim "not without merit." This directly contradicts the representation in the IP assignment. | B4 vs B5 |
| **I6** | REVENUE_RECOGNITION | HIGH | Board minutes (B5) — CFO flags that 3 multi-year Q4 2023 contracts bundle implementation with SaaS and were recognized as SaaS revenue. Customer contract (B6) — shows exactly this pattern: 450k implementation fee alongside 1.25M SaaS. If implementation revenue is reclassified, 2023 SaaS revenue drops. Auditors were not consulted on bundling treatment. | B5 + B6 |
| **I7** | CONTINGENT_LIABILITY | HIGH | Audited financials (B2) Note 11 — describes IP dispute as "risk assessed as low, no provision made." Board minutes (B5) — Setterwalls assesses the same claim as "not without merit." Material difference in risk characterization between audited statements and board's own legal counsel. | B2 vs B5 |
| **I8** | OBLIGATION | MEDIUM | CTO non-compete (B3) is 18 months for competing business, 24 months for customer solicitation. If Helena leaves post-acquisition, acquirer must pay 60% of salary (~684k SEK/year) for up to 24 months. This obligation survives the transaction. | B3 |
| **I9** | RELATED_PARTY | MEDIUM | CEO's company Lindgren Fastigheter AB leases office to DataSyn for 840 kSEK/year (B2 Note 12). Board states "assessed as market-rate" but no independent valuation referenced. If acquisition proceeds, this lease is a related-party transaction requiring scrutiny. | B2 |

### Extraction Difficulty per Challenge

| Challenge | Difficulty | Why |
|-----------|-----------|-----|
| Entity extraction | EASY | Named individuals with explicit roles |
| Financial extraction | MEDIUM | Numbers in different formats (MSEK vs kSEK vs %) |
| Equity ownership % | EASY to extract, HARD to cross-reference | Must compare across docs |
| Obligation detection | HARD | Must parse legal language for non-compete, IP assignment terms |
| Risk characterization comparison | HARD | "risk assessed as low" vs "not without merit" — requires semantic reasoning |
| Revenue recognition issue | HARD | Must connect CFO's admission (B5) with the actual contract pattern (B6) |
| Related party detection | MEDIUM | Must flag CEO-owned entity leasing to company |
| Temporal obligation tracking | MEDIUM | Non-compete periods, contract terms, fund exit horizons |

---
