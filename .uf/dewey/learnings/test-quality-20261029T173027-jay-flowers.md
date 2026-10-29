---
tag: test-quality
author: jay-flowers
category: gotcha
created_at: 2026-10-29T17:30:27Z
identity: test-quality-20261029T173027-jay-flowers
tier: draft
---

Byte-identity mirror tests (e.g., TestEmbeddedAssetsMatchSource comparing .opencode/ scaffold sources to internal/scaffold/assets/ embedded copies) verify copy synchronization, not content correctness — a contract violation present in both copies passes CI. When a delta spec defines MUST scenarios for a scaffolded prompt asset's frontmatter (no model pin, temperature value, permission deny block), those scenarios need their own parse-and-assert test anchored on the source file; the mirror test then extends the guarantee to the embedded copy for free.
