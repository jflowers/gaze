---
tag: scaffold-config
author: jay-flowers
category: pattern
created_at: 2026-10-29T17:30:22Z
identity: scaffold-config-20261029T173022-jay-flowers
tier: draft
---

When a repository's checked-in config must equal the end-state that a sync/merge function produces (e.g., gaze's hand-edited .uf review config vs. the syncUFReviewConfig function that `gaze init` runs in target projects), guard the hand edit with an end-state test: copy the repo's config files into a temp dir, run the sync against them, and assert zero rewrites and zero warnings (gaze's TestSyncUFReviewConfig_RepoConfigIsEndState). This catches drift between direct git edits and the code that would regenerate them — a hand-edited file silently diverging from what the tool produces is otherwise invisible until a consumer re-runs the tool and gets a surprise rewrite.
