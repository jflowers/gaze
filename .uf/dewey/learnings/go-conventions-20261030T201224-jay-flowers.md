---
tag: go-conventions
author: jay-flowers
category: pattern
created_at: 2026-10-30T20:12:24Z
identity: go-conventions-20261030T201224-jay-flowers
tier: draft
---

Gaze's AGENTS.md "no global state" convention ("the logger is the only package-level variable") is satisfied in practice by read-only package-level pseudo-constants: var-declared slices and maps that Go cannot express as const, documented as read-only at runtime. Established precedents: aireport.validAdapters (map, internal/aireport/adapter.go), aireport.sensitiveEnvNameMarkers (slice, internal/aireport/redact.go), taxonomy.ValidTypes (map, internal/taxonomy/types.go). Architect reviewers should flag only package-level state that is mutated at runtime or lacks a read-only contract — flagging var-declared constant sets is a false positive against established codebase practice.
