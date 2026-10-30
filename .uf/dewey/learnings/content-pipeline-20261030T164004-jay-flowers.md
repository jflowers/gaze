---
tag: content-pipeline
author: jay-flowers
category: pattern
created_at: 2026-10-30T16:40:04Z
identity: content-pipeline-20261030T164004-jay-flowers
tier: draft
---

Borderline Website Documentation Gate calls in spec artifacts should be recorded as explicit, defeasible decisions rather than silent skips: state the N/A assessment with reasoning in proposal.md's Impact section and repeat it as an Open Question in design.md so a human can override at PR review. Demonstrated by gaze's redact-env-values-from-adapter-stderr OpenSpec change (observable [REDACTED:<VAR>] error-output markers, but no new command, flag, installation step, or workflow; consumer guidance lands in docs/guides/ci-integration.md — the file AGENTS.md designates as the reusable-workflow consumer contract). Curator reviews can then verify the in-repo docs target actually exists (section present, no duplicated content) and concur with or challenge the N/A call, instead of flagging an untracked gate skip or trusting the spec's claim unverified.
