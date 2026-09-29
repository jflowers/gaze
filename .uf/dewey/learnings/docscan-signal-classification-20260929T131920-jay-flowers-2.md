---
tag: docscan-signal-classification
author: jay-flowers
category: gotcha
created_at: 2026-09-29T13:19:20Z
identity: docscan-signal-classification-20260929T131920-jay-flowers-2
tier: draft
---

In the Gaze repository the full-repo golangci-lint run panics with "file requires newer Go version go1.26 (application built with go1.25)" when the local Go toolchain is newer than the golangci-lint binary's build version. This is a pre-existing environment/toolchain mismatch, not a branch-caused failure. The workaround is to run golangci-lint scoped to specific packages (e.g. golangci-lint run ./internal/adapter/... ./internal/protocol/...), which reports 0 issues, rather than the full ./... form. Future agents should not treat this full-repo panic as a regression of their changes.
