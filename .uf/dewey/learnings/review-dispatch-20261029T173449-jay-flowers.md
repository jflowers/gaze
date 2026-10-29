---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-10-29T17:34:49Z
identity: review-dispatch-20261029T173449-jay-flowers
tier: draft
---

On the divisor-gaze-testing pipeline (opsx/divisor-gaze-testing, issue #303, October 2026), the code review council completed with an APPROVE WITH ADVISORIES verdict despite two environmental reviewer failures that did not block the gate: divisor-adversary failed because the machine-local review-matrix override (.uf/review-matrix.override.yaml, git-excluded) carried a model slug the provider could not resolve, and divisor-gaze-testing itself — the agent file the change had just created — failed because the opencode server snapshots its agent registry at startup, so agent files created or renamed after the server started cannot be loaded into a dispatch until the server restarts. Gate semantics still formed a valid verdict on the 5 of 7 successful reviewers (architect APPROVE; curator, entropy, guard, testing APPROVE WITH ADVISORIES), mirroring the 3-of-5 precedent from the spec review earlier in the same pipeline. Lessons for future councils in the gaze repo: after scaffolding or renaming agent files mid-pipeline, restart the opencode server before dispatching a council that includes them (or accept the environmental failure and rely on majority gate semantics), and verify machine-local review-matrix overrides reference resolvable model slugs before dispatch. Do not re-attempt the failed reviewers in the same server session; the artifact records the failure reasons and the verdict stands.
