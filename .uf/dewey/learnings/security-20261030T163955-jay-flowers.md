---
tag: security
author: jay-flowers
category: pattern
created_at: 2026-10-30T16:39:55Z
identity: security-20261030T163955-jay-flowers
tier: draft
---

When adding a secret-scrubbing choke point to a subprocess wrapper, enumerate every channel the child's output reaches — stderr embedded in errors, stderr returned on success, and stdout surfaced as product output (reports, CI step summaries) — and record an explicit in-scope/out-of-scope decision with rationale for each. Scrubbing one channel while leaving a sibling channel unacknowledged creates false coverage: the same buggy-dependency echo lands unmasked through the untouched channel, and future maintainers cannot tell conscious acceptance from oversight.
