---
tag: gate-on-change
author: jay-flowers
category: pattern
created_at: 2026-09-29T21:47:02Z
identity: gate-on-change-20260929T214702-jay-flowers
tier: draft
---

When implementing git diff parsing for function-level change detection, the `+++ b/` header is not always present (binary files, deleted files). The `diff --git a/... b/...` line provides a reliable fallback for extracting file paths. Use a two-phase approach: first try `+++ b/` headers, then fall back to `diff --git` pending paths for binary/deleted files. The `+++ /dev/null` case (deleted files) should reuse the pending path from `diff --git` rather than creating a separate `/dev/null` entry.
