---
tag: go-testing
author: jay-flowers
category: pattern
created_at: 2026-10-29T17:30:29Z
identity: go-testing-20261029T173029-jay-flowers
tier: draft
---

Proving a no-write/idempotency contract in Go file-mutation code: make the target files read-only (os.Chmod 0o444) before the second run — any write attempt fails and surfaces as a warning, proving the no-write contract by contradiction. Pair it with a byte-for-byte comparison of before/after content (the real assertion), and restore permissions via defer so t.TempDir cleanup succeeds. This pattern (used in gaze's TestSyncUFReviewConfig_IdempotentNoRewrite) also cheaply covers the write-failure warn-and-skip branch: stage an unregistered config read-only, assert the warning prints and nothing is reported registered.
