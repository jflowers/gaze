---
tag: review-hallucination-gotcha
author: jay-flowers
created_at: 2026-09-28T22:32:15Z
identity: review-hallucination-gotcha-20260928T223215-jay-flowers
tier: draft
---

When auto-fixing review-council findings, do NOT trust a reviewer agent's file/line references without verifying against disk. During the exclude-external-test-functions change, an earlier exploration concluded that TestQualityWithExternalAnalyzer_HappyPath and cmd/gaze/external_analyzer_test.go did not exist (a hallucinated negative), but they actually did exist and hard-coded want-3 counts that broke when the shared fake analyzer fixture gained a 4th test_mapping entry — discovered only at the full-suite run. Conversely, the reviewer correctly identified TestExternalComplexityProvider (adapter_test.go:45) as a real hard-coded-count test needing repair. Shared test fixtures (the fake analyzer) are the highest blast-radius surface; always grep for every hard-coded count assertion before changing a fixture, and verify reviewer-cited test paths with glob/read before editing.
