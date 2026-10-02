---
tag: gaze-payload-compaction
author: jay-flowers
category: gotcha
created_at: 2026-10-02T15:54:55Z
identity: gaze-payload-compaction-20261002T155455-jay-flowers
tier: draft
---

When compacting an AI-bound payload in the gaze repo, the previous `report-payload-reduction` change bounded the wrong fields: it kept three unbounded arrays (`crap.scores`, `quality.quality_reports`, `classify.results`) that scale with the total number of functions/tests/effects analyzed, while stripping the bounded worst-offender lists (`worst_crap` top-5, `worst_gaze_crap` top-5, `recommended_actions` top-20, `worst_coverage_tests` bottom-5) that the gaze-reporter prompt actually reads. This caused `gaze report --ai=opencode` to produce a 3.2MB compact payload (~1.05M tokens) that exceeded opencode's 1M-token limit. The correct inversion is to drop/bound the unbounded full arrays and preserve the bounded worst-offender lists, projecting quality gaps to self-contained side-effect objects (id/type/tier/location/description/target). The compact-size budget test must be calibrated against the REAL repo scale (thousands of functions), not a synthetic 200-function fixture.
