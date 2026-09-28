---
tag: external-analyzer-filtering
author: jay-flowers
created_at: 2026-09-28T22:32:15Z
identity: external-analyzer-filtering-20260928T223215-jay-flowers
tier: draft
---

When an external analyzer instruments only source files (e.g. Python --cov=src), its test functions receive no coverage entry and therefore score 0% coverage by default, silently inflating CRAP scores, CRAPload, quadrant counts, and phantom add_tests flags (issue #284: 724/993 functions were tests, 172/188 Q4 entries were tests, 174 phantom add_tests). The correct fix is to filter test files out of scoring at the adapter/provider layer — not in the language-neutral scoring engine (crap.computeScores) — using the analyzer's own `discover` protocol method which already returns a source_files/test_files split. Graceful degradation: when discover is absent or fails, leave the filter set nil (no-op) and log a warning, preserving prior behavior. Go mode is unaffected because -coverprofile instruments _test.go.
