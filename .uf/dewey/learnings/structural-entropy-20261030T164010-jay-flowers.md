---
tag: structural-entropy
author: jay-flowers
category: pattern
created_at: 2026-10-30T16:40:10Z
identity: structural-entropy-20261030T164010-jay-flowers
tier: draft
---

Specs-mode entropy reviews of not-yet-implemented OpenSpec changes: run the full vibe-check delta workflow anyway (detached base worktree at the shared SHA vs the repo working tree when input_context has head_sha == base_sha with uncommitted:true) — Markdown-only diffs yield all-zero package deltas and verdict APPROVE (confirmed on redact-env-values-from-adapter-stderr across all 22 gaze packages, matching the earlier divisor-gaze-testing instance). The design-claim audit has a precise trigger: a LOW design-precision advisory applies only when a planned file introduces the package's FIRST import of a dependency (module-true but package-false claims, e.g. internal/scaffold gaining its first yaml.v3 edge). When the planned imports are a subset of the package's existing import set (redact.go needs strings/sort/os, all already imported by internal/aireport), the planned ΔCe is exactly zero and no advisory applies — verify with a quick import grep before flagging.
