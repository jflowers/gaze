---
tag: uf-unleash
author: jay-flowers
category: gotcha
created_at: 2026-10-29T17:34:52Z
identity: uf-unleash-20261029T173452-jay-flowers
tier: draft
---

When an entire OpenSpec change is still uncommitted (HEAD sitting at the base commit with the whole change as working-tree state, pending /uf.finale for staging and commit), uf_unleash_detect_resume can false-negative on the spec_review step — reporting it incomplete even though the uf.unleash checkpoint lists spec-review in completed_steps with an approving verdict and the dispatch artifact under .uf/artifacts/dispatch/ records the same head_sha as current HEAD with an APPROVE verdict. The spec-mode freshness probe expects clean spec artifacts, which an uncommitted tree cannot demonstrate. The correct re-anchor on resume is get_progress plus the on-disk dispatch artifacts, not the detect_resume step_status alone; the orchestrator's durable state and the artifact's recorded head_sha/verdict are authoritative. Encountered on opsx/divisor-gaze-testing (issue #303, re-entry window 7) where the pipeline correctly skipped re-running the spec review and proceeded straight to banking the code-review gate, retrospective, and demo.
