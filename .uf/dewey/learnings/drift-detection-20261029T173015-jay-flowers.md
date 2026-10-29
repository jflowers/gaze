---
tag: drift-detection
author: jay-flowers
category: pattern
created_at: 2026-10-29T17:30:15Z
identity: drift-detection-20261029T173015-jay-flowers
tier: draft
---

Drift-detection pattern for a repo that consumes its own tool's output: when a repository's checked-in config files (the gaze repo's .uf/reviewer-capabilities.yaml and .uf/review-matrix.yaml) must equal the exact end-state a tool function produces for external target projects, add an invariant test that stages the repo files in a t.TempDir(), runs the function, and asserts zero rewrites and zero warnings (TestSyncUFReviewConfig_RepoConfigIsEndState in internal/scaffold/ufconfig_test.go). Direct git edits that drift from tool output then fail CI immediately — the same pattern gaze already uses for embedded-asset byte identity (TestEmbeddedAssetsMatchSource). The repo .uf files are edited directly to the end-state (design D6) rather than by running gaze init against the repo itself, because repo .opencode/ sources are deliberately marker-free.
