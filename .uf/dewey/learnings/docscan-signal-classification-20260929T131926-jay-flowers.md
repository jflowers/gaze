---
tag: docscan-signal-classification
author: jay-flowers
category: context
created_at: 2026-09-29T13:19:26Z
identity: docscan-signal-classification-20260929T131926-jay-flowers
tier: draft
---

For deterministic doc-derived classification signal weighting, a filename-based source-classification rule is more implementable and testable than a priority-tier-based rule. The docscan-signal-classification change initially specified weights by proximity priority (PrioritySamePackage/PriorityModuleRoot vs PriorityOther), which produced contradictory acceptance scenarios because the same document class was assigned two different weights. The fix was a single filename-based rule: files named README* map to the readme source (±15) regardless of priority tier, and all other docs map to architecture_doc (±25), with sidecar overrides at ±30. This eliminated the ambiguity and let every acceptance scenario agree with the weight table.
