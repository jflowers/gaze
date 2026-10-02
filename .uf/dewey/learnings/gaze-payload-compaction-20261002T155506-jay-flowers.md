---
tag: gaze-payload-compaction
author: jay-flowers
category: gotcha
created_at: 2026-10-02T15:55:06Z
identity: gaze-payload-compaction-20261002T155506-jay-flowers
tier: draft
---

The review council (Code Review Mode) caught that the canonical main spec `openspec/specs/payload-compaction/spec.md` still described pre-change semantics after an earlier spec-review step had "synced" it from a predecessor delta. When a change INVERTS behavior (e.g., flips what a compaction step strips vs preserves), syncing the canonical spec from the predecessor captures the OLD state, and the canonical spec ends up diametrically contradicting the new delta spec and implementation. The fix is to run `openspec-sync-specs` (agent-driven merge) to apply the delta's ADDED and MODIFIED requirements into the canonical spec, removing "Previously:" lines and reflecting the new inverted state. Canonical specs must never describe removed behavior. Also: GoDoc comments on exported types (e.g., `ReportPayload` in `internal/aireport/payload.go`) are a common place for stale pre-change semantics to linger — grep the doc comments for the old keywords when inverting compaction behavior.
