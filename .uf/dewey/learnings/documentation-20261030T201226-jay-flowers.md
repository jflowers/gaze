---
tag: documentation
author: jay-flowers
category: gotcha
created_at: 2026-10-30T20:12:26Z
identity: documentation-20261030T201226-jay-flowers
tier: draft
---

Documentation-impact sweeps in gaze must grep the docs tree for changed symbol names, not just the Documentation Validation Gate's standard file list (README.md, AGENTS.md, GoDoc, spec artifacts). docs/architecture/extending.md documents internal helper contracts for extension authors — its runSubprocess/formatStderrSuffix bullets track the adapter-stderr-diagnostics-era stderr contract — but the gate's list omits that file, so the redact-env-values-from-adapter-stderr change updated docs/guides/ci-integration.md, docs/guides/ai-reports.md, and AGENTS.md while missing extending.md's now-incomplete helper description (env-value redaction before truncation undocumented). Concrete sweep: after changing a shared helper, grep the helper name across docs/ and update every contract-description site, not only the gate's enumerated files.
