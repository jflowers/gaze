---
tag: scaffold
author: jay-flowers
category: decision
created_at: 2026-10-29T17:30:13Z
identity: scaffold-20261029T173013-jay-flowers
tier: draft
---

Cross-tool registration guard (divisor-gaze-testing change, design D5): gaze init never creates .uf files when the target project lacks a .uf directory — uf's parseReviewerManifest enforces its canonical KNOWN_REVIEWERS list, so a minimal generated capabilities file would be rejected by the consumer tool that inspired it. Merge into existing config only, print a notice when the directory is absent, and warn-and-skip when files are malformed. Never manufacturing a config the downstream tool would reject is the Minimal Assumptions principle applied to tool interop: optional registration intelligence must degrade to a notice, never fail the host operation (gaze init exit 0).
