# Remediation Priority Model

Score each gap from `1` to `5` on each axis:

- `risk`: security/compliance/protocol correctness exposure
- `impact`: user/business impact
- `urgency`: immediate delivery pressure
- `effort`: engineering effort (inverse priority)

## Formula

`priority_score = (risk * 0.35) + (impact * 0.35) + (urgency * 0.2) + ((6 - effort) * 0.1)`

## Priority Bands

- `P0`: score >= 4.2 or any critical security/protocol gap
- `P1`: 3.2 to 4.19
- `P2`: < 3.2

## Required Fields Per Backlog Item

- Requirement reference
- Gap summary
- Owner role
- Dependencies
- Deliverables (endpoint/page/job/test)
- Acceptance criteria
- Verification command(s)
