# Post-Incident Review Template

## Incident Summary

| Field           | Details                              |
| --------------- | ------------------------------------ |
| **Incident ID** | INC-YYYY-NNN                         |
| **Date**        | YYYY-MM-DD                           |
| **Duration**    | Start time → End time (total: Xh Xm) |
| **Severity**    | P1 / P2 / P3 / P4                    |
| **Impact**      | Description of user impact           |
| **Responders**  | Names of team members involved       |
| **Status**      | Resolved / Monitoring                |

## Timeline

| Time (UTC) | Event                             |
| ---------- | --------------------------------- |
| HH:MM      | Alert triggered — [alert name]    |
| HH:MM      | Incident acknowledged by [person] |
| HH:MM      | Root cause identified             |
| HH:MM      | Mitigation applied                |
| HH:MM      | Service restored                  |
| HH:MM      | Incident resolved                 |

## Root Cause Analysis

### What happened?

> Describe the technical root cause in detail

### Why did it happen?

> Describe the underlying factors (use 5 Whys technique)

1. **Why?** — [Direct cause]
2. **Why?** — [Underlying cause]
3. **Why?** — [Deeper cause]
4. **Why?** — [Systemic cause]
5. **Why?** — [Root cause]

## Detection

- **How was the incident detected?** Alert / User report / Monitoring
- **Time to detect**: X minutes
- **Could we have detected it earlier?** Yes/No — explain

## Response

- **Time to acknowledge**: X minutes
- **Time to mitigate**: X minutes
- **Time to resolve**: X minutes
- **Was the runbook followed?** Yes/No
- **Was escalation needed?** Yes/No

## Action Items

| #   | Action                  | Owner  | Priority | Due Date   | Status |
| --- | ----------------------- | ------ | -------- | ---------- | ------ |
| 1   | [Preventive action]     | [Name] | High     | YYYY-MM-DD | Open   |
| 2   | [Detective improvement] | [Name] | Medium   | YYYY-MM-DD | Open   |
| 3   | [Process improvement]   | [Name] | Low      | YYYY-MM-DD | Open   |

## Lessons Learned

### What went well?

-

### What could be improved?

-

### Where did we get lucky?

-

---

## Example: Filled Template

### Incident: Database Connection Pool Exhaustion

| Field           | Details                                                |
| --------------- | ------------------------------------------------------ |
| **Incident ID** | INC-2026-042                                           |
| **Date**        | 2026-02-20                                             |
| **Duration**    | 14:30 → 15:45 UTC (1h 15m)                             |
| **Severity**    | P2                                                     |
| **Impact**      | 30% of API requests returned 500 errors for 75 minutes |
| **Responders**  | Hamfa, DevOps Team                                     |
| **Status**      | Resolved                                               |

**Root Cause**: A slow query caused by missing database index led to connection pool exhaustion, causing cascading failures.

**Action Items**:

1. ✅ Add missing index on `tasks.created_at` column
2. ✅ Implement connection pool monitoring alert (threshold: 80%)
3. 🔲 Add query timeout configuration (default: 30s)
4. 🔲 Implement circuit breaker pattern for database calls
