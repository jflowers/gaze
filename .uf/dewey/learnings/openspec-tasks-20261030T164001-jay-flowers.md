---
tag: openspec-tasks
author: jay-flowers
category: gotcha
created_at: 2026-10-30T16:40:01Z
identity: openspec-tasks-20261030T164001-jay-flowers
tier: draft
---

In gaze's OpenSpec/Speckit tasks.md files, the [P] marker means "parallel-eligible relative to [P] siblings in the same phase" — the documented precedent (openspec/changes/adapter-call-unmarshal-helper/tasks.md header comment) requires [P] tasks to touch distinct files and share only already-completed prerequisites. A test task that compile-depends on a preceding implementation task (e.g., redact_test.go referencing symbols that task 1.1 defines in redact.go) must not be marked [P]: a lone [P] in a phase is vacuous, and a dispatcher that batches [P] tasks concurrently fails compilation in the test worker. Correct [P] uses: mutually independent verification checks (validate / CI parity / docs-impact) and same-phase migrations across distinct files.
